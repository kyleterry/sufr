package migrations

import (
	"encoding/binary"
	"encoding/json"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/kyleterry/sufr/config"
	"github.com/kyleterry/sufr/data"
	"github.com/pkg/errors"
)

type itemSet map[uint64]struct{}

func (i *itemSet) UnmarshalJSON(b []byte) error {
	*i = make(itemSet)
	j := make(map[string]interface{})
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	for k := range j {
		ui, err := strconv.ParseUint(k, 10, 64)
		if err != nil {
			return err
		}
		(*i)[ui] = struct{}{}
	}
	return nil
}

type oldSettingsModel struct {
	ID          uint64    `json:"id"`
	Visibility  string    `json:"visibility"`
	EmbedPhotos bool      `json:"embed_photos"`
	EmbedVideos bool      `json:"embed_videos"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type oldURLModel struct {
	ID        uint64    `json:"id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Tags      itemSet   `json:"tags"`
	Private   bool      `json:"private"`
	Favorite  bool      `json:"favorite"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type oldTagModel struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	URLs      itemSet   `json:"urls"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type oldUserModel struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

var tagCache = make(map[uint64]*data.Tag)

type MigrateOldDatabase struct{}

func (MigrateOldDatabase) Description() string {
	return "migrating legacy database"
}

func (MigrateOldDatabase) Migrate(tx *bolt.Tx) error {
	var (
		root  = []byte("sufr")
		urlB  = []byte("url")
		tagB  = []byte("tag")
		userB = []byte("user")
	)

	// 1: grab settings from the root bucket and save new settings record
	err := data.RunWithBucketForType(tx, data.Settings{}, func(bucket *bolt.Bucket) error {
		rootBucket := tx.Bucket(root)
		rawSettings := rootBucket.Get(itob(1))
		oldSettings := oldSettingsModel{}

		if err := json.Unmarshal(rawSettings, &oldSettings); err != nil {
			return errors.Wrap(err, "failed to decode old settings")
		}

		newSettings := data.Settings{
			Private:     oldSettings.Visibility == "private",
			EmbedPhotos: oldSettings.EmbedPhotos,
			EmbedVideos: oldSettings.EmbedVideos,
			PerPage:     config.DefaultPerPage,
			CreatedAt:   oldSettings.CreatedAt,
			UpdatedAt:   time.Now(),
		}

		b, err := json.Marshal(newSettings)
		if err != nil {
			return err
		}

		if err := bucket.Put([]byte("settings"), b); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to transfer settings object")
	}

	// 2: grab user from the user bucket and save new user record
	err = data.RunWithBucketForType(tx, data.User{}, func(bucket *bolt.Bucket) error {
		rootBucket := tx.Bucket(root)
		userBucket := rootBucket.Bucket(userB)
		rawUser := userBucket.Get(itob(1))
		oldUser := oldUserModel{}

		if err := json.Unmarshal(rawUser, &oldUser); err != nil {
			return err
		}

		newUser := data.User{
			ID:        uuid.New(),
			Email:     oldUser.Email,
			Password:  oldUser.Password,
			CreatedAt: oldUser.CreatedAt,
			UpdatedAt: oldUser.UpdatedAt,
		}

		b, err := json.Marshal(newUser)
		if err != nil {
			return err
		}

		if err := bucket.Put([]byte("user"), b); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to transfer user object")
	}

	// 3: grab all tags and cache them in a map[<name>]tmptagstruct
	err = data.RunWithBucketForType(tx, data.Tag{}, func(bucket *bolt.Bucket) error {
		rootBucket := tx.Bucket(root)
		tagBucket := rootBucket.Bucket(tagB)

		err := tagBucket.ForEach(func(_, v []byte) error {
			var oldTag oldTagModel

			if err := json.Unmarshal(v, &oldTag); err != nil {
				return err
			}

			newTag := data.Tag{
				ID:        uuid.New(),
				Name:      oldTag.Name,
				CreatedAt: oldTag.CreatedAt,
				UpdatedAt: oldTag.UpdatedAt,
			}

			b, err := json.Marshal(newTag)
			if err != nil {
				return err
			}

			id, _ := newTag.ID.MarshalText()
			if err := bucket.Put(id, b); err != nil {
				return err
			}

			tagCache[oldTag.ID] = &newTag

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to transfer tag objects")
	}

	// 4: grab all urls from the url bucket and save them.
	// 4.5: on each iteration, look up the tags and save those.
	err = data.RunWithBucketForType(tx, data.URL{}, func(bucket *bolt.Bucket) error {
		rootBucket := tx.Bucket(root)
		urlBucket := rootBucket.Bucket(urlB)

		err := urlBucket.ForEach(func(_, v []byte) error {
			var oldURL oldURLModel

			if err := json.Unmarshal(v, &oldURL); err != nil {
				return err
			}

			newURL := data.URL{
				ID:         uuid.New(),
				URL:        oldURL.URL,
				Title:      oldURL.Title,
				StatusCode: 200,
				Private:    oldURL.Private,
				Favorite:   oldURL.Favorite,
				CreatedAt:  oldURL.CreatedAt,
				UpdatedAt:  oldURL.UpdatedAt,
			}

			for k := range oldURL.Tags {
				tag := tagCache[k]
				tag.AddURL(&newURL, tx)
				newURL.TagIDs = append(newURL.TagIDs, tag.ID)
			}

			b, err := json.Marshal(newURL)
			if err != nil {
				return err
			}

			id, _ := newURL.ID.MarshalText()
			if err := bucket.Put(id, b); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to transfer url objects")
	}

	// 5: delete all the old buckets

	err = tx.DeleteBucket(legacyBucket)
	if err != nil {
		return errors.Wrap(err, "failed to delete legacy bucket")
	}

	return nil
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
