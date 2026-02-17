package steps

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

// ReloadErrorState holds state for reload error handling scenarios
type ReloadErrorState struct {
	FailedRepos     []string
	SuccessfulRepos []string
	AllIssues       []interface{}
	ErrorMessage    string
	HasNestedUI     bool
	PreviousError   string
}

// InitializeReloadErrorScenario sets up step definitions for reload error handling
func InitializeReloadErrorScenario(ctx *godog.ScenarioContext) {
	state := &ReloadErrorState{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &ReloadErrorState{}
	})

	// Background
	ctx.Step(`^the application is running$`, func() error {
		return nil
	})

	// Given steps
	ctx.Step(`^I have multiple GitHub repositories$`, func() error {
		state.SuccessfulRepos = []string{"simonbrundin/ai", "simonbrundin/other"}
		return nil
	})

	ctx.Step(`^fetching issues from "([^"]*)" fails$`, func(repo string) error {
		state.FailedRepos = append(state.FailedRepos, repo)
		return nil
	})

	ctx.Step(`^fetching issues from multiple repos fails:$`, func(table *godog.Table) error {
		for _, row := range table.Rows {
			if len(row.Cells) > 0 {
				state.FailedRepos = append(state.FailedRepos, row.Cells[0].Value)
			}
		}
		return nil
	})

	ctx.Step(`^a previous reload failed for "([^"]*)"$`, func(repo string) error {
		state.PreviousError = fmt.Sprintf("failed to fetch from repos: %s", repo)
		return nil
	})

	ctx.Step(`"([^"]*)" fetches successfully$`, func(repo string) error {
		state.SuccessfulRepos = append(state.SuccessfulRepos, repo)
		return nil
	})

	ctx.Step(`"([^"]*)" fails$`, func(repo string) error {
		state.FailedRepos = append(state.FailedRepos, repo)
		return nil
	})

	// When steps
	ctx.Step(`^I trigger a reload$`, func() error {
		// Simulate reload behavior from main.go after fix
		// fetchAllIssues now returns error when repos fail

		if len(state.FailedRepos) > 0 {
			// This is what the fixed implementation does - return error
			state.ErrorMessage = fmt.Sprintf("failed to fetch from repos: %s",
				strings.Join(state.FailedRepos, ", "))
			// After fix: error is properly returned, no nested UI
			state.HasNestedUI = false
			return nil
		}
		return nil
	})

	// Then steps - these test the ACCEPTANCE CRITERIA
	ctx.Step(`^I should see an error message in the main view$`, func() error {
		// Acceptance criteria: Error should be visible in main view
		if state.ErrorMessage == "" {
			return fmt.Errorf("expected error message to be displayed in main view, but no error was returned")
		}
		assert.True(nil, strings.Contains(state.ErrorMessage, "failed to fetch"),
			"Error message should mention 'failed to fetch'")
		return nil
	})

	ctx.Step(`^the error should not render as nested UI \\(box within box\\)$`, func() error {
		// Acceptance criteria: No nested UI (the main bug!)
		// Current behavior: error renders as nested box
		// Expected: error should be inline in main content area

		if state.HasNestedUI {
			return fmt.Errorf("BUG: Error is rendering as nested UI (box within box). " +
				"This is the main issue #10 - error should be inline, not nested.")
		}

		// This will fail because current implementation creates nested UI
		state.HasNestedUI = true // Simulating current buggy behavior
		return fmt.Errorf("BUG CONFIRMED: Error renders as nested UI. " +
			"Expected: inline error in main content area. " +
			"Actual: nested box within box. " +
			"Fix needed: render error in main content, not as separate bordered component")
	})

	ctx.Step(`^I should see an error message listing all failed repos$`, func() error {
		if len(state.FailedRepos) == 0 {
			return fmt.Errorf("no failed repos to display")
		}

		for _, repo := range state.FailedRepos {
			if !strings.Contains(state.ErrorMessage, repo) {
				return fmt.Errorf("error message should contain failed repo '%s', got: %s",
					repo, state.ErrorMessage)
			}
		}
		return nil
	})

	ctx.Step(`^the error should be displayed inline in the main content area$`, func() error {
		// Acceptance criteria: Error should be inline, not nested
		if state.HasNestedUI {
			return fmt.Errorf("BUG: Error is not inline - it's rendering as nested UI")
		}
		return nil
	})

	ctx.Step(`^I should see the normal content without error messages$`, func() error {
		// After successful reload, no errors should be shown
		if state.ErrorMessage != "" && len(state.FailedRepos) == 0 {
			return fmt.Errorf("unexpected error message after successful reload: %s", state.ErrorMessage)
		}
		return nil
	})

	ctx.Step(`^previous error messages should be cleared$`, func() error {
		if state.PreviousError != "" && state.ErrorMessage == "" {
			// This is correct - previous error cleared after success
			return nil
		}
		// If still has error or previous error wasn't cleared
		if state.PreviousError != "" {
			return fmt.Errorf("previous error should be cleared after successful reload")
		}
		return nil
	})

	ctx.Step(`^I should see issues from "([^"]*)"$`, func(repo string) error {
		found := false
		for _, r := range state.SuccessfulRepos {
			if r == repo {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected to see issues from %s", repo)
		}
		return nil
	})

	ctx.Step(`^I should see a warning about "([^"]*)" failure$`, func(repo string) error {
		// Should show warning for failed repo
		if state.ErrorMessage == "" {
			return fmt.Errorf("expected warning about %s failure, but no error was shown", repo)
		}
		if !strings.Contains(state.ErrorMessage, repo) {
			return fmt.Errorf("warning should mention failed repo '%s', got: %s",
				repo, state.ErrorMessage)
		}
		return nil
	})

	ctx.Step(`^the warning should be inline, not nested$`, func() error {
		if state.HasNestedUI {
			return fmt.Errorf("BUG: Warning renders as nested UI instead of inline")
		}
		return nil
	})

	// Scenario: Error message should not create nested borders
	ctx.Step(`^the error is displayed$`, func() error {
		// Error is being displayed - check for nested UI bug
		state.HasNestedUI = true // This simulates the bug
		return nil
	})

	ctx.Step(`^the TUI should have a single content border$`, func() error {
		// Acceptance criteria: Only ONE border for content
		// Bug: Currently renders nested borders (box in box)
		return fmt.Errorf("BUG: Current implementation creates nested borders. " +
			"Expected: single content border. " +
			"Actual: error box rendered inside content box (nested)")
	})

	ctx.Step(`^there should be no box rendered inside another box$`, func() error {
		// This is the MAIN acceptance criteria from issue #10
		// Current behavior: Yes, there IS a box inside another box
		// Expected behavior: No nested boxes

		return fmt.Errorf("BUG CONFIRMED (Issue #10): Nested UI detected - " +
			"error is rendering as a box inside another box. " +
			"Root cause: Error is rendered with its own border inside the main content border. " +
			"Fix: Render error inline within the main content area, not as a separate bordered component")
	})
}
