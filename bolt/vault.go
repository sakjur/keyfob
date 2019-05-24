package bolt

import (
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

func (v *vault) ListKeys(userid uuid.UUID) ([]*keyfob.StoredKey, error) {
	keys := []*keyfob.StoredKey{}

	err := v.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketUserKeyStore)
		if bucket == nil {
			return errNoSuchKey
		}

		userbucket := bucket.Bucket(userid[:])
		if userbucket == nil {
			return errNoSuchKey
		}

		err := userbucket.ForEach(func(k, v []byte) error {
			keys = append(keys, &keyfob.StoredKey{
				Category: string(k),
				Key:      v,
			})
			return nil
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (v *vault) GetKey(userid uuid.UUID, category string) (*keyfob.StoredKey, error) {
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

		storedKey := userbucket.Get([]byte(category))
		if storedKey == nil {
			return errNoSuchKey
		}
		key = storedKey
		return nil
	})
	return &keyfob.StoredKey{
		Key:      key,
		Category: category,
	}, err
}

func (v *vault) InsertKey(userid uuid.UUID, category string, key []byte) error {
	return v.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketUserKeyStore)
		if err != nil {
			return err
		}

		userbucket, err := bucket.CreateBucketIfNotExists(userid[:])
		if err != nil {
			return err
		}

		row := []byte(category)
		curr := userbucket.Get(row)
		if curr != nil {
			// Key already exists, skip insertion.
			return nil
		}

		return userbucket.Put(row, key)
	})
}

func (v *vault) DeleteKey(userid uuid.UUID, category string) error {
	return v.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketUserKeyStore)
		if bucket == nil {
			return nil
		}

		userbucket := bucket.Bucket(userid[:])
		if userbucket == nil {
			return nil
		}

		return userbucket.Delete([]byte(category))
	})
}
