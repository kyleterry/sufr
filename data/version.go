package data

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

var versionKey = []byte("version")

type Version struct {
	Version uint `json:"version"`
}

func CreateVersion(start uint, tx *bolt.Tx) (*Version, error) {
	version := &Version{}

	bucket := tx.Bucket(appBucket)

	version.Version = start

	b, err := json.Marshal(version)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize version object")
	}

	err = bucket.Put(versionKey, b)
	if err != nil {
		return nil, errors.Wrap(err, "failed to store version object")
	}

	return version, nil
}

func (v *Version) Increment(tx *bolt.Tx) error {
	bucket := tx.Bucket(appBucket)

	v.Version += 1

	b, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "failed to serialize version object")
	}

	if err := bucket.Put(versionKey, b); err != nil {
		return errors.Wrap(err, "failed to store version object")
	}

	return nil
}

func GetVersion(tx *bolt.Tx) (*Version, error) {
	var ver Version

	bucket := tx.Bucket(appBucket)

	rawVersion := bucket.Get(versionKey)
	if len(rawVersion) == 0 {
		return nil, ErrNotFound
	}

	if err := json.Unmarshal(rawVersion, &ver); err != nil {
		return nil, errors.Wrap(err, "failed to decode version object")
	}

	return &ver, nil
}
