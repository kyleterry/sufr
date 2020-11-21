package data

import (
	"encoding/binary"
	"encoding/json"
	"math"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// PageIndex is a quick and dirty hack to cache the URL pages in bolt
type PageIndex struct {
	URLIDs     []uuid.UUID `json:"url_ids"`
	TotalPages int         `json:"total_pages"`
	PerPage    int         `json:"per_page"`

	URLs []*URL `json:"-"`
}

func (pi *PageIndex) fillURLS(tx *bolt.Tx) error {
	for _, id := range pi.URLIDs {
		url, err := getURL(id, tx)
		if err != nil {
			return errors.Wrap(err, "failed to get url")
		}

		pi.URLs = append(pi.URLs, url)
	}

	return nil
}

func GetPageIndexes() ([]*PageIndex, error) {
	var pis []*PageIndex

	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		pis, err = getPageIndexes(tx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return pis, nil
}

func getPageIndexes(tx *bolt.Tx) ([]*PageIndex, error) {
	var pis []*PageIndex

	bucket := tx.Bucket(buckets[pageIndexKey])

	bucket.ForEach(func(_, v []byte) error {
		pi := PageIndex{}

		if err := json.Unmarshal(v, &pi); err != nil {
			return err
		}

		pis = append(pis, &pi)

		return nil
	})

	return pis, nil
}

func GetPageIndexByPage(page int) (*PageIndex, error) {
	var pi *PageIndex

	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		pi, err = getPageIndexByPage(page, tx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return pi, nil
}

func getPageIndexByPage(page int, tx *bolt.Tx) (*PageIndex, error) {
	var pi PageIndex

	bucket := tx.Bucket(buckets[pageIndexKey])

	idb := make([]byte, 8)
	binary.BigEndian.PutUint64(idb, uint64(page-1))

	rawPageIndex := bucket.Get(idb)
	if len(rawPageIndex) == 0 {
		return nil, ErrNotFound
	}

	if err := json.Unmarshal(rawPageIndex, &pi); err != nil {
		return nil, errors.Wrap(err, "failed to decode object")
	}

	return &pi, nil
}

func CreatePageIndexes(perPage int) error {
	tx, err := db.bolt.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := createPageIndexes(perPage, tx); err != nil {
		return errors.Wrap(err, "failed to create page indexes")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

func createPageIndexes(perPage int, tx *bolt.Tx) error {
	var pis []*PageIndex

	urls, err := getURLs(tx)
	if err != nil {
		return err
	}

	totalPages := int(math.Ceil(float64(len(urls)) / float64(perPage)))

	pi := &PageIndex{
		TotalPages: totalPages,
		PerPage:    perPage,
	}

	for i, u := range urls {
		pi.URLIDs = append(pi.URLIDs, u.ID)

		if len(pi.URLIDs) == perPage || i == len(urls)-1 {
			pis = append(pis, pi)

			pi = &PageIndex{
				TotalPages: totalPages,
				PerPage:    perPage,
			}
		}
	}

	bucketKey := buckets[pageIndexKey]

	if err := tx.DeleteBucket(bucketKey); err != nil {
		return err
	}

	bucket, err := tx.CreateBucket(bucketKey)
	if err != nil {
		return err
	}

	var id uint64
	for _, pi := range pis {
		b, err := json.Marshal(pi)
		if err != nil {
			return err
		}

		idb := make([]byte, 8)
		binary.BigEndian.PutUint64(idb, id)

		if err := bucket.Put(idb, b); err != nil {
			return err
		}

		id++
	}

	return nil
}

func HACKCreatePageIndexes(perPage int, tx *bolt.Tx) error {
	return createPageIndexes(perPage, tx)
}
