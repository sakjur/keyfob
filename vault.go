package keyfob

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/hkdf"
)

var errServiceKeyTooShort = errors.New("service key must be at least 128 bits")

// A KeyVault is an abstract description of a secure storage location for
// root keys.
type KeyVault interface {
	GetKey(userid uuid.UUID, namespace string) ([]byte, error)
	// Repeated insertion of key should not update the key. Only the first
	// key inserted should be persisted. Collision on persistence is not
	// considered an error.
	InsertKey(userid uuid.UUID, namespace string, key []byte) error
	DeleteKey(userid uuid.UUID, namespace string) error
}

// UserKey is a struct which maps out a single cryptographic key from a
// KeyVault.
type UserKey struct {
	UserID     uuid.UUID
	Namespace  string
	ServiceKey []byte
}

// DeriveKey fetches the root key for the user within the data category and
// combines it with the service key to derive a new key which the service
// can use to encrypt data securely.
//
// The service key must be longer than 128 bits.
func (k *UserKey) DeriveKey(vault KeyVault) ([]byte, error) {
	if len(k.ServiceKey) < 16 {
		return nil, errServiceKeyTooShort
	}

	rootKey, err := vault.GetKey(k.UserID, k.Namespace)
	if err != nil {
		return nil, err
	}

	// sanity check the root key.
	if len(rootKey) < 16 {
		return nil, errors.New("root key is too short")
	}

	derivedKey := bytes.Join([][]byte{rootKey, k.ServiceKey}, []byte{})

	key := make([]byte, 32)
	keyReader := hkdf.New(sha256.New, derivedKey, nil, []byte(k.Namespace))
	_, err = keyReader.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// DeleteKey permanently deletes a key from the vault.
func (k *UserKey) DeleteKey(vault KeyVault) error {
	err := vault.DeleteKey(k.UserID, k.Namespace)
	return err
}

// CreateKey generates and inserts a new root key for the user.
func (k *UserKey) CreateKey(vault KeyVault) error {
	rootKey, err := k.generateKey()
	if err != nil {
		return err
	}

	err = vault.InsertKey(k.UserID, k.Namespace, rootKey)
	return err
}

// generateKey creates a 256-bit random byte array.
func (k *UserKey) generateKey() ([]byte, error) {
	// 256 bit
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}
