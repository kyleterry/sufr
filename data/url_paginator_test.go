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
		paginator, err := NewURLPaginator(1, c.perPage, MockURLGetter{c.num})
		assert.NoError(t, err)
		if c.shouldHavePagination {
			assert.True(t, paginator.HasPagination())
		} else {
			assert.False(t, paginator.HasPagination())
		}
		assert.Equal(t, c.shouldTotalPages, paginator.TotalPages())
	}
}
