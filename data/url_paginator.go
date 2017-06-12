package data

import (
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

// interface for fetching different sets of URLs for use in the paginator
type urlGetter interface {
	GetURLs(*bolt.Tx) ([]*URL, error)
}

// URLPaginator represents a paginated collection of URLs sorted by CreatedAt desc
type URLPaginator struct {
	NumRecords int
	Page       int
	PerPage    int
	URLs       []*URL
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

	p.NumRecords = count
	p.Page = page
	p.PerPage = perPage

	sub, err := p.urlSubset(urls)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get url subset")
	}

	p.URLs = sub

	return p, nil
}

func (p *URLPaginator) urlSubset(urls []*URL) ([]*URL, error) {
	var offset int
	if p.Page > 1 {
		offset = p.PerPage * (p.Page - 1)
	}

	end := offset + p.PerPage
	if end > p.NumRecords {
		end = p.NumRecords
	}

	return urls[offset:end], nil
}

func (p URLPaginator) HasPagination() bool { return p.NumRecords > p.PerPage }
func (p URLPaginator) TotalPages() int     { return p.NumRecords / p.PerPage }
func (p URLPaginator) CurrentPage() int    { return p.Page }
func (p URLPaginator) HasPrevious() bool   { return p.Page > 1 }
func (p URLPaginator) HasNext() bool       { return p.Page < p.TotalPages() }

func (p URLPaginator) PreviousPage() int {
	if p.Page == 1 {
		return p.Page
	}
	return p.Page - 1
}

func (p URLPaginator) NextPage() int {
	if p.Page == p.TotalPages() {
		return p.Page
	}
	return p.Page + 1
}

func (p URLPaginator) Pages() []int {
	pgs := []int{}
	i := 1
	for i <= p.TotalPages() {
		pgs = append(pgs, i)
		i++
	}

	return pgs
}
