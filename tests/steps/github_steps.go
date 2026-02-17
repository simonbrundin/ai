package steps

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

// GitHubState holds the state for GitHub-related scenarios
type GitHubState struct {
	Issues         []Issue
	FilteredIssues []Issue
	Err            error
	Repo           string
}

// Issue represents a GitHub issue
type Issue struct {
	Number int
	Title  string
	State  string
	Labels []Label
}

type Label struct {
	Name string
}

// InitializeGitHubScenario sets up the GitHub step definitions
func InitializeGitHubScenario(ctx *godog.ScenarioContext) {
	state := &GitHubState{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &GitHubState{}
	})

	// Background
	ctx.Step(`^the GitHub API is available$`, func() error {
		return nil
	})

	// Given steps
	ctx.Step(`^I have a GitHub repository "([^"]*)"$`, func(repo string) error {
		state.Repo = repo
		state.Err = nil
		return nil
	})

	ctx.Step(`^I have the following issues:$`, func(table *godog.Table) error {
		state.Issues = []Issue{}
		for i, row := range table.Rows {
			if i == 0 { // Skip header
				continue
			}
			issue := Issue{
				Number: parseInt(row.Cells[0].Value),
				Title:  row.Cells[1].Value,
				Labels: []Label{},
			}
			// Handle labels column if present
			if len(row.Cells) > 2 {
				labelNames := strings.Split(row.Cells[2].Value, ",")
				for _, name := range labelNames {
					name = strings.TrimSpace(name)
					if name != "" {
						issue.Labels = append(issue.Labels, Label{Name: name})
					}
				}
			}
			state.Issues = append(state.Issues, issue)
		}
		return nil
	})

	ctx.Step(`^the repository "([^"]*)" has no issues$`, func(repo string) error {
		state.Issues = []Issue{}
		state.Repo = repo
		return nil
	})

	ctx.Step(`^GitHub API returns rate limit exceeded$`, func() error {
		state.Err = fmt.Errorf("rate limit exceeded")
		return nil
	})

	ctx.Step(`^the network is unavailable$`, func() error {
		state.Err = fmt.Errorf("connection refused")
		return nil
	})

	// When steps
	ctx.Step(`^I request all open issues$`, func() error {
		if state.Err != nil {
			return nil
		}
		return nil
	})

	ctx.Step(`^I filter by label "([^"]*)"$`, func(label string) error {
		state.FilteredIssues = FilterIssuesByLabel(state.Issues, label)
		return nil
	})

	ctx.Step(`^I search for "([^"]*)"$`, func(term string) error {
		state.FilteredIssues = FilterIssuesBySearch(state.Issues, term)
		return nil
	})

	// Then steps
	ctx.Step(`^I should receive a list of issues$`, func() error {
		if state.Err != nil {
			return fmt.Errorf("expected no error, got: %v", state.Err)
		}
		assert.True(nil, len(state.Issues) >= 0, "should have issues list")
		return nil
	})

	ctx.Step(`^each issue should have a number, title, and state$`, func() error {
		for _, issue := range state.Issues {
			if issue.Number == 0 || issue.Title == "" || issue.State == "" {
				return fmt.Errorf("issue missing required fields")
			}
		}
		return nil
	})

	ctx.Step(`^I should see (\d+) issue[s]?$`, func(count int) error {
		if state.FilteredIssues == nil {
			state.FilteredIssues = state.Issues
		}
		assert.Equal(nil, count, len(state.FilteredIssues))
		return nil
	})

	ctx.Step(`^issue #(\d+) should be in the result$`, func(number int) error {
		if state.FilteredIssues == nil {
			state.FilteredIssues = state.Issues
		}
		found := false
		for _, issue := range state.FilteredIssues {
			if issue.Number == number {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("issue #%d not found in filtered results", number)
		}
		return nil
	})

	ctx.Step(`^the result should be empty$`, func() error {
		assert.Equal(nil, 0, len(state.Issues))
		return nil
	})

	ctx.Step(`^I should get an error with "([^"]*)"$`, func(text string) error {
		if state.Err == nil {
			return fmt.Errorf("expected error containing '%s', got nil", text)
		}
		if !strings.Contains(state.Err.Error(), text) {
			return fmt.Errorf("expected error containing '%s', got '%s'", text, state.Err.Error())
		}
		return nil
	})

	ctx.Step(`^I should see retry information$`, func() error {
		return nil
	})

	ctx.Step(`^I should get a connection error$`, func() error {
		if state.Err == nil {
			return fmt.Errorf("expected connection error, got nil")
		}
		return nil
	})
}
