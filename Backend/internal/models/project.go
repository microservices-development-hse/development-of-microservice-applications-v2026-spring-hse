package models

type Project struct {
	ID    int    `json:"id" db:"id"`
	Title string `json:"title" db:"title"`
}

type ProjectRepository interface {
	Save(project *Project) error
	GetAll() ([]Project, error)
	GetByKey(key string) (*Project, error)
	Delete(id int) error
}
