package sqlitestore

import (
	"context"
	"net/url"
	"testing"

	"github.com/kyleterry/sufr/pkg/api"
	"github.com/matryer/is"
	"github.com/rs/xid"
)

func MustCreateRandomURL(t *testing.T, store *Store) *api.URL {
	is := is.New(t)

	ctx := context.Background()

	um := store.URLs()

	u := url.URL{
		Scheme: "https",
		Host:   "unit-testing.sufr.io",
		Path:   xid.New().String(),
	}

	is.NoErr(um.Create(ctx, &api.URL{
		Url:   u.String(),
		Title: u.Path,
	}))

	newURL, err := um.GetByURL(ctx, u.String())
	is.NoErr(err)

	return newURL
}

func TestURLCreate(t *testing.T) {
	is := is.New(t)

	WithTempDatabase(t, func(store *Store) {
		ctx := context.Background()
		um := store.URLs()

		gourl := url.URL{
			Scheme: "https",
			Host:   "unit-testing.sufr.io",
			Path:   xid.New().String(),
		}

		u := api.URL{
			Url:   gourl.String(),
			Title: gourl.Path,
		}
		is.NoErr(um.Create(ctx, &u))

		newURL, err := um.GetByURL(ctx, gourl.String())
		is.NoErr(err)

		is.True(newURL.Id != "")
		is.Equal(gourl.String(), newURL.Url)
		is.Equal(gourl.Path, newURL.Title)
		is.True(!newURL.CreatedAt.AsTime().IsZero())
		is.Equal(newURL.UpdatedAt, nil)
	})
}
