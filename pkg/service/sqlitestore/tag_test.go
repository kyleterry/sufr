package sqlitestore

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyleterry/sufr/pkg/api"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
)

func MustCreateRandomTag(t *testing.T, store *Store) *api.Tag {
	ctx := context.Background()

	tm := store.Tags()

	name := fmt.Sprintf("tag-%s", xid.New().String())
	require.NoError(t, tm.Create(ctx, &api.Tag{
		Name: name,
	}))

	newTag, err := tm.GetByName(ctx, name)
	require.NoError(t, err)

	return newTag
}

func TestCreateTag(t *testing.T) {
	WithTempDatabase(t, func(store *Store) {
		ctx := context.Background()

		tm := store.Tags()

		tag := api.Tag{
			Name: "test-tag",
		}

		require.NoError(t, tm.Create(ctx, &tag))

		newTag, err := tm.GetByName(ctx, "test-tag")
		require.NoError(t, err)

		require.NotEmpty(t, newTag.Id)
		require.Equal(t, "test-tag", newTag.Name)
		require.NotEmpty(t, newTag.Id)
		require.NotZero(t, newTag.CreatedAt.AsTime())
		require.Nil(t, newTag.UpdatedAt)
	})
}
