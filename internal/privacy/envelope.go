package privacy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"sync"
)

type Envelope struct {
	Version    int    `json:"version"`
	Algorithm  string `json:"algorithm"`
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
	Digest     string `json:"digest"`
}
type KeyRing struct {
	mu     sync.RWMutex
	keys   map[string][]byte
	active string
}

func NewKeyRing() *KeyRing { return &KeyRing{keys: map[string][]byte{}} }
func (k *KeyRing) Put(id string, key []byte) error {
	if id == "" || len(key) != 32 {
		return errors.New("key id and 32-byte key required")
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	k.keys[id] = append([]byte(nil), key...)
	if k.active == "" {
		k.active = id
	}
	return nil
}
func (k *KeyRing) Activate(id string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if _, ok := k.keys[id]; !ok {
		return errors.New("key not found")
	}
	k.active = id
	return nil
}
func (k *KeyRing) Remove(id string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if id == k.active {
		return errors.New("active key cannot be removed")
	}
	if _, ok := k.keys[id]; !ok {
		return errors.New("key not found")
	}
	delete(k.keys, id)
	return nil
}
func (k *KeyRing) Encrypt(plaintext []byte) (Envelope, error) {
	k.mu.RLock()
	id, key := k.active, append([]byte(nil), k.keys[k.active]...)
	k.mu.RUnlock()
	if id == "" {
		return Envelope{}, errors.New("no active key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return Envelope{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return Envelope{}, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return Envelope{}, err
	}
	sealed := gcm.Seal(nil, nonce, plaintext, nil)
	digest := sha256.Sum256(plaintext)
	return Envelope{Version: 1, Algorithm: "AES-256-GCM/" + id, Nonce: base64.RawStdEncoding.EncodeToString(nonce), Ciphertext: base64.RawStdEncoding.EncodeToString(sealed), Digest: fmt.Sprintf("%x", digest[:])}, nil
}
func (k *KeyRing) Decrypt(e Envelope) ([]byte, error) {
	if e.Version != 1 {
		return nil, errors.New("unsupported envelope version")
	}
	parts := splitAlgorithm(e.Algorithm)
	k.mu.RLock()
	key := append([]byte(nil), k.keys[parts]...)
	k.mu.RUnlock()
	if len(key) == 0 {
		return nil, errors.New("encryption key unavailable")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce, err := base64.RawStdEncoding.DecodeString(e.Nonce)
	if err != nil {
		return nil, err
	}
	sealed, err := base64.RawStdEncoding.DecodeString(e.Ciphertext)
	if err != nil {
		return nil, err
	}
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, errors.New("ciphertext authentication failed")
	}
	digest := sha256.Sum256(plain)
	if fmt.Sprintf("%x", digest[:]) != e.Digest {
		return nil, errors.New("plaintext digest mismatch")
	}
	return plain, nil
}
func splitAlgorithm(value string) string {
	for i := len(value) - 1; i >= 0; i-- {
		if value[i] == '/' {
			return value[i+1:]
		}
	}
	return value
}
