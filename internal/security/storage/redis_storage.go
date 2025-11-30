// Package storage provides Redis-based secret storage implementation.
package storage

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisSecretStorage struct {
	Client    *redis.Client
	KeyPrefix string
	MasterKey []byte
	Ctx       context.Context
}

func NewRedisSecretStorage(addr, password, prefix string, masterKey []byte) *RedisSecretStorage {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &RedisSecretStorage{
		Client:    rdb,
		KeyPrefix: prefix,
		MasterKey: masterKey,
		Ctx:       context.Background(),
	}
}

func (r *RedisSecretStorage) key(name string) string {
	return fmt.Sprintf("%s:%s", r.KeyPrefix, name)
}

func (r *RedisSecretStorage) StorePassword(password string) error {
	hash := sha256.Sum256(r.MasterKey)
	block, _ := aes.NewCipher(hash[:])
	aead, _ := cipher.NewGCM(block)
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	ct := aead.Seal(nil, nonce, []byte(password), nil)
	payload := base64.StdEncoding.EncodeToString(append(nonce, ct...))
	return r.Client.Set(r.Ctx, r.key("secret"), payload, 0).Err()
}

func (r *RedisSecretStorage) RetrievePassword() (string, error) {
	data, err := r.Client.Get(r.Ctx, r.key("secret")).Result()
	if err != nil {
		return "", err
	}
	raw, _ := base64.StdEncoding.DecodeString(data)
	hash := sha256.Sum256(r.MasterKey)
	block, _ := aes.NewCipher(hash[:])
	aead, _ := cipher.NewGCM(block)
	nonceSize := aead.NonceSize()
	nonce, ct := raw[:nonceSize], raw[nonceSize:]
	plain, err := aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
