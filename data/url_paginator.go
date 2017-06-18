package data

import (
	"math"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

// interface for fetching different sets of URLs for use in the paginator
type urlGetter interface {
	GetURLs(*bolt.Tx) ([]*URL, error)
}

// URLPaginator represents a paginated collection of URLs sorted by CreatedAt desc
type URLPaginator struct {
	URLs []*URL

	numRecords int
	page       int
	perPage    int
}

// NewURLPaginator returns a filled-out *URLPaginator.
// returns an error if something during the read transaction fails.
func NewURLPaginator(page int, perPage int, getter urlGetter) (*URLPaginator, error) {
	var p *URLPaginator

	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		p, err = newURLPaginator(page, perPage, getter, tx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return p, nil
}

func newURLPaginator(page, perPage int, getter urlGetter, tx *bolt.Tx) (*URLPaginator, error) {
	p := &URLPaginator{}

	urls, err := getter.GetURLs(tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get URLs")
	}

	count := len(urls)

	p.numRecords = count
	p.page = page
	p.perPage = perPage

	sub, err := p.urlSubset(urls)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get url subset")
	}

	p.URLs = sub

	return p, nil
}

func (p *URLPaginator) urlSubset(urls []*URL) ([]*URL, error) {
	var offset int
	if p.page > 1 {
		offset = p.perPage * (p.page - 1)
	}

	end := offset + p.perPage
	if end > p.numRecords {
		end = p.numRecords
	}

	return urls[offset:end], nil
}

func (p URLPaginator) HasPagination() bool { return p.numRecords > p.perPage }
func (p URLPaginator) CurrentPage() int    { return p.page }
func (p URLPaginator) HasPrevious() bool   { return p.page > 1 }
func (p URLPaginator) HasNext() bool       { return p.page < p.TotalPages() }

func (p URLPaginator) TotalPages() int {
	return int(math.Ceil(float64(p.numRecords) / float64(p.perPage)))
}
func (p URLPaginator) PreviousPage() int {
	if p.page == 1 {
		return p.page
	}
	return p.page - 1
}

func (p URLPaginator) NextPage() int {
	if p.page == p.TotalPages() {
		return p.page
	}
	return p.page + 1
}

func (p URLPaginator) Pages() []int {
	pgs := []int{}
	i := 0
	for i < p.TotalPages() {
		pgs = append(pgs, i+1)
		i++
	}

	return pgs
}
