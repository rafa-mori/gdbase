package external

import (
	sci "github.com/kubex-ecosystem/gdbase/internal/security/interfaces"
	"github.com/kubex-ecosystem/gdbase/internal/security/storage"
)

type SecretStorageAdapter struct {
	backend storage.ISecretStorage
}

func NewSecretStorageAdapter(backend storage.ISecretStorage) sci.IKeyringService {
	return &SecretStorageAdapter{backend: backend}
}

func (a *SecretStorageAdapter) StorePassword(password string) error {
	return a.backend.StorePassword(password)
}

func (a *SecretStorageAdapter) RetrievePassword() (string, error) {
	return a.backend.RetrievePassword()
}
