package integration

type PageInfo struct {
	CurrentPage   int `json:"currentPage"`
	PageCount     int `json:"pageCount"`
	ProjectsCount int `json:"projectsCount"`
}

type ProjectItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}
