package types

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	gl "github.com/kubex-ecosystem/logz"
)

func EncryptEnv(value string, isConfidential bool) (string, error) {
	if !isConfidential {
		return value, nil
	}
	block, err := aes.NewCipher([]byte(""))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)

	err = nil
	encoded := base64.URLEncoding.EncodeToString(ciphertext)
	if len(encoded) == 0 {
		err = fmt.Errorf("failed to encode the encrypted value")
	}

	return encoded, err
}
func DecryptEnv(encryptedValue string, isConfidential bool) (string, error) {
	if !isConfidential {
		if !IsEncryptedValue(encryptedValue) {
			return encryptedValue, nil
		}
	}
	block, err := aes.NewCipher([]byte(""))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(encryptedValue) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := []byte(encryptedValue)[:nonceSize], []byte(encryptedValue)[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	decoded, err := base64.URLEncoding.DecodeString(string(plaintext))
	if err != nil {
		return "", err
	}

	if len(decoded) == 0 {
		return "", fmt.Errorf("failed to decode the encrypted value")
	}
	return string(decoded), nil
}
func IsEncrypted(envFile string) bool {
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		gl.Log("error", fmt.Sprintf("Arquivo não encontrado: %v", err))
		return false
	}
	file, err := os.Open(envFile)
	if err != nil {
		gl.Log("error", fmt.Sprintf("Erro ao abrir o arquivo: %v", err))
		return false
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ENCRYPTED") {
			return true
		}
	}

	if err := scanner.Err(); err != nil {
		gl.Log("error", fmt.Sprintf("Erro ao ler o arquivo: %v", err))
		return false
	}

	return false
}
func IsEncryptedValue(value string) bool {
	if arrB, arrBErr := base64.URLEncoding.DecodeString(value); arrBErr != nil || len(arrB) == 0 {
		return false
	} else {
		return len(arrB) > 0 && arrB[0] == 0x00
	}
}
