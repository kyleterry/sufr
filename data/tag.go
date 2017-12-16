package data

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type CreateTagOptions struct {
	Name string
}

// Tag holds the little information we have about url tags
type Tag struct {
	ID        uuid.UUID   `json:"id"`
	Name      string      `json:"name"`
	URLs      []*URL      `json:"-"`
	URLIDs    []uuid.UUID `json:"url_ids"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// TODO: made public to help with the migration from the old db.
// This shouldn't be like this and I would like to find a solution
func (t *Tag) AddURL(url *URL, tx *bolt.Tx) error {
	for _, id := range t.URLIDs {
		if id == url.ID {
			return nil
		}
	}

	t.URLIDs = append(t.URLIDs, url.ID)

	bucket := tx.Bucket(buckets[tagKey])

	b, err := json.Marshal(t)
	if err != nil {
		return errors.Wrap(err, "failed to update tag")
	}

	id, _ := t.ID.MarshalText()
	if err := bucket.Put(id, b); err != nil {
		return errors.Wrap(err, "failed to put tag")
	}

	return nil
}

func (t *Tag) removeURL(url *URL, tx *bolt.Tx) error {
	urlIDs := []uuid.UUID{}

	for _, id := range t.URLIDs {
		if id != url.ID {
			urlIDs = append(urlIDs, id)
		}
	}

	t.URLIDs = urlIDs

	bucket := tx.Bucket(buckets[tagKey])

	b, err := json.Marshal(t)
	if err != nil {
		return errors.Wrap(err, "failed to update tag")
	}

	id, _ := t.ID.MarshalText()
	if err := bucket.Put(id, b); err != nil {
		return errors.Wrap(err, "failed to put tag")
	}

	return nil
}

// GetURLs satisfies the urlGetter interface for the URLPaginator
func (t *Tag) GetURLs(_ *bolt.Tx) ([]*URL, error) {
	sort.Sort(URLsByDateDesc(t.URLs))
	return t.URLs, nil
}

func CreateTag(opts CreateTagOptions) (*Tag, error) {
	tx, err := db.bolt.Begin(true)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	tag, err := createTag(opts, tx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create tag")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return tag, nil
}

func createTag(opts CreateTagOptions, tx *bolt.Tx) (*Tag, error) {
	if opts.Name == "" {
		return nil, errors.New("validation errors: tag name cannot be blank")
	}

	now := time.Now()

	tag := &Tag{
		ID:        uuid.New(),
		Name:      opts.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	bucket := tx.Bucket(buckets[tagKey])

	b, err := json.Marshal(tag)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize tag")
	}

	id, _ := tag.ID.MarshalText()
	if err := bucket.Put(id, b); err != nil {
		return nil, errors.Wrap(err, "boltdb put failed")
	}

	return tag, nil
}

func GetTag(id uuid.UUID) (*Tag, error) {
	var tag *Tag

	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		tag, err = getTag(id, tx)
		if err != nil {
			return err
		}

		for _, urlID := range tag.URLIDs {
			url, err := getURL(urlID, tx)
			if err != nil {
				return errors.Wrap(err, "failed to get url")
			}

			if err = url.fillTags(tx); err != nil {
				return err
			}

			tag.URLs = append(tag.URLs, url)
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return tag, nil
}

func getTag(tagID uuid.UUID, tx *bolt.Tx) (*Tag, error) {
	var tag Tag
	bucket := tx.Bucket(buckets[tagKey])

	id, _ := tagID.MarshalText()
	rawTag := bucket.Get(id)

	if len(rawTag) == 0 {
		return nil, ErrNotFound
	}

	if err := json.Unmarshal(rawTag, &tag); err != nil {
		return nil, errors.Wrap(err, "failed to decode object")
	}

	return &tag, nil
}

func getOrCreateTag(name string, tx *bolt.Tx) (*Tag, bool, error) {
	var created bool

	tag, err := getTagByName(name, tx)
	if err != nil {
		if err == ErrNotFound {
			created = true

			tag, err := createTag(CreateTagOptions{name}, tx)
			if err != nil {
				return nil, created, errors.Wrap(err, "failed to create tag")
			}

			return tag, created, nil
		}

		return nil, created, errors.Wrap(err, "failed to get tag")
	}

	return tag, created, nil
}

func getTagByName(name string, tx *bolt.Tx) (*Tag, error) {
	var tag *Tag

	bucket := tx.Bucket(buckets[tagKey])

	err := bucket.ForEach(func(_, v []byte) error {
		var t Tag
		if err := json.Unmarshal(v, &t); err != nil {
			return err
		}

		if t.Name == name {
			tag = &t
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	if tag == nil {
		return nil, ErrNotFound
	}

	return tag, nil
}
