package app

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/kyleterry/sufr/config"
)

type itemSet map[uint64]struct{}

func (i itemSet) MarshalJSON() ([]byte, error) {
	jsonmap := make(map[string]interface{})
	for k := range i {
		jsonmap[ui64toa(k)] = true
	}
	return json.Marshal(jsonmap)
}

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

// URL is the model for a url object
type URL struct {
	ID        uint64    `json:"id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Tags      itemSet   `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	tagvalues []*Tag
}

// DeserializeURL will convert raw bytes of JSON into a URL struct pointer
// Returns a *URL
func DeserializeURL(b []byte) *URL {
	url := &URL{}
	if err := json.Unmarshal(b, url); err != nil {
		return nil
	}
	return url
}

// DeserializeURLs takes a slice of []byte and calls DeserializeURL on each one
// Returns a slice of *URL
func DeserializeURLs(b ...[]byte) []*URL {
	var urls []*URL
	for _, rb := range b {
		urls = append(urls, DeserializeURL(rb))
	}
	return urls
}

// Type returns object type (such as bucket name)
func (u *URL) Type() string {
	return config.BucketNameURL
}

// Serialize returns a json []byte slice of the struct
func (u *URL) Serialize() ([]byte, error) {
	return json.Marshal(u)
}

// GetID returns a uint64 id of the record
func (u *URL) GetID() uint64 {
	return u.ID
}

// SetID sets a uint64 id of the record
func (u *URL) SetID(id uint64) {
	u.ID = id
}

// FormattedCreatedAt is used in the template to display a human readable timestamp
// Returns a string
func (u *URL) FormattedCreatedAt() string {
	return u.CreatedAt.Format(time.RFC1123)
}

// HasTags returns a bool true if the url has tags assigned to it
func (u *URL) HasTags() bool {
	return len(u.Tags) > 0
}

// AddTag adds a tag to the record
// This method does not save.
func (u *URL) AddTag(tag *Tag) {
	if u.Tags == nil {
		u.Tags = itemSet{}
	}
	u.Tags[tag.ID] = struct{}{}
	//tag.AddURL(u)
}

// RemoveTag removes a tag from the record
// This method does not save.
func (u *URL) RemoveTag(tag *Tag) {
	if u.Tags == nil {
		return
	}
	delete(u.Tags, tag.ID)
	//tag.RemoveURL(u)
}

// ClearTags removes all tags from the record
// This method does not save.
func (u *URL) ClearTags() {
	u.Tags = itemSet{}
}

// GetTags will use the Tags field to fetch Tag objects from the DB
// This method also memoizes the tags into the tagvalues field for later use.
// Returns a []*Tag slice
func (u *URL) GetTags() []*Tag {
	if len(u.tagvalues) == 0 {
		for i := range u.Tags {
			b, err := database.Get(i, config.BucketNameTag)
			if err != nil {
				continue
			}
			u.tagvalues = append(u.tagvalues, DeserializeTag(b))
		}
	}
	return u.tagvalues
}

// GetTagsForDisplay will return a string of space separated tags
func (u *URL) GetTagsForDisplay() string {
	var tags []string
	for _, t := range u.GetTags() {
		tags = append(tags, t.Name)
	}
	return strings.Join(tags, " ")
}

// Save saves the URL to the database
func (u *URL) Save() error {
	err := database.Put(u)
	if err != nil {
		return err
	}
	return nil
}

// SaveWithTags calls Save and saves the URL to the database but parses a string of tags as well
func (u *URL) SaveWithTags(tagsstring string) error {
	err := u.Save()
	if err != nil {
		return err
	}
	// Parse Tags if there are any
	if tagsstring != "" {
		splittags := parseTags(tagsstring)
		tags := []*Tag{}
		tagbytes, notfoundtags, err := database.GetValuesByField("name", config.BucketNameTag, splittags...)
		if err != nil {
			return err
		}
		for _, notfound := range notfoundtags {
			tags = append(tags, &Tag{Name: notfound})
		}
		tags = append(tags, DeserializeTags(tagbytes...)...)

		u.ClearTags()

		for _, tag := range tags {
			tag.AddURL(u)
			database.Put(tag)
			u.AddTag(tag)
		}

	}
	return u.Save()
}

func (u *URL) Delete() error {
	for _, tag := range u.GetTags() {
		tag.RemoveURL(u)
		tag.Save()
	}
	return database.Delete(u.ID, config.BucketNameURL)
}

// Tag holds the little information we have about url tags
type Tag struct {
	ID        uint64  `json:"id"`
	Name      string  `json:"name"`
	URLs      itemSet `json:"urls"`
	urlvalues []*URL
}

// Type returns object type (such as bucket name)
func (t *Tag) Type() string {
	return config.BucketNameTag
}

// Serialize returns a json []byte slice of the struct
func (t *Tag) Serialize() ([]byte, error) {
	return json.Marshal(t)
}

// GetID returns a uint64 id of the record
func (t *Tag) GetID() uint64 {
	return t.ID
}

// SetID sets a uint64 id of the record
func (t *Tag) SetID(id uint64) {
	t.ID = id
}

// AddURL adds a url to the record
// This method does not save.
func (t *Tag) AddURL(url *URL) {
	if t.URLs == nil {
		t.URLs = itemSet{}
	}
	t.URLs[url.ID] = struct{}{}
}

//RemoveURL removed a url from the record
// This method does not save.
func (t *Tag) RemoveURL(url *URL) {
	if t.URLs == nil {
		return
	}
	delete(t.URLs, url.ID)
}

// GetURLs will use the URLs field to fetch URL objects from the DB
// This method also memoizes the tags into the urlvalues field for later use.
// Returns a []*URL slice
func (t *Tag) GetURLs() []*URL {
	if len(t.urlvalues) == 0 {
		for i := range t.URLs {
			b, err := database.Get(i, config.BucketNameURL)
			if err != nil {
				continue
			}
			t.urlvalues = append(t.urlvalues, DeserializeURL(b))
		}
	}
	return t.urlvalues
}

// URLCount returns the URL count
func (t *Tag) URLCount() int {
	return len(t.URLs)
}

// Save saves the URL to the database
func (t *Tag) Save() error {
	err := database.Put(t)
	if err != nil {
		return err
	}
	return nil
}

// DeserializeTag will convert raw bytes of JSON into a URL struct pointer
// Returns a *Tag
func DeserializeTag(b []byte) *Tag {
	tag := &Tag{}
	if err := json.Unmarshal(b, tag); err != nil {
		return nil
	}
	return tag
}

// DeserializeTags takes a slice of []byte and calls DeserializeTag on each one
// Returns a slice of *Tag
func DeserializeTags(b ...[]byte) []*Tag {
	var tags []*Tag
	for _, rb := range b {
		tags = append(tags, DeserializeTag(rb))
	}
	return tags
}
