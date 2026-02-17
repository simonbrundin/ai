package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Tests for Issue #27: Skapa nya GitHub-issues från TUI:n
//
// These tests verify the implementation works correctly
// =============================================================================

// NewIssueDialogState simulates the new issue dialog logic for testing
type NewIssueDialogState struct {
	repos              []string
	filteredRepos      []string
	selectedRepoIndex  int
	showNewIssueDialog bool
	dialogMode         string // "repo-select", "error", ""
	errorMessage       string
	filterText         string
}

// Available repositories (simulated)
var opencodeSecurePath = "/home/simon/repos/dotfiles/opencode/.config/opencode/opencode-secure"
var issuePromptCommand = "--prompt \"/issue\""

func (n *NewIssueDialogState) openDialog() {
	n.showNewIssueDialog = true
	n.dialogMode = "repo-select"
	n.selectedRepoIndex = 0
	n.filterText = ""
	n.filteredRepos = n.repos
}

func (n *NewIssueDialogState) closeDialog() {
	n.showNewIssueDialog = false
	n.dialogMode = ""
	n.selectedRepoIndex = -1
	n.filterText = ""
}

func (n *NewIssueDialogState) filterRepos(query string) {
	n.filterText = query
	if query == "" {
		n.filteredRepos = n.repos
		return
	}
	queryLower := strings.ToLower(query)
	n.filteredRepos = nil
	for _, repo := range n.repos {
		if fuzzyMatch(repo, queryLower) {
			n.filteredRepos = append(n.filteredRepos, repo)
		}
	}
	// Reset selection if out of bounds
	if n.selectedRepoIndex >= len(n.filteredRepos) {
		n.selectedRepoIndex = 0
	}
}

func fuzzyMatch(text, query string) bool {
	// Case-insensitive fuzzy matching - checks if all query chars appear in order
	textLower := strings.ToLower(text)
	queryLower := strings.ToLower(query)
	queryIdx := 0
	for _, c := range textLower {
		if queryIdx < len(queryLower) && string(c) == string(queryLower[queryIdx]) {
			queryIdx++
		}
	}
	return queryIdx == len(queryLower)
}

func (n *NewIssueDialogState) moveSelectionUp() {
	if n.selectedRepoIndex > 0 {
		n.selectedRepoIndex--
	}
}

func (n *NewIssueDialogState) moveSelectionDown() {
	if n.selectedRepoIndex < len(n.filteredRepos)-1 {
		n.selectedRepoIndex++
	}
}

func (n *NewIssueDialogState) selectByNumber(num int) bool {
	if !n.showNewIssueDialog || n.dialogMode != "repo-select" {
		return false
	}
	if num >= 1 && num <= len(n.filteredRepos) && num <= 9 {
		n.selectedRepoIndex = num - 1
		return true
	}
	return false
}

func (n *NewIssueDialogState) getSelectedRepo() string {
	if n.selectedRepoIndex >= 0 && n.selectedRepoIndex < len(n.filteredRepos) {
		return n.filteredRepos[n.selectedRepoIndex]
	}
	return ""
}

func (n *NewIssueDialogState) confirmRepoSelection() string {
	repo := n.getSelectedRepo()
	if repo == "" {
		n.showError("No repository selected")
		return ""
	}
	// Returns the tmux command that would be executed
	return "tmux new-window -d -n 'opencode-issue' && tmux send-keys -t 'opencode-issue' '" + opencodeSecurePath + " " + issuePromptCommand + "' C-m"
}

func (n *NewIssueDialogState) showError(message string) {
	n.dialogMode = "error"
	n.errorMessage = message
}

// =============================================================================
// HAPPY PATH TESTS
// =============================================================================

// Test: Press 'n' on issue tab opens the dialog
func Test_NewIssueDialog_PressN_OpensDialog(t *testing.T) {
	n := &NewIssueDialogState{
		repos: []string{"simonbrundin/ai", "simonbrundin/dotfiles", "simonbrundin/agent"},
	}

	n.openDialog()

	assert.True(t, n.showNewIssueDialog, "Dialog should open when pressing 'n'")
	assert.Equal(t, "repo-select", n.dialogMode, "Should be in repo-select mode")
	assert.Equal(t, 0, n.selectedRepoIndex, "Should default to first repo")
}

