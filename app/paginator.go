package app

type Paginator struct {
	NumRecords int
	Page       int
	PerPage    int
}

func NewPaginator(num int, page int, perPage int) *Paginator {
	p := &Paginator{}
	p.NumRecords = num
	p.Page = page
	p.PerPage = perPage
	return p
}

func (p Paginator) GetObjects(bucket string) ([][]byte, error) {
	var offset int
	if p.Page == 1 {
		offset = 0
	} else {
		offset = p.PerPage * p.Page
	}
	return database.GetSubset(uint64(offset), uint64(p.PerPage), bucket)
}

func (p Paginator) TotalPages() int {
	return p.NumRecords / p.PerPage
}

func (p Paginator) CurrentPage() int {
	return p.Page
}

func (p Paginator) HasPrevious() bool {
	return p.Page > 1
}

func (p Paginator) PreviousPage() int {
	if p.Page == 1 {
		return p.Page
	}
	return p.Page - 1
}

func (p Paginator) HasNext() bool {
	return p.Page < p.TotalPages()
}

func (p Paginator) NextPage() int {
	if p.Page == p.TotalPages() {
		return p.Page
	}
	return p.Page + 1
}

func (p Paginator) Pages() []int {
	pgs := []int{}
	i := 1
	for i <= p.TotalPages() {
		pgs = append(pgs, i)
		i++
	}

	return pgs
}
