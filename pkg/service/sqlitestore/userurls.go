package sqlitestore

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/kyleterry/sufr/pkg/api"
	"github.com/kyleterry/sufr/pkg/store"
	"github.com/rs/xid"
)

const pageSize = 100

type userURLManager struct {
	statementLoader
	user  *api.User
	store *Store
}

func (m *userURLManager) Create(ctx context.Context, userURL *api.UserURL) error {
	return m.store.withTx(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		st, err := m.getStatement("Create")
		if err != nil {
			return err
		}

		userURL.Id = xid.New().String()
		userURL.User = m.user

		if _, err := tx.NamedExecContext(ctx, st, userURL); err != nil {
			return fmt.Errorf("failed to create UserURL: %w", mapError(err))
		}

		tagFn := m.updateTagsFunc(userURL.Id, userURL.Tags)

		if err := tagFn(ctx, tx); err != nil {
			return fmt.Errorf("failed to create tags for new UserURL: %w", mapError(err))
		}

		return nil
	})
}

func (m *userURLManager) Update(ctx context.Context, userURL *api.UserURL) error {
	return m.store.withTx(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		st, err := m.getStatement("Update")
		if err != nil {
			return err
		}

		if _, err := tx.NamedExecContext(ctx, st, userURL); err != nil {
			return fmt.Errorf("failed to update UserURL: %w", mapError(err))
		}

		tagFn := m.updateTagsFunc(userURL.Id, userURL.Tags)

		if err := tagFn(ctx, tx); err != nil {
			return fmt.Errorf("failed to create tags for new UserURL: %w", mapError(err))
		}

		return nil
	})
}

func (m *userURLManager) updateTagsFunc(id string, tags *api.TagList) txFunc {
	return func(ctx context.Context, tx *sqlx.Tx) error {
		st, err := m.getStatement("clearTags")
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, st, id); err != nil {
			return err
		}

		st, err = m.getStatement("updateTags")
		if err != nil {
			return err
		}

		for _, tag := range tags.Items {
			if _, err = tx.ExecContext(ctx, st, id, tag.Id); err != nil {
				return err
			}
		}

		return nil
	}
}

func (m *userURLManager) GetAll(ctx context.Context, filters ...store.FilterOption) ([]*api.UserURL, error) {
	st, err := m.getStatement("GetAll")
	if err != nil {
		return nil, err
	}

	opts := store.FilterOptions{}

	for _, filter := range filters {
		filter.Apply(&opts)
	}

	uus := []*api.UserURL{}
	if err := m.store.db.SelectContext(ctx, &uus, st, m.user.Id); err != nil {
		return nil, fmt.Errorf("failed to get UserURLs: %w", mapError(err))
	}

	// sub := selectQuery(
	// 	columns(
	// 		as(window("row_number()", orderBy("user_url.id")), "row"),
	// 		column("user_url.id"),
	// 		column("url.id"),
	// 		column("url.url"),
	// 		column("url.title"),
	// 		as(column("user_url.user_id"), "user.id"),
	// 		column("user_url.favorite"),
	// 		column("user_url.created_at"),
	// 		column("user_url.updated_at"),
	// 		as(
	// 			function(
	// 				"coalesce",
	// 				function("nullif", column("user_url.title"), sqlstring("''")),
	// 				column("user_url.title"),
	// 			),
	// 			"derived_title",
	// 		),
	// 	),
	// 	from(tableAs(table("user_urls"), "user_url")),
	// 	join("urls url", on("url.id", "=", "user_url.url_id")),
	// 	where("user_url.user_id", "=", "?"),
	// )

	// search := selectQuery(
	// 	columns(
	// 		column("uu.*"),
	// 		as(function("json_object", sqlstring("items"),
	// 			function("json_group_array",
	// 				function("json_object",
	// 					sqlstring("id"),
	// 					column("t.id"),
	// 					sqlstring("name"),
	// 					column("t.name"),
	// 				),
	// 			),
	// 		), "tags"),
	// 	),
	// 	from(tableAs(sub, "uu")),
	// 	join("user_url_tags ut", on("ut.user_url_id", "=", "uu.id")),
	// 	join("tags tag", on("tag.id", "=", "tag.tag_id")),
	// 	where("row", ">", "?"),
	// 	groupBy("uu.id"),
	// )

	// st := search.build()

	return uus, nil
}

// func (m *userURLManager) GetAllAfter(ctx context.Context, after int64) ([]*api.UserURL, error) {
// 	st, err := m.getStatement("GetAllAfter")
// 	if err != nil {
// 		return nil, err
// 	}

// 	uus := []*api.UserURL{}
// 	if err := m.store.db.SelectContext(ctx, &uus, st, m.user.Id, after, after+pageSize); err != nil {
// 		return nil, err
// 	}

// 	return uus, nil
// }

// func (m *userURLManager) GetAllByTags(ctx context.Context, tags *api.TagList) ([]*api.UserURL, error) {
// 	st, err := m.getStatement("GetAllByTags")
// 	if err != nil {
// 		return nil, err
// 	}

// 	uus := []*api.UserURL{}

// 	q, args, err := sqlx.In(st, m.user.Id, tags)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := m.store.db.SelectContext(ctx, &uus, q, args...); err != nil {
// 		return nil, err
// 	}

// 	return uus, nil
// }

func (m *userURLManager) GetByURLID(ctx context.Context, urlID string) (*api.UserURL, error) {
	st, err := m.getStatement("GetByURLID")
	if err != nil {
		return nil, err
	}

	uu := api.UserURL{}

	if err := m.store.db.GetContext(ctx, &uu, st, m.user.Id, urlID); err != nil {
		return nil, err
	}

	return &uu, nil
}

func newUserURLManager(store *Store, user *api.User) *userURLManager {
	return &userURLManager{
		statementLoader: statementLoader{
			dialect:     store.db.DriverName(),
			managerName: "UserURLManager",
		},
		store: store,
		user:  user,
	}
}
