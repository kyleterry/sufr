package data

import (
	"encoding/json"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var pinnedTagsKey = []byte("pinned_tags")

type PinnedTag struct {
	TagID   uuid.UUID `json:"tag_id"`
	Ordinal int       `json:"orginal"`
	Tag     *Tag      `json:"-"`
}

type PinnedTags []PinnedTag

func PinTag(tag *Tag) (*PinnedTags, error) {
	var pinnedTags *PinnedTags

	err := db.bolt.Update(func(tx *bolt.Tx) error {
		var err error

		pinnedTags, err = pinTag(tag, tx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return pinnedTags, nil
}

func pinTag(tag *Tag, tx *bolt.Tx) (*PinnedTags, error) {
	pinnedTags, err := getPinnedTags(tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pinned tags")
	}

	var found bool
	for _, pt := range *pinnedTags {
		if pt.TagID == tag.ID {
			found = true
		}
	}

	if !found {
		*pinnedTags = append(*pinnedTags, PinnedTag{TagID: tag.ID})
	}

	bucket := tx.Bucket(buckets[appKey])

	b, err := json.Marshal(pinnedTags)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode pinned tags")
	}

	if err := bucket.Put(pinnedTagsKey, b); err != nil {
		return nil, errors.Wrap(err, "failed to store pinned tags")
	}

	return pinnedTags, nil
}

func UnpinTag(tag *Tag) (*PinnedTags, error) {
	var pinnedTags *PinnedTags

	err := db.bolt.Update(func(tx *bolt.Tx) error {
		var err error

		// TODO: ordinals
		pinnedTags, err = unpinTag(tag, tx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return pinnedTags, nil
}

func unpinTag(tag *Tag, tx *bolt.Tx) (*PinnedTags, error) {
	pinnedTags, err := getPinnedTags(tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get pinned tags")
	}

	newPinnedTags := &PinnedTags{}

	for _, pt := range *pinnedTags {
		// TODO: reorder the ordinals
		if pt.TagID != tag.ID {
			*newPinnedTags = append(*newPinnedTags, pt)
		}
	}

	bucket := tx.Bucket(buckets[appKey])

	b, err := json.Marshal(newPinnedTags)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode pinned tags")
	}

	if err := bucket.Put(pinnedTagsKey, b); err != nil {
		return nil, errors.Wrap(err, "failed to store pinned tags")
	}

	return newPinnedTags, nil
}

func GetPinnedTags() (*PinnedTags, error) {
	var pinnedTags *PinnedTags

	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		pinnedTags, err = getPinnedTags(tx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return pinnedTags, nil
}

func getPinnedTags(tx *bolt.Tx) (*PinnedTags, error) {
	var pinnedTags PinnedTags

	bucket := tx.Bucket(buckets[appKey])

	b := bucket.Get(pinnedTagsKey)

	if len(b) == 0 {
		return &pinnedTags, nil
	}

	if err := json.Unmarshal(b, &pinnedTags); err != nil {
		return nil, errors.Wrap(err, "failed to decode object")
	}

	for i := range pinnedTags {
		tag, err := getTag(pinnedTags[i].TagID, tx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get tag %s", pinnedTags[i].TagID)
		}

		pinnedTags[i].Tag = tag
	}

	return &pinnedTags, nil
}
