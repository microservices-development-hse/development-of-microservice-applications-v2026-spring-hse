package jira

import (
	"strings"
	"time"
)

type JTime struct {
	time.Time
}

func (jt *JTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		return nil
	}

	formats := []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000Z",
		time.RFC3339,
	}

	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			jt.Time = t
			return nil
		}
	}

	return nil
}

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
	Summary string `json:"summary"`
	Status  Status `json:"status"`
	Created JTime  `json:"created"`
	Updated JTime  `json:"updated"`
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
	ID      string `json:"id"`
	Author  Author `json:"author"`
	Created JTime  `json:"created"`
	Items   []Item `json:"items"`
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
