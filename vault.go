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
	ListKeys(userid uuid.UUID) ([]*StoredKey, error)
	GetKey(userid uuid.UUID, category string) (*StoredKey, error)
	// Repeated insertion of key should not update the key. Only the first
	// key inserted should be persisted. Collision on persistence is not
	// considered an error.
	InsertKey(userid uuid.UUID, category string, key []byte) error
	DeleteKey(userid uuid.UUID, category string) error
}

type StoredKey struct {
	Key      []byte
	User     uuid.UUID
	Category string
}

// UserKeyPointer is a struct which maps out a single cryptographic key from a
// KeyVault.
type UserKeyPointer struct {
	UserID     uuid.UUID
	Category   string
	ServiceKey []byte
}

// ListUserKeys lists all keys for a user combined with the service key to
// lower the number of round-trips that a service needs to do in order to
// decrypt user data.
func (k *UserKeyPointer) ListUserKeys(vault KeyVault) ([]*StoredKey, error) {
	rootKeys, err := vault.ListKeys(k.UserID)
	if err != nil {
		return nil, err
	}

	var keys []*StoredKey
	for _, root := range rootKeys {
		key, err := (&UserKeyPointer{
			UserID:     k.UserID,
			Category:   root.Category,
			ServiceKey: k.ServiceKey,
		}).deriveKey(root)
		if err != nil {
			return nil, err
		}

		keys = append(keys, &StoredKey{
			Key:      key,
			User:     k.UserID,
			Category: root.Category,
		})
	}

	return keys, nil
}

// DeriveKey fetches the root key for the user within the data category and
// combines it with the service key to derive a new key which the service
// can use to encrypt data securely.
//
// The service key must be longer than 128 bits.
func (k *UserKeyPointer) DeriveKey(vault KeyVault) ([]byte, error) {
	root, err := vault.GetKey(k.UserID, k.Category)
	if err != nil {
		return nil, err
	}

	return k.deriveKey(root)
}

func (k *UserKeyPointer) deriveKey(root *StoredKey) ([]byte, error) {
	if len(k.ServiceKey) < 16 {
		return nil, errServiceKeyTooShort
	}

	// sanity check the root key.
	if len(root.Key) < 16 {
		return nil, errors.New("root key is too short")
	}

	derivedKey := bytes.Join([][]byte{root.Key, k.ServiceKey}, []byte{})

	key := make([]byte, 32)
	keyReader := hkdf.New(sha256.New, derivedKey, nil, []byte(k.Category))
	_, err := keyReader.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// DeleteKey permanently deletes a key from the vault.
func (k *UserKeyPointer) DeleteKey(vault KeyVault) error {
	err := vault.DeleteKey(k.UserID, k.Category)
	return err
}

// CreateKey generates and inserts a new root key for the user.
func (k *UserKeyPointer) CreateKey(vault KeyVault) error {
	rootKey, err := k.generateKey()
	if err != nil {
		return err
	}

	err = vault.InsertKey(k.UserID, k.Category, rootKey)
	return err
}

// generateKey creates a 256-bit random byte array.
func (k *UserKeyPointer) generateKey() ([]byte, error) {
	// 256 bit
	key := make([]byte, 32)

	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}
