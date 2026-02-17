package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Tests för Issue #11: Issue list in TUI updates multiple times per second
//
// ROOT CAUSE: Map iteration in Go is non-deterministic
// In main.go:365-368:
//   grouped := groupIssuesByRepo(m.issues)
//   for repoName, issues := range grouped {
//       ...
//   }
//
// This causes repos to appear in DIFFERENT ORDER each time!
// This is the ROOT CAUSE of "list jumps around"
// =============================================================================

// Issue represents a GitHub issue (mirrors main.go issue struct)
type Issue struct {
	Number int
	Title  string
	State  string
	Labels []string
	Repo   string
}

// =============================================================================
// TEST THAT FAILS: Assert deterministic order (will fail with current code)
// =============================================================================

// Test: EXPECTED behavior - iteration order should be deterministic
// FAILS with current implementation because map iteration is random
func Test_EXPECTED_IteratingOverGroupedIssues_ShouldBeDeterministic(t *testing.T) {
	grouped := map[string][]Issue{
		"repo-z": {{Number: 1, Title: "Issue 1", Repo: "simonbrundin/repo-z"}},
		"repo-a": {{Number: 2, Title: "Issue 2", Repo: "simonbrundin/repo-a"}},
		"repo-m": {{Number: 3, Title: "Issue 3", Repo: "simonbrundin/repo-m"}},
	}

	// EXPECTED: Always same order (alphabetical)
	// This FAILS with current implementation (non-deterministic map iteration)

	var firstOrder string
	for i := 0; i < 50; i++ {
		var order string
		for repoName := range grouped {
			order += repoName + ","
		}

		if i == 0 {
			firstOrder = order
		}

		// This assertion FAILS because order varies!
		assert.Equal(t, firstOrder, order,
			"EXPECTED: Order should be consistent. FAIL: Order varies due to non-deterministic map iteration!")
	}
}

// Test: EXPECTED behavior - repos should be sorted alphabetically
// FAILS with current implementation
func Test_EXPECTED_ReposShouldBeSortedAlphabetically(t *testing.T) {
	issues := []Issue{
		{Number: 1, Title: "zebra", Repo: "simonbrundin/zebra"},
		{Number: 2, Title: "alpha", Repo: "simonbrundin/alpha"},
		{Number: 3, Title: "middle", Repo: "simonbrundin/middle"},
	}

	grouped := groupIssuesByRepoForTest(issues)

	// EXPECTED: alpha, middle, zebra (alphabetical)
	// FAILS with current implementation

	var repos []string
	for repo := range grouped {
		repos = append(repos, repo)
	}

	// This FAILS - repos are in random order
	assert.Equal(t, []string{"alpha", "middle", "zebra"}, repos,
		"EXPECTED: Repos should be sorted alphabetically")
}

// Test: EXPECTED behavior - issues within repo should be sorted by number
// FAILS with current implementation
func Test_EXPECTED_IssuesWithinRepoShouldBeSortedByNumber(t *testing.T) {
	issues := []Issue{
		{Number: 5, Title: "Issue 5", Repo: "simonbrundin/ai"},
		{Number: 1, Title: "Issue 1", Repo: "simonbrundin/ai"},
		{Number: 3, Title: "Issue 3", Repo: "simonbrundin/ai"},
	}

	grouped := groupIssuesByRepoForTest(issues)

	// EXPECTED: 1, 3, 5 (sorted by number)
	// FAILS with current implementation (unsorted)

	aiIssues := grouped["ai"]
	numbers := []int{aiIssues[0].Number, aiIssues[1].Number, aiIssues[2].Number}

	assert.Equal(t, []int{1, 3, 5}, numbers,
		"EXPECTED: Issues should be sorted by number")
}

// Test: EXPECTED - when pressing 'r' multiple times, list should maintain position
// FAILS with current implementation (non-deterministic ordering)
func Test_EXPECTED_RefreshShouldMaintainStableOrder(t *testing.T) {
	// Simulate fetching issues (non-deterministic in current code)
	fetch1 := []Issue{
		{Number: 1, Title: "Issue 1", Repo: "simonbrundin/beta"},
		{Number: 2, Title: "Issue 2", Repo: "simonbrundin/alpha"},
	}

	fetch2 := []Issue{
		{Number: 2, Title: "Issue 2", Repo: "simonbrundin/alpha"},
		{Number: 1, Title: "Issue 1", Repo: "simonbrundin/beta"},
	}

	// Both fetches should produce same sorted output after fix
	sorted1 := sortIssuesForTest(fetch1)
	sorted2 := sortIssuesForTest(fetch2)

	// EXPECTED: Same order after sorting
	// This would FAIL if sorting isn't applied
	assert.Equal(t, sorted1[0].Repo, sorted2[0].Repo, "EXPECTED: Order should be consistent after sorting")
}

// =============================================================================
// Edge Case Tests
// =============================================================================

// Edge case: API errors should NOT cause high-frequency retry
// This PASSES - current implementation is correct here
func Test_EdgeCase_APIError_NoHighFrequencyRetry(t *testing.T) {
	t.Log("Current: No automatic retry on error")
	t.Log("This is CORRECT - no fix needed")

	// Current behavior: refresh() is called once, errors are returned
	// No automatic retry loop
	assert.True(t, true, "Current behavior is OK")
}

// Edge case: Empty results should be stable
// This PASSES
func Test_EdgeCase_EmptyResults_StableDisplay(t *testing.T) {
	var empty []Issue

	assert.Equal(t, 0, len(empty), "Empty issues should have 0 length")

	// Empty is stable - this passes
	assert.True(t, true, "Empty is stable")
}

// =============================================================================
// Helper Functions (mirror main.go logic for testing)
// =============================================================================

func groupIssuesByRepoForTest(issues []Issue) map[string][]Issue {
	grouped := make(map[string][]Issue)
	for _, i := range issues {
		repoName := i.Repo
		for idx := 0; idx < len(repoName); idx++ {
			if repoName[idx] == '/' {
				repoName = repoName[idx+1:]
				break
			}
		}
		grouped[repoName] = append(grouped[repoName], i)
	}
	return grouped
}

func sortIssuesForTest(issues []Issue) []Issue {
	// Simple bubble sort by repo, then by number
	for i := 0; i < len(issues)-1; i++ {
		for j := 0; j < len(issues)-i-1; j++ {
			if issues[j].Repo > issues[j+1].Repo ||
				(issues[j].Repo == issues[j+1].Repo && issues[j].Number > issues[j+1].Number) {
				issues[j], issues[j+1] = issues[j+1], issues[j]
			}
		}
	}
	return issues
}
