package migrations

import (
	"github.com/boltdb/bolt"
	"github.com/kyleterry/sufr/pkg/config"
	"github.com/kyleterry/sufr/pkg/data"
)

type CreateInitialPageIndexes struct{}

func (CreateInitialPageIndexes) Description() string {
	return "creates initial page indexes"
}

func (CreateInitialPageIndexes) Migrate(cfg *config.Config, tx *bolt.Tx) error {
	return data.HACKCreatePageIndexes(cfg.ResultsPerPage, tx)
}
