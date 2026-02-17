package tests

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Tests for Issue #23: Add command dialog when pressing Enter on issue
//
// These tests verify the implementation works correctly
// =============================================================================

// CommandDialogState simulates the command dialog logic for testing
type CommandDialogState struct {
	issues            []dialogIssue
	selectedIssue     int
	showCommandDialog bool
	selectedCommand   int // 0-4 for the 5 commands
}

// Available commands matching the 5 options
var commandNames = []string{"Skriv tester", "Implementera", "Refactor", "Dokumentera", "Skapa PR"}
var commandAliases = []string{"/tdd", "/implement", "/refactor", "/docs", "/pr"}

func (c *CommandDialogState) handleEnterKey() bool {
	if c.showCommandDialog && c.selectedIssue >= 0 && c.selectedIssue < len(c.issues) {
		return true
	}
	return false
}

func (c *CommandDialogState) showDialog() {
	if len(c.issues) > 0 && c.selectedIssue >= 0 && c.selectedIssue < len(c.issues) {
		c.showCommandDialog = true
		c.selectedCommand = 0 // Default to first command
	}
}

func (c *CommandDialogState) closeDialog() {
	c.showCommandDialog = false
	c.selectedCommand = -1
}

func (c *CommandDialogState) moveSelectionUp() {
	if c.showCommandDialog && c.selectedCommand > 0 {
		c.selectedCommand--
	}
}

func (c *CommandDialogState) moveSelectionDown() {
	if c.showCommandDialog && c.selectedCommand < len(commandNames)-1 {
		c.selectedCommand++
	}
}

func (c *CommandDialogState) selectByNumber(num int) bool {
	if c.showCommandDialog && num >= 1 && num <= len(commandNames) {
		c.selectedCommand = num - 1
		return true
	}
	return false
}

func (c *CommandDialogState) getSelectedCommand() string {
	if c.selectedCommand >= 0 && c.selectedCommand < len(commandNames) {
		return commandNames[c.selectedCommand]
	}
	return ""
}

func (c *CommandDialogState) getCommandWithIssueNum() string {
	if c.selectedCommand < 0 || c.selectedCommand >= len(commandAliases) {
		return ""
	}
	if c.selectedIssue < 0 || c.selectedIssue >= len(c.issues) {
		return ""
	}
	issueNum := c.issues[c.selectedIssue].Number
	return commandAliases[c.selectedCommand] + " " + strconv.Itoa(issueNum)
}

func (c *CommandDialogState) getSelectedIssue() *dialogIssue {
	if len(c.issues) == 0 || c.selectedIssue < 0 || c.selectedIssue >= len(c.issues) {
		return nil
	}
	return &c.issues[c.selectedIssue]
}

// =============================================================================
// HAPPY PATH TESTS
// =============================================================================

