package sqlitestore

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kyleterry/sufr/pkg/api"
	"github.com/rs/xid"
)

type tagManager struct {
	statementLoader
	store *Store
}

func (m *tagManager) Create(ctx context.Context, tag *api.Tag) error {
	err := m.store.withTx(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		st, err := m.getStatement("Create")
		if err != nil {
			return err
		}

		tag.Id = xid.New().String()

		if _, err := tx.NamedExecContext(ctx, st, tag); err != nil {
			return fmt.Errorf("failed to create Tag: %w", mapError(err))
		}

		return nil
	})

	return err
}

func (m *tagManager) GetByID(ctx context.Context, id string) (*api.Tag, error) {
	st, err := m.getStatement("GetByID")
	if err != nil {
		return nil, err
	}

	tag := api.Tag{}

	err = m.store.db.GetContext(ctx, &tag, st, id)
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

func (m *tagManager) GetByName(ctx context.Context, name string) (*api.Tag, error) {
	st, err := m.getStatement("GetByName")
	if err != nil {
		return nil, err
	}

	tag := api.Tag{}

	err = m.store.db.GetContext(ctx, &tag, st, name)
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

func newTagManager(store *Store) *tagManager {
	return &tagManager{
		statementLoader: statementLoader{
			dialect:     store.db.DriverName(),
			managerName: "TagManager",
		},
		store: store,
	}
}
