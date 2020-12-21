package sqlitestore

import (
	"context"
	"testing"

	"github.com/kyleterry/sufr/pkg/api"
	"github.com/stretchr/testify/require"
)

const (
	BasicTestUserPassword = "password"
	BasicTestUserEmail    = "basic-test-user@unit-testing.sufr.io"
)

func MustCreateBasicTestUser(t *testing.T, store *Store) *api.User {
	ctx := context.Background()
	um := store.Users()

	ph, err := api.GeneratePasswordHash(BasicTestUserPassword)
	require.NoError(t, err)

	require.NoError(t, um.Create(ctx, &api.User{
		Email:        BasicTestUserEmail,
		PasswordHash: ph,
	}))

	newUser, err := um.GetByEmail(ctx, BasicTestUserEmail)
	require.NoError(t, err)

	return newUser
}

func TestCreateUser(t *testing.T) {
	WithTempDatabase(t, func(store *Store) {
		ctx := context.Background()

		um := store.Users()

		ph, err := api.GeneratePasswordHash(BasicTestUserPassword)
		require.NoError(t, err)

		user := api.User{
			Email:        BasicTestUserEmail,
			PasswordHash: ph,
		}

		require.NoError(t, um.Create(ctx, &user))

		t.Run("can get created user by email", func(t *testing.T) {
			newUser, err := um.GetByEmail(ctx, BasicTestUserEmail)
			require.NoError(t, err)
			require.NotEmpty(t, newUser.Id)
			require.Equal(t, BasicTestUserEmail, newUser.Email)
			require.NoError(t, api.CompareHashAndPassword(newUser, BasicTestUserPassword))
			require.False(t, newUser.EmbedContent)
			require.NotZero(t, newUser.CreatedAt.AsTime())
			require.Nil(t, newUser.UpdatedAt)
		})

		t.Run("can get created user by email and password", func(t *testing.T) {
			newUser, err := um.GetByEmailAndPassword(ctx, user.Email, BasicTestUserPassword)
			require.NoError(t, err)
			require.NotEmpty(t, newUser.Id)
			require.Equal(t, BasicTestUserEmail, newUser.Email)
			require.False(t, newUser.EmbedContent)
			require.NotZero(t, newUser.CreatedAt.AsTime())
			require.Nil(t, newUser.UpdatedAt)
		})

		t.Run("can get created user by id with sensitive info missing", func(t *testing.T) {
			newUser, err := um.GetByID(ctx, user.Id)
			require.NoError(t, err)
			require.NotEmpty(t, newUser.Id)
			require.Equal(t, BasicTestUserEmail, newUser.Email)
			require.Empty(t, newUser.PasswordHash)
			require.Empty(t, newUser.ApiToken)
			require.False(t, newUser.EmbedContent)
			require.NotZero(t, newUser.CreatedAt.AsTime())
			require.Nil(t, newUser.UpdatedAt)
		})
	})
}

func TestUserPinnedCategories(t *testing.T) {
	WithTempDatabase(t, func(store *Store) {
		ctx := context.Background()
		um := store.Users()
		tagset1 := &api.TagList{
			Items: []*api.Tag{
				MustCreateRandomTag(t, store),
				MustCreateRandomTag(t, store),
				MustCreateRandomTag(t, store),
			},
		}
		tagset2 := &api.TagList{
			Items: []*api.Tag{
				MustCreateRandomTag(t, store),
				MustCreateRandomTag(t, store),
			},
		}

		ph, err := api.GeneratePasswordHash(BasicTestUserPassword)
		require.NoError(t, err)

		user := api.User{
			Email:        BasicTestUserEmail,
			PasswordHash: ph,
			PinnedCategories: []*api.Category{
				{
					Label: "tagset1",
					Tags:  tagset1,
				},
				{
					Label: "tagset2",
					Tags:  tagset2,
				},
			},
		}

		require.NoError(t, um.Create(ctx, &user))

		newUser, err := um.GetByEmail(ctx, BasicTestUserEmail)
		require.NoError(t, err)

		require.Len(t, newUser.PinnedCategories, 2)

		newCat := api.Category{
			Label: "tagset3",
			Tags:  tagset2,
		}

		cats := api.CategoryListInsert(newUser.PinnedCategories, &newCat, 0)

		newUser.PinnedCategories = cats

		require.NoError(t, um.UpdatePinnedCategories(ctx, newUser))

		{
			user, err := um.GetByEmail(ctx, newUser.Email)
			require.NoError(t, err)

			require.Len(t, user.PinnedCategories, 3)
			// require.Equal(t, newCat.Label, user.PinnedCategories[0].Label)
		}
	})
}
