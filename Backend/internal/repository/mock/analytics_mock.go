package mock

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/microservices-development-hse/backend/internal/domain"
)

type AnalyticsMock struct {
	mu       sync.Mutex
	projects []domain.Project
	graphs   map[string]domain.GraphResponse
	nextID   int
	delay    time.Duration
	err      error
}

func NewAnalyticsMock() *AnalyticsMock {
	return &AnalyticsMock{
		projects: []domain.Project{
			{ID: 1, Key: "PROJ1", Name: "Project One"},
			{ID: 2, Key: "PROJ2", Name: "Project Two"},
		},
		graphs: make(map[string]domain.GraphResponse),
		nextID: 3,
	}
}

func (m *AnalyticsMock) SetDelay(d time.Duration) { m.delay = d }
func (m *AnalyticsMock) SetError(err error)       { m.err = err }

func (m *AnalyticsMock) applyDelay() {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
}

func (m *AnalyticsMock) GetAllProjects(page int, limit int, search string) (domain.ProjectsResponse, error) {
	if m.err != nil {
		return domain.ProjectsResponse{}, m.err
	}
	m.applyDelay()
	m.mu.Lock()
	defer m.mu.Unlock()

	if limit <= 0 {
		limit = 20
	}

	search = strings.ToLower(search)
	filtered := []domain.Project{}
	for _, p := range m.projects {
		if search == "" || strings.Contains(strings.ToLower(p.Name), search) || strings.Contains(strings.ToLower(p.Key), search) {
			filtered = append(filtered, p)
		}
	}

	total := len(filtered)
	start := (page - 1) * limit
	if start < 0 { start = 0 }
	if start > total { start = total }
	end := start + limit
	if end > total { end = total }

	return domain.ProjectsResponse{
		Data: filtered[start:end],
		PageInfo: domain.PageInfo{
			CurrentPage: page,
			Total:       total,
			PerPage:     limit,
		},
	}, nil
}

func (m *AnalyticsMock) AddProjectFromJira(key string) (domain.OperationResult, error) {
	if m.err != nil {
		return domain.OperationResult{}, m.err
	}
	m.applyDelay()
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.projects {
		if strings.EqualFold(p.Key, key) {
			return domain.OperationResult{Message: "exists"}, nil
		}
	}

	p := domain.Project{ID: m.nextID, Key: strings.ToUpper(key), Name: "Mock " + key}
	m.nextID++
	m.projects = append(m.projects, p)
	k := "1|" + p.Key
	m.graphs[k] = domain.GraphResponse{Task: "1", Project: p.Key, Data: []interface{}{}}
	return domain.OperationResult{Message: "queued"}, nil
}

func (m *AnalyticsMock) DeleteProjectByID(id int) (domain.OperationResult, error) {
	if m.err != nil {
		return domain.OperationResult{}, m.err
	}
	m.applyDelay()
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, p := range m.projects {
		if p.ID == id {
			m.projects = append(m.projects[:i], m.projects[i+1:]...)
			return domain.OperationResult{Message: "deleted"}, nil
		}
	}
	return domain.OperationResult{}, errors.New("not found")
}

func (m *AnalyticsMock) GetProjectStatByID(id string) (domain.ProjectStat, error) {
	if m.err != nil {
		return domain.ProjectStat{}, m.err
	}
	m.applyDelay()
	return domain.ProjectStat{
		TotalIssues:           100,
		OpenIssues:            12,
		AvgTimeHours:          48.5,
		CreatedPerDayLastWeek: []int{1, 2, 3, 0, 1, 0, 2},
	}, nil
}

func (m *AnalyticsMock) MakeGraph(task string, project string) (domain.GraphJob, error) {
	if m.err != nil {
		return domain.GraphJob{}, m.err
	}
	m.applyDelay()
	m.mu.Lock()
	defer m.mu.Unlock()
	job := domain.GraphJob{JobID: "job-" + task + "-" + project, Status: "queued"}
	key := task + "|" + project
	m.graphs[key] = domain.GraphResponse{Task: task, Project: project, Data: []interface{}{"pt1", "pt2"}}
	return job, nil
}

func (m *AnalyticsMock) GetGraph(task string, project string) (domain.GraphResponse, error) {
	if m.err != nil {
		return domain.GraphResponse{}, m.err
	}
	m.applyDelay()
	m.mu.Lock()
	defer m.mu.Unlock()
	key := task + "|" + project
	if g, ok := m.graphs[key]; ok {
		return g, nil
	}
	return domain.GraphResponse{Task: task, Project: project, Data: []interface{}{}}, nil
}

func (m *AnalyticsMock) CompareGraphs(task string, projects []string) (domain.GraphResponse, error) {
	if m.err != nil {
		return domain.GraphResponse{}, m.err
	}
	m.applyDelay()
	data := []interface{}{map[string]interface{}{"projects": projects}}
	return domain.GraphResponse{Task: task, Project: "compare", Data: data}, nil
}

func (m *AnalyticsMock) DeleteGraphs(project string) (domain.OperationResult, error) {
	if m.err != nil {
		return domain.OperationResult{}, m.err
	}
	m.applyDelay()
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.graphs {
		if strings.HasSuffix(k, "|"+project) {
			delete(m.graphs, k)
		}
	}
	return domain.OperationResult{Message: "deleted"}, nil
}

func (m *AnalyticsMock) IsAnalyzed(project string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	m.applyDelay()
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.graphs {
		if strings.HasSuffix(k, "|"+project) {
			return true, nil
		}
	}
	return false, nil
}

func (m *AnalyticsMock) IsEmpty(project string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	m.applyDelay()
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.projects {
		if strings.EqualFold(p.Key, project) {
			return false, nil
		}
	}
	return true, nil
}
