package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockMetadataFetcher struct {
	title      string
	statusCode int
}

func (m mockMetadataFetcher) FetchMetadata(_ string) (PageMeta, error) {
	if m.title == "" {
		m.title = "no title"
	}

	if m.statusCode == 0 {
		m.statusCode = 200
	}

	return PageMeta{m.title, m.statusCode}, nil
}

func TestCreateURL(t *testing.T) {
	opts := CreateURLOptions{
		URL:  "https://google.com",
		Tags: "search engine",
	}

	url, err := CreateURL(opts, mockMetadataFetcher{title: "Google"})

	assert.NoError(t, err)
	assert.Equal(t, url.URL, "https://google.com")
	assert.Equal(t, url.Title, "Google")
	assert.Equal(t, url.StatusCode, 200)
}

func TestGetURL(t *testing.T) {
	opts := CreateURLOptions{
		URL:  "https://yahoo.com",
		Tags: "search engine",
	}

	url, err := CreateURL(opts, mockMetadataFetcher{title: "Yahoo!"})

	assert.NoError(t, err)

	newURL, err := GetURL(url.ID)

	assert.NoError(t, err)
	assert.Equal(t, newURL.URL, "https://yahoo.com")
	assert.Equal(t, newURL.Title, "Yahoo!")
}

func TestParseTags(t *testing.T) {
	var cases = []struct {
		tagStr string
		tags   []string
	}{
		{"computer-science testing foobar", []string{"computer-science", "testing", "foobar"}},
	}

	for _, c := range cases {
		assert.Equal(t, parseTags(c.tagStr), c.tags)
	}
}
