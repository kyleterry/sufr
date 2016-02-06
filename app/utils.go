package app

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

func getPageTitle(url string) (string, error) {
	doc, err := goquery.NewDocument(url)

	title := doc.Find("title").Text()
	return title, err
}

func parseTags(tagsstring string) []string {
	return strings.FieldsFunc(tagsstring, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
	})
}

func parseTagsMap(tagsstring string) map[string]struct{} {
	var m = make(map[string]struct{})
	for _, t := range parseTags(tagsstring) {
		m[t] = struct{}{}
	}
	return m
}

func ui64toa(v uint64) string {
	return strconv.FormatUint(v, 10)
}
