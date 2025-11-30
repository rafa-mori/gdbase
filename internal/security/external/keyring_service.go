// Package external provides implementations for external services
package external

import (
	"errors"
	"fmt"
	"os"

	ci "github.com/kubex-ecosystem/gdbase/internal/interfaces"
	// gl "github.com/kubex-ecosystem/logz"
	svc "github.com/kubex-ecosystem/gdbase/internal/globals"
	sci "github.com/kubex-ecosystem/gdbase/internal/security/interfaces"
	"github.com/zalando/go-keyring"
)

type KeyringService struct {
	keyringService ci.IProperty[string]
	keyringName    ci.IProperty[string]
}

func newKeyringService(service, name string) *KeyringService {
	return &KeyringService{
		keyringService: svc.NewPropertyType("keyringService", &service, false, nil),
		keyringName:    svc.NewPropertyType("keyringName", &name, false, nil),
	}
}
func NewKeyringService(service, name string) sci.IKeyringService {
	return newKeyringService(service, name)
}
func NewKeyringServiceType(service, name string) *KeyringService {
	return newKeyringService(service, name)
}

func (k *KeyringService) StorePassword(password string) error {
	if password == "" {
		// gl.Log("error", "key cannot be empty")
		return fmt.Errorf("key cannot be empty")
	}
	if err := keyring.Set(k.keyringService.GetValue(), k.keyringName.GetValue(), password); err != nil {
		return fmt.Errorf("error storing key: %v", err)
	}
	// gl.Log("debug", fmt.Sprintf("key stored successfully: %s", k.keyringName.GetValue()))
	return nil
}
func (k *KeyringService) RetrievePassword() (string, error) {
	if password, err := keyring.Get(k.keyringService.GetValue(), k.keyringName.GetValue()); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", os.ErrNotExist
		}
		// gl.Log("debug", fmt.Sprintf("error retrieving key: %v", err))
		return "", fmt.Errorf("error retrieving key: %v", err)
	} else {
		return password, nil
	}
}
