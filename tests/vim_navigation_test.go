package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Tests för Issue #16: vim-liknande navigering i issue-listan med j, k, o
//
// ACCEPTANCE CRITERIA:
// - j navigerar till nästa issue i listan
// - k navigerar till föregående issue i listan
// - o öppnar den valda issue:n i standardwebbläsaren (GitHub)
//
// EDGE CASES:
// - j på sista issue → stanna på sista
// - k på första issue → stanna på första
// - Navigering med tom lista → inget krasch
// - Växla mellan tabs → selection bevaras
// =============================================================================

// testIssue mirrors the issue struct from main.go
type testIssue struct {
	Number int
	Title  string
	State  string
	Labels []string
	Repo   string
}

// vimNavigator simulates the vim navigation logic for testing
// This is a pure function that can be tested without the full TUI
type vimNavigator struct {
	issues        []testIssue
	selectedIssue int
	issueURL      string
	repo          string
}

// moveNext moves selection to next issue
func (v *vimNavigator) moveNext() {
	if len(v.issues) == 0 {
		return
	}
	if v.selectedIssue < len(v.issues)-1 {
		v.selectedIssue++
	}
}

// movePrev moves selection to previous issue
func (v *vimNavigator) movePrev() {
	if len(v.issues) == 0 {
		return
	}
	if v.selectedIssue > 0 {
		v.selectedIssue--
	}
}

// openSelectedIssue stores the URL of the currently selected issue
func (v *vimNavigator) openSelectedIssue() {
	if len(v.issues) == 0 || v.selectedIssue >= len(v.issues) {
		v.issueURL = ""
		return
	}
	issue := v.issues[v.selectedIssue]
	v.issueURL = "https://github.com/" + v.repo + "/issues/" + intToString(issue.Number)
}

// intToString converts int to string
func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// =============================================================================
// HAPPY PATH TESTS
// =============================================================================

