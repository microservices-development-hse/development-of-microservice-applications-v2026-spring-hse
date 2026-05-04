package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAnalyticsRepository_Integration(t *testing.T) {
	env := SetupRepoTestEnv(t)
	repo := NewAnalyticsRepository(env.DB)
	ctx := context.Background()

	setup := func(t *testing.T) (models.Project, models.Author) {

		env.DB.Exec("TRUNCATE TABLE status_changes, issues, projects, authors, analytics_snapshots RESTART IDENTITY CASCADE")

		author := models.Author{ExternalID: "author_" + t.Name(), Name: "Analyst"}
		if err := env.DB.Create(&author).Error; err != nil {
			t.Fatalf("failed to create author: %v", err)
		}

		proj := models.Project{Key: "PRJ", Title: "Analytics Project"}
		if err := env.DB.Create(&proj).Error; err != nil {
			t.Fatalf("failed to create project: %v", err)
		}

		return proj, author
	}

	t.Run("CalculateTimeInState_Success", func(t *testing.T) {
		proj, author := setup(t)
		issue := models.Issue{
			ProjectID:  proj.ID,
			AuthorID:   author.ID,
			AssigneeID: author.ID,
			Key:        "AN-1",
			Summary:    "Time Test",
			ExternalID: "ext-1",
		}
		env.DB.Create(&issue)

		startTime := time.Now().Add(-10 * time.Hour).UTC().Truncate(time.Second)

		env.DB.Create(&models.StatusChanges{IssueID: issue.ID, AuthorID: author.ID, ToStatus: "To Do", ChangeTime: startTime})
		env.DB.Create(&models.StatusChanges{IssueID: issue.ID, AuthorID: author.ID, ToStatus: "In Progress", ChangeTime: startTime.Add(2 * time.Hour)})
		env.DB.Create(&models.StatusChanges{IssueID: issue.ID, AuthorID: author.ID, ToStatus: "Done", ChangeTime: startTime.Add(5 * time.Hour)})

		res, err := repo.CalculateTimeInState(ctx, proj.ID)
		assert.NoError(t, err)
		if assert.NotEmpty(t, res) {
			assert.Contains(t, res, "To Do")
			assert.InDelta(t, 2.0, res["To Do"][0], 0.001)
		}
	})

	t.Run("GetProjectComplexity_Full_Coverage", func(t *testing.T) {
		proj, author := setup(t)
		now := time.Now().UTC()
		issue := models.Issue{
			ProjectID:   proj.ID,
			AuthorID:    author.ID,
			AssigneeID:  author.ID,
			Key:         "COMP-1",
			Summary:     "Complexity Test",
			ExternalID:  "ext-comp-1",
			CreatedTime: now.Add(-10 * time.Hour),
			ClosedTime:  &now,
		}
		env.DB.Create(&issue)

		env.DB.Create(&models.StatusChanges{IssueID: issue.ID, AuthorID: author.ID, ToStatus: "In Progress", ChangeTime: now.Add(-5 * time.Hour)})

		results, err := repo.GetProjectComplexity(ctx, proj.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.GreaterOrEqual(t, results[0].MoveCount, 1)
	})

	t.Run("GetOpenTasksBottlenecks_Success", func(t *testing.T) {
		proj, author := setup(t)
		env.DB.Create(&models.Issue{
			ProjectID:  proj.ID,
			AuthorID:   author.ID,
			AssigneeID: author.ID,
			Key:        "OPEN-1",
			Summary:    "Bottleneck Test",
			ExternalID: "e1",
			Status:     "In Progress",
		})

		results, err := repo.GetOpenTasksBottlenecks(ctx, proj.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})

	t.Run("Distributions_Success", func(t *testing.T) {
		proj, author := setup(t)

		issue := models.Issue{
			ProjectID:  proj.ID,
			AuthorID:   author.ID,
			AssigneeID: author.ID,
			Key:        "D-1",
			Summary:    "S1",
			ExternalID: "dist-ext-1",
			Status:     "Open",
			Priority:   "High",
		}
		err := env.DB.Create(&issue).Error
		assert.NoError(t, err)

		distStatus, err := repo.GetTaskStatusDistribution(ctx, proj.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, distStatus)

		distPriority, err := repo.GetTaskPriorityDistribution(ctx, proj.ID)
		assert.NoError(t, err)
		assert.NotEmpty(t, distPriority)
	})

	t.Run("Snapshot_Lifecycle", func(t *testing.T) {
		proj, _ := setup(t)
		snapshot := &models.AnalyticsSnapshot{
			ProjectID: proj.ID,
			Type:      "velocity",
			Data:      []byte(`{"score": 100}`),
		}
		err := repo.SaveSnapshot(ctx, snapshot)
		assert.NoError(t, err)

		latest, err := repo.GetLatestSnapshot(ctx, proj.ID, "velocity")
		assert.NoError(t, err)
		assert.NotNil(t, latest)

		none, err := repo.GetLatestSnapshot(ctx, proj.ID, "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, none)
	})

	t.Run("GetOpenTasksBottlenecks_With_Filters", func(t *testing.T) {
		proj, author := setup(t)

		env.DB.Create(&models.Issue{
			ProjectID: proj.ID, AuthorID: author.ID, AssigneeID: author.ID,
			Key: "OPEN-1", Summary: "In Work", ExternalID: "e1", Status: "In Progress",
		})

		env.DB.Create(&models.Issue{
			ProjectID: proj.ID, AuthorID: author.ID, AssigneeID: author.ID,
			Key: "CLOSED-1", Summary: "Finished", ExternalID: "e2", Status: "Done",
		})

		results, err := repo.GetOpenTasksBottlenecks(ctx, proj.ID)
		assert.NoError(t, err)

		for _, r := range results {
			assert.NotEqual(t, "Done", r.CurrentStatus)
		}
	})

	t.Run("Distributions_Multiple_Items", func(t *testing.T) {
		proj, author := setup(t)

		env.DB.Create(&models.Issue{
			ProjectID:  proj.ID,
			AuthorID:   author.ID,
			AssigneeID: author.ID,
			Key:        "M-1",
			Summary:    "S1",
			ExternalID: "em1",
			Status:     "StatusA",
			Priority:   "High",
		})
		env.DB.Create(&models.Issue{
			ProjectID:  proj.ID,
			AuthorID:   author.ID,
			AssigneeID: author.ID,
			Key:        "M-2",
			Summary:    "S2",
			ExternalID: "em2",
			Status:     "StatusB",
			Priority:   "Low",
		})

		distStatus, err := repo.GetTaskStatusDistribution(ctx, proj.ID)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(distStatus))

		distPriority, err := repo.GetTaskPriorityDistribution(ctx, proj.ID)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(distPriority))
	})

	t.Run("Snapshot_Operations_And_Errors", func(t *testing.T) {
		proj, _ := setup(t)

		snapshot := &models.AnalyticsSnapshot{
			ProjectID: proj.ID,
			Type:      "velocity",
			Data:      []byte(`{"val": 10}`),
		}
		assert.NoError(t, repo.SaveSnapshot(ctx, snapshot))

		latest, err := repo.GetLatestSnapshot(ctx, proj.ID, "velocity")
		assert.NoError(t, err)
		assert.NotNil(t, latest)

		none, err := repo.GetLatestSnapshot(ctx, proj.ID, "non-existent-type")
		assert.NoError(t, err)
		assert.Nil(t, none)
	})
}
