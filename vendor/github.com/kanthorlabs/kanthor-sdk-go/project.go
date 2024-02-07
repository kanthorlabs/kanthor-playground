package kanthorsdk

import _ "embed"

//go:embed project.json
var project []byte

type Project struct {
	Version string            `json:"version"`
	Host    string            `json:"host"`
	Hosts   map[string]string `json:"hosts"`
}

var DefaultPagingLimitMin int64 = 5
var DefaultPagingLimitMax int64 = 100
var DefaultPagingPageMin int64 = 1
var DefaultPagingPageMax int64 = 100

type Queries struct {
	Search string   `json:"_q"`
	Limit  int64    `json:"_limit"`
	Page   int64    `json:"_page"`
	Id     []string `json:"id"`
}

type Query func(*Queries)

func WithSearch(keyword string) Query {
	return func(q *Queries) {
		q.Search = keyword
	}
}

func WithLimit(limit int64) Query {
	return func(q *Queries) {
		q.Limit = min(max(limit, DefaultPagingLimitMin), DefaultPagingLimitMax)
	}
}

func WithPage(limit int64) Query {
	return func(q *Queries) {
		q.Page = min(max(limit, DefaultPagingPageMin), DefaultPagingPageMax)
	}
}

func WithIds(ids []string) Query {
	return func(q *Queries) {
		if len(ids) > 0 {
			q.Id = ids
		}
	}
}
