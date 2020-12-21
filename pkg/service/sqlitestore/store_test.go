package sqlitestore

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func WithTempDatabase(t *testing.T, fn func(store *Store)) {
	tempdir, err := ioutil.TempDir("", "sufr-test-*")
	require.NoError(t, err)

	database := filepath.Join(tempdir, "sufr.db")
	t.Logf("temporary database: %s", database)

	s, err := New(WithPath(database))
	require.NoError(t, err)
	require.NoError(t, s.Migrate(context.Background()))

	fn(s)

	require.NoError(t, os.RemoveAll(tempdir))
}
