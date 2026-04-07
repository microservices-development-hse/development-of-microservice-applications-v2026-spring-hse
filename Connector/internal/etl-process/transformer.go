package etlprocess

import (
	"fmt"
	"strconv"

	dbmodels "github.com/microservices-development-hse/connector/internal/models/db"
	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
)

func TransformProject(jp jiramodels.ProjectResponse) (dbmodels.Project, error) {
	id, err := strconv.Atoi(jp.ID)
	if err != nil {
		return dbmodels.Project{}, fmt.Errorf("transform project: invalid id %q: %w", jp.ID, err)
	}

	return dbmodels.Project{
		ID:    id,
		Key:   jp.Key,
		Title: jp.Name,
		URL:   jp.Self,
	}, nil
}

func TransformIssue(ji jiramodels.Issue, projectID int, authorID, assigneeID *int) (dbmodels.Issue, error) {
	return dbmodels.Issue{
		ExternalID:  ji.ID,
		ProjectID:   projectID,
		AuthorID:    authorID,
		AssigneeID:  assigneeID,
		Key:         ji.Key,
		Summary:     ji.Fields.Summary,
		Priority:    ji.Fields.Priority.Name,
		Status:      ji.Fields.Status.Name,
		CreatedTime: ji.Fields.Created.Time,
		UpdatedTime: ji.Fields.Updated.Time,
		TimeSpent:   ji.Fields.TimeTracking.TimeSpentSeconds,
	}, nil
}

func TransformStatusChanges(changelog *jiramodels.Changelog, issueID int, authorIDs map[string]int) []dbmodels.StatusChange {
	var changes []dbmodels.StatusChange

	for _, h := range changelog.Histories {
		authorID, ok := authorIDs[h.Author.Name]
		if !ok || authorID == 0 {
			continue
		}

		for _, item := range h.Items {
			if item.Field == "status" {
				changes = append(changes, dbmodels.StatusChange{
					IssueID:    issueID,
					AuthorID:   authorID,
					ChangeTime: h.Created.Time,
					FromStatus: item.From,
					ToStatus:   item.To,
				})
			}
		}
	}

	return changes
}
