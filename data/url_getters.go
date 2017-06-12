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
