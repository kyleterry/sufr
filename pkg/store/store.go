package store

import (
	"context"

	"github.com/kyleterry/sufr/pkg/api"
)

type Manager interface {
	URLs() URLManager
	Tags() TagManager
	UserURLs(*api.User) UserURLManager
	Users() UserManager
}

type URLManager interface {
	Create(ctx context.Context, url *api.URL) error
	GetByURL(ctx context.Context, url string) (*api.URL, error)
}

type TagManager interface {
	Create(ctx context.Context, tag *api.Tag) error
	GetByID(ctx context.Context, id string) (*api.Tag, error)
	GetByName(ctx context.Context, name string) (*api.Tag, error)
}

type UserURLManager interface {
	Create(ctx context.Context, userURL *api.UserURL) error
	Update(ctx context.Context, userURL *api.UserURL) error
	GetAll(ctx context.Context, filters ...FilterOption) ([]*api.UserURL, error)
	GetByURLID(ctx context.Context, urlID string) (*api.UserURL, error)
}

type UserManager interface {
	Create(ctx context.Context, user *api.User) error
	UpdatePinnedCategories(ctx context.Context, user *api.User) error
	GetByID(ctx context.Context, id string) (*api.User, error)
	GetByEmail(ctx context.Context, email string) (*api.User, error)
	GetByEmailAndPassword(ctx context.Context, email string, password string) (*api.User, error)
}

type FilterOptions struct {
	Search string
	Tags   []string
	After  int64
}

type FilterOption interface {
	Apply(*FilterOptions)
}

type FilterOptionFunc struct {
	f func(*FilterOptions)
}

func (s *FilterOptionFunc) Apply(opts *FilterOptions) {
	s.f(opts)
}

func WithResultsAfter(i int64) FilterOption {
	return &FilterOptionFunc{
		f: func(opts *FilterOptions) {
			opts.After = i
		},
	}
}

func WithSearchTerm(term string) FilterOption {
	return &FilterOptionFunc{
		f: func(opts *FilterOptions) {
			opts.Search = term
		},
	}
}

func WithTags(tags []string) FilterOption {
	return &FilterOptionFunc{
		f: func(opts *FilterOptions) {
			opts.Tags = tags
		},
	}
}