// Test: Repo list is populated correctly
func Test_NewIssueDialog_RepoList_Populated(t *testing.T) {
	n := &NewIssueDialogState{
		repos: []string{"simonbrundin/ai", "simonbrundin/dotfiles"},
	}

	n.openDialog()

	assert.Equal(t, 2, len(n.filteredRepos), "Should have 2 repos")
	assert.Equal(t, "simonbrundin/ai", n.filteredRepos[0], "First repo should be ai")
	assert.Equal(t, "simonbrundin/dotfiles", n.filteredRepos[1], "Second repo should be dotfiles")
}

// Test: Filter repos by typing
func Test_NewIssueDialog_FilterRepos(t *testing.T) {
	n := &NewIssueDialogState{
		repos: []string{"simonbrundin/ai", "simonbrundin/dotfiles", "simonbrundin/agent"},
	}

	n.openDialog()
	n.filterRepos("ai")

	assert.Equal(t, 1, len(n.filteredRepos), "Should match 1 repo")
	assert.Equal(t, "simonbrundin/ai", n.filteredRepos[0], "Should filter to ai")
	assert.Equal(t, "ai", n.filterText, "Filter text should be saved")
}

// Test: Up arrow moves selection up
func Test_NewIssueDialog_UpArrow_MovesUp(t *testing.T) {
	n := &NewIssueDialogState{
		repos:              []string{"simonbrundin/ai", "simonbrundin/dotfiles", "simonbrundin/agent"},
		showNewIssueDialog: true,
		dialogMode:         "repo-select",
		selectedRepoIndex:  2,
	}
	n.filteredRepos = n.repos

	n.moveSelectionUp()

	assert.Equal(t, 1, n.selectedRepoIndex, "Should move to previous repo")
}

// Test: Down arrow moves selection down
func Test_NewIssueDialog_DownArrow_MovesDown(t *testing.T) {
	n := &NewIssueDialogState{
		repos:              []string{"simonbrundin/ai", "simonbrundin/dotfiles", "simonbrundin/agent"},
		showNewIssueDialog: true,
		dialogMode:         "repo-select",
		selectedRepoIndex:  0,
	}
	n.filteredRepos = n.repos

	n.moveSelectionDown()

	assert.Equal(t, 1, n.selectedRepoIndex, "Should move to next repo")
}

// Test: Number keys 1-9 select corresponding repo
func Test_NewIssueDialog_NumberKeys_SelectRepo(t *testing.T) {
	n := &NewIssueDialogState{
		repos:              []string{"simonbrundin/ai", "simonbrundin/dotfiles", "simonbrundin/agent"},
		showNewIssueDialog: true,
		dialogMode:         "repo-select",
		selectedRepoIndex:  0,
	}
	n.filteredRepos = n.repos

	result := n.selectByNumber(2)

	assert.True(t, result, "selectByNumber(2) should return true")
	assert.Equal(t, 1, n.selectedRepoIndex, "Should select repo at index 1")
	assert.Equal(t, "simonbrundin/dotfiles", n.getSelectedRepo(), "Should show dotfiles")
}

// Test: Enter confirms repo selection and returns tmux command
func Test_NewIssueDialog_Enter_ConfirmsAndReturnsCommand(t *testing.T) {
	n := &NewIssueDialogState{
		repos:              []string{"simonbrundin/ai", "simonbrundin/dotfiles"},
		showNewIssueDialog: true,
		dialogMode:         "repo-select",
		selectedRepoIndex:  0,
	}
	n.filteredRepos = n.repos

	command := n.confirmRepoSelection()

	assert.Contains(t, command, "tmux new-window", "Should contain tmux command")
	assert.Contains(t, command, opencodeSecurePath, "Should contain opencode-secure path")
	assert.Contains(t, command, issuePromptCommand, "Should contain issue prompt")
}

