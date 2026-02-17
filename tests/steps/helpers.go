package steps

import (
	"strings"
)

func FilterIssuesByLabel(issues []Issue, label string) []Issue {
	var result []Issue
	for _, issue := range issues {
		for _, l := range issue.Labels {
			if strings.EqualFold(l.Name, label) {
				result = append(result, issue)
				break
			}
		}
	}
	return result
}

func FilterIssuesBySearch(issues []Issue, term string) []Issue {
	var result []Issue
	term = strings.ToLower(term)
	for _, issue := range issues {
		if strings.Contains(strings.ToLower(issue.Title), term) {
			result = append(result, issue)
		}
	}
	return result
}

// GroupIssuesByRepo groups issues by their repository name (without owner prefix)
func GroupIssuesByRepo(issues []Issue) map[string][]Issue {
	grouped := make(map[string][]Issue)
	for _, issue := range issues {
		repoName := issue.Repo
		// Extract just the repo name (after the /)
		if idx := strings.Index(repoName, "/"); idx > 0 {
			repoName = repoName[idx+1:]
		}
		grouped[repoName] = append(grouped[repoName], issue)
	}
	return grouped
}

// GetRepoName extracts just the repo name from a full repo path (e.g., "simonbrundin/ai" -> "ai")
func GetRepoName(repoPath string) string {
	if idx := strings.Index(repoPath, "/"); idx > 0 {
		return repoPath[idx+1:]
	}
	return repoPath
}

// IssuesByRepoCount holds the count of issues per repo
type IssuesByRepoCount struct {
	Repo   string
	Count  int
	Issues []Issue
}

// GetRepoIssueCounts returns a slice of repo names with their issue counts
func GetRepoIssueCounts(issues []Issue) []IssuesByRepoCount {
	grouped := GroupIssuesByRepo(issues)
	var result []IssuesByRepoCount
	for repo, issueList := range grouped {
		result = append(result, IssuesByRepoCount{
			Repo:   repo,
			Count:  len(issueList),
			Issues: issueList,
		})
	}
	return result
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