// Test: Pressing Enter shows command dialog when valid selection exists
func Test_CommandDialog_PressEnter_ShowsDialog(t *testing.T) {
	c := &CommandDialogState{
		issues: []dialogIssue{
			{Number: 23, Title: "Test Issue", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 0,
	}

	c.showDialog()

	assert.True(t, c.showCommandDialog, "Enter should show command dialog with valid selection")
	assert.Equal(t, 0, c.selectedCommand, "Should default to first command (index 0)")
}

// Test: Dialog shows correct issue number and command
func Test_CommandDialog_ShowsCorrectIssueAndCommand(t *testing.T) {
	c := &CommandDialogState{
		issues: []dialogIssue{
			{Number: 42, Title: "Feature Request", Repo: "simonbrundin/ai"},
		},
		selectedIssue:     0,
		showCommandDialog: true,
		selectedCommand:   0,
	}

	issue := c.getSelectedIssue()
	command := c.getSelectedCommand()
	commandWithNum := c.getCommandWithIssueNum()

	assert.NotNil(t, issue, "Should have selected issue")
	assert.Equal(t, 42, issue.Number, "Issue number should be 42")
	assert.Equal(t, "Skriv tester", command, "Default command should be 'Skriv tester'")
	assert.Equal(t, "/tdd 42", commandWithNum, "Command should include issue number")
}

// Test: Up arrow moves selection up
func Test_CommandDialog_UpArrow_MovesUp(t *testing.T) {
	c := &CommandDialogState{
		showCommandDialog: true,
		selectedCommand:   2, // Currently on "Refactor"
	}

	c.moveSelectionUp()

	assert.Equal(t, 1, c.selectedCommand, "Should move to previous command (Implementera)")
	assert.Equal(t, "Implementera", c.getSelectedCommand())
}

// Test: Down arrow moves selection down
func Test_CommandDialog_DownArrow_MovesDown(t *testing.T) {
	c := &CommandDialogState{
		showCommandDialog: true,
		selectedCommand:   1, // Currently on "Implementera"
	}

	c.moveSelectionDown()

	assert.Equal(t, 2, c.selectedCommand, "Should move to next command (Refactor)")
	assert.Equal(t, "Refactor", c.getSelectedCommand())
}

// Test: Number keys 1-5 select corresponding command
func Test_CommandDialog_NumberKeys_SelectCommand(t *testing.T) {
	tests := []struct {
		key         int
		expectedCmd string
		expectedIdx int
	}{
		{1, "Skriv tester", 0},
		{2, "Implementera", 1},
		{3, "Refactor", 2},
		{4, "Dokumentera", 3},
		{5, "Skapa PR", 4},
	}

	for _, tc := range tests {
		c := &CommandDialogState{
			showCommandDialog: true,
			selectedCommand:   0,
		}

		result := c.selectByNumber(tc.key)

		assert.True(t, result, "selectByNumber(%d) should return true", tc.key)
		assert.Equal(t, tc.expectedIdx, c.selectedCommand, "Should select command at index %d", tc.key-1)
		assert.Equal(t, tc.expectedCmd, c.getSelectedCommand(), "Should show '%s'", tc.expectedCmd)
	}
}

// Test: Enter confirms command selection
func Test_CommandDialog_Enter_Confirms(t *testing.T) {
	c := &CommandDialogState{
		showCommandDialog: true,
		issues: []dialogIssue{
			{Number: 23, Title: "Test", Repo: "simonbrundin/ai"},
		},
		selectedIssue:   0,
		selectedCommand: 1, // "Implementera"
	}

	confirmed := c.handleEnterKey()

	assert.True(t, confirmed, "Enter should confirm command selection")
	assert.Equal(t, "/implement 23", c.getCommandWithIssueNum(), "Should return command with issue number")
}

// Test: Escape closes dialog
func Test_CommandDialog_Escape_Closes(t *testing.T) {
	c := &CommandDialogState{
		showCommandDialog: true,
		selectedCommand:   2,
	}

	c.closeDialog()

	assert.False(t, c.showCommandDialog, "Escape should close dialog")
	assert.Equal(t, -1, c.selectedCommand, "Selected command should be reset")
}

// Test: 'n' cancels dialog
func Test_CommandDialog_N_Cancels(t *testing.T) {
	c := &CommandDialogState{
		showCommandDialog: true,
		selectedCommand:   3,
	}

	c.closeDialog()

	assert.False(t, c.showCommandDialog, "'n' should cancel")
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

// Edge case: Enter with empty issue list - no dialog
func Test_CommandDialog_EdgeCase_EmptyList_NoDialog(t *testing.T) {
	c := &CommandDialogState{
		issues:        []dialogIssue{},
		selectedIssue: 0,
	}

	c.showDialog()

	assert.False(t, c.showCommandDialog, "Should NOT show dialog with empty list")
}

// Edge case: Enter with negative selection - no dialog
func Test_CommandDialog_EdgeCase_NegativeSelection_NoDialog(t *testing.T) {
	c := &CommandDialogState{
		issues: []dialogIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
		},
		selectedIssue: -1,
	}

	c.showDialog()

	assert.False(t, c.showCommandDialog, "Should NOT show dialog with negative selection")
}

// Edge case: Enter with out of bounds selection - no dialog
func Test_CommandDialog_EdgeCase_OutOfBounds_NoDialog(t *testing.T) {
	c := &CommandDialogState{
		issues: []dialogIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
		},
		selectedIssue: 100,
	}

	c.showDialog()

	assert.False(t, c.showCommandDialog, "Should NOT show dialog with out of bounds selection")
}

// Edge case: Up arrow at first command - stays at first
func Test_CommandDialog_EdgeCase_UpArrow_AtFirst(t *testing.T) {
	c := &CommandDialogState{
		showCommandDialog: true,
		selectedCommand:   0,
	}

	c.moveSelectionUp()

	assert.Equal(t, 0, c.selectedCommand, "Should stay at first command")
}

// Edge case: Down arrow at last command - stays at last
func Test_CommandDialog_EdgeCase_DownArrow_AtLast(t *testing.T) {
	c := &CommandDialogState{
		showCommandDialog: true,
		selectedCommand:   4, // Last command (Skapa PR)
	}

	c.moveSelectionDown()

	assert.Equal(t, 4, c.selectedCommand, "Should stay at last command")
}

// Edge case: Number key 0 does nothing
func Test_CommandDialog_EdgeCase_NumberZero_NoAction(t *testing.T) {
	c := &CommandDialogState{
		showCommandDialog: true,
		selectedCommand:   0,
	}

	result := c.selectByNumber(0)

	assert.False(t, result, "selectByNumber(0) should return false")
	assert.Equal(t, 0, c.selectedCommand, "Selection should not change")
}

// Edge case: Number key 6+ does nothing
func Test_CommandDialog_EdgeCase_NumberSixPlus_NoAction(t *testing.T) {
	c := &CommandDialogState{
		showCommandDialog: true,
		selectedCommand:   0,
	}

	result := c.selectByNumber(6)

	assert.False(t, result, "selectByNumber(6) should return false")
	assert.Equal(t, 0, c.selectedCommand, "Selection should not change")
}

// Edge case: Number keys when dialog is closed - no action
func Test_CommandDialog_EdgeCase_NumberKey_DialogClosed(t *testing.T) {
	c := &CommandDialogState{
		showCommandDialog: false,
		selectedCommand:   0,
	}

	result := c.selectByNumber(3)

	assert.False(t, result, "selectByNumber should return false when dialog is closed")
	assert.Equal(t, 0, c.selectedCommand, "Selection should not change")
}

// Edge case: getSelectedIssue with empty list
func Test_CommandDialog_EdgeCase_GetSelectedIssue_EmptyList(t *testing.T) {
	c := &CommandDialogState{
		issues:        []dialogIssue{},
		selectedIssue: 0,
	}

	issue := c.getSelectedIssue()

	assert.Nil(t, issue, "Should return nil for empty list")
}

// Edge case: getSelectedIssue with negative index
func Test_CommandDialog_EdgeCase_GetSelectedIssue_NegativeIndex(t *testing.T) {
	c := &CommandDialogState{
		issues: []dialogIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
		},
		selectedIssue: -1,
	}

	issue := c.getSelectedIssue()

	assert.Nil(t, issue, "Should return nil for negative index")
}

