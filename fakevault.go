package keyfob

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

var errUserNotFound = errors.New("User not found")
var errKeyNotFound = errors.New("Key not found")

func NewVault() KeyVault {
	return &fakeKeyVault{
		Keys: make(map[uuid.UUID]map[string][]byte),
	}
}

type fakeKeyVault struct {
	Keys map[uuid.UUID]map[string][]byte
	lock sync.RWMutex
}

func (v *fakeKeyVault) ListKeys(userid uuid.UUID) ([]*StoredKey, error) {
	userKeys := v.Keys[userid]
	if userKeys == nil {
		return nil, errUserNotFound
	}

	keys := make([]*StoredKey, len(userKeys))
	i := 0
	for category, key := range userKeys {
		keys[i] = &StoredKey{
			Key:      key,
			User:     userid,
			Category: category,
		}
		i++
	}

	return keys, nil
}

func (v *fakeKeyVault) GetKey(userid uuid.UUID, category string) (*StoredKey, error) {
	userKeys := v.Keys[userid]
	if userKeys == nil {
		return nil, errUserNotFound
	}

	key := userKeys[category]
	if key == nil {
		return nil, errKeyNotFound
	}

	return &StoredKey{
		Key:      key,
		User:     userid,
		Category: category,
	}, nil
}

func (v *fakeKeyVault) InsertKey(userid uuid.UUID, category string, key []byte) error {
	v.lock.Lock()
	if _, found := v.Keys[userid]; !found {
		v.Keys[userid] = make(map[string][]byte)
	}
	if _, found := v.Keys[userid][category]; !found {
		v.Keys[userid][category] = key
	}
	v.lock.Unlock()

	return nil
}

func (v *fakeKeyVault) DeleteKey(userid uuid.UUID, category string) error {
	userKeys := v.Keys[userid]
	if userKeys == nil {
		return nil
	}

	userKeys[category] = nil
	return nil
}
