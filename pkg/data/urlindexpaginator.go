package data

import (
	"github.com/boltdb/bolt"
)

const (
	defaultPerPage     = 40
	defaultPagePadding = 3
)

type pageIndexPaginator struct {
	index       *PageIndex
	currentPage int
	pagePadding int
}

func (p *pageIndexPaginator) CurrentPage() int {
	return p.currentPage
}

func (p *pageIndexPaginator) HasPagination() bool {
	return p.index.TotalPages > 1
}

func (p *pageIndexPaginator) HasNext() bool {
	return p.currentPage < p.index.TotalPages
}

func (p *pageIndexPaginator) HasPrevious() bool {
	return p.currentPage > 1
}

func (p *pageIndexPaginator) Pages() []int {
	pgs := []int{}

	var i int

	for i < p.TotalPages() {
		pgs = append(pgs, i+1)

		i++
	}

	maxPages := (p.pagePadding * 2) + 1

	var start, stop int

	if p.index.TotalPages > maxPages {
		start = p.currentPage - (p.pagePadding + 1)
		stop = p.currentPage + p.pagePadding

		if start < 0 {
			start, stop = 0, maxPages
		} else if (p.index.TotalPages - p.currentPage) < p.pagePadding {
			start, stop = p.index.TotalPages-maxPages, p.index.TotalPages
		}
	}

	return pgs[start:stop]
}

func (p *pageIndexPaginator) NextPage() int {
	if p.currentPage == p.index.TotalPages {
		return p.currentPage
	}

	return p.currentPage + 1
}

func (p *pageIndexPaginator) PreviousPage() int {
	if p.currentPage == 1 {
		return p.currentPage
	}

	return p.currentPage - 1
}

func (p *pageIndexPaginator) TotalPages() int {
	return p.index.TotalPages
}

func (p *pageIndexPaginator) URLs() []*URL {
	return p.index.URLs
}

func NewPageIndexPaginator(page int) (*pageIndexPaginator, error) {
	pi := &PageIndex{}

	if page == 0 {
		page = 1
	}

	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		pi, err = GetPageIndexByPage(page)
		if err != nil {
			return err
		}

		if err := pi.fillURLS(tx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &pageIndexPaginator{
		index:       pi,
		currentPage: page,
		pagePadding: defaultPagePadding,
	}, nil
}
