package sqlitestore

import (
	"context"
	"errors"
	"testing"

	"github.com/kyleterry/sufr/pkg/api"
	"github.com/kyleterry/sufr/pkg/store"
	"github.com/stretchr/testify/require"
)

func TestUserURLCreate(t *testing.T) {
	WithTempDatabase(t, func(db *Store) {
		ctx := context.Background()

		user := MustCreateBasicTestUser(t, db)
		url := MustCreateRandomURL(t, db)
		tag := MustCreateRandomTag(t, db)
		tag2 := MustCreateRandomTag(t, db)

		uum := db.UserURLs(user)

		t.Run("can create a user's url and get by url_id", func(t *testing.T) {
			tags := api.TagList{
				Items: []*api.Tag{tag, tag2},
			}

			uu := &api.UserURL{
				Url:   url,
				User:  user,
				Title: "user's title",
				Tags:  &tags,
			}

			require.NoError(t, uum.Create(ctx, uu))

			newUserURL, err := uum.GetByURLID(ctx, url.Id)
			require.NoError(t, err)

			require.NotEmpty(t, newUserURL.Id)
			require.Equal(t, user.Id, newUserURL.User.Id)
			require.Equal(t, url.Id, newUserURL.Url.Id)
			require.Equal(t, uu.Title, newUserURL.DerivedTitle)
			require.Len(t, newUserURL.Tags.Items, 2)
			require.False(t, newUserURL.Favorite)
			require.NotZero(t, newUserURL.CreatedAt.AsTime())
			require.Nil(t, newUserURL.UpdatedAt)
		})

		t.Run("can't create a user's url with a non-existent url_id", func(t *testing.T) {
			urlCopy := *url
			urlCopy.Id = "invalid"

			uu := &api.UserURL{
				Url:  &urlCopy,
				User: user,
			}

			err := uum.Create(ctx, uu)
			require.Error(t, err)
			require.True(t, errors.Is(err, store.ErrInvalidDependency), err)

			t.Log(err)
		})
		// TODO test these
		// all, err := uum.GetAll(ctx)
		// require.NoError(t, err)

		// require.Len(t, all, 1)

		// tagged, err := uum.GetAllByTags(ctx, &tags)
		// require.NoError(t, err)
		// require.Len(t, tagged, 1)
	})
}

func TestUserURLUpdate(t *testing.T) {
	WithTempDatabase(t, func(db *Store) {
		ctx := context.Background()

		user := MustCreateBasicTestUser(t, db)
		url := MustCreateRandomURL(t, db)
		tag := MustCreateRandomTag(t, db)
		tags := api.TagList{
			Items: []*api.Tag{tag},
		}

		uum := db.UserURLs(user)

		uu := &api.UserURL{
			Url:  url,
			User: user,
			Tags: &tags,
		}

		require.NoError(t, uum.Create(ctx, uu))

		newUserURL, err := uum.GetByURLID(ctx, url.Id)

		require.NoError(t, err)
		require.Equal(t, url.Title, newUserURL.DerivedTitle)
		require.Len(t, newUserURL.Tags.Items, 1)

		// change newUserURL and see if it persists
		tag2 := MustCreateRandomTag(t, db)
		newUserURL.Title = "our new title"
		newUserURL.Favorite = true
		newUserURL.Tags.Items = append(newUserURL.Tags.Items, tag2)

		require.NoError(t, uum.Update(ctx, newUserURL))

		updated, err := uum.GetByURLID(ctx, url.Id)
		require.NoError(t, err)

		require.Equal(t, newUserURL.Title, updated.DerivedTitle)
		require.True(t, updated.Favorite)
		require.Len(t, updated.Tags.Items, 2)
		require.NotZero(t, updated.CreatedAt.AsTime())
		require.NotZero(t, updated.UpdatedAt.AsTime())

		for _, tag := range updated.Tags.Items {
			require.NotEmpty(t, tag.Id)
			require.NotEmpty(t, tag.Name)
		}
	})
}
