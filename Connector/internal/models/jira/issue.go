package jira

import "time"

type IssueSearchResponse struct {
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}

type Issue struct {
	ID        string     `json:"id"`
	Key       string     `json:"key"`
	Self      string     `json:"self"`
	Fields    Fields     `json:"fields"`
	Changelog *Changelog `json:"changelog,omitempty"`
}

type Fields struct {
	Summary string    `json:"summary"`
	Status  Status    `json:"status"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type Status struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Changelog struct {
	Total     int       `json:"total"`
	Histories []History `json:"histories"`
}

type History struct {
	ID      string    `json:"id"`
	Author  Author    `json:"author"`
	Created time.Time `json:"created"`
	Items   []Item    `json:"items"`
}

type Author struct {
	Self        string `json:"self"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type Item struct {
	Field string `json:"field"`
	From  string `json:"fromString"`
	To    string `json:"toString"`
}
