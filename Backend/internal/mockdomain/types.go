package mockdomain

type Project struct {
	ID   int    `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type PageInfo struct {
	CurrentPage int `json:"currentPage"`
	Total       int `json:"total"`
	PerPage     int `json:"perPage"`
}

type ProjectsResponse struct {
	Data     []Project `json:"data"`
	PageInfo PageInfo  `json:"pageInfo"`
}

type OperationResult struct {
	Status  bool   `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

type ProjectStat struct {
	TotalIssues           int     `json:"totalIssues"`
	OpenIssues            int     `json:"openIssues"`
	AvgTimeHours          float64 `json:"avgTimeHours"`
	CreatedPerDayLastWeek []int   `json:"createdPerDayLastWeek"`
}

type GraphResponse struct {
	Task    string        `json:"task"`
	Project string        `json:"project"`
	Data    []interface{} `json:"data"`
}

type GraphJob struct {
	JobID  string `json:"jobId"`
	Status string `json:"status"`
}
