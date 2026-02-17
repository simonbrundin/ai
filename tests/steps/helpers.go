package steps

import "strings"

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

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
