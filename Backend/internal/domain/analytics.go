package domain

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
	Message string `json:"message,omitempty"`
}

type ProjectStat struct {
	TotalIssues           int     `json:"totalIssues"`
	OpenIssues            int     `json:"openIssues"`
	AvgTimeHours          float64 `json:"avgTimeHours"`
	CreatedPerDayLastWeek []int   `json:"createdPerDayLastWeek"`
}

type GraphJob struct {
	JobID  string `json:"jobId"`
	Status string `json:"status"`
}

type GraphResponse struct {
	Task    string        `json:"task"`
	Project string        `json:"project"`
	Data    []interface{} `json:"data"`
}

type AnalyticsRepository interface {
	GetAllProjects(page int, limit int, search string) (ProjectsResponse, error)
	AddProjectFromJira(key string) (OperationResult, error)
	DeleteProjectByID(id int) (OperationResult, error)

	GetProjectStatByID(id string) (ProjectStat, error)
	MakeGraph(task string, project string) (GraphJob, error)
	GetGraph(task string, project string) (GraphResponse, error)
	CompareGraphs(task string, projects []string) (GraphResponse, error)
	DeleteGraphs(project string) (OperationResult, error)
	IsAnalyzed(project string) (bool, error)
	IsEmpty(project string) (bool, error)
}
