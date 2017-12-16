package data

type query struct {
	tags    []string
	domains []string
}

func NewQuery(query string) *query {
	return &query{}
}
