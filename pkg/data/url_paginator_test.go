package data

import (
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

type MockURLGetter struct {
	gimme int
}

func (m MockURLGetter) GetURLs(_ *bolt.Tx) ([]*URL, error) {
	urls := []*URL{}
	for i := 0; i < m.gimme; i++ {
		urls = append(urls, &URL{})
	}

	return urls, nil
}

func TestCanPaginateURLs(t *testing.T) {
	var cases = []struct {
		num                  int
		perPage              int
		shouldHavePagination bool
		shouldTotalPages     int
	}{
		{40, 40, false, 1},
		{60, 40, true, 2},
		{120, 40, true, 3},
	}

	for _, c := range cases {
		paginator, err := NewURLPaginator(1, c.perPage, 3, MockURLGetter{c.num})
		assert.NoError(t, err)
		if c.shouldHavePagination {
			assert.True(t, paginator.HasPagination())
		} else {
			assert.False(t, paginator.HasPagination())
		}
		assert.Equal(t, c.shouldTotalPages, paginator.TotalPages())
	}
}

func TestPaginatorCanScrubPages(t *testing.T) {
	var cases = []struct {
		numItems    int
		perPage     int
		currentPage int
		pagePadding int
		pages       []int
	}{
		{480, 40, 1, 3, []int{1, 2, 3, 4, 5, 6, 7}},
		{480, 40, 12, 3, []int{6, 7, 8, 9, 10, 11, 12}},
		{480, 40, 6, 3, []int{3, 4, 5, 6, 7, 8, 9}},
		{480, 40, 3, 3, []int{1, 2, 3, 4, 5, 6, 7}},
	}

	for _, c := range cases {
		paginator, err := NewURLPaginator(
			c.currentPage,
			c.perPage,
			c.pagePadding,
			MockURLGetter{c.numItems})

		assert.NoError(t, err)
		assert.Equal(t, c.pages, paginator.Pages())
	}
}