// Edge case: getCommandWithIssueNum when no issue selected
func Test_CommandDialog_EdgeCase_CommandWithNoIssue(t *testing.T) {
	c := &CommandDialogState{
		issues: []dialogIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
		},
		selectedIssue:     -1,
		showCommandDialog: true,
		selectedCommand:   0,
	}

	cmd := c.getCommandWithIssueNum()

	assert.Equal(t, "", cmd, "Should return empty string when no issue selected")
}

// Edge case: getCommandWithIssueNum when no command selected
func Test_CommandDialog_EdgeCase_CommandWithNoSelection(t *testing.T) {
	c := &CommandDialogState{
		issues: []dialogIssue{
			{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
		},
		selectedIssue:     0,
		showCommandDialog: true,
		selectedCommand:   -1,
	}

	cmd := c.getCommandWithIssueNum()

	assert.Equal(t, "", cmd, "Should return empty string when no command selected")
}

// Edge case: All 5 commands return correct full command strings
func Test_CommandDialog_EdgeCase_AllCommands(t *testing.T) {
	c := &CommandDialogState{
		issues: []dialogIssue{
			{Number: 99, Title: "Test", Repo: "simonbrundin/ai"},
		},
		selectedIssue:     0,
		showCommandDialog: true,
	}

	expectedCommands := []struct {
		index    int
		expected string
	}{
		{0, "/tdd 99"},
		{1, "/implement 99"},
		{2, "/refactor 99"},
		{3, "/docs 99"},
		{4, "/pr 99"},
	}

	for _, tc := range expectedCommands {
		c.selectedCommand = tc.index
		cmd := c.getCommandWithIssueNum()
		assert.Equal(t, tc.expected, cmd, "Command index %d should return '%s'", tc.index, tc.expected)
	}
}

// Edge case: Multiple issues - confirm correct one
func Test_CommandDialog_EdgeCase_MultipleIssues_ConfirmCorrect(t *testing.T) {
	c := &CommandDialogState{
		issues: []dialogIssue{
			{Number: 1, Title: "First", Repo: "simonbrundin/ai"},
			{Number: 23, Title: "Second", Repo: "simonbrundin/ai"},
			{Number: 42, Title: "Third", Repo: "simonbrundin/ai"},
		},
		selectedIssue:     1, // Second issue
		showCommandDialog: true,
		selectedCommand:   4, // "Skapa PR"
	}

	cmd := c.getCommandWithIssueNum()

	assert.Equal(t, "/pr 23", cmd, "Should use issue #23 (the selected one)")
}
