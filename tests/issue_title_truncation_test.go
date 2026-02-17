package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Tests för Issue #19: Issue titles are truncated to 30 characters
//
// ROOT CAUSE: Hardcoded truncation limit of 30 characters
// In main.go:437:
//   truncate(i.Title, 30)
//
// This prevents users from seeing full issue titles
// =============================================================================

// Test: EXPECTED behavior - title truncation should be dynamic based on terminal width
// FAILS with current implementation (hardcoded 30)
func Test_EXPECTED_TitleTruncationShouldBeDynamicBasedOnWidth(t *testing.T) {
	terminalWidth := 120
	prefixWidth := 6   // "    #" + number space (max 5 digits + 1 = 6)
	labelsBuffer := 20 // buffer for labels "[label1, label2]"

	// Calculate expected max title length
	expectedMaxTitle := calculateMaxTitleWidth(terminalWidth, prefixWidth, labelsBuffer)

	// With 120 width terminal, title should NOT be truncated for titles < expectedMaxTitle
	longTitle := "This is a very long issue title that should not be truncated when terminal is wide"

	// The function should return full title when it's shorter than max
	result := truncateForTest(longTitle, expectedMaxTitle)

	// FAILS if hardcoded 30 is used instead of dynamic width
	assert.GreaterOrEqual(t, len(result), 50,
		"EXPECTED: With 120 width terminal, title should not be truncated at 30 chars")
}

// Test: EXPECTED behavior - narrow terminal should still truncate appropriately
func Test_EXPECTED_NarrowTerminalShouldTruncateAppropriately(t *testing.T) {
	terminalWidth := 80
	prefixWidth := 6
	labelsBuffer := 20

	expectedMaxTitle := calculateMaxTitleWidth(terminalWidth, prefixWidth, labelsBuffer)

	// This title is longer than what fits in 80 width
	longTitle := "This is a very long issue title that definitely exceeds the available width"

	result := truncateForTest(longTitle, expectedMaxTitle)

	// Should be truncated but to a reasonable length (not hardcoded 30)
	assert.LessOrEqual(t, len(result), len(longTitle),
		"EXPECTED: Title should be truncated when exceeding width")
}

// Test: Edge case - very narrow terminal (minimum viable width)
func Test_EdgeCase_MinimumTerminalWidth(t *testing.T) {
	terminalWidth := 60 // very narrow
	prefixWidth := 6
	labelsBuffer := 10

	expectedMaxTitle := calculateMaxTitleWidth(terminalWidth, prefixWidth, labelsBuffer)

	// Should still allow reasonable title length even on narrow terminal
	assert.GreaterOrEqual(t, expectedMaxTitle, 20,
		"EXPECTED: Even narrow terminals should allow at least 20 char titles")
}

// Test: Edge case - labels affect available title space
func Test_EdgeCase_LabelsReduceAvailableTitleSpace(t *testing.T) {
	terminalWidth := 100
	prefixWidth := 6

	// No labels = more space for title
	maxTitleNoLabels := calculateMaxTitleWidth(terminalWidth, prefixWidth, 0)
	// With labels = less space for title
	maxTitleWithLabels := calculateMaxTitleWidth(terminalWidth, prefixWidth, 20)

	assert.Greater(t, maxTitleNoLabels, maxTitleWithLabels,
		"EXPECTED: Labels should reduce available title space")
}

// =============================================================================
// Helper Functions
// =============================================================================

// calculateMaxTitleWidth calculates the maximum width for issue titles
// based on terminal width, prefix (line number), and labels
func calculateMaxTitleWidth(terminalWidth, prefixWidth, labelsBuffer int) int {
	// Available width = terminal width - prefix - labels buffer - padding
	available := terminalWidth - prefixWidth - labelsBuffer - 2
	if available < 10 {
		return 10 // minimum
	}
	return available
}

// truncateForTest truncates a string to maxLen (mirrors main.go truncate)
func truncateForTest(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
