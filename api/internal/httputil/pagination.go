package httputil

import (
	"net/http"
	"strconv"
	"strings"
)

type PageSearch struct {
	PageIndex int
	PageSize  int
	// Search is a free-text filter applied server-side when supported by the list endpoint.
	Search string
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
	q := r.URL.Query().Get("q")
	if q == "" {
		q = r.URL.Query().Get("search")
	}
	return PageSearch{PageIndex: idx, PageSize: size, Search: strings.TrimSpace(q)}
}

func (p PageSearch) Offset() int {
	return p.PageIndex * p.PageSize
}

func (p PageSearch) Limit() int {
	return p.PageSize
}
