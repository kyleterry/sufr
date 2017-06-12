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

var (
	appBucket = []byte("_app")
	urlBucket = []byte("_urls")
	tagBucket = []byte("_tags")
)

// SufrDB is a BoltDB wrapper that provides a SUFR specific interface to the DB
type SufrDB struct {
	path string
	bolt *bolt.DB

	sync.Mutex
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

// DBWithLock runs func fn with the global db object. Locked so nothing else can use
// the DB while migrations are running.
func DBWithLock(fn func(*bolt.DB)) {
	db.Lock()
	defer db.Unlock()
	fn(db.bolt)
}

// RunWithBucketForType takes a *bolt.Tx and will find the bucket for type m
// and then run will run the func fn with the bucket (if found). If no bucket is
// registered for type m, an error is returned.
func RunWithBucketForType(tx *bolt.Tx, m interface{}, fn func(*bolt.Bucket) error) error {
	switch m.(type) {
	case URL:
		return fn(tx.Bucket(urlBucket))
	case Tag:
		return fn(tx.Bucket(tagBucket))
	case Settings, PinnedTag, PinnedTags, User:
		return fn(tx.Bucket(appBucket))
	}

	return errors.Errorf("no such bucket for type %v", m)
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
		log.Println("initializing database")
		_, err := tx.CreateBucketIfNotExists(appBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(urlBucket)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(tagBucket)
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
