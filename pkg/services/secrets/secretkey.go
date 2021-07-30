package secrets

import "github.com/grafana/grafana/pkg/services/encryption"

type settingsSecretKey struct {
	key func() []byte
	e   encryption.EncryptionService
}

func (s *settingsSecretKey) Encrypt(blob []byte) ([]byte, error) {
	return s.e.Encrypt(blob, s.key())
}

func (s *settingsSecretKey) Decrypt(blob []byte) ([]byte, error) {
	return s.e.Decrypt(blob, s.key())
}
