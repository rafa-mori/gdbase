// Package kbx has default configuration values
package kbx

import (
	"fmt"
	"os"

	crp "github.com/kubex-ecosystem/gdbase/internal/security/crypto"
	krg "github.com/kubex-ecosystem/gdbase/internal/security/external"
	gl "github.com/kubex-ecosystem/logz"
)

const (
	KeyringService        = "kubex"
	DefaultKubexConfigDir = "$HOME/.kubex"

	DefaultGoBEKeyPath    = "$HOME/.kubex/gobe/gobe-key.pem"
	DefaultGoBECertPath   = "$HOME/.kubex/gobe/gobe-cert.pem"
	DefaultGoBECAPath     = "$HOME/.kubex/gobe/ca-cert.pem"
	DefaultGoBEConfigPath = "$HOME/.kubex/gobe/config/config.json"

	DefaultConfigDir        = "$HOME/.kubex/gdbase/config"
	DefaultConfigFile       = "$HOME/.kubex/gdbase/config.json"
	DefaultGDBaseConfigPath = "$HOME/.kubex/gdbase/config/config.json"
)

const (
	DefaultVolumesDir     = "$HOME/.kubex/volumes"
	DefaultRedisVolume    = "$HOME/.kubex/volumes/redis"
	DefaultPostgresVolume = "$HOME/.kubex/volumes/postgresql"
	DefaultMongoVolume    = "$HOME/.kubex/volumes/mongo"
	DefaultRabbitMQVolume = "$HOME/.kubex/volumes/rabbitmq"
)

const (
	DefaultRateLimitLimit  = 100
	DefaultRateLimitBurst  = 100
	DefaultRequestWindow   = 1 * 60 * 1000 // 1 minute
	DefaultRateLimitJitter = 0.1
)

const (
	DefaultMaxRetries = 3
	DefaultRetryDelay = 1 * 1000 // 1 second
)

const (
	DefaultMaxIdleConns          = 100
	DefaultMaxIdleConnsPerHost   = 100
	DefaultIdleConnTimeout       = 90 * 1000 // 90 seconds
	DefaultTLSHandshakeTimeout   = 10 * 1000 // 10 seconds
	DefaultExpectContinueTimeout = 1 * 1000  // 1 second
	DefaultResponseHeaderTimeout = 5 * 1000  // 5 seconds
	DefaultTimeout               = 30 * 1000 // 30 seconds
	DefaultKeepAlive             = 30 * 1000 // 30 seconds
	DefaultMaxConnsPerHost       = 100
)

const (
	DefaultLLMProvider    = "gemini"
	DefaultLLMModel       = "gemini-2.0-flash"
	DefaultLLMMaxTokens   = 1024
	DefaultLLMTemperature = 0.3
)

const (
	DefaultApprovalRequireForResponses = false
	DefaultApprovalTimeoutMinutes      = 15
)

const (
	DefaultServerPort = "8088"
	DefaultServerHost = "0.0.0.0"
)

type ValidationError struct {
	Field   string
	Message string
}

func (v *ValidationError) Error() string {
	return v.Message
}
func (v *ValidationError) FieldError() map[string]string {
	return map[string]string{v.Field: v.Message}
}
func (v *ValidationError) FieldsError() map[string]string {
	return map[string]string{v.Field: v.Message}
}
func (v *ValidationError) ErrorOrNil() error {
	return v
}

var (
	ErrUsernameRequired = &ValidationError{Field: "username", Message: "Username is required"}
	ErrPasswordRequired = &ValidationError{Field: "password", Message: "Password is required"}
	ErrEmailRequired    = &ValidationError{Field: "email", Message: "Email is required"}
	ErrDBNotProvided    = &ValidationError{Field: "db", Message: "Database not provided"}
	ErrModelNotFound    = &ValidationError{Field: "model", Message: "Model not found"}
)

var (
	KubexKeyringName = "kubex"
	KubexKeyringKey  string
)

func init() {
	var err error
	if KubexKeyringKey == "" {
		KubexKeyringKey, err = GetOrGenPasswordKeyringPass(KubexKeyringName)
		if err != nil {
			gl.Log("fatal", fmt.Sprintf("Error initializing keyring: %v", err))
		}
	}
}

// GetOrGenPasswordKeyringPass retrieves the password from the keyring or generates a new one if it doesn't exist
// It uses the keyring service name to store and retrieve the password
// These methods aren't exposed to the outside world, only accessible through the package main logic
func GetOrGenPasswordKeyringPass(name string) (string, error) {
	// Try to retrieve the password from the keyring
	krPass, krPassErr := krg.NewKeyringService(KeyringService, fmt.Sprintf("gdbase-%s", name)).RetrievePassword()
	if krPassErr != nil && krPassErr == os.ErrNotExist {
		gl.Log("debug", fmt.Sprintf("Key found for %s", name))
		// If the error is "keyring: item not found", generate a new key
		gl.Log("debug", fmt.Sprintf("Key not found, generating new key for %s", name))
		krPassKey, krPassKeyErr := crp.NewCryptoServiceType().GenerateKey()
		if krPassKeyErr != nil {
			gl.Log("error", fmt.Sprintf("Error generating key: %v", krPassKeyErr))
			return "", krPassKeyErr
		}
		krPass = string(krPassKey)

		// Store the password in the keyring and return the encoded password
		return storeKeyringPassword(name, []byte(krPass))
	} else if krPassErr != nil {
		gl.Log("error", fmt.Sprintf("Error retrieving key: %v", krPassErr))
		return "", krPassErr
	}

	if !crp.IsBase64String(krPass) {
		krPass = crp.NewCryptoService().EncodeBase64([]byte(krPass))
	}

	return krPass, nil
}

// storeKeyringPassword stores the password in the keyring
// It will check if data is encoded, if so, will decode, store and then
// encode again or encode for the first time, returning always a portable data for
// the caller/logic outside this package be able to use it better and safer
// This method is not exposed to the outside world, only accessible through the package main logic
func storeKeyringPassword(name string, pass []byte) (string, error) {
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
			gl.Log("error", fmt.Sprintf("Error decoding password: %v", decodeErr))
			return "", decodeErr
		}
	} else {
		decodedPass = copyPass
	}

	// Store the password in the keyring decoded to avoid storing the encoded password
	// locally are much better for security keep binary static and encoded to handle with transport
	// integration and other utilities
	storeErr := krg.NewKeyringService(KeyringService, fmt.Sprintf("gdbase-%s", name)).StorePassword(string(decodedPass))
	if storeErr != nil {
		gl.Log("error", fmt.Sprintf("Error storing key: %v", storeErr))
		return "", storeErr
	}

	// Handle with logging here for getOrGenPasswordKeyringPass output
	encodedPass, encodeErr := cryptoService.EncodeIfDecoded(decodedPass)
	if encodeErr != nil {
		gl.Log("error", fmt.Sprintf("Error encoding password: %v", encodeErr))
		return "", encodeErr
	}

	// Return the encoded password to be used by the caller/logic outside this package
	return encodedPass, nil
}
