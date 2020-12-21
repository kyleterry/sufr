package data

import (
	"encoding/binary"
	"encoding/json"
	"math"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type PageIndexType int

const (
	PageIndexTypeAllURLs PageIndexType = iota
	PageIndexTypeFavoriteURLs
)

// PageIndex is a quick and dirty hack to cache the URL pages in bolt
type PageIndex struct {
	URLIDs       []uuid.UUID `json:"url_ids"`
	TotalPages   int         `json:"total_pages"`
	TotalRecords int         `json:"total_records"`
	PerPage      int         `json:"per_page"`

	URLs []*URL `json:"-"`
}

func (pi *PageIndex) fillURLS(tx *bolt.Tx) error {
	for _, id := range pi.URLIDs {
		url, err := getURL(id, tx)
		if err != nil {
			return errors.Wrap(err, "failed to get url")
		}

		if err := url.fillTags(tx); err != nil {
			return errors.Wrap(err, "failed to fill url tags")
		}

		pi.URLs = append(pi.URLs, url)
	}

	return nil
}

type PageIndexManager struct {
	bucketKey []byte
	urlGetter urlGetter
}

func (pim *PageIndexManager) GetAllPageIndexes() ([]*PageIndex, error) {
	var pis []*PageIndex

	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		pis, err = pim.getAllPageIndexes(tx)
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

func (pim *PageIndexManager) getAllPageIndexes(tx *bolt.Tx) ([]*PageIndex, error) {
	var pis []*PageIndex

	bucket := tx.Bucket(pim.bucketKey)

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

func (pim *PageIndexManager) GetPageIndexByPage(page int) (*PageIndex, error) {
	var pi *PageIndex

	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		pi, err = pim.getPageIndexByPage(page, tx)
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

func (pim *PageIndexManager) getPageIndexByPage(page int, tx *bolt.Tx) (*PageIndex, error) {
	var pi PageIndex

	bucket := tx.Bucket(pim.bucketKey)

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

func (pim *PageIndexManager) CreatePageIndexes(perPage int) error {
	tx, err := db.bolt.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := pim.createPageIndexes(perPage, tx); err != nil {
		return errors.Wrap(err, "failed to create page indexes")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

func (pim *PageIndexManager) createPageIndexes(perPage int, tx *bolt.Tx) error {
	var pis []*PageIndex

	urls, err := pim.urlGetter.GetURLs(tx)
	if err != nil {
		return err
	}

	totalURLs := len(urls)

	totalPages := int(math.Ceil(float64(totalURLs) / float64(perPage)))

	newPageIndex := func() *PageIndex {
		return &PageIndex{
			TotalPages:   totalPages,
			TotalRecords: totalURLs,
			PerPage:      perPage,
		}
	}

	pi := newPageIndex()

	for i, u := range urls {
		pi.URLIDs = append(pi.URLIDs, u.ID)

		if len(pi.URLIDs) == perPage || i == totalURLs-1 {
			pis = append(pis, pi)

			pi = newPageIndex()
		}
	}

	if err := tx.DeleteBucket(pim.bucketKey); err != nil {
		return err
	}

	bucket, err := tx.CreateBucket(pim.bucketKey)
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

func NewPageIndexManager(pit PageIndexType) *PageIndexManager {
	var bucketKey []byte
	var ug urlGetter

	switch pit {
	case PageIndexTypeAllURLs:
		bucketKey = buckets[allURLsIndex]
		ug = AllURLGetter{}
	case PageIndexTypeFavoriteURLs:
		bucketKey = buckets[favoriteURLsIndex]
		ug = FavURLGetter{}
	}

	return &PageIndexManager{
		bucketKey: bucketKey,
		urlGetter: ug,
	}
}

func HACKCreatePageIndexes(perPage int, tx *bolt.Tx) error {
	if err := NewPageIndexManager(PageIndexTypeAllURLs).createPageIndexes(perPage, tx); err != nil {
		return err
	}

	return nil
}
