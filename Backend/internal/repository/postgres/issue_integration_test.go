package postgres

import (
	"testing"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestIssueRepository_Full(t *testing.T) {
	env := SetupRepoTestEnv(t)
	repo := NewIssueRepository(env.DB)

	author := models.Author{ExternalID: "issue_auth", Name: "Ivan"}
	env.DB.Create(&author)

	proj := models.Project{Key: "ISS", Title: "Issue Proj"}
	env.DB.Create(&proj)

	t.Run("Issue_Lifecycle_And_History", func(t *testing.T) {
		issue := &models.Issue{
			ProjectID: proj.ID, AuthorID: author.ID, AssigneeID: author.ID,
			Key: "ISS-1", Summary: "Test Issue", ExternalID: "ext_1", Status: "Open",
		}

		err := repo.CreateIssue(issue)
		assert.NoError(t, err)

		found, err := repo.GetIssueByKey("ISS-1")
		assert.NoError(t, err)
		assert.Equal(t, "ext_1", found.ExternalID)

		issue.Status = "In Progress"
		err = repo.UpdateIssueWithHistory(issue, "Open")
		assert.NoError(t, err)

		var change models.StatusChanges

		err = env.DB.Where("issue_id = ?", issue.ID).First(&change).Error
		assert.NoError(t, err)
		assert.Equal(t, "Open", change.FromStatus)
		assert.Equal(t, "In Progress", change.ToStatus)
	})

	t.Run("Issue_Extended_Coverage", func(t *testing.T) {
		issue := &models.Issue{
			ProjectID: proj.ID, AuthorID: author.ID, AssigneeID: author.ID,
			Key: "EXT-1", ExternalID: "unique_ext_id", Summary: "Sum",
		}

		err := repo.CreateIssue(issue)
		assert.NoError(t, err)

		issue.Summary = "New Summary"

		issues, total, err := repo.GetIssuesByProjectID(proj.ID, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)
		assert.NotEmpty(t, issues)
	})

	t.Run("GetIssueByKey_NotFound", func(t *testing.T) {
		res, err := repo.GetIssueByKey("NON-EXISTENT")

		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("CreateIssue_FK_Error", func(t *testing.T) {
		issue := &models.Issue{
			ProjectID:  9999,
			AuthorID:   9999,
			AssigneeID: 9999,
			Key:        "ERR-1",
		}

		err := repo.CreateIssue(issue)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "repository error")
	})

	t.Run("UpdateIssueWithHistory_Errors", func(t *testing.T) {
		issue := &models.Issue{
			ProjectID:  proj.ID,
			AuthorID:   author.ID,
			AssigneeID: author.ID,
			Key:        "ERR-UPD",
			Status:     "Open",
		}

		err := repo.CreateIssue(issue)
		require.NoError(t, err)

		issue.ProjectID = 99999
		err = repo.UpdateIssueWithHistory(issue, "Open")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transaction failed")

		err = env.DB.Transaction(func(tx *gorm.DB) error {
			if execErr := tx.Exec("SELECT 1/0").Error; execErr != nil {
				return execErr
			}

			r := NewIssueRepository(tx)
			issue.ProjectID = proj.ID
			updateErr := r.UpdateIssueWithHistory(issue, "Closed")

			return updateErr
		})

		assert.Error(t, err)
	})

	t.Run("GetIssuesByProjectID_SQL_Errors", func(t *testing.T) {
		err := env.DB.Transaction(func(tx *gorm.DB) error {
			tx.Exec("SELECT 'invalid'::integer")

			r := NewIssueRepository(tx)
			_, _, err := r.GetIssuesByProjectID(proj.ID, 10, 0)

			assert.Error(t, err)

			return err
		})

		require.NoError(t, err)
	})
}
