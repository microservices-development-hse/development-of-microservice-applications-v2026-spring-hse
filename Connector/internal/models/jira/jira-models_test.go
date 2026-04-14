package jira

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJTime_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "full format with timezone",
			input:    `"2023-01-15T10:30:00.000+0300"`,
			expected: "2023-01-15T10:30:00+03:00",
		},
		{
			name:     "Z format",
			input:    `"2023-01-15T10:30:00.000Z"`,
			expected: "2023-01-15T10:30:00Z",
		},
		{
			name:     "RFC3339",
			input:    `"2023-01-15T10:30:00Z"`,
			expected: "2023-01-15T10:30:00Z",
		},
		{
			name:     "null value",
			input:    `null`,
			expected: "",
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jt JTime

			err := json.Unmarshal([]byte(tt.input), &jt)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expected == "" && !jt.IsZero() {
				t.Errorf("expected zero time, got %v", jt.Time)
			}

			if tt.expected != "" {
				got := jt.Format("2006-01-02T15:04:05Z07:00")
				if got != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, got)
				}
			}
		})
	}
}

func TestJTime_UnmarshalJSON_Invalid(t *testing.T) {
	var jt JTime

	err := json.Unmarshal([]byte(`"invalid-date"`), &jt)
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestIssueSearchResponse_MarshalUnmarshal(t *testing.T) {
	original := IssueSearchResponse{
		StartAt:    0,
		MaxResults: 50,
		Total:      100,
		Issues: []Issue{
			{
				ID:  "1",
				Key: "TEST-1",
				Fields: Fields{
					Summary: "Test issue",
					Status: Status{
						ID:   "1",
						Name: "Open",
					},
					Priority: Priority{Name: "High"},
					Creator: Author{
						Name:        "john",
						DisplayName: "John Doe",
					},
				},
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var result IssueSearchResponse

	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if result.Total != original.Total {
		t.Errorf("expected Total %d, got %d", original.Total, result.Total)
	}

	if len(result.Issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(result.Issues))
	}

	if result.Issues[0].Key != "TEST-1" {
		t.Errorf("expected TEST-1, got %s", result.Issues[0].Key)
	}
}

func TestIssue_WithChangelog(t *testing.T) {
	issue := Issue{
		ID:  "123",
		Key: "PROJ-123",
		Changelog: &Changelog{
			Total: 2,
			Histories: []History{
				{
					ID:      "1",
					Created: JTime{Time: time.Now()},
					Items: []Item{
						{Field: "status", From: "Open", To: "In Progress"},
					},
				},
			},
		},
	}

	data, err := json.Marshal(issue)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var result Issue

	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if result.Changelog == nil {
		t.Error("changelog is nil after unmarshal")
	}

	if result.Changelog.Total != 2 {
		t.Errorf("expected Total 2, got %d", result.Changelog.Total)
	}
}

func TestProjectResponse_Unmarshal(t *testing.T) {
	jsonData := `{"id":"10000","key":"TEST","name":"Test Project","self":"http://jira/rest/api/2/project/10000"}`

	var resp ProjectResponse

	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if resp.ID != "10000" {
		t.Errorf("expected ID 10000, got %s", resp.ID)
	}

	if resp.Key != "TEST" {
		t.Errorf("expected Key TEST, got %s", resp.Key)
	}

	if resp.Name != "Test Project" {
		t.Errorf("expected Name Test Project, got %s", resp.Name)
	}
}

func TestPriority_Unmarshal(t *testing.T) {
	data := []byte(`{"name":"Critical"}`)

	var p Priority

	err := json.Unmarshal(data, &p)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if p.Name != "Critical" {
		t.Errorf("expected Critical, got %s", p.Name)
	}
}

func TestTimeTracking_Unmarshal(t *testing.T) {
	data := []byte(`{"timeSpentSeconds":3600}`)

	var tt TimeTracking

	err := json.Unmarshal(data, &tt)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if tt.TimeSpentSeconds != 3600 {
		t.Errorf("expected 3600, got %d", tt.TimeSpentSeconds)
	}
}

func TestAuthor_Unmarshal(t *testing.T) {
	data := []byte(`{"self":"http://...","accountId":"123","name":"john","displayName":"John Smith"}`)

	var a Author

	err := json.Unmarshal(data, &a)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if a.Name != "john" {
		t.Errorf("expected john, got %s", a.Name)
	}

	if a.DisplayName != "John Smith" {
		t.Errorf("expected John Smith, got %s", a.DisplayName)
	}
}

func TestHistory_WithItems(t *testing.T) {
	data := []byte(`{
		"id":"1",
		"author":{"name":"admin"},
		"created":"2023-01-01T00:00:00.000Z",
		"items":[{"field":"status","fromString":"Open","toString":"Closed"}]
	}`)

	var h History

	err := json.Unmarshal(data, &h)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if len(h.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(h.Items))
	}

	if h.Items[0].Field != "status" {
		t.Errorf("expected status, got %s", h.Items[0].Field)
	}
}
