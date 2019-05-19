package fake

import (
	"sync"

	"github.com/google/uuid"
	"github.com/sakjur/keyfob"
)

type rowKey struct {
	userID    uuid.UUID
	namespace string
}

func NewVault() keyfob.KeyVault {
	return &KeyVault{
		Keys: make(map[rowKey][]byte),
	}
}

type KeyVault struct {
	Keys map[rowKey][]byte
	lock sync.RWMutex
}

type notFoundErr struct {
	msg string
}

func (e notFoundErr) Error() string {
	return e.msg
}

func (notFoundErr) KeyNotFound() {
}

func (v *KeyVault) GetKey(userid uuid.UUID, namespace string) ([]byte, error) {
	row := rowKey{userID: userid, namespace: namespace}

	key := v.Keys[row]
	if key == nil {
		return nil, notFoundErr{msg: "Key not found"}
	}

	return key, nil
}

func (v *KeyVault) InsertKey(userid uuid.UUID, namespace string, key []byte) error {
	row := rowKey{userID: userid, namespace: namespace}

	v.lock.Lock()
	if _, found := v.Keys[row]; !found {
		v.Keys[row] = key
	}
	v.lock.Unlock()

	return nil
}

func (v *KeyVault) DeleteKey(userid uuid.UUID, namespace string) error {
	row := rowKey{userID: userid, namespace: namespace}
	v.Keys[row] = nil
	return nil
}
