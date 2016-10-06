package app

type Paginator struct {
	NumRecords int
	Page       int
	TotalPages int
}

func NewPaginator(num int, page int) *Paginator {
	p := &Paginator{}
	p.NumRecords = num
	p.Page = page
	return p
}
