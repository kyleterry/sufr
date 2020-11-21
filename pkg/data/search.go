package data

type query struct {
	tags    []string
	domains []string
}

func NewQuery(q string) *query {
	return &query{}
}
