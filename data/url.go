package data

import (
	"encoding/json"
	"net/http"
	stdurl "net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/kyleterry/sufr/config"
	"github.com/pkg/errors"
)

// URLBucket is the name of the bucket to store URL objects in
const (
	URLBucket       = "_urls"
	DefaultURLTitle = "No title (edit to change)"
)

type URLMetadataFetcher interface {
	FetchMetadata(string) (PageMeta, error)
}

// CreateURLOptions is passed into CreateURL from the http handler to initiate a url creation
type CreateURLOptions struct {
	URL     string
	Tags    string
	Private bool
}

// UpdateURLOptions is passed into UpdateURL from the http handler to update a url
type UpdateURLOptions struct {
	ID       uuid.UUID
	Title    string
	Private  bool
	Favorite bool
	Tags     string
}

type PageMeta struct {
	Title  string
	Status int
}

type URLsByDateDesc []*URL

func (u URLsByDateDesc) Len() int           { return len(u) }
func (u URLsByDateDesc) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u URLsByDateDesc) Less(i, j int) bool { return u[i].CreatedAt.After(u[j].CreatedAt) }

// URL is the model for a url object
type URL struct {
	ID         uuid.UUID   `json:"id"`
	URL        string      `json:"url"`
	Title      string      `json:"title"`
	StatusCode int         `json:"status_code"`
	Private    bool        `json:"private"`
	Favorite   bool        `json:"favorite"`
	Tags       []*Tag      `json:"-"`
	TagIDs     []uuid.UUID `json:"tag_ids"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// helpers

// IsPublic will return true if the url is visible to everyone (even those not logged in)
// This option doesn't matter is the SUFR global setting for visibility is private.
func (u *URL) IsPublic() bool {
	return !u.Private
}

// HasTags returns a bool true if the url has tags assigned to it
func (u *URL) HasTags() bool {
	return len(u.Tags) > 0
}

// FormattedCreatedAt is used in the template to display a human readable timestamp
// Returns a string
func (u *URL) FormattedCreatedAt() string {
	return u.CreatedAt.Format(time.RFC1123)
}

func (u *URL) GetTagsForDisplay() string {
	s := []string{}
	for _, tag := range u.Tags {
		s = append(s, tag.Name)
	}

	return strings.Join(s, " ")
}

func (u *URL) ToggleFavorite() error {
	u.Favorite = !u.Favorite

	err := db.bolt.Update(func(tx *bolt.Tx) error {
		id, _ := u.ID.MarshalText()

		bucket := tx.Bucket(urlBucket)

		b, err := json.Marshal(u)
		if err != nil {
			return errors.Wrap(err, "failed to serialize object")
		}

		err = bucket.Put(id, b)
		if err != nil {
			return errors.Wrap(err, "failed to put object to database")
		}

		return nil
	})

	if err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

func (u *URL) fillTags(tx *bolt.Tx) error {
	for _, tagID := range u.TagIDs {
		tag, err := getTag(tagID, tx)
		if err != nil {
			return errors.Wrap(err, "failed to get tag")
		}

		u.Tags = append(u.Tags, tag)
	}

	return nil
}

func CreateURL(opts CreateURLOptions, fetcher URLMetadataFetcher) (*URL, error) {
	tx, err := db.bolt.Begin(true)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	url, err := createURL(opts, fetcher, tx)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create url")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return url, nil
}

func createURL(opts CreateURLOptions, fetcher URLMetadataFetcher, tx *bolt.Tx) (*URL, error) {
	if _, err := stdurl.Parse(opts.URL); err != nil {
		return nil, errors.Wrap(err, "failed to parse url")
	}

	if u, _ := getURLByURL(opts.URL, tx); u != nil {
		return nil, ErrDuplicateKey
	}

	pm, err := fetcher.FetchMetadata(opts.URL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch page")
	}

	if pm.Title == "" {
		pm.Title = DefaultURLTitle
	}

	now := time.Now()

	url := &URL{
		ID:         uuid.New(),
		URL:        opts.URL,
		Title:      pm.Title,
		StatusCode: pm.Status,
		CreatedAt:  now,
		UpdatedAt:  now,
		TagIDs:     []uuid.UUID{},
	}

	tagNames := parseTags(opts.Tags)

	for _, tagName := range tagNames {
		// Not really sure what can go wrong here, so we just skip if there's an error
		tag, _, err := getOrCreateTag(tagName, tx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to search for tag")
		}

		url.TagIDs = append(url.TagIDs, tag.ID)

		if err := tag.AddURL(url, tx); err != nil {
			return nil, errors.Wrap(err, "failed to add url to tag")
		}
	}

	bucket := tx.Bucket(urlBucket)

	b, err := json.Marshal(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize url")
	}

	id, _ := url.ID.MarshalText()
	if err := bucket.Put(id, b); err != nil {
		return nil, errors.Wrap(err, "boltdb put failed")
	}

	return url, nil
}

func UpdateURL(opts UpdateURLOptions) (*URL, error) {
	tx, err := db.bolt.Begin(true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transaction")
	}

	defer tx.Rollback()

	url, err := updateURL(opts, tx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update url")
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return url, nil
}

func updateURL(opts UpdateURLOptions, tx *bolt.Tx) (*URL, error) {
	url, err := getURL(opts.ID, tx)
	if err != nil {
		return nil, err
	}

	// TODO: validate this in the handler instead
	if opts.Title != "" {
		url.Title = opts.Title
	}

	url.Private = opts.Private
	url.Favorite = opts.Favorite

	tagNames := parseTags(opts.Tags)

	type tagState struct {
		found bool
		tag   Tag
	}
	oldTags := make(map[string]tagState)
	for _, tag := range url.Tags {
		oldTags[tag.Name] = tagState{tag: *tag}
	}

	newTags := []*Tag{}
	for _, name := range tagNames {
		tag, _, err := getOrCreateTag(name, tx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to search for tag")
		}
		newTags = append(newTags, tag)

		if oldTag, ok := oldTags[tag.Name]; ok {
			oldTag.found = true
		}
	}

	for _, old := range oldTags {
		if !old.found {
			old.tag.removeURL(url, tx)
		}
	}

	url.TagIDs = []uuid.UUID{}
	for _, new := range newTags {
		url.Tags = append(url.Tags, new)
		url.TagIDs = append(url.TagIDs, new.ID)
		if err := new.AddURL(url, tx); err != nil {
			return nil, errors.Wrap(err, "failed to add url to tag")
		}
	}

	url.UpdatedAt = time.Now()

	bucket := tx.Bucket(urlBucket)

	b, err := json.Marshal(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize url")
	}

	id, _ := url.ID.MarshalText()
	if err := bucket.Put(id, b); err != nil {
		return nil, errors.Wrap(err, "boltdb put failed")
	}

	return url, nil
}

func DeleteURL(url *URL) error {
	tx, err := db.bolt.Begin(true)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction")
	}

	defer tx.Rollback()

	if err := deleteURL(url, tx); err != nil {
		return errors.Wrap(err, "failed to delete url")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}

func deleteURL(url *URL, tx *bolt.Tx) error {
	bucket := tx.Bucket(urlBucket)

	for _, tag := range url.Tags {
		if err := tag.removeURL(url, tx); err != nil {
			return errors.Wrap(err, "failed to remove url from tag")
		}
	}

	id, _ := url.ID.MarshalText()
	if err := bucket.Delete(id); err != nil {
		return err
	}

	return nil
}

// GetURLs will return a slice of *URL sorted by URL.CreatedAt desc
func GetURLs() ([]*URL, error) {
	var urls []*URL
	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		urls, err = getURLs(tx)
		if err != nil {
			return err
		}

		// for _, url := range urls {
		// 	if err := url.fillTags(tx); err != nil {
		// 		return errors.Wrap(err, "failed to fill tags")
		// 	}
		// }

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return urls, nil
}

func getURLs(tx *bolt.Tx) ([]*URL, error) {
	var urls []*URL
	bucket := tx.Bucket(urlBucket)

	bucket.ForEach(func(_, v []byte) error {
		url := URL{}
		if err := json.Unmarshal(v, &url); err != nil {
			return err
		}

		if err := url.fillTags(tx); err != nil {
			return errors.Wrap(err, "failed to fill tags")
		}

		urls = append(urls, &url)

		return nil
	})

	sort.Sort(URLsByDateDesc(urls))

	return urls, nil
}

func getFavoriteURLs(tx *bolt.Tx) ([]*URL, error) {
	var urls []*URL
	bucket := tx.Bucket(urlBucket)

	bucket.ForEach(func(_, v []byte) error {
		url := URL{}
		if err := json.Unmarshal(v, &url); err != nil {
			return err
		}

		if err := url.fillTags(tx); err != nil {
			return errors.Wrap(err, "failed to fill tags")
		}

		if url.Favorite {
			urls = append(urls, &url)
		}

		return nil
	})

	sort.Sort(URLsByDateDesc(urls))

	return urls, nil
}

func GetURL(id uuid.UUID) (*URL, error) {
	var url *URL

	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error

		url, err = getURL(id, tx)
		if err != nil {
			return err
		}

		return url.fillTags(tx)
	})

	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return url, nil
}

func getURL(urlID uuid.UUID, tx *bolt.Tx) (*URL, error) {
	var url URL
	bucket := tx.Bucket(urlBucket)

	id, _ := urlID.MarshalText()
	rawURL := bucket.Get(id)
	if len(rawURL) == 0 {
		return nil, ErrNotFound
	}

	if err := json.Unmarshal(rawURL, &url); err != nil {
		return nil, errors.Wrap(err, "failed to decode object")
	}

	return &url, nil
}

func GetURLByURL(urlstr string) (*URL, error) {
	var url *URL

	err := db.bolt.View(func(tx *bolt.Tx) error {
		var err error
		url, err = getURLByURL(urlstr, tx)
		if err != nil {
			return err
		}

		for _, tagID := range url.TagIDs {
			tag, err := getTag(tagID, tx)
			if err != nil {
				return errors.Wrap(err, "failed to get tag")
			}

			url.Tags = append(url.Tags, tag)
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return url, nil

}

func getURLByURL(urlstr string, tx *bolt.Tx) (*URL, error) {
	var url *URL
	bucket := tx.Bucket(urlBucket)

	err := bucket.ForEach(func(_, v []byte) error {
		var u URL
		if err := json.Unmarshal(v, &u); err != nil {
			return errors.Wrap(err, "failed to decode object")
		}

		if u.URL == urlstr {
			url = &u
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return url, nil
}

func getURLCount(tx *bolt.Tx) (int, error) {
	var i int

	bucket := tx.Bucket(urlBucket)

	err := bucket.ForEach(func(k, _ []byte) error {
		i++
		return nil
	})
	if err != nil {
		return 0, errors.Wrap(err, "transaction failed")
	}

	return i, nil
}

// takes a string like "computer-science interesting books" and turns it into []string{""computer-science", "interesting", "books"}
func parseTags(tagsstring string) []string {
	return strings.FieldsFunc(tagsstring, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
	})
}

type HTTPMetadataFetcher struct{}

// Returns the page title or an error. If there is an error, the url is returned as well.
func (HTTPMetadataFetcher) FetchMetadata(url string) (PageMeta, error) {
	var pm PageMeta

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return pm, err
	}

	req.Header.Set("User-Agent", config.SUFRUserAgent)

	res, err := client.Do(req)
	if err != nil {
		return pm, err
	}

	defer res.Body.Close()

	pm.Status = res.StatusCode

	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		return pm, err
	}

	pm.Title = doc.Find("title").Text()
	return pm, nil
}
