// Package external implements a file-based, AES-GCM encrypted replacement for go-keyring.
package external

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	sci "github.com/kubex-ecosystem/gdbase/internal/security/interfaces"
	t "github.com/kubex-ecosystem/gdbase/internal/types"
	logz "github.com/kubex-ecosystem/logz"
)

// FileKeyringService is a drop-in replacement for KeyringService,
// maintaining the same contract and method signatures.
type FileKeyringService struct {
	keyringService *t.Property[string]
	keyringName    *t.Property[string]
	masterKey      []byte
	baseDir        string
}

// NewFileKeyringService creates a new encrypted file-based secret store.
func NewFileKeyringService(service, name string) sci.IKeyringService {
	return newFileKeyringService(service, name)
}

func newFileKeyringService(service, name string) *FileKeyringService {
	mk := os.Getenv("APP_MASTER_KEY")
	if mk == "" {
		logz.Log("warn", "APP_MASTER_KEY not set; using ephemeral random key (not persistent!)")
		tmp := make([]byte, 32)
		_, _ = rand.Read(tmp)
		mk = base64.StdEncoding.EncodeToString(tmp)
	}
	raw, _ := base64.StdEncoding.DecodeString(mk)
	dir := os.Getenv("APP_SECRETS_DIR")
	if dir == "" {
		dir = "/var/lib/kubex/secrets"
	}
	_ = os.MkdirAll(dir, 0o700)

	return &FileKeyringService{
		keyringService: t.NewProperty("keyringService", &service, false, nil),
		keyringName:    t.NewProperty("keyringName", &name, false, nil),
		masterKey:      raw,
		baseDir:        dir,
	}
}

func (k *FileKeyringService) StorePassword(password string) error {
	if password == "" {
		logz.Log("error", "key cannot be empty")
		return fmt.Errorf("key cannot be empty")
	}
	enc, err := k.encrypt([]byte(password))
	if err != nil {
		return fmt.Errorf("error encrypting password: %v", err)
	}
	path := filepath.Join(k.baseDir, fmt.Sprintf("%s_%s.secret", k.keyringService.GetValue(), k.keyringName.GetValue()))
	if err := os.WriteFile(path, []byte(enc), 0o600); err != nil {
		return fmt.Errorf("error storing key: %v", err)
	}
	logz.Log("debug", fmt.Sprintf("key stored successfully: %s", k.keyringName.GetValue()))
	return nil
}

func (k *FileKeyringService) RetrievePassword() (string, error) {
	path := filepath.Join(k.baseDir, fmt.Sprintf("%s_%s.secret", k.keyringService.GetValue(), k.keyringName.GetValue()))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", os.ErrNotExist
		}
		logz.Log("debug", fmt.Sprintf("error reading key: %v", err))
		return "", fmt.Errorf("error retrieving key: %v", err)
	}
	plain, err := k.decrypt(string(data))
	if err != nil {
		return "", fmt.Errorf("error decrypting key: %v", err)
	}
	return string(plain), nil
}

// --- internal helpers ---

func (k *FileKeyringService) encrypt(plain []byte) (string, error) {
	hash := sha256.Sum256(k.masterKey)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := aead.Seal(nil, nonce, plain, nil)
	full := append(nonce, ct...)
	return base64.StdEncoding.EncodeToString(full), nil
}

func (k *FileKeyringService) decrypt(ciphertext string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(k.masterKey)
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(data) < aead.NonceSize() {
		return nil, errors.New("invalid ciphertext")
	}
	nonce, ct := data[:aead.NonceSize()], data[aead.NonceSize():]
	return aead.Open(nil, nonce, ct, nil)
}
