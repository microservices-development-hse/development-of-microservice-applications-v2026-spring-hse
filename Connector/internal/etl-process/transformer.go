package etlprocess

import (
	"encoding/json"
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
		ID:   id,
		Key:  jp.Key,
		Name: jp.Name,
		URL:  jp.Self,
	}, nil
}

func TransformIssue(ji jiramodels.Issue, projectID int) (dbmodels.Issue, []dbmodels.User, error) {
	var changelogJSON interface{}

	if ji.Changelog != nil {
		raw, err := json.Marshal(ji.Changelog)
		if err != nil {
			return dbmodels.Issue{}, nil, fmt.Errorf("transform issue %s: marshal changelog: %w", ji.Key, err)
		}

		changelogJSON = raw
	}

	issue := dbmodels.Issue{
		ProjectID: projectID,
		Key:       ji.Key,
		Summary:   ji.Fields.Summary,
		Status:    ji.Fields.Status.Name,
		Created:   ji.Fields.Created.Time,
		Updated:   ji.Fields.Updated.Time,
		Changelog: changelogJSON,
	}

	var users []dbmodels.User

	if ji.Changelog != nil {
		seen := make(map[string]struct{})
		for _, h := range ji.Changelog.Histories {
			if _, ok := seen[h.Author.Name]; ok {
				continue
			}

			seen[h.Author.Name] = struct{}{}
			users = append(users, dbmodels.User{
				Username:    h.Author.Name,
				DisplayName: h.Author.DisplayName,
			})
		}
	}

	return issue, users, nil
}
