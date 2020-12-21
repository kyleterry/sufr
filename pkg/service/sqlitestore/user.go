package sqlitestore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kyleterry/sufr/pkg/api"
	"github.com/rs/xid"
)

type userManager struct {
	statementLoader
	store *Store
}

func (m *userManager) Create(ctx context.Context, user *api.User) error {
	return m.store.withTx(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		st, err := m.getStatement("Create")
		if err != nil {
			return err
		}

		user.Id = xid.New().String()

		_, err = tx.NamedExecContext(ctx, st, user)
		if err != nil {
			return err
		}

		return m.updatePinnedCategories(ctx, tx, user)
	})
}

func (m *userManager) UpdatePinnedCategories(ctx context.Context, user *api.User) error {
	return m.store.withTx(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		return m.updatePinnedCategories(ctx, tx, user)
	})
}

func (m *userManager) GetByID(ctx context.Context, id string) (*api.User, error) {
	st, err := m.getStatement("GetByID")
	if err != nil {
		return nil, err
	}

	user := api.User{}

	if err := m.store.db.GetContext(ctx, &user, st, id); err != nil {
		return nil, err
	}

	user.PinnedCategories, err = m.getPinnedCategories(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pinned categories: %w", err)
	}

	return &user, nil
}

func (m *userManager) GetByEmailAndPassword(ctx context.Context, email, password string) (*api.User, error) {
	user, err := m.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := api.CompareHashAndPassword(user, password); err != nil {
		return nil, err
	}

	return user, nil
}

func (m *userManager) GetByEmail(ctx context.Context, email string) (*api.User, error) {
	st, err := m.getStatement("GetByEmail")
	if err != nil {
		return nil, err
	}

	user := api.User{}

	if err := m.store.db.GetContext(ctx, &user, st, email); err != nil {
		return nil, err
	}

	user.PinnedCategories, err = m.getPinnedCategories(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to get pinned categories: %w", err)
	}

	return &user, nil
}

func (m *userManager) getPinnedCategories(ctx context.Context, user *api.User) ([]*api.Category, error) {
	st, err := m.getStatement("getPinnedCategories")
	if err != nil {
		return nil, err
	}

	cats := []*api.Category{}

	if err := m.store.db.SelectContext(ctx, &cats, st, user.Id); err != nil {
		return nil, err
	}

	return cats, nil
}

func (m *userManager) updatePinnedCategories(ctx context.Context, tx *sqlx.Tx, user *api.User) error {
	st, err := m.getStatement("UpdatePinnedCategories")
	if err != nil {
		return err
	}

	b, err := json.Marshal(user.PinnedCategories)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, st, b, user.Id)

	return err
}

func newUserManager(store *Store) *userManager {
	return &userManager{
		statementLoader: statementLoader{
			dialect:     store.db.DriverName(),
			managerName: "UserManager",
		},
		store: store,
	}
}
