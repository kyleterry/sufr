package sqlitestore

import (
	"context"
	"database/sql"
	"log"
	"path/filepath"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kyleterry/sufr/pkg/api"
	"github.com/rs/xid"
)

var migrations = []Migration{
	InitializeDatabase{},
	AddInitialAdminUser{},
}

type Migration interface {
	Run(ctx context.Context, tx *sqlx.Tx) error
	Version() string
	Description() string
}

type MigrationRecord struct {
	Version   string     `json:"version"`
	CreatedAt *time.Time `json:"created_at"`
}

func runAllMigrations(ctx context.Context, store *Store) error {
	return store.withTx(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		if err := createMigrationsTable(ctx, tx); err != nil {
			return err
		}

		for _, m := range migrations {
			log.Printf("migration: %s - %s", m.Version(), m.Description())

			hasRan, err := migrationHasRan(ctx, tx, m)
			if err != nil {
				return err
			}

			if !hasRan {
				if err := m.Run(ctx, tx); err != nil {
					return err
				}

				if err := recordMigration(ctx, tx, m); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func createMigrationsTable(ctx context.Context, tx *sqlx.Tx) error {
	statement, err := getSQL(filepath.Join("migrations", "migrations-table"))
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, statement); err != nil {
		return err
	}

	return nil
}

func migrationHasRan(ctx context.Context, tx *sqlx.Tx, m Migration) (bool, error) {
	st := `select version, created_at from migrations where version = ?`

	if err := tx.GetContext(ctx, &MigrationRecord{}, st, m.Version()); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func recordMigration(ctx context.Context, tx *sqlx.Tx, m Migration) error {
	mr := MigrationRecord{
		Version: m.Version(),
	}

	st := `insert into migrations(version) values(?)`

	if _, err := tx.ExecContext(ctx, st, mr.Version); err != nil {
		return err
	}

	return nil
}

// InitializeDatabase is a migration that creates all the initial database tables.
type InitializeDatabase struct{}

func (m InitializeDatabase) Description() string {
	return "initializing database"
}

func (m InitializeDatabase) Version() string {
	return "001"
}

func (m InitializeDatabase) Run(ctx context.Context, tx *sqlx.Tx) error {
	st, err := getSQL(filepath.Join("migrations", "001-init"))
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, st); err != nil {
		return err
	}

	return nil
}

type AddInitialAdminUser struct{}

func (m AddInitialAdminUser) Description() string {
	return "adding initial admin user"
}

func (m AddInitialAdminUser) Version() string {
	return "002"
}

func (m AddInitialAdminUser) Run(ctx context.Context, tx *sqlx.Tx) error {
	st, err := getSQL(filepath.Join("sqlite3", "UserManager.Create.generated"))
	if err != nil {
		return err
	}

	pw, err := api.GeneratePasswordHash("admin")
	if err != nil {
		return err
	}

	admin := &api.User{
		Id:           xid.New().String(),
		Email:        "admin@localhost",
		PasswordHash: pw,
	}

	if _, err := tx.NamedExecContext(ctx, st, admin); err != nil {
		return err
	}

	return nil
}
