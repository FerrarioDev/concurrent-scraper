package domain

import "time"

type Params struct {
	BaseURL  string
	Workers  int
	MaxPages int
	Timeout  time.Duration
}

type Site struct {
	ID    int
	Link  string
	Title string
	Links []string
}

type SiteRequest struct {
	ID       *int   `json:"id" db:"id"` // unique id autoincremented by db
	URL      string `json:"url" db:"url"`
	Title    string `json:"title" db:"title"`
	Links    int    `json:"links" db:"links"`         // count of links
	FatherID *int   `json:"father_id" db:"father_id"` // if it has a father it will store its id
}
