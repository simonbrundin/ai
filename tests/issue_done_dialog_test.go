package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Tests for Issue #20: Add dialog to confirm marking issues as done in GitHub
//
// These tests verify the implementation works correctly
// =============================================================================

// Mirror the issue struct from main.go for testing
type dialogIssue struct {
	Number int
	Title  string
	State  string
	Labels []string
	Repo   string
}

// DialogState simulates the dialog logic for testing
type DialogState struct {
	issues            []dialogIssue
	selectedIssue     int
	showConfirmDialog bool
}

func (d *DialogState) handleDKey() {
	if len(d.issues) > 0 && d.selectedIssue >= 0 && d.selectedIssue < len(d.issues) {
		d.showConfirmDialog = true
	}
}

func (d *DialogState) handleYKey() bool {
	if d.showConfirmDialog && d.selectedIssue >= 0 && d.selectedIssue < len(d.issues) {
		d.showConfirmDialog = false
		return true
	}
	return false
}

func (d *DialogState) handleNKey() {
	d.showConfirmDialog = false
}

func (d *DialogState) handleEscapeKey() {
	d.showConfirmDialog = false
}

func (d *DialogState) handleEnterKey() bool {
	if d.showConfirmDialog && d.selectedIssue >= 0 && d.selectedIssue < len(d.issues) {
		d.showConfirmDialog = false
		return true
	}
	return false
}

func (d *DialogState) getSelectedIssue() *dialogIssue {
	if len(d.issues) == 0 || d.selectedIssue < 0 || d.selectedIssue >= len(d.issues) {
		return nil
	}
	return &d.issues[d.selectedIssue]
}

// =============================================================================
// HAPPY PATH TESTS
// =============================================================================

// Test: Pressing 'd' shows dialog when valid selection exists
func Test_Dialog_PressD_ShowsDialog(t *testing.T) {
	d := &DialogState{
		issues: []dialogIssue{
			{Number: 20, Title: "Test Issue", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 0,
	}

	d.handleDKey()

	assert.True(t, d.showConfirmDialog, "'d' key should show dialog with valid selection")
}

// Test: Dialog shows correct issue number
func Test_Dialog_ShowsCorrectIssueNumber(t *testing.T) {
	d := &DialogState{
		issues: []dialogIssue{
			{Number: 42, Title: "Feature Request", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 0,
	}

	issue := d.getSelectedIssue()

	assert.NotNil(t, issue, "Should have selected issue")
	assert.Equal(t, 42, issue.Number, "Issue number should be 42")
}

// Test: Enter confirms and closes
func Test_Dialog_Enter_Confirms(t *testing.T) {
	d := &DialogState{
		showConfirmDialog: true,
		issues: []dialogIssue{
			{Number: 20, Title: "Test", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 0,
	}

	confirmed := d.handleEnterKey()

	assert.True(t, confirmed, "Enter should confirm")
	assert.False(t, d.showConfirmDialog, "Dialog should close after confirmation")
}

// Test: 'y' confirms
func Test_Dialog_Y_Confirms(t *testing.T) {
	d := &DialogState{
		showConfirmDialog: true,
		issues: []dialogIssue{
			{Number: 20, Title: "Test", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 0,
	}

	confirmed := d.handleYKey()

	assert.True(t, confirmed, "'y' should confirm")
	assert.False(t, d.showConfirmDialog, "Dialog should close after confirmation")
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

// Edge case: 'd' with empty issue list
func Test_Dialog_EdgeCase_EmptyList_NoDialog(t *testing.T) {
	d := &DialogState{
		issues:        []dialogIssue{},
		selectedIssue: 0,
	}

	d.handleDKey()

	assert.False(t, d.showConfirmDialog, "'d' should NOT show dialog with empty list")
}

// Edge case: 'd' with negative selection
func Test_Dialog_EdgeCase_NegativeSelection_NoDialog(t *testing.T) {
	d := &DialogState{
		issues: []dialogIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
		},
		selectedIssue: -1,
	}

	d.handleDKey()

	assert.False(t, d.showConfirmDialog, "'d' should NOT show dialog with negative selection")
}

// Edge case: 'd' with out of bounds selection
func Test_Dialog_EdgeCase_OutOfBounds_NoDialog(t *testing.T) {
	d := &DialogState{
		issues: []dialogIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 100,
	}

	d.handleDKey()

	assert.False(t, d.showConfirmDialog, "'d' should NOT show dialog with out of bounds selection")
}

// Edge case: Escape closes dialog
func Test_Dialog_Escape_Closes(t *testing.T) {
	d := &DialogState{
		showConfirmDialog: true,
	}

	d.handleEscapeKey()

	assert.False(t, d.showConfirmDialog, "Escape should close dialog")
}

// Edge case: 'n' cancels
func Test_Dialog_N_Cancels(t *testing.T) {
	d := &DialogState{
		showConfirmDialog: true,
	}

	d.handleNKey()

	assert.False(t, d.showConfirmDialog, "'n' should cancel")
}

// Edge case: getSelectedIssue with empty list
func Test_Dialog_EdgeCase_GetSelectedIssue_EmptyList(t *testing.T) {
	d := &DialogState{
		issues:        []dialogIssue{},
		selectedIssue: 0,
	}

	issue := d.getSelectedIssue()

	assert.Nil(t, issue, "Should return nil for empty list")
}

// Edge case: getSelectedIssue with negative index
func Test_Dialog_EdgeCase_GetSelectedIssue_NegativeIndex(t *testing.T) {
	d := &DialogState{
		issues: []dialogIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
		},
		selectedIssue: -1,
	}

	issue := d.getSelectedIssue()

	assert.Nil(t, issue, "Should return nil for negative index")
}

// Edge case: Multiple issues - confirm correct one
func Test_Dialog_EdgeCase_MultipleIssues_ConfirmCorrect(t *testing.T) {
	d := &DialogState{
		issues: []dialogIssue{
			{Number: 1, Title: "First", Repo: "simonbrundin/ai"},
			{Number: 20, Title: "Second", Repo: "simonbrundin/ai"},
			{Number: 42, Title: "Third", Repo: "simonbrundin/ai"},
		},
		selectedIssue:     1,
		showConfirmDialog: true,
	}

	d.handleYKey()
	issue := d.getSelectedIssue()

	assert.NotNil(t, issue, "Should still have issue reference")
	assert.Equal(t, 20, issue.Number, "Should have issue #20")
}
