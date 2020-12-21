package sqlitestore

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kyleterry/sufr/pkg/api"
	"github.com/rs/xid"
)

type urlManager struct {
	statementLoader
	store *Store
}

func (m *urlManager) Create(ctx context.Context, u *api.URL) error {
	return m.store.withTx(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		statement, err := m.getStatement("Create")
		if err != nil {
			return err
		}

		u.Id = xid.New().String()

		_, err = tx.NamedExecContext(ctx, statement, u)
		if err != nil {
			return fmt.Errorf("failed to create URL: %w", mapError(err))
		}

		return nil
	})
}

func (m *urlManager) GetByURL(ctx context.Context, us string) (*api.URL, error) {
	statement, err := m.getStatement("GetByURL")
	if err != nil {
		return nil, err
	}

	u := api.URL{}

	err = m.store.db.GetContext(ctx, &u, statement, us)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func newURLManager(store *Store) *urlManager {
	return &urlManager{
		statementLoader: statementLoader{
			dialect:     store.db.DriverName(),
			managerName: "URLManager",
		},
		store: store,
	}
}
