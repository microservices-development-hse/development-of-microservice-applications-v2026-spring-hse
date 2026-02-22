package models

type Project struct {
	ID    int    `json:"id" gorm:"primaryKey;column:id"`
	Key   string `json:"key" gorm:"column:key;unique"`
	Title string `json:"title" gorm:"column:title"`
}

type ProjectRepository interface {
	CreateProject(project *Project) error
	UpdateProject(project *Project) error
	GetAllProjects() ([]Project, error)
	GetProjectByKey(key string) (*Project, error)
	GetProjectByID(id int) (*Project, error)
	DeleteProject(id int) error
}

func (Project) TableName() string {
	return "Project"
}
