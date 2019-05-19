package bolt

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/sakjur/keyfob"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

const unixPermOwnerRW = 0600

var bucketUserKeyStore = []byte("user_key_store")
var errNoSuchKey = errors.New("secret key not found")

// NewVault creates a KeyVault which is suitable for use in locally running
// vaults without high availability requirements. Bolt only allows a single
// connection to the database at a time, so only a single Keyfob may use a
// given Bolt vault at a time.
func NewVault() (keyfob.KeyVault, error) {
	db, err := bbolt.Open("dead.bolt", unixPermOwnerRW, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	return &vault{
		db: db,
	}, nil
}

type vault struct {
	db *bbolt.DB
}

func (v *vault) Close() {
	_ = v.db.Close()
}

func (v *vault) GetKey(userid uuid.UUID, namespace string) ([]byte, error) {
	var key []byte

	err := v.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketUserKeyStore)
		if bucket == nil {
			return errNoSuchKey
		}

		userbucket := bucket.Bucket(userid[:])
		if userbucket == nil {
			return errNoSuchKey
		}

		storedKey := userbucket.Get([]byte(namespace))
		if storedKey == nil {
			return errNoSuchKey
		}
		key = storedKey
		return nil
	})
	return key, err
}

func (v *vault) InsertKey(userid uuid.UUID, namespace string, key []byte) error {
	return v.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketUserKeyStore)
		if err != nil {
			return err
		}

		userbucket, err := bucket.CreateBucketIfNotExists(userid[:])
		if err != nil {
			return err
		}

		row := []byte(namespace)
		curr := userbucket.Get(row)
		if curr != nil {
			// Key already exists, skip insertion.
			return nil
		}

		return userbucket.Put(row, key)
	})
}

func (v *vault) DeleteKey(userid uuid.UUID, namespace string) error {
	return v.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketUserKeyStore)
		if bucket == nil {
			return nil
		}

		userbucket := bucket.Bucket(userid[:])
		if userbucket == nil {
			return nil
		}

		return userbucket.Delete([]byte(namespace))
	})
}

func serializeRowKey(userid uuid.UUID, namespace string) ([]byte, error) {
	type rowKey struct {
		userid    string
		namespace string
	}

	return json.Marshal(rowKey{
		userid:    userid.String(),
		namespace: namespace,
	})
}
