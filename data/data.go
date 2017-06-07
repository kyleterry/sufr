package data

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/kyleterry/sufr/config"
	"github.com/pkg/errors"
)

var (
	db   *SufrDB
	once sync.Once

	ErrDatabaseAlreadyOpen = errors.New("database is already open")
	ErrNotFound            = errors.New("object not found")
	ErrDuplicateKey        = errors.New("duplicate key")
)

// SufrDB is a BoltDB wrapper that provides a SUFR specific interface to the DB
type SufrDB struct {
	path string
	bolt *bolt.DB
}

func MustInit() {
	var err error

	once.Do(func() {
		db, err = New(config.DatabaseFile)
		if err != nil {
			panic(errors.Wrap(err, "failed to open database"))
		}
	})
}

// New creates and returns a new pointer to a SufrDB struct
func New(path string) (*SufrDB, error) {
	db := &SufrDB{path: path}

	if err := db.Open(); err != nil {
		return nil, err
	}

	return db, nil
}

//Open will open the bolt database and panic on error
func (s *SufrDB) Open() error {
	var err error

	if s.bolt != nil {
		return ErrDatabaseAlreadyOpen
	}

	if s.bolt, err = bolt.Open(s.path, 0600, nil); err != nil {
		return err
	}

	// Make sure the sufr bucket exists so we can use it later
	err = s.bolt.Update(func(tx *bolt.Tx) error {
		// TODO better logging
		log.Println("Creating database buckets")
		_, err := tx.CreateBucketIfNotExists([]byte(AppBucket))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(URLBucket))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte(TagBucket))
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
func (s *SufrDB) Close() {
	s.bolt.Close()
}

// Runs in it's own goroutine if debug is on
func (s *SufrDB) Statsdumper() {
	prev := s.bolt.Stats()
	for {
		time.Sleep(10 * time.Second)
		stats := s.bolt.Stats()
		diff := stats.Sub(&prev)
		json.NewEncoder(os.Stderr).Encode(diff)
		prev = stats
	}
}

func (s *SufrDB) GetSubset(offset uint64, n uint64, bucket string) ([][]byte, error) {
	items := [][]byte{}
	err := s.bolt.View(func(tx *bolt.Tx) error {
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

func BucketLength(bucket string) (int, error) {
	var count int
	err := db.bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		b.ForEach(func(_, v []byte) error {
			count++
			return nil
		})

		return nil
	})

	return count, err
}

func BackupHandler(w http.ResponseWriter, req *http.Request) error {
	err := db.bolt.View(func(tx *bolt.Tx) error {
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
