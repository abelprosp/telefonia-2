package httputil

import (
	"net/http"
	"strconv"
)

type PageSearch struct {
	PageIndex int
	PageSize  int
}

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

func ParsePagination(r *http.Request) PageSearch {
	idx, _ := strconv.Atoi(r.URL.Query().Get("page_index"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if idx < 0 {
		idx = 0
	}
	if size <= 0 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	return PageSearch{PageIndex: idx, PageSize: size}
}

func (p PageSearch) Offset() int {
	return p.PageIndex * p.PageSize
}

func (p PageSearch) Limit() int {
	return p.PageSize
}
