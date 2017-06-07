package data

import "testing"

func TestParseTags(t *testing.T) {
	var cases = struct {
		input  string
		output []string
	}{
		{"tag tags taggy", []string{"tag", "tags", "taggy"}},
		{"tag-tag tags1 taggy>", []string{"tag-tag", "tags1", "taggy>"}},
	}

	for _, c := range cases {
		assert.Equal(t, c.output, parseTags(c.input))
	}
}
