package app

import (
	"encoding/json"
	"time"

	"github.com/kyleterry/sufr/config"
)

// URL is the model for a url object
type URL struct {
	ID        uint64    `json:"id"`
	URL       string    `json:"url"`
	URLTitle  string    `json:"url_title"`
	Tags      []Tag     `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *URL) Type() string {
	return config.BucketNameURL
}

func (u *URL) Serialize() ([]byte, error) {
	return json.Marshal(u)
}

func (u *URL) GetID() uint64 {
	return u.ID
}

type Tag struct {
	Name string
	URLs []URL
}

func (t *Tag) Type() string {
	return config.BucketNameTag
}

func (t *Tag) Serialize() ([]byte, error) {
	return json.Marshal(t)
}
