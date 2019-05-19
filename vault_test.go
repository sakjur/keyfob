package keyfob

import (
	"bytes"
	"testing"

	"github.com/google/uuid"

	"github.com/sakjur/keyfob/fake"
)

var serviceKey = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
var serviceKey2 = []byte{0xff, 0xfe, 0xfd, 0xfc, 0xfb, 0xfa, 0xf9, 0xf8, 0xf7, 0xf6, 0xf5, 0xf4, 0xf3, 0xf2, 0xf1, 0xf0}

func TestUserKey_DeriveKey_givenShortServiceKey(t *testing.T) {
	userKey := getUserKey()
	userKey.ServiceKey = []byte{0x00, 0x01, 0x02, 0x03}

	vault := fake.NewVault()

	key, err := userKey.DeriveKey(vault)

	if err != errServiceKeyTooShort {
		t.Errorf("expected %s, got %s", errServiceKeyTooShort, err)
	}
	if key != nil {
		t.Error("expected key to be nil, got ", key)
	}
}

func TestUserKey_DeriveKey_simple(t *testing.T) {
	userKey := getUserKey()
	vault := fake.NewVault()

	err := userKey.CreateKey(vault)
	if err != nil {
		t.Error("expected err to be nil, got ", err)
	}

	key, err := userKey.DeriveKey(vault)
	if err != nil {
		t.Error("expected err to be nil, got ", err)
	}
	if key == nil {
		t.Error("expected key to be set, got nil")
	}

	sameKey, _ := userKey.DeriveKey(vault)
	if !bytes.Equal(sameKey, key) {
		t.Errorf("expected %x, got %x", key, sameKey)
	}
}

func TestUserKey_DeriveKey_differentServices(t *testing.T) {
	userKey := getUserKey()
	vault := fake.NewVault()

	err := userKey.CreateKey(vault)
	if err != nil {
		t.Error("expected err to be nil, got ", err)
	}

	key, err := userKey.DeriveKey(vault)
	if err != nil {
		t.Error("expected err to be nil, got ", err)
	}
	if key == nil {
		t.Error("expected key to be set, got nil")
	}

	userKey.ServiceKey = serviceKey2
	anotherKey, _ := userKey.DeriveKey(vault)
	if bytes.Equal(anotherKey, key) {
		t.Errorf("expected the new service to have another key, but it was equal to %x", key)
	}
}

func getUserKey() *UserKey {
	return &UserKey{
		UserID:     uuid.New(),
		Namespace:  "contact",
		ServiceKey: serviceKey,
	}
}
