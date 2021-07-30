package secrets

import (
	"fmt"
	"testing"

	"github.com/grafana/grafana/pkg/services/sqlstore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana/pkg/setting"
)

func TestSecrets_Encrypt(t *testing.T) {
	s := SecretsService{
		Store: sqlstore.InitTestDB(t),
	}

	require.NoError(t, s.Init())

	{
		old := setting.SecretKey
		defer func() {
			setting.SecretKey = old
		}()
		setting.SecretKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	}

	//t.Run("getting encryption key", func(t *testing.T) {
	//	key, err := encryption.encryptionKeyToBytes([]byte("secret"), []byte("salt"))
	//	require.NoError(t, err)
	//	assert.Len(t, key, 32)
	//
	//	key, err = encryption.encryptionKeyToBytes([]byte("a very long secret key that is larger then 32bytes"), []byte("salt"))
	//	require.NoError(t, err)
	//	assert.Len(t, key, 32)
	//})

	plaintexts := [][]byte{
		[]byte("hello, world"),
		[]byte("grafana"),
	}

	for _, plaintext := range plaintexts {
		t.Run(fmt.Sprintf("encrypting and decrypting %s", string(plaintext)), func(t *testing.T) {
			encrypted, err := s.Encrypt(plaintext)
			require.NoError(t, err)
			decrypted, err := s.Decrypt(encrypted)
			require.NoError(t, err)

			assert.Equal(t, plaintext, decrypted)
		})
	}

	t.Run("decrypting empty payload should not fail", func(t *testing.T) {
		_, err := s.Decrypt([]byte(""))
		require.Error(t, err)

		assert.Equal(t, "unable to compute salt", err.Error())
	})

	t.Run("old test", func(t *testing.T) {
		encrypted, err := s.Encryption.Encrypt([]byte("my very long text"), []byte(setting.SecretKey))
		decrypted, err := s.Encryption.Decrypt(encrypted, []byte(setting.SecretKey))
		require.NoError(t, err)

		assert.Equal(t, []byte("my very long text"), decrypted)
	})
}
