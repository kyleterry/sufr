package data

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var ErrUserExists = errors.New("user already exists")

type UserOptions struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUser will create the global user account. This is used to login and run api queries.
// There can only be one user in the system, so if an account exists, it will return an error.
func CreateUser(opts UserOptions) (*User, error) {
	user, err := GetUser()
	if err == nil {
		return nil, errors.Wrap(ErrUserExists, "failed to create user")
	}

	if errors.Cause(err) != ErrNotFound {
		return nil, errors.Wrap(err, "failed to create user")
	}

	tx, err := db.bolt.Begin(true)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	user, err = createUser(opts, tx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create user")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return user, nil
}

func createUser(opts UserOptions, tx *bolt.Tx) (*User, error) {
	bucket := tx.Bucket([]byte(AppBucket))

	now := time.Now()

	user := &User{
		ID:        uuid.New(),
		Email:     opts.Email,
		Password:  opts.Password,
		CreatedAt: now,
		UpdatedAt: now,
	}

	b, err := json.Marshal(user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize user")
	}

	if err := bucket.Put([]byte("user"), b); err != nil {
		return nil, errors.Wrap(err, "boltdb put failed")
	}

	return user, nil
}

//GetUser returns the sufr user used to login and perform api queries
func GetUser() (*User, error) {
	tx, err := db.bolt.Begin(true)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	user, err := getUser(tx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return user, nil
}

func getUser(tx *bolt.Tx) (*User, error) {
	var user User
	bucket := tx.Bucket([]byte(AppBucket))

	rawUser := bucket.Get([]byte("user"))
	if len(rawUser) == 0 {
		return nil, ErrNotFound
	}

	if err := json.Unmarshal(rawUser, &user); err != nil {
		return nil, errors.Wrap(err, "failed to decode object")
	}

	return &user, nil
}
