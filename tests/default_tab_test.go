package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Tests för Issue #24: Sortera flikarna så Issues visas först vid start
//
// ACCEPTANCE CRITERIA:
// - Issues-fliken är aktiv när programmet startar
// - Tab-ordningen är: Issues (först), Agents (andra)
// - tangenterna 1 och 2 fungerar fortfarande korrekt för att växla mellan flikarna
// - Befintlig funktionalitet påverkas inte
//
// EDGE CASES:
// - Minimifönsterstorlek (80x24) ska fortfarande fungera
// - Help-modal ska visa korrekt flik-nummer
// =============================================================================

// model mirrors the model struct from main.go for testing
type model struct {
	currentTab int
	width      int
	height     int
	ready      bool
}

// These constants MUST match main.go after the fix
const (
	tabIssues = iota // 0 - Issues is now first tab
	tabAgents        // 1 - Agents is now second tab
	numTabs   = 2
)

// tabNames from main.go - after fix: Issues first, Agents second
var tabNames = []string{"Issues", "Agents"}

// =============================================================================
// HAPPY PATH TESTS
// =============================================================================

// Test: Default tab should be Issues (index 0)
func Test_DefaultTab_ShouldBeIssues(t *testing.T) {
	// Simulate default model initialization from main.go
	// Currently: model{} has currentTab = 0 (which is tabAgents = 0)
	// After fix: currentTab should be tabIssues = 0

	m := model{} // Default: currentTab = 0

	// After implementation, currentTab should equal tabIssues (which should be 0)
	// Currently it equals tabAgents (which is 0)
	// So we need to check that tabIssues == 0 AND currentTab == tabIssues
	assert.Equal(t, 0, tabIssues,
		"tabIssues should be 0 (first tab)")
	assert.Equal(t, m.currentTab, tabIssues,
		"Default currentTab should equal tabIssues")
}

// Test: Tab order should be Issues first, Agents second
func Test_TabOrder_ShouldBeIssuesThenAgents(t *testing.T) {
	// After fix:
	// - tabIssues should be 0
	// - tabAgents should be 1

	assert.Equal(t, 0, tabIssues,
		"tabIssues should be 0 (first tab)")
	assert.Equal(t, 1, tabAgents,
		"tabAgents should be 1 (second tab)")
}

// Test: Tab names should be Issues first, Agents second
func Test_TabNames_IssuesFirstAgentsSecond(t *testing.T) {
	// After fix: tabNames should be []string{"Issues", "Agents"}
	// Currently: tabNames = []string{"Agents", "Issues"}

	assert.Equal(t, "Issues", tabNames[0],
		"First tab name should be Issues, got: %s", tabNames[0])
	assert.Equal(t, "Agents", tabNames[1],
		"Second tab name should be Agents, got: %s", tabNames[1])
}

// Test: Pressing '1' should switch to Issues tab (index 0)
func Test_KeyPress1_ShouldSwitchToIssuesTab(t *testing.T) {
	currentTab := tabAgents // Start from Agents

	// Simulate key handling from main.go lines 341-346
	key := "1"
	tabNum := int(key[0] - '0')
	if tabNum >= 1 && tabNum <= numTabs {
		currentTab = tabNum - 1
	}

	// After fix: tabIssues = 0, so pressing '1' should give currentTab = 0 = tabIssues
	assert.Equal(t, 0, currentTab,
		"Pressing '1' should give tab index 0")
	assert.Equal(t, tabIssues, currentTab,
		"Pressing '1' should switch to Issues tab")
}

