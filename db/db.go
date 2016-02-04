package db

import (
	"encoding/binary"
	"log"

	"github.com/boltdb/bolt"
	"github.com/kyleterry/sufr/config"
)

// SufrDB is a BoltDB wrapper that provides a SUFR specific interface to the DB
type SufrDB struct {
	path     string
	database *bolt.DB
}

// SufrItem is an interface to describe how to serialize a storable item for the
// Sufr Database
type SufrItem interface {
	GetID() uint64
	SetID(uint64)
	Type() string
	Serialize() []byte
}

// New creates and returns a new pointer to a SufrDB struct
func New(path string) *SufrDB {
	return &SufrDB{path: path}
}

//Open will open the bolt database and panic on error
func (sdb *SufrDB) Open() error {
	if sdb.database != nil {
		return config.ErrDatabaseAlreadyOpen
	}
	db, err := bolt.Open(sdb.path, 0600, nil)
	if err != nil {
		return err
	}

	sdb.database = db

	// Make sure the sufr bucket exists so we can use it later
	err = sdb.database.Update(func(tx *bolt.Tx) error {
		log.Println("Creating database buckets")
		b, err := tx.CreateBucketIfNotExists([]byte(config.BucketNameRoot))
		if err != nil {
			return err
		}

		_, err = b.CreateBucketIfNotExists([]byte(config.BucketNameURL))
		if err != nil {
			return err
		}

		_, err = b.CreateBucketIfNotExists([]byte(config.BucketNameTag))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Close will close the BoltDB instance
func (sdb *SufrDB) Close() {
	sdb.database.Close()
}

// Put will create or update a SufrItem
func (sdb *SufrDB) Put(item SufrItem) error {
	err := sdb.database.Update(func(tx *bolt.Tx) error {
		rootBucket := tx.Bucket([]byte(config.BucketNameRoot))
		b := rootBucket.Bucket([]byte(item.Type()))
		id := item.GetID()
		if id == 0 {
			id, _ = b.NextSequence()
			item.SetID(id)
		}
		b.Put(itob(id), item.Serialize())
		return nil
	})
	return err
}

// Get will return raw bytes for an item at `id` or return an error
func (sdb *SufrDB) Get(id int, bucket string) ([]byte, error) {
	return []byte{}, nil
}

// GetAll is used to fetch all of the recods for a particular bucket
// Returns a slice of []byte or an error
func (sdb *SufrDB) GetAll(bucket string) ([][]byte, error) {
	items := [][]byte{}
	err := sdb.database.View(func(tx *bolt.Tx) error {
		rootBucket := tx.Bucket([]byte(config.BucketNameRoot))
		b := rootBucket.Bucket([]byte(bucket))
		b.ForEach(func(_, v []byte) error {
			items = append(items, v)
			return nil
		})

		return nil
	})
	return items, err
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