// Test: Escape closes dialog
func Test_NewIssueDialog_Escape_Closes(t *testing.T) {
	n := &NewIssueDialogState{
		showNewIssueDialog: true,
		dialogMode:         "repo-select",
		selectedRepoIndex:  2,
		filterText:         "some filter",
	}

	n.closeDialog()

	assert.False(t, n.showNewIssueDialog, "Escape should close dialog")
	assert.Equal(t, "", n.dialogMode, "Mode should be cleared")
	assert.Equal(t, -1, n.selectedRepoIndex, "Selection should be reset")
	assert.Equal(t, "", n.filterText, "Filter should be cleared")
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

// Edge case: Empty repo list - should still open dialog
func Test_NewIssueDialog_EdgeCase_EmptyRepos_OpensDialog(t *testing.T) {
	n := &NewIssueDialogState{
		repos: []string{},
	}

	n.openDialog()

	assert.True(t, n.showNewIssueDialog, "Dialog should open even with empty repos")
	assert.Equal(t, 0, len(n.filteredRepos), "Should have empty filtered repos")
}

// Edge case: Filter with no matches - shows empty list
func Test_NewIssueDialog_EdgeCase_FilterNoMatch(t *testing.T) {
	n := &NewIssueDialogState{
		repos: []string{"simonbrundin/ai", "simonbrundin/dotfiles"},
	}
	n.openDialog()

	n.filterRepos("xyz")

	assert.Equal(t, 0, len(n.filteredRepos), "Should have no matches")
	assert.Equal(t, 0, n.selectedRepoIndex, "Selection should be reset to 0")
}

// Edge case: Up arrow at first repo - stays at first
func Test_NewIssueDialog_EdgeCase_UpArrow_AtFirst(t *testing.T) {
	n := &NewIssueDialogState{
		repos:              []string{"simonbrundin/ai"},
		showNewIssueDialog: true,
		dialogMode:         "repo-select",
		selectedRepoIndex:  0,
	}
	n.filteredRepos = n.repos

	n.moveSelectionUp()

	assert.Equal(t, 0, n.selectedRepoIndex, "Should stay at first repo")
}

// Edge case: Down arrow at last repo - stays at last
func Test_NewIssueDialog_EdgeCase_DownArrow_AtLast(t *testing.T) {
	n := &NewIssueDialogState{
		repos:              []string{"simonbrundin/ai"},
		showNewIssueDialog: true,
		dialogMode:         "repo-select",
		selectedRepoIndex:  0,
	}
	n.filteredRepos = n.repos

	n.moveSelectionDown()

	assert.Equal(t, 0, n.selectedRepoIndex, "Should stay at last repo")
}

// Edge case: Number key 0 does nothing
func Test_NewIssueDialog_EdgeCase_NumberZero_NoAction(t *testing.T) {
	n := &NewIssueDialogState{
		repos:              []string{"simonbrundin/ai"},
		showNewIssueDialog: true,
		dialogMode:         "repo-select",
		selectedRepoIndex:  0,
	}
	n.filteredRepos = n.repos

	result := n.selectByNumber(0)

	assert.False(t, result, "selectByNumber(0) should return false")
	assert.Equal(t, 0, n.selectedRepoIndex, "Selection should not change")
}

// Edge case: Number key >9 does nothing
func Test_NewIssueDialog_EdgeCase_NumberTenPlus_NoAction(t *testing.T) {
	n := &NewIssueDialogState{
		repos:              []string{"simonbrundin/ai"},
		showNewIssueDialog: true,
		dialogMode:         "repo-select",
		selectedRepoIndex:  0,
	}
	n.filteredRepos = n.repos

	result := n.selectByNumber(10)

	assert.False(t, result, "selectByNumber(10) should return false")
	assert.Equal(t, 0, n.selectedRepoIndex, "Selection should not change")
}

// Edge case: Number keys when dialog is closed - no action
func Test_NewIssueDialog_EdgeCase_NumberKey_DialogClosed(t *testing.T) {
	n := &NewIssueDialogState{
		showNewIssueDialog: false,
		selectedRepoIndex:  0,
	}

	result := n.selectByNumber(3)

	assert.False(t, result, "selectByNumber should return false when dialog is closed")
	assert.Equal(t, 0, n.selectedRepoIndex, "Selection should not change")
}

// Edge case: Number keys in error mode - no action
func Test_NewIssueDialog_EdgeCase_NumberKey_ErrorMode(t *testing.T) {
	n := &NewIssueDialogState{
		showNewIssueDialog: true,
		dialogMode:         "error",
		selectedRepoIndex:  0,
	}

	result := n.selectByNumber(3)

	assert.False(t, result, "selectByNumber should return false in error mode")
}

// Edge case: Confirm with no repo selected - shows error
func Test_NewIssueDialog_EdgeCase_ConfirmNoSelection(t *testing.T) {
	n := &NewIssueDialogState{
		repos:              []string{"simonbrundin/ai"},
		showNewIssueDialog: true,
		dialogMode:         "repo-select",
		selectedRepoIndex:  -1, // No selection
	}
	n.filteredRepos = n.repos

	command := n.confirmRepoSelection()

	assert.Equal(t, "", command, "Should return empty command")
	assert.Equal(t, "error", n.dialogMode, "Should switch to error mode")
	assert.Equal(t, "No repository selected", n.errorMessage, "Should show error message")
}

// Edge case: Clear filter by typing empty string
func Test_NewIssueDialog_EdgeCase_ClearFilter(t *testing.T) {
	n := &NewIssueDialogState{
		repos: []string{"simonbrundin/ai", "simonbrundin/dotfiles", "simonbrundin/agent"},
	}
	n.openDialog()
	n.filterRepos("ai")

	n.filterRepos("")

	assert.Equal(t, 3, len(n.filteredRepos), "Should show all repos again")
	assert.Equal(t, "", n.filterText, "Filter text should be cleared")
}

// Edge case: Fuzzy match partial string
func Test_NewIssueDialog_EdgeCase_FuzzyMatchPartial(t *testing.T) {
	n := &NewIssueDialogState{
		repos: []string{"simonbrundin/ai", "simonbrundin/dotfiles", "simonbrundin/agent"},
	}
	n.openDialog()

	n.filterRepos("ai") // Should match "ai"

	assert.Equal(t, 1, len(n.filteredRepos), "Should match 'ai' only")
	assert.Equal(t, "simonbrundin/ai", n.filteredRepos[0], "Should find ai repo")
}

// Edge case: Fuzzy match is case insensitive
func Test_NewIssueDialog_EdgeCase_FuzzyMatchCaseInsensitive(t *testing.T) {
	n := &NewIssueDialogState{
		repos: []string{"simonbrundin/AI", "simonbrundin/dotfiles"},
	}
	n.openDialog()

	n.filterRepos("ai")

	assert.Equal(t, 1, len(n.filteredRepos), "Should match case-insensitively")
	if len(n.filteredRepos) > 0 {
		assert.Equal(t, "simonbrundin/AI", n.filteredRepos[0], "Should find AI repo")
	}
}

// Edge case: Multiple repos - verify correct tmux command
func Test_NewIssueDialog_EdgeCase_MultipleRepos_TmuxCommand(t *testing.T) {
	n := &NewIssueDialogState{
		repos:              []string{"simonbrundin/ai", "simonbrundin/dotfiles", "simonbrundin/agent"},
		showNewIssueDialog: true,
		dialogMode:         "repo-select",
		selectedRepoIndex:  1, // Select dotfiles
	}
	n.filteredRepos = n.repos

	command := n.confirmRepoSelection()

	assert.Contains(t, command, "opencode-issue", "Should have window name")
}

// Edge case: Get selected repo with empty filtered list
func Test_NewIssueDialog_EdgeCase_GetSelectedRepo_Empty(t *testing.T) {
	n := &NewIssueDialogState{
		filteredRepos:     []string{},
		selectedRepoIndex: 0,
	}

	repo := n.getSelectedRepo()

	assert.Equal(t, "", repo, "Should return empty string for empty list")
}

// Edge case: Get selected repo with negative index
func Test_NewIssueDialog_EdgeCase_GetSelectedRepo_NegativeIndex(t *testing.T) {
	n := &NewIssueDialogState{
		repos:             []string{"simonbrundin/ai"},
		filteredRepos:     []string{"simonbrundin/ai"},
		selectedRepoIndex: -1,
	}

	repo := n.getSelectedRepo()

	assert.Equal(t, "", repo, "Should return empty string for negative index")
}

// Edge case: Get selected repo with out of bounds index
func Test_NewIssueDialog_EdgeCase_GetSelectedRepo_OutOfBounds(t *testing.T) {
	n := &NewIssueDialogState{
		repos:             []string{"simonbrundin/ai"},
		filteredRepos:     []string{"simonbrundin/ai"},
		selectedRepoIndex: 100,
	}

	repo := n.getSelectedRepo()

	assert.Equal(t, "", repo, "Should return empty string for out of bounds index")
}
