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
	"github.com/kyleterry/sufr/pkg/config"
	"github.com/pkg/errors"
)

var (
	db   *SufrDB
	once sync.Once

	ErrDatabaseAlreadyOpen = errors.New("database is already open")
	ErrNotFound            = errors.New("object not found")
	ErrDuplicateKey        = errors.New("duplicate key")
)

type bucketKey int

const (
	appKey bucketKey = iota
	urlKey
	tagKey
	apiTokenKey
)

var (
	buckets = map[bucketKey][]byte{
		appKey:      []byte("_app"),
		urlKey:      []byte("_urls"),
		tagKey:      []byte("_tags"),
		apiTokenKey: []byte("_api_tokens"),
	}
)

// SufrDB is a BoltDB wrapper that provides a SUFR specific interface to the DB
type SufrDB struct {
	path string
	bolt *bolt.DB

	sync.Mutex
}

func MustInit(cfg *config.Config) {
	var err error

	once.Do(func() {
		db, err = New(cfg.DatabaseFile())
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
	var key bucketKey

	switch m.(type) {
	case URL:
		key = urlKey
	case Tag:
		key = tagKey
	case APIToken:
		key = apiTokenKey
	case Settings, PinnedTag, PinnedTags, User:
		key = appKey
	default:
		return errors.Errorf("no such bucket for type %v", m)
	}

	return fn(tx.Bucket(buckets[key]))
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
		// TODO make buckets a map[string][]byte or something
		log.Println("initializing database")

		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return err
			}
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
