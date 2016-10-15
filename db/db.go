package db

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

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
	Serialize() ([]byte, error)
}

// New creates and returns a new pointer to a SufrDB struct
func New(path string) *SufrDB {
	return &SufrDB{path: path}
}

// Runs in it's own goroutine if debug is on
func (sdb *SufrDB) Statsdumper() {
	prev := sdb.database.Stats()
	for {
		time.Sleep(10 * time.Second)
		stats := sdb.database.Stats()
		diff := stats.Sub(&prev)
		json.NewEncoder(os.Stderr).Encode(diff)
		prev = stats
	}
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

		_, err = b.CreateBucketIfNotExists([]byte(config.BucketNameUser))
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
		b := tx.Bucket([]byte(config.BucketNameRoot))
		if item.Type() != config.BucketNameRoot {
			b = b.Bucket([]byte(item.Type()))
		}
		id := item.GetID()
		if id == 0 {
			id, _ = b.NextSequence()
			item.SetID(id)
		}
		content, err := item.Serialize()
		if err != nil {
			return err
		}
		b.Put(itob(id), content)
		return nil
	})
	return err
}

// Get will return raw bytes for an item at `id` or return an error
func (sdb *SufrDB) Get(id uint64, bucket string) ([]byte, error) {
	var item []byte
	err := sdb.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.BucketNameRoot))
		if bucket != config.BucketNameRoot {
			b = b.Bucket([]byte(bucket))
		}
		item = b.Get(itob(id))
		return nil
	})
	return item, err
}

// GetAll is used to fetch all of the recods for a particular bucket
// Returns a slice of []byte or an error
func (sdb *SufrDB) GetAll(bucket string) ([][]byte, error) {
	items := [][]byte{}
	err := sdb.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.BucketNameRoot))
		if bucket != config.BucketNameRoot {
			b = b.Bucket([]byte(bucket))
		}
		b.ForEach(func(_, v []byte) error {
			items = append(items, v)
			return nil
		})

		return nil
	})
	return items, err
}

func (sdb *SufrDB) GetDesc(bucket string) ([][]byte, error) {
	items := [][]byte{}
	err := sdb.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.BucketNameRoot))
		if bucket != config.BucketNameRoot {
			b = b.Bucket([]byte(bucket))
		}
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			items = append(items, v)
		}
		return nil
	})
	return items, err
}

func (sdb *SufrDB) GetSubset(offset uint64, n uint64, bucket string) ([][]byte, error) {
	items := [][]byte{}
	err := sdb.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.BucketNameRoot))
		if bucket != config.BucketNameRoot {
			b = b.Bucket([]byte(bucket))
		}

		c := b.Cursor()

		var k, v []byte
		k, v = c.Last()

		for offset > 0 {
			k, v = c.Prev()
			if k == nil {
				return nil
			}
			offset--
		}

		if k != nil {
			items = append(items, v)
		}

		for n-1 > 0 {
			k, v := c.Prev()
			if k == nil {
				return nil
			}
			items = append(items, v)
			n--
		}

		return nil
	})
	return items, err
}

func (sdb *SufrDB) LatestItem(bucket string) ([]byte, error) {
	item := []byte{}
	err := sdb.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.BucketNameRoot))
		if bucket != config.BucketNameRoot {
			b = b.Bucket([]byte(bucket))
		}

		c := b.Cursor()

		_, item = c.Last()

		return nil
	})
	return item, err
}

// Delete will return raw bytes for an item at `id` or return an error
func (sdb *SufrDB) Delete(id uint64, bucket string) error {
	err := sdb.database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.BucketNameRoot))
		if bucket != config.BucketNameRoot {
			b = b.Bucket([]byte(bucket))
		}
		return b.Delete(itob(id))
	})
	return err
}

// GetValuesByField will find objects who's `fieldname` value matches any of the `values`
// If any of the objects are lacking `fieldname` when deserialized, then it returns an error
// Return [][]byte, a []string of objects not found or an error
func (sdb *SufrDB) GetValuesByField(fieldname, bucket string, values ...string) ([][]byte, []string, error) {
	items := [][]byte{}
	err := sdb.database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.BucketNameRoot))
		if bucket != config.BucketNameRoot {
			b = b.Bucket([]byte(bucket))
		}
		err := b.ForEach(func(k []byte, v []byte) error {
			j := make(map[string]interface{})
			if err := json.Unmarshal(v, &j); err != nil {
				return err
			}
			if _, ok := j[fieldname]; !ok {
				return fmt.Errorf("Field `%s` doesn't exist", fieldname)
			}
			valuestring := j[fieldname].(string)
			for i, need := range values {
				if need == valuestring {
					items = append(items, v)
					// If we find a match, remove this one from the values slice so we can return
					// it as notfound values
					values = append(values[:i], values[i+1:]...)
				}
			}
			return nil
		})
		return err
	})
	return items, values, err
}

func (sdb *SufrDB) BucketLength(bucket string) (int, error) {
	var count int
	err := sdb.database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(config.BucketNameRoot))
		if bucket != config.BucketNameRoot {
			b = b.Bucket([]byte(bucket))
		}

		b.ForEach(func(_, v []byte) error {
			count++
			return nil
		})

		return nil
	})

	return count, err
}

func (sdb *SufrDB) BackupHandler(w http.ResponseWriter, req *http.Request) error {
	err := sdb.database.View(func(tx *bolt.Tx) error {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="sufr.db"`)
		w.Header().Set("Content-Length", strconv.Itoa(int(tx.Size())))
		_, err := tx.WriteTo(w)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
