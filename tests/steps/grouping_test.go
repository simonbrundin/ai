package steps

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupIssuesByRepo(t *testing.T) {
	tests := []struct {
		name     string
		issues   []Issue
		expected map[string][]Issue
	}{
		{
			name: "group by repo",
			issues: []Issue{
				{Number: 1, Title: "Fix bug", Repo: "simonbrundin/ai"},
				{Number: 2, Title: "Add feature", Repo: "simonbrundin/ai"},
				{Number: 3, Title: "Update docs", Repo: "simonbrundin/web"},
			},
			expected: map[string][]Issue{
				"ai": {
					{Number: 1, Title: "Fix bug", Repo: "simonbrundin/ai"},
					{Number: 2, Title: "Add feature", Repo: "simonbrundin/ai"},
				},
				"web": {
					{Number: 3, Title: "Update docs", Repo: "simonbrundin/web"},
				},
			},
		},
		{
			name: "single repo",
			issues: []Issue{
				{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
				{Number: 2, Title: "Issue 2", Repo: "simonbrundin/ai"},
			},
			expected: map[string][]Issue{
				"ai": {
					{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
					{Number: 2, Title: "Issue 2", Repo: "simonbrundin/ai"},
				},
			},
		},
		{
			name:     "empty issues",
			issues:   []Issue{},
			expected: map[string][]Issue{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GroupIssuesByRepo(tt.issues)
			assert.Equal(t, len(tt.expected), len(result))
			for repo, expectedIssues := range tt.expected {
				resultIssues, ok := result[repo]
				assert.True(t, ok, "repo %s should exist", repo)
				assert.Equal(t, len(expectedIssues), len(resultIssues))
			}
		})
	}
}

func TestGetRepoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simonbrundin/ai", "ai"},
		{"simonbrundin/web", "web"},
		{"owner/repo-name", "repo-name"},
		{"just-name", "just-name"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GetRepoName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
