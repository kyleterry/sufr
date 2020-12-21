package sqlitestore

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
	"github.com/kyleterry/sufr/pkg/api"
	"github.com/kyleterry/sufr/pkg/store"
	"github.com/mattn/go-sqlite3"
)

type storeOptions struct {
	path string
}

type storeOptionFunc struct {
	f func(*storeOptions)
}

func (s *storeOptionFunc) apply(opts *storeOptions) {
	s.f(opts)
}

type StoreOption interface {
	apply(*storeOptions)
}

func WithPath(path string) StoreOption {
	return &storeOptionFunc{
		f: func(opts *storeOptions) {
			opts.path = path
		},
	}
}

type Store struct {
	db *sqlx.DB
}

func (s *Store) URLs() store.URLManager {
	return newURLManager(s)
}

func (s *Store) Tags() store.TagManager {
	return newTagManager(s)
}

func (s *Store) UserURLs(user *api.User) store.UserURLManager {
	return newUserURLManager(s, user)
}

func (s *Store) Users() store.UserManager {
	return newUserManager(s)
}

func (s *Store) Migrate(ctx context.Context) error {
	return runAllMigrations(ctx, s)
}

func (s *Store) withTx(ctx context.Context, fn txFunc) (err error) {
	var tx *sqlx.Tx

	tx, err = s.db.BeginTxx(ctx, nil)
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("tx callback error: %w", err)

			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("tx rollback error: %w; original error: %v", rollbackErr, err)
			}
		} else {
			err = tx.Commit()

			if err != nil {
				err = fmt.Errorf("tx commit error: %w", err)
			}
		}
	}()

	err = fn(ctx, tx)

	return
}

// New takes a path to a sqlite database file and returns a Store
func New(opts ...StoreOption) (*Store, error) {
	so := storeOptions{}

	for _, opt := range opts {
		opt.apply(&so)
	}

	dbOptions := url.Values{}
	dbOptions.Set("_foreign_keys", "true")

	dbURL := url.URL{
		Scheme:   "file",
		Path:     so.path,
		RawQuery: dbOptions.Encode(),
	}

	db, err := sqlx.Connect("sqlite3", dbURL.String())
	if err != nil {
		return nil, err
	}

	db.Mapper = reflectx.NewMapperFunc("json", strings.ToLower)

	return &Store{
		db: db,
	}, nil
}

type txFunc func(ctx context.Context, tx *sqlx.Tx) error

type statementLoader struct {
	managerName string
	dialect     string
}

func (l statementLoader) getStatement(name string) (string, error) {
	filename := fmt.Sprintf("%s.%s.generated", l.managerName, name)
	p := filepath.Join(l.dialect, filename)

	return getSQL(p)
}

func getSQL(name string) (string, error) {
	f, err := assets.Open(filepath.Join("/sql", name+".sql"))
	if err != nil {
		return "", err
	}

	defer f.Close()

	b := &bytes.Buffer{}

	if _, err := b.ReadFrom(f); err != nil {
		return "", err
	}

	return b.String(), nil
}

func mapError(err error) error {
	if slErr, ok := err.(sqlite3.Error); ok {
		if slErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			err = store.ErrAlreadyExists
		} else if slErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
			err = store.ErrInvalidDependency
		}
	} else {
		if errors.Is(err, sql.ErrNoRows) {
			err = store.ErrNotFound
		}
	}

	return err
}
