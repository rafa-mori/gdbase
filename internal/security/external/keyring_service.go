// Package external provides implementations for external services
package external

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kubex-ecosystem/gdbase/internal/module/kbx"
	crp "github.com/kubex-ecosystem/gdbase/internal/security/crypto"
	sci "github.com/kubex-ecosystem/gdbase/internal/security/interfaces"

	logz "github.com/kubex-ecosystem/logz"
)

var (
	KubexKeyringName = "canalize"
	KubexKeyringKey  string
)

func init() {
	var err error
	if KubexKeyringKey == "" {
		KubexKeyringKey, err = GetOrGenerateFromKeyring(KubexKeyringName)
		if err != nil {
			logz.Log("warn", fmt.Sprintf("Key storage unavailable, generated in-memory only: %v", err))
		}
	}
}

type KeyringService struct {
	keyringService string
	keyringName    string
}

func newKeyringService(service, name string) *KeyringService {
	return &KeyringService{
		keyringService: service,
		keyringName:    name,
	}
}
func NewKeyringService(service, name string) sci.IKeyringService {
	return newKeyringService(service, name)
}
func NewKeyringServiceType(service, name string) *KeyringService {
	return newKeyringService(service, name)
}

// StorePassword tries to store the password in the keyring
func (k *KeyringService) StorePassword(password string) error {
	if password == "" {
		return fmt.Errorf("key cannot be empty")
	}

	path := secretPath(k.keyringService, k.keyringName)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("error creating secrets directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(password), 0o600); err != nil {
		return fmt.Errorf("error storing key: %w", err)
	}

	return nil
}

func secretPath(service, name string) string {
	baseDir := os.ExpandEnv(kbx.DefaultKubexConfigDir)
	return filepath.Join(baseDir, "secrets", service, name)
}

// RetrievePassword tries to get the password from the keyring
func (k *KeyringService) RetrievePassword() (string, error) {
	path := secretPath(k.keyringService, k.keyringName)
	password, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", os.ErrNotExist
		}
		return "", fmt.Errorf("error retrieving key: %w", err)
	}

	return strings.TrimSpace(string(password)), nil
}

// GetOrGenPasswordKeyringPass tries to get the password from the keyring
// It uses the keyring service name to store and retrieve the password
// These methods aren't exposed to the outside world, only accessible through the package main logic
func (k *KeyringService) GetOrGenPasswordKeyringPass(name string) (string, error) {
	// Try to retrieve the password from the keyring
	krPass, krPassErr := k.RetrievePassword()
	if krPassErr != nil && krPassErr == os.ErrNotExist {
		logz.Log("debug", fmt.Sprintf("Key not found, generating new key for %s", name))
		krPassKey, krPassKeyErr := crp.NewCryptoServiceType().GenerateKey()
		if krPassKeyErr != nil {
			logz.Log("error", fmt.Sprintf("Error generating key: %v", krPassKeyErr))
			return "", krPassKeyErr
		}
		krPass = crp.NewCryptoService().EncodeBase64([]byte(krPassKey))

		// Store the password in the keyring and return the encoded password
		return StoreKeyringPassword(name, []byte(krPass))
	} else if krPassErr != nil {
		logz.Log("error", fmt.Sprintf("Error retrieving key: %v", krPassErr))
		return "", krPassErr
	}

	if !crp.IsBase64String(krPass) {
		krPass = crp.NewCryptoService().EncodeBase64([]byte(krPass))
	}

	return krPass, nil
}

// StoreKeyringPassword (Stateless function) stores the password in the keyring
// It will check if data is encoded, if so, will decode, store and then
// encode again or encode for the first time, returning always a portable data for
// the caller/logic outside this package be able to use it better and safer
// This method is not exposed to the outside world, only accessible through the package main logic
func StoreKeyringPassword(name string, pass []byte) (string, error) {
	cryptoService := crp.NewCryptoServiceType()
	// Will decode if encoded, but only if the password is not empty, not nil and not ENCODED
	copyPass := make([]byte, len(pass))
	copy(copyPass, pass)

	var decodedPass []byte
	if crp.IsBase64String(string(copyPass)) {
		var decodeErr error
		// Will decode if encoded, but only if the password is not empty, not nil and not ENCODED
		decodedPass, decodeErr = cryptoService.DecodeIfEncoded(copyPass)
		if decodeErr != nil {
			logz.Log("error", fmt.Sprintf("Error decoding password: %v", decodeErr))
			return "", decodeErr
		}
	} else {
		decodedPass = copyPass
	}

	// Store the password in the keyring decoded to avoid storing the encoded password
	// locally are much better for security keep binary static and encoded to handle with transport
	// integration and other utilities
	storeErr := NewKeyringService(kbx.KeyringService, fmt.Sprintf("canalizedb-%s", name)).StorePassword(string(decodedPass))
	if storeErr != nil {
		logz.Log("error", fmt.Sprintf("Error storing key: %v", storeErr))
		return "", storeErr
	}

	// Handle with logging here for getOrGenPasswordKeyringPass output
	encodedPass, encodeErr := cryptoService.EncodeIfDecoded(decodedPass)
	if encodeErr != nil {
		logz.Log("error", fmt.Sprintf("Error encoding password: %v", encodeErr))
		return "", encodeErr
	}

	// Return the encoded password to be used by the caller/logic outside this package
	if err := NewKeyringService(kbx.KeyringService, fmt.Sprintf("canalizedb-%s", name)).StorePassword(encodedPass); err != nil {
		logz.Log("warn", fmt.Sprintf("Error persisting generated credential %s: %v", name, err))
	}

	return encodedPass, nil
}

// GetOrGenerateFromKeyring (Stateless function) tries to get the password from the keyring
// If not found, it will generate a new one, store it and return the encoded password
// This method is not exposed to the outside world, only accessible through the package main logic
func GetOrGenerateFromKeyring(name string) (string, error) {
	krPass, pgPassErr := NewKeyringService(kbx.KeyringService, fmt.Sprintf("canalizedb-%s", name)).RetrievePassword()
	if pgPassErr != nil && pgPassErr != os.ErrNotExist {
		return "", pgPassErr
	}
	if pgPassErr == nil && krPass != "" {
		if !crp.IsBase64String(krPass) {
			krPass = crp.NewCryptoService().EncodeBase64([]byte(krPass))
		}
		return krPass, nil
	}

	krPassKey, krPassKeyErr := crp.NewCryptoServiceType().GenerateKey()
	if krPassKeyErr != nil {
		logz.Log("error", fmt.Sprintf("Error generating key: %v", krPassKeyErr))
		return "", krPassKeyErr
	}

	encoded := crp.NewCryptoService().EncodeBase64([]byte(krPassKey))
	if err := NewKeyringService(kbx.KeyringService, fmt.Sprintf("canalizedb-%s", name)).StorePassword(encoded); err != nil {
		logz.Log("warn", fmt.Sprintf("Error storing key locally: %v", err))
	}

	return encoded, nil
}
