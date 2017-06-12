// Package migrations contains code that modifies data in the boltdb database
// as application logic changes. If a new field is added to a model, there is a good
// chance there will be an accompanying migration to help facilitate the change.
// Everything in here is ran in a single transaction, so any errors will rollback
// all changes made by migrations in this file.
package migrations

import (
	"log"

	"github.com/boltdb/bolt"
	"github.com/kyleterry/sufr/data"
	"github.com/pkg/errors"
)

// legacyBucket is the bucket that stored the original db objects
// since I moved to a new "schema", I need to check if this bucket exists.
// if it does, then I need to run the initial migration.
var legacyBucket = []byte("sufr")

// Migration is an interface type that represents a single migration action.
// Types that satisfy this interface can modify data using the transaction passed
// into it as a parameter.
type Migration interface {
	Description() string
	Migrate(tx *bolt.Tx) error
}

// Migrations holds the list of migrations that need to be ran.
// Version.Version is a count of migrations that have been ran so far and it's
// used as an index into this slice to find pending migrations.
var Migrations = []Migration{
	MigrateOldDatabase{},
}

func run(tx *bolt.Tx) error {
	defer tx.Rollback()

	var version *data.Version

	version, err := data.GetVersion(tx)
	if err != nil {
		if errors.Cause(err) == data.ErrNotFound {
			migrationLen := uint(0)
			bucket := tx.Bucket(legacyBucket)
			// if we are new, we just skip the migrations
			if bucket == nil {
				migrationLen = uint(len(Migrations))
			}

			version, err = data.CreateVersion(migrationLen, tx)
			if err != nil {
				log.Println("creating initial migations failed")
				return errors.Wrap(err, "failed to create version")
			}

			log.Println("new instance; no migrations to run")

			// if the old bucket doesn't exist, go ahead and commit and skip migrations
			if bucket == nil {
				goto Commit
			}
		} else {
			return errors.Wrap(err, "failed to get version")
		}
	}

	if len(Migrations[version.Version:]) == 0 {
		log.Println("no pending migrations")
		goto Commit
	}

	log.Println("starting migrations")
	for _, migration := range Migrations[version.Version:] {
		log.Printf("running migration: %s, version %d...", migration.Description(), version.Version+1)
		if err := migration.Migrate(tx); err != nil {
			log.Println("error migrating ; rolling back.", "error", err)
			tx.Rollback()

			return err
		}
		log.Println("done")

		err := version.Increment(tx)
		if err != nil {
			return errors.Wrap(err, "failed to increment version")
		}
	}

Commit:
	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

func MustMigrate() {
	data.DBWithLock(func(db *bolt.DB) {
		tx, err := db.Begin(true)
		if err != nil {
			panic(errors.Wrap(err, "failed to start transaction"))
		}

		if err := run(tx); err != nil {
			panic(err)
		}

	})
}