// Test: Pressing 'j' should move to next issue
func Test_VIM_PressJ_MoveToNextIssue(t *testing.T) {
	v := &vimNavigator{
		issues: []testIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
			{Number: 2, Title: "Issue 2", Repo: "simonbrundin/ai"},
			{Number: 3, Title: "Issue 3", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 0,
		repo:          "simonbrundin/ai",
	}

	v.moveNext()

	assert.Equal(t, 1, v.selectedIssue, "Pressing 'j' should move to next issue (index 1)")
}

// Test: Pressing 'k' should move to previous issue
func Test_VIM_PressK_MoveToPreviousIssue(t *testing.T) {
	v := &vimNavigator{
		issues: []testIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
			{Number: 2, Title: "Issue 2", Repo: "simonbrundin/ai"},
			{Number: 3, Title: "Issue 3", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 2,
		repo:          "simonbrundin/ai",
	}

	v.movePrev()

	assert.Equal(t, 1, v.selectedIssue, "Pressing 'k' should move to previous issue (index 1)")
}

// Test: Pressing 'o' should store URL to open
func Test_VIM_PressO_StoreURLToOpen(t *testing.T) {
	v := &vimNavigator{
		issues: []testIssue{
			{Number: 16, Title: "Test Issue", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 0,
		repo:          "simonbrundin/ai",
	}

	v.openSelectedIssue()

	assert.Equal(t, "https://github.com/simonbrundin/ai/issues/16", v.issueURL,
		"Pressing 'o' should store the GitHub URL for the selected issue")
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

// Edge case: Pressing 'j' on last issue should stay on last
func Test_VIM_EdgeCase_PressJOnLastIssue_StayOnLast(t *testing.T) {
	v := &vimNavigator{
		issues: []testIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
			{Number: 2, Title: "Issue 2", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 1, // Already on last
		repo:          "simonbrundin/ai",
	}

	v.moveNext()

	assert.Equal(t, 1, v.selectedIssue,
		"Pressing 'j' on last issue should stay on last (index 1)")
}

// Edge case: Pressing 'k' on first issue should stay on first
func Test_VIM_EdgeCase_PressKOnFirstIssue_StayOnFirst(t *testing.T) {
	v := &vimNavigator{
		issues: []testIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
			{Number: 2, Title: "Issue 2", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 0, // Already on first
		repo:          "simonbrundin/ai",
	}

	v.movePrev()

	assert.Equal(t, 0, v.selectedIssue,
		"Pressing 'k' on first issue should stay on first (index 0)")
}

// Edge case: Navigating with empty issue list should not crash
func Test_VIM_EdgeCase_EmptyIssueList_NoCrash(t *testing.T) {
	v := &vimNavigator{
		issues:        []testIssue{},
		selectedIssue: 0,
		repo:          "simonbrundin/ai",
	}

	// Press j on empty list - should not panic
	v.moveNext()

	assert.Equal(t, 0, v.selectedIssue,
		"Pressing 'j' on empty list should stay at 0")

	// Press k on empty list - should not panic
	v.movePrev()

	assert.Equal(t, 0, v.selectedIssue,
		"Pressing 'k' on empty list should stay at 0")
}

// Edge case: 'o' with empty list should not crash
func Test_VIM_EdgeCase_PressOOnEmptyList_NoCrash(t *testing.T) {
	v := &vimNavigator{
		issues:        []testIssue{},
		selectedIssue: 0,
		repo:          "simonbrundin/ai",
	}

	// Should not panic
	v.openSelectedIssue()

	assert.Equal(t, "", v.issueURL,
		"Pressing 'o' on empty list should result in empty URL")
}

// Edge case: Selection out of bounds should be handled gracefully
// (stays at invalid index, but openSelectedIssue handles it gracefully)
func Test_VIM_EdgeCase_InvalidSelection_HandledGracefully(t *testing.T) {
	v := &vimNavigator{
		issues: []testIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 5, // Invalid index (list only has 1 item)
		repo:          "simonbrundin/ai",
	}

	// After any vim operation, should handle gracefully (not panic)
	v.moveNext()
	v.movePrev()
	v.openSelectedIssue()

	// URL should be empty since selection was invalid (out of bounds)
	assert.Equal(t, "", v.issueURL,
		"Invalid selection should result in empty URL")
}

// Edge case: Multiple rapid navigations
func Test_VIM_EdgeCase_MultipleRapidNavigations(t *testing.T) {
	v := &vimNavigator{
		issues: []testIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
			{Number: 2, Title: "Issue 2", Repo: "simonbrundin/ai"},
			{Number: 3, Title: "Issue 3", Repo: "simonbrundin/ai"},
			{Number: 4, Title: "Issue 4", Repo: "simonbrundin/ai"},
			{Number: 5, Title: "Issue 5", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 0,
		repo:          "simonbrundin/ai",
	}

	// j j j k k j
	v.moveNext() // 1
	v.moveNext() // 2
	v.moveNext() // 3
	v.movePrev() // 2
	v.movePrev() // 1
	v.moveNext() // 2

	assert.Equal(t, 2, v.selectedIssue,
		"Multiple navigations should result in correct position")
}

// Edge case: URL generation for different repos
func Test_VIM_EdgeCase_URLGeneration_DifferentRepos(t *testing.T) {
	testCases := []struct {
		repo     string
		number   int
		expected string
	}{
		{"simonbrundin/ai", 1, "https://github.com/simonbrundin/ai/issues/1"},
		{"simonbrundin/other-repo", 42, "https://github.com/simonbrundin/other-repo/issues/42"},
		{"simonbrundin/my-project", 100, "https://github.com/simonbrundin/my-project/issues/100"},
	}

	for _, tc := range testCases {
		v := &vimNavigator{
			issues: []testIssue{
				{Number: tc.number, Title: "Test", Repo: tc.repo},
			},
			selectedIssue: 0,
			repo:          tc.repo,
		}

		v.openSelectedIssue()

		assert.Equal(t, tc.expected, v.issueURL,
			"URL should be generated correctly for repo %s", tc.repo)
	}
}
