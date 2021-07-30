package secrets

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/services/encryption"

	"github.com/grafana/grafana/pkg/registry"

	"github.com/grafana/grafana/pkg/util"

	"github.com/grafana/grafana/pkg/bus"

	"github.com/grafana/grafana/pkg/models"

	"github.com/grafana/grafana/pkg/services/sqlstore"

	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/setting"
)

var logger = log.New("secrets") // TODO: should it be at the package level?

func init() {
	registry.Register(&registry.Descriptor{
		Name:         "SecretsService",
		Instance:     &SecretsService{},
		InitPriority: registry.High,
	})
}

type SecretsService struct {
	Store      *sqlstore.SQLStore           `inject:""`
	Bus        bus.Bus                      `inject:""`
	Encryption encryption.EncryptionService `inject:""`

	//defaultEncryptionKey string // TODO: Where should it be initialized? It looks like key id/name
	defaultProvider string
	providers       map[string]Provider
	dataKeyCache    map[string]dataKeyCacheItem
}

type dataKeyCacheItem struct {
	expiry  time.Time
	dataKey []byte
}

type Provider interface {
	Encrypt(blob []byte) ([]byte, error)
	Decrypt(blob []byte) ([]byte, error)
}

func (s *SecretsService) Init() error {
	s.providers = map[string]Provider{
		"": &settingsSecretKey{
			e: encryption.OSSEncryptionService{},
			key: func() []byte {
				return []byte(setting.SecretKey)
			},
		},
	}
	s.defaultProvider = "" // should be read from settings

	//ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	//defer cancel()

	//baseKey := "root"
	//_, err := s.Store.GetDataKey(ctx, baseKey)
	//if err != nil {
	//	if errors.Is(err, models.ErrDataKeyNotFound) {
	//		err = s.newRandomDataKey(ctx, baseKey)
	//		if err != nil {
	//			return err
	//		}
	//	} else {
	//		return err
	//	}
	//}

	util.Encrypt = s.Encrypt
	util.Decrypt = s.Decrypt

	return nil
}

func (s *SecretsService) newRandomDataKey(ctx context.Context, name string) error {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}

	provider, exists := s.providers[s.defaultProvider]
	if !exists {
		return fmt.Errorf("could not find encryption provider '%s'", s.defaultProvider)
	}
	encrypted, err := provider.Encrypt(b)
	if err != nil {
		return err
	}

	err = s.Store.CreateDataKey(ctx, models.DataKey{
		Active:        true, // TODO: how do we manage active/deactivated DEKs?
		Name:          name,
		Provider:      s.defaultProvider,
		EncryptedData: encrypted,
	})
	return nil
}

var b64 = base64.RawStdEncoding

func (s *SecretsService) Encrypt(payload []byte) ([]byte, error) {
	key := "root"                  // TODO: some logic to figure out what DEK identifier to use
	dataKey, err := s.dataKey(key) // remove key from args
	// if no active key, create a new and cache createDEK()
	if err != nil {
		return nil, err
	}

	encrypted, err := s.Encryption.Encrypt(payload, dataKey)
	if err != nil {
		return nil, err
	}

	prefix := make([]byte, b64.EncodedLen(len(key))+2)
	b64.Encode(prefix[1:], []byte(key))
	prefix[0] = '#'
	prefix[len(prefix)-1] = '#'

	blob := make([]byte, len(prefix)+len(encrypted))
	copy(blob, prefix)
	copy(blob[len(prefix):], encrypted)

	return blob, nil
}

func (s *SecretsService) Decrypt(payload []byte) ([]byte, error) {
	if len(payload) == 0 {
		return []byte{}, nil // TODO: Not sure if it should return error like decrypt does (also see tests)
	}

	var dataKey []byte

	if payload[0] != '#' {
		dataKey = []byte(setting.SecretKey)
	} else {
		payload = payload[1:]
		endOfKey := bytes.Index(payload, []byte{'#'})
		if endOfKey == -1 {
			return nil, fmt.Errorf("could not find valid key in encrypted payload")
		}
		b64Key := payload[:endOfKey]
		payload = payload[endOfKey+1:]
		key := make([]byte, b64.DecodedLen(len(b64Key)))
		_, err := b64.Decode(key, b64Key)
		if err != nil {
			return nil, err
		}

		dataKey, err = s.dataKey(string(key)) // TODO: key is the identifier of the key
		if err != nil {
			return nil, err
		}
	}

	return s.Encryption.Decrypt(payload, dataKey)
}

// dataKey decrypts and caches DEK
func (s *SecretsService) dataKey(key string) ([]byte, error) {
	if key == "" {
		return []byte(setting.SecretKey), nil // TODO: not sure what case this condition handles
	}

	if item, exists := s.dataKeyCache[key]; exists {
		if item.expiry.Before(time.Now()) && !item.expiry.IsZero() {
			delete(s.dataKeyCache, key)
		} else {
			return item.dataKey, nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 1. get encrypted data key from database
	dataKey, err := s.Store.GetDataKey(ctx, key)
	if err != nil {
		return nil, err
	}

	// 2. decrypt data key
	provider, exists := s.providers[dataKey.Provider]
	if !exists {
		return nil, fmt.Errorf("could not find encryption provider '%s'", dataKey.Provider)
	}

	decrypted, err := provider.Decrypt(dataKey.EncryptedData)
	if err != nil {
		return nil, err
	}

	// 3. cache data key
	s.dataKeyCache[key] = dataKeyCacheItem{
		expiry:  time.Now().Add(15 * time.Minute),
		dataKey: decrypted,
	}

	return decrypted, nil
}
