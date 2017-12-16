package data

import "github.com/boltdb/bolt"

type AllURLGetter struct{}

func (AllURLGetter) GetURLs(tx *bolt.Tx) ([]*URL, error) {
	return getURLs(tx)
}

type FavURLGetter struct{}

func (FavURLGetter) GetURLs(tx *bolt.Tx) ([]*URL, error) {
	return getFavoriteURLs(tx)
}

type SearchURLGetter struct {
	query string
}

func NewSearchURLGetter(query string) SearchURLGetter {
	return &SearchURLGetter{query}
}

func (s SearchURLGetter) GetURLs(tx *bolt.Tx) ([]*URL, error) {
	return getResultingURLsFromSearch(s.query)
}