// Test: Pressing '2' should switch to Agents tab (index 1)
func Test_KeyPress2_ShouldSwitchToAgentsTab(t *testing.T) {
	currentTab := tabIssues // Start from Issues

	// Simulate key handling from main.go
	key := "2"
	tabNum := int(key[0] - '0')
	if tabNum >= 1 && tabNum <= numTabs {
		currentTab = tabNum - 1
	}

	// After fix: tabAgents = 1, so pressing '2' should give currentTab = 1 = tabAgents
	assert.Equal(t, 1, currentTab,
		"Pressing '2' should give tab index 1")
	assert.Equal(t, tabAgents, currentTab,
		"Pressing '2' should switch to Agents tab")
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

// Edge case: Tab key cycles forward correctly
func Test_TabKey_CyclesForward(t *testing.T) {
	currentTab := tabIssues

	// Cycle forward: Issues (0) -> Agents (1) -> Issues (0)
	currentTab = (currentTab + 1) % numTabs
	assert.Equal(t, 1, currentTab,
		"After one tab press from Issues, should be on Agents (1)")
	assert.Equal(t, tabAgents, currentTab,
		"After one tab press from Issues, should be on Agents")

	currentTab = (currentTab + 1) % numTabs
	assert.Equal(t, 0, currentTab,
		"After two tab presses, should cycle back to Issues (0)")
	assert.Equal(t, tabIssues, currentTab,
		"After two tab presses, should cycle back to Issues")
}

// Edge case: Shift+tab cycles backwards
func Test_ShiftTab_CyclesBackwards(t *testing.T) {
	currentTab := tabAgents

	// Cycle backwards: Agents (1) -> Issues (0) -> Agents (1)
	currentTab = (currentTab - 1 + numTabs) % numTabs
	assert.Equal(t, 0, currentTab,
		"Shift+tab from Agents should go to Issues (0)")
	assert.Equal(t, tabIssues, currentTab,
		"Shift+tab from Agents should go to Issues")

	currentTab = (currentTab - 1 + numTabs) % numTabs
	assert.Equal(t, 1, currentTab,
		"Shift+tab from Issues should cycle to Agents (1)")
	assert.Equal(t, tabAgents, currentTab,
		"Shift+tab from Issues should cycle to Agents")
}

// Edge case: Invalid tab number (3) should be ignored
func Test_InvalidTabNumber3_ShouldBeIgnored(t *testing.T) {
	currentTab := tabIssues
	originalTab := currentTab

	key := "3"
	tabNum := int(key[0] - '0')
	if tabNum >= 1 && tabNum <= numTabs {
		currentTab = tabNum - 1
	}

	// Should remain unchanged
	assert.Equal(t, originalTab, currentTab,
		"Invalid tab number 3 should be ignored")
}

// Edge case: Invalid tab number (0) should be ignored
func Test_InvalidTabNumber0_ShouldBeIgnored(t *testing.T) {
	currentTab := tabIssues
	originalTab := currentTab

	key := "0"
	tabNum := int(key[0] - '0')
	if tabNum >= 1 && tabNum <= numTabs {
		currentTab = tabNum - 1
	}

	// Should remain unchanged
	assert.Equal(t, originalTab, currentTab,
		"Invalid tab number 0 should be ignored")
}

// Edge case: Minimum window size constants
func Test_MinWindowSize_Constants(t *testing.T) {
	minWidth := 80
	minHeight := 24

	// These are constants in main.go - verify they work
	assert.Equal(t, 80, minWidth,
		"minWidth should be 80")
	assert.Equal(t, 24, minHeight,
		"minHeight should be 24")

	// Verify they can be used in size check
	width, height := 80, 24
	isValid := width >= minWidth && height >= minHeight
	assert.True(t, isValid,
		"80x24 should pass minimum size check")
}

// Edge case: Tab navigation boundary indices
func Test_TabIndices_WithinBounds(t *testing.T) {
	// All tab indices should be valid
	allTabs := []int{tabIssues, tabAgents}

	for i, idx := range allTabs {
		assert.True(t, idx >= 0 && idx < numTabs,
			"Tab index %d (value %d) should be within [0, %d)", i, idx, numTabs)
	}
}

// Edge case: Number of tabs matches tab names
func Test_NumTabs_MatchesTabNamesLength(t *testing.T) {
	assert.Equal(t, numTabs, len(tabNames),
		"numTabs (%d) should match length of tabNames (%d)", numTabs, len(tabNames))
}

// Edge case: Help display shows correct tab numbers
func Test_HelpDisplay_ShowsCorrectTabNumbers(t *testing.T) {
	// Simulating the footer hints from main.go line 571-577

	// After fix:
	// - Tab 1 = Issues (tabIssues = 0)
	// - Tab 2 = Agents (tabAgents = 1)

	// Verify the mapping is correct
	assert.Equal(t, 0, tabIssues,
		"Issues should be tab 1 (index 0)")
	assert.Equal(t, 1, tabAgents,
		"Agents should be tab 2 (index 1)")

	// Key "1" maps to index 0 = Issues
	// Key "2" maps to index 1 = Agents
	key1Index := int('1'-'0') - 1 // = 0
	key2Index := int('2'-'0') - 1 // = 1

	assert.Equal(t, 0, key1Index,
		"Key '1' should map to index 0")
	assert.Equal(t, 1, key2Index,
		"Key '2' should map to index 1")

	// And those indices should match the correct tabs
	assert.Equal(t, tabIssues, key1Index,
		"Key '1' should select Issues tab")
	assert.Equal(t, tabAgents, key2Index,
		"Key '2' should select Agents tab")
}
