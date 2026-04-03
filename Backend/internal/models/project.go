package models

type Project struct {
	ID    int    `json:"id" gorm:"primaryKey;column:id"`
	Key   string `json:"key" gorm:"column:key;unique"`
	Title string `json:"title" gorm:"column:title"`
	URL   string `json:"url" gorm:"column:url"`
}

type ProjectRepository interface {
	CreateProject(project *Project) error
	UpdateProject(project *Project) error
	GetAllProjects(limit, offset int) ([]Project, int, error)
	GetProjectByKey(key string) (*Project, error)
	GetProjectByID(id int) (*Project, error)
	GetBasicStats(projectID int) (map[string]interface{}, error)
	DeleteProject(id int) error
}

func (Project) TableName() string {
	return "projects"
}
