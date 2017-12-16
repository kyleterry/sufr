package data

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type APIToken struct {
	ID        uuid.UUID `json:"id"`
	Token     uuid.UUID `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateAPIToken() (*APIToken, error) {
	tx, err := db.bolt.Begin(true)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	token, err := createAPIToken(tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create api token")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return token, nil
}

func createAPIToken(tx *bolt.Tx) (*APIToken, error) {
	token := uuid.New()
	now := time.Now()

	at := &APIToken{
		ID:        uuid.New(),
		Token:     token,
		CreatedAt: now,
		UpdatedAt: now,
	}

	bucket := tx.Bucket(buckets[apiTokenKey])

	b, err := json.Marshal(at)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize api token")
	}

	id, _ := at.ID.MarshalText()

	if err := bucket.Put(id, b); err != nil {
		return nil, errors.Wrap(err, "boltdb put failed")
	}

	return at, nil
}

func GetAPIToken() (*APIToken, error) {
	var token *APIToken

	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		token, err = getAPIToken(tx)
		if err != nil {
			return err
		}

		if token == nil {
			return ErrNotFound
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

func getAPIToken(tx *bolt.Tx) (*APIToken, error) {
	var token *APIToken

	bucket := tx.Bucket(buckets[apiTokenKey])

	err := bucket.ForEach(func(_, v []byte) error {
		var t APIToken
		if err := json.Unmarshal(v, &t); err != nil {
			return err
		}

		token = &t

		return nil
	})
	if err != nil {
		return nil, err
	}

	if token == nil {
		return nil, ErrNotFound
	}

	return token, nil
}

func DeleteAPITokens() error {
	tx, err := db.bolt.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := deleteAPITokens(tx); err != nil {
		return errors.Wrap(err, "failed to delete api tokens")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

func deleteAPITokens(tx *bolt.Tx) error {
	bucket := tx.Bucket(buckets[apiTokenKey])

	err := bucket.ForEach(func(k, _ []byte) error {
		if err := bucket.Delete(k); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
