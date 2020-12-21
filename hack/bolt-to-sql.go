// This package loads all the objects from the old boltdb database and creates
// new ones in a sql store.
package main

import (
	"context"
	"log"

	"github.com/kyleterry/sufr/pkg/api"
	"github.com/kyleterry/sufr/pkg/config"
	"github.com/kyleterry/sufr/pkg/data"
	"github.com/kyleterry/sufr/pkg/data/migrations"
	"github.com/kyleterry/sufr/pkg/service/sqlitestore"
	"github.com/kyleterry/sufr/pkg/store"
)

func main() {
	ctx := context.Background()

	cfg := &config.Config{}
	config.SetDefaults(cfg)

	data.MustInit(cfg)
	migrations.MustMigrate(cfg)

	sqlstore, err := sqlitestore.New(sqlitestore.WithPath("/home/kyle/sufr/sufr.db"))
	if err != nil {
		log.Fatalln(err)
	}

	if err := sqlstore.Migrate(ctx); err != nil {
		log.Fatalln(err)
	}

	tags, err := mapTags(ctx, sqlstore)
	if err != nil {
		log.Fatalln(err)
	}

	user, err := mapUser(ctx, sqlstore, tags)
	if err != nil {
		log.Fatalln(err)
	}

	urls, err := mapURLs(ctx, sqlstore, tags, user)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("imported %d urls for %s\n", len(urls), user.Email)
}

func mapUser(ctx context.Context, db store.Manager, tags map[string]*api.Tag) (*api.User, error) {
	boltUser, err := data.GetUser()
	if err != nil {
		return nil, err
	}

	boltPinnedTags, err := data.GetPinnedTags()
	if err != nil {
		return nil, err
	}

	pc := []*api.Category{}
	for _, bpt := range *boltPinnedTags {
		pc = append(pc, &api.Category{
			Label: bpt.Tag.Name,
			Tags: &api.TagList{
				Items: []*api.Tag{
					tags[bpt.Tag.Name],
				},
			},
		})
	}

	ca := &api.Timestamp{}
	ca.SetFromGoTime(boltUser.CreatedAt)

	ua := &api.Timestamp{}
	ua.SetFromGoTime(boltUser.UpdatedAt)

	sqluser := api.User{
		Email:            boltUser.Email,
		PasswordHash:     []byte(boltUser.Password),
		PinnedCategories: pc,
		CreatedAt:        ca,
		UpdatedAt:        ua,
	}

	if err := db.Users().Create(ctx, &sqluser); err != nil {
		return nil, err
	}

	return &sqluser, nil
}

func mapTags(ctx context.Context, db store.Manager) (map[string]*api.Tag, error) {
	tags := map[string]*api.Tag{}

	boltTags, err := data.GetTags()
	if err != nil {
		return nil, err
	}

	for _, bt := range boltTags {
		if _, ok := tags[bt.Name]; ok {
			log.Printf("found duplicate Tag in boltdb: %s", bt.Name)

			continue
		}
		ca := &api.Timestamp{}
		ca.SetFromGoTime(bt.CreatedAt)

		ua := &api.Timestamp{}
		ua.SetFromGoTime(bt.UpdatedAt)

		sqltag := api.Tag{
			Name:      bt.Name,
			CreatedAt: ca,
			UpdatedAt: ua,
		}

		if err := db.Tags().Create(ctx, &sqltag); err != nil {
			return nil, err
		}

		tags[sqltag.Name] = &sqltag
	}

	return tags, nil
}

func mapURLs(ctx context.Context, db store.Manager, tags map[string]*api.Tag, user *api.User) (map[string]*api.UserURL, error) {
	boltURLs, err := data.GetURLs()
	if err != nil {
		return nil, err
	}

	userURLs := map[string]*api.UserURL{}

	for _, bu := range boltURLs {
		if _, ok := userURLs[bu.URL]; ok {
			log.Printf("found duplicate URL in boltdb: %s", bu.URL)

			continue
		}

		ca := &api.Timestamp{}
		ca.SetFromGoTime(bu.CreatedAt)

		ua := &api.Timestamp{}
		ua.SetFromGoTime(bu.UpdatedAt)

		sqlurl := api.URL{
			Title:     bu.Title,
			Url:       bu.URL,
			CreatedAt: ca,
			UpdatedAt: ua,
		}

		if err := db.URLs().Create(ctx, &sqlurl); err != nil {
			return nil, err
		}

		userurl := api.UserURL{
			Url:       &sqlurl,
			User:      user,
			Favorite:  bu.Favorite,
			CreatedAt: sqlurl.CreatedAt,
			UpdatedAt: sqlurl.UpdatedAt,
		}

		tl := api.TagList{}
		for _, bt := range bu.Tags {
			tl.Items = append(tl.Items, tags[bt.Name])
		}

		userurl.Tags = &tl

		if err := db.UserURLs(user).Create(ctx, &userurl); err != nil {
			return nil, err
		}

		userURLs[sqlurl.Url] = &userurl
	}

	return userURLs, nil
}
