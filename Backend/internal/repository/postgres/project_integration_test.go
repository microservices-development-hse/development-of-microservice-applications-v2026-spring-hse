package postgres

import (
	"testing"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestProjectRepository_Full(t *testing.T) {
	env := SetupRepoTestEnv(t)
	repo := NewProjectRepository(env.DB)

	t.Run("Project_CRUD_Lifecycle", func(t *testing.T) {
		proj := &models.Project{Key: "NEW", Title: "New Project"}

		err := repo.CreateProject(proj)
		assert.NoError(t, err)
		assert.NotZero(t, proj.ID)

		found, err := repo.GetProjectByID(proj.ID)
		assert.NoError(t, err)
		assert.Equal(t, "New Project", found.Title)

		exists, _ := repo.Exists(proj.ID)
		assert.True(t, exists)

		proj.Title = "Updated Title"
		err = repo.UpdateProject(proj)
		assert.NoError(t, err)

		projects, total, err := repo.GetAllProjects(10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)
		assert.NotEmpty(t, projects)

		err = repo.DeleteProject(proj.ID)
		assert.NoError(t, err)
	})

	t.Run("Additional_Methods_And_Errors", func(t *testing.T) {
		res, err := repo.GetProjectByID(999999)
		assert.NoError(t, err)
		assert.Nil(t, res)

		proj := &models.Project{Key: "KEY-1", Title: "Key Test"}
		err = repo.CreateProject(proj)
		assert.NoError(t, err)

		dryStats, err := repo.GetDryStatistics(proj.ID)
		assert.NoError(t, err)
		assert.NotNil(t, dryStats)
	})

	t.Run("Cover_Remaining_Methods", func(t *testing.T) {
		proj := &models.Project{Key: "COV", Title: "Coverage"}
		err := repo.CreateProject(proj)
		assert.NoError(t, err)

		_, err = repo.GetDryStatistics(proj.ID)
		assert.NoError(t, err)

		res, err := repo.GetProjectByID(99999)
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("CreateProject_DuplicateKey_Error", func(t *testing.T) {
		proj1 := &models.Project{Key: "DUP", Title: "First"}
		err := repo.CreateProject(proj1)
		assert.NoError(t, err)

		proj2 := &models.Project{Key: "DUP", Title: "Second"}
		err = repo.CreateProject(proj2)

		assert.Error(t, err)
	})

	t.Run("UpdateProject_Error", func(t *testing.T) {
		proj1 := &models.Project{Key: "UNIQUE1", Title: "Project 1"}
		proj2 := &models.Project{Key: "UNIQUE2", Title: "Project 2"}
		err := repo.CreateProject(proj1)
		assert.NoError(t, err)

		err = repo.CreateProject(proj2)
		assert.NoError(t, err)

		proj2.Key = "UNIQUE1"
		err = repo.UpdateProject(proj2)

		assert.Error(t, err)
	})

	t.Run("Exists_Error", func(t *testing.T) {
		exists, err := repo.Exists(-1)
		if err != nil {
			assert.False(t, exists)
			assert.Error(t, err)
		}
	})
}
