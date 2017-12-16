package data

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
	"github.com/kyleterry/sufr/config"
	"github.com/pkg/errors"
)

var ErrSettingsExists = errors.New("settings object already exists")

type InitializeInstanceOptions struct {
	Email       string `schema:"email"`
	Password    string `schema:"password"`
	Private     bool   `schema:"private"`
	EmbedPhotos bool   `schema:"embedphotos"`
	EmbedVideos bool   `schema:"embedvideos"`
	PerPage     int    `schema:"perpage"`
}

type SettingsOptions struct {
	Private     bool `schema:"private"`
	EmbedPhotos bool `schema:"embedphotos"`
	EmbedVideos bool `schema:"embedvideos"`
	PerPage     int  `schema:"perpage"`
}

type Settings struct {
	Private     bool      `json:"private"`
	EmbedPhotos bool      `json:"embed_photos"`
	EmbedVideos bool      `json:"embed_videos"`
	PerPage     int       `json:"per_page"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func SaveSettings(opts SettingsOptions) (*Settings, error) {
	tx, err := db.bolt.Begin(true)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	settings, err := saveSettings(opts, tx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create settings")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return settings, nil
}

func saveSettings(opts SettingsOptions, tx *bolt.Tx) (*Settings, error) {
	bucket := tx.Bucket(buckets[appKey])

	now := time.Now()

	perPage := opts.PerPage
	if perPage == 0 {
		perPage = config.DefaultPerPage
	}

	settings := &Settings{
		Private:     opts.Private,
		EmbedVideos: opts.EmbedVideos,
		EmbedPhotos: opts.EmbedPhotos,
		PerPage:     perPage,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	b, err := json.Marshal(settings)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize settings")
	}

	if err := bucket.Put([]byte("settings"), b); err != nil {
		return nil, errors.Wrap(err, "boltdb put failed")
	}

	return settings, nil
}

func GetSettings() (*Settings, error) {
	tx, err := db.bolt.Begin(true)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	settings, err := getSettings(tx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get settings")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return settings, nil
}

func getSettings(tx *bolt.Tx) (*Settings, error) {
	var settings Settings
	bucket := tx.Bucket(buckets[appKey])

	rawSettings := bucket.Get([]byte("settings"))
	if len(rawSettings) == 0 {
		return nil, ErrNotFound
	}

	if err := json.Unmarshal(rawSettings, &settings); err != nil {
		return nil, errors.Wrap(err, "failed to decode object")
	}

	return &settings, nil
}
