package etlprocess

import (
	"testing"
	"time"

	dbmodels "github.com/microservices-development-hse/connector/internal/models/db"
	jiramodels "github.com/microservices-development-hse/connector/internal/models/jira"
)

// ─── TransformProject ───────────────────────────────────────────────────────

func TestTransformProject_Valid(t *testing.T) {
	jp := jiramodels.ProjectResponse{
		ID:   "123",
		Key:  "TEST",
		Name: "Test Project",
		Self: "https://jira/rest/api/2/project/123",
	}

	p, err := TransformProject(jp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if p.ID != 123 {
		t.Errorf("expected ID=123, got %d", p.ID)
	}

	if p.Key != "TEST" {
		t.Errorf("expected Key=TEST, got %s", p.Key)
	}

	if p.Title != "Test Project" {
		t.Errorf("expected Title=Test Project, got %s", p.Title)
	}

	if p.URL != "https://jira/rest/api/2/project/123" {
		t.Errorf("expected URL mismatch, got %s", p.URL)
	}
}

func TestTransformProject_InvalidID(t *testing.T) {
	jp := jiramodels.ProjectResponse{
		ID:   "not-a-number",
		Key:  "TEST",
		Name: "Test Project",
	}

	_, err := TransformProject(jp)
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestTransformProject_EmptyID(t *testing.T) {
	jp := jiramodels.ProjectResponse{
		ID:   "",
		Key:  "TEST",
		Name: "Test Project",
	}

	_, err := TransformProject(jp)
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestTransformProject_ZeroID(t *testing.T) {
	jp := jiramodels.ProjectResponse{
		ID:   "0",
		Key:  "ZERO",
		Name: "Zero Project",
	}

	p, err := TransformProject(jp)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if p.ID != 0 {
		t.Errorf("expected ID=0, got %d", p.ID)
	}
}

// ─── TransformIssue ─────────────────────────────────────────────────────────

func makeJiraIssue() jiramodels.Issue {
	now := jiramodels.JTime{Time: time.Now()}

	return jiramodels.Issue{
		ID:  "10001",
		Key: "TEST-1",
		Fields: jiramodels.Fields{
			Summary:      "Test issue",
			Status:       jiramodels.Status{Name: "Open"},
			Priority:     jiramodels.Priority{Name: "High"},
			Created:      now,
			Updated:      now,
			TimeTracking: jiramodels.TimeTracking{TimeSpentSeconds: 3600},
		},
	}
}

func intPtr(i int) *int { return &i }

func TestTransformIssue_WithAuthorAndAssignee(t *testing.T) {
	ji := makeJiraIssue()
	authorID := intPtr(1)
	assigneeID := intPtr(2)

	issue, err := TransformIssue(ji, 42, authorID, assigneeID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if issue.ExternalID != "10001" {
		t.Errorf("expected ExternalID=10001, got %s", issue.ExternalID)
	}

	if issue.ProjectID != 42 {
		t.Errorf("expected ProjectID=42, got %d", issue.ProjectID)
	}

	if issue.Key != "TEST-1" {
		t.Errorf("expected Key=TEST-1, got %s", issue.Key)
	}

	if issue.Summary != "Test issue" {
		t.Errorf("expected Summary=Test issue, got %s", issue.Summary)
	}

	if issue.Status != "Open" {
		t.Errorf("expected Status=Open, got %s", issue.Status)
	}

	if issue.Priority != "High" {
		t.Errorf("expected Priority=High, got %s", issue.Priority)
	}

	if issue.TimeSpent != 3600 {
		t.Errorf("expected TimeSpent=3600, got %d", issue.TimeSpent)
	}

	if issue.AuthorID != authorID {
		t.Errorf("expected AuthorID=%v, got %v", authorID, issue.AuthorID)
	}

	if issue.AssigneeID != assigneeID {
		t.Errorf("expected AssigneeID=%v, got %v", assigneeID, issue.AssigneeID)
	}
}

func TestTransformIssue_NilAuthorAndAssignee(t *testing.T) {
	ji := makeJiraIssue()

	issue, err := TransformIssue(ji, 1, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if issue.AuthorID != nil {
		t.Errorf("expected nil AuthorID, got %v", issue.AuthorID)
	}

	if issue.AssigneeID != nil {
		t.Errorf("expected nil AssigneeID, got %v", issue.AssigneeID)
	}
}

func TestTransformIssue_NilAssigneeOnly(t *testing.T) {
	ji := makeJiraIssue()
	authorID := intPtr(5)

	issue, err := TransformIssue(ji, 1, authorID, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if issue.AuthorID == nil || *issue.AuthorID != 5 {
		t.Errorf("expected AuthorID=5, got %v", issue.AuthorID)
	}

	if issue.AssigneeID != nil {
		t.Errorf("expected nil AssigneeID, got %v", issue.AssigneeID)
	}
}

func TestTransformIssue_EmptyFields(t *testing.T) {
	ji := jiramodels.Issue{
		ID:     "99",
		Key:    "EMPTY-1",
		Fields: jiramodels.Fields{},
	}

	issue, err := TransformIssue(ji, 1, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if issue.Summary != "" {
		t.Errorf("expected empty Summary, got %s", issue.Summary)
	}

	if issue.Priority != "" {
		t.Errorf("expected empty Priority, got %s", issue.Priority)
	}

	if issue.Status != "" {
		t.Errorf("expected empty Status, got %s", issue.Status)
	}

	if issue.TimeSpent != 0 {
		t.Errorf("expected TimeSpent=0, got %d", issue.TimeSpent)
	}
}

// ─── TransformStatusChanges ─────────────────────────────────────────────────

func makeChangelog(authorName, field, from, to string) *jiramodels.Changelog {
	return &jiramodels.Changelog{
		Histories: []jiramodels.History{
			{
				Author:  jiramodels.Author{Name: authorName},
				Created: jiramodels.JTime{Time: time.Now()},
				Items: []jiramodels.Item{
					{Field: field, From: from, To: to},
				},
			},
		},
	}
}

func TestTransformStatusChanges_SingleStatusChange(t *testing.T) {
	changelog := makeChangelog("user1", "status", "Open", "Closed")
	authorIDs := map[string]int{"user1": 10}

	changes := TransformStatusChanges(changelog, 42, authorIDs)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}

	if changes[0].IssueID != 42 {
		t.Errorf("expected IssueID=42, got %d", changes[0].IssueID)
	}

	if changes[0].AuthorID != 10 {
		t.Errorf("expected AuthorID=10, got %d", changes[0].AuthorID)
	}

	if changes[0].FromStatus != "Open" {
		t.Errorf("expected FromStatus=Open, got %s", changes[0].FromStatus)
	}

	if changes[0].ToStatus != "Closed" {
		t.Errorf("expected ToStatus=Closed, got %s", changes[0].ToStatus)
	}
}

func TestTransformStatusChanges_SkipsNonStatusFields(t *testing.T) {
	changelog := makeChangelog("user1", "priority", "Low", "High")
	authorIDs := map[string]int{"user1": 10}

	changes := TransformStatusChanges(changelog, 1, authorIDs)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for non-status field, got %d", len(changes))
	}
}

func TestTransformStatusChanges_SkipsUnknownAuthor(t *testing.T) {
	changelog := makeChangelog("unknown_user", "status", "Open", "Closed")
	authorIDs := map[string]int{"user1": 10}

	changes := TransformStatusChanges(changelog, 1, authorIDs)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for unknown author, got %d", len(changes))
	}
}

func TestTransformStatusChanges_SkipsZeroAuthorID(t *testing.T) {
	changelog := makeChangelog("user1", "status", "Open", "Closed")
	authorIDs := map[string]int{"user1": 0}

	changes := TransformStatusChanges(changelog, 1, authorIDs)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for zero authorID, got %d", len(changes))
	}
}

func TestTransformStatusChanges_EmptyChangelog(t *testing.T) {
	changelog := &jiramodels.Changelog{Histories: []jiramodels.History{}}
	authorIDs := map[string]int{"user1": 10}

	changes := TransformStatusChanges(changelog, 1, authorIDs)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for empty changelog, got %d", len(changes))
	}
}

func TestTransformStatusChanges_MultipleItemsInHistory(t *testing.T) {
	changelog := &jiramodels.Changelog{
		Histories: []jiramodels.History{
			{
				Author:  jiramodels.Author{Name: "user1"},
				Created: jiramodels.JTime{Time: time.Now()},
				Items: []jiramodels.Item{
					{Field: "priority", From: "Low", To: "High"},
					{Field: "status", From: "Open", To: "In Progress"},
					{Field: "status", From: "In Progress", To: "Closed"},
				},
			},
		},
	}
	authorIDs := map[string]int{"user1": 10}

	changes := TransformStatusChanges(changelog, 1, authorIDs)
	if len(changes) != 2 {
		t.Errorf("expected 2 status changes, got %d", len(changes))
	}
}

func TestTransformStatusChanges_MultipleHistories(t *testing.T) {
	changelog := &jiramodels.Changelog{
		Histories: []jiramodels.History{
			{
				Author:  jiramodels.Author{Name: "user1"},
				Created: jiramodels.JTime{Time: time.Now()},
				Items:   []jiramodels.Item{{Field: "status", From: "Open", To: "In Progress"}},
			},
			{
				Author:  jiramodels.Author{Name: "user2"},
				Created: jiramodels.JTime{Time: time.Now()},
				Items:   []jiramodels.Item{{Field: "status", From: "In Progress", To: "Closed"}},
			},
			{
				Author:  jiramodels.Author{Name: "unknown"},
				Created: jiramodels.JTime{Time: time.Now()},
				Items:   []jiramodels.Item{{Field: "status", From: "Closed", To: "Reopened"}},
			},
		},
	}
	authorIDs := map[string]int{"user1": 1, "user2": 2}

	changes := TransformStatusChanges(changelog, 5, authorIDs)
	if len(changes) != 2 {
		t.Errorf("expected 2 changes (unknown author skipped), got %d", len(changes))
	}
}

func TestTransformStatusChanges_EmptyAuthorName(t *testing.T) {
	changelog := makeChangelog("", "status", "Open", "Closed")
	authorIDs := map[string]int{"user1": 10}

	changes := TransformStatusChanges(changelog, 1, authorIDs)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for empty author name, got %d", len(changes))
	}
}

func TestTransformStatusChanges_CorrectChangeTime(t *testing.T) {
	expectedTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	changelog := &jiramodels.Changelog{
		Histories: []jiramodels.History{
			{
				Author:  jiramodels.Author{Name: "user1"},
				Created: jiramodels.JTime{Time: expectedTime},
				Items:   []jiramodels.Item{{Field: "status", From: "Open", To: "Closed"}},
			},
		},
	}
	authorIDs := map[string]int{"user1": 1}

	changes := TransformStatusChanges(changelog, 1, authorIDs)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}

	if !changes[0].ChangeTime.Equal(expectedTime) {
		t.Errorf("expected ChangeTime=%v, got %v", expectedTime, changes[0].ChangeTime)
	}
}

// ─── helpers ────────────────────────────────────────────────────────────────

var _ = dbmodels.StatusChange{}
