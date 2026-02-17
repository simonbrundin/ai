package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"ai-tui/agent"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("141"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
)

const (
	specialPathStart    = "start"
	specialPathStdio    = "--stdio"
	specialPathWildcard = "**"
)

var spinners = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type model struct {
	agents  []agent.Agent
	issues  []issue
	loading bool
	err     error
	repo    string
	spinner int
}

type issue struct {
	Number int
	Title  string
	State  string
	Labels []string
	Repo   string
}

func main() {
	p := tea.NewProgram(&model{repo: "simonbrundin/ai"})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.refresh, tick())
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return t
	})
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case time.Time:
		m.spinner = (m.spinner + 1) % len(spinners)
		return m, tick()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, m.refresh
		}
	case refreshComplete:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		}
		m.agents = msg.agents
		m.issues = msg.issues
	}
	return m, nil
}

func (m *model) View() string {
	s := titleStyle.Render("🤖 AI Monitor") + "\n\n"

	if m.loading {
		spinner := spinners[m.spinner]
		s += statusStyle.Render(spinner + " Loading...")
		return s
	}

	s += sectionTitleStyle.Render("🤖 Running Agents") + "\n"
	if len(m.agents) == 0 && m.err == nil {
		s += itemStyle.Render("  No agents running") + "\n"
	} else if len(m.agents) > 0 {
		seen := make(map[string]bool)
		for _, a := range m.agents {
			key := a.Name + ":" + a.WorkingDir
			if seen[key] {
				continue
			}
			seen[key] = true
			repoName := getRepoName(a.WorkingDir)
			s += itemStyle.Render(fmt.Sprintf("  • OpenCode @ %s", repoName)) + "\n"
		}
	}

	s += "\n" + sectionTitleStyle.Render("📋 GitHub Issues (alla repos)") + "\n"
	if len(m.issues) == 0 && m.err == nil {
		s += itemStyle.Render("  No issues found") + "\n"
	} else if len(m.issues) > 0 {
		for _, i := range m.issues {
			labels := ""
			if len(i.Labels) > 0 {
				labels = " " + labelStyle.Render(fmt.Sprintf("[%s]", strings.Join(i.Labels, ", ")))
			}
			repoName := i.Repo
			if idx := strings.Index(repoName, "/"); idx > 0 {
				repoName = repoName[idx+1:]
			}
			s += itemStyle.Render(fmt.Sprintf("  #%d %s%s (%s)", i.Number, truncate(i.Title, 30), labels, repoName)) + "\n"
		}
	}

	if m.err != nil {
		s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n"
	}

	s += "\n" + statusStyle.Render("r: refresh | q: quit")
	return s
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getRepoName(path string) string {
	if path == "" || path == specialPathWildcard || path == specialPathStart || path == specialPathStdio {
		return path
	}
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

type refreshComplete struct {
	agents []agent.Agent
	issues []issue
	err    error
}

func (m *model) refresh() tea.Msg {
	agents, err := agent.DetectAgents()
	issues, fetchErr := fetchAllIssues()

	if err != nil {
		return refreshComplete{agents: agents, issues: issues, err: fmt.Errorf("agent detection failed: %w", err)}
	}

	if fetchErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", fetchErr)
	}

	return refreshComplete{agents: agents, issues: issues, err: nil}
}

func fetchAllIssues() ([]issue, error) {
	out, err := runGHCommand("repo", "list", "--limit", "50", "--json", "nameWithOwner")
	if err != nil {
		return nil, formatGHError(err)
	}

	var repos []struct {
		NameWithOwner string `json:"nameWithOwner"`
	}
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, err
	}

	var allIssues []issue
	var failedRepos []string
	for _, repo := range repos {
		out, cmdErr := runGHCommand("issue", "list", "--repo", repo.NameWithOwner, "--limit", "10")
		if cmdErr != nil {
			failedRepos = append(failedRepos, repo.NameWithOwner)
			continue
		}
		issues := parseIssues(string(out))
		for i := range issues {
			issues[i].Repo = repo.NameWithOwner
		}
		allIssues = append(allIssues, issues...)
	}

	if len(failedRepos) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: failed to fetch from repos: %s\n", strings.Join(failedRepos, ", "))
	}

	return allIssues, nil
}

func fetchGitHubIssues(repo string) ([]issue, error) {
	out, err := runGHCommand("issue", "list", "--repo", repo, "--limit", "20")
	if err != nil {
		return nil, formatGHError(err)
	}
	return parseIssues(string(out)), nil
}

func runGHCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	return cmd.Output()
}

func parseIssues(output string) []issue {
	var issues []issue
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 3 {
			num := 0
			fmt.Sscanf(parts[0], "%d", &num)
			labels := []string{}
			if len(parts) >= 4 && parts[3] != "" {
				labels = strings.Split(parts[3], ",")
			}
			issues = append(issues, issue{
				Number: num,
				Title:  strings.TrimSpace(parts[2]),
				State:  strings.TrimSpace(parts[1]),
				Labels: labels,
			})
		}
	}
	return issues
}

func formatGHError(err error) error {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "exec format error"),
		strings.Contains(errStr, "not found"),
		strings.Contains(errStr, "no such file"):
		return fmt.Errorf("gh CLI not found. Please install GitHub CLI: https://cli.github.com")

	case strings.Contains(errStr, "authentication"),
		strings.Contains(errStr, "Auth"),
		strings.Contains(errStr, "not authenticated"),
		strings.Contains(errStr, "could not read"):
		return fmt.Errorf("GitHub not authenticated. Run 'gh auth login'")

	case strings.Contains(errStr, "rate limit"),
		strings.Contains(errStr, "Rate limit"):
		return fmt.Errorf("GitHub API rate limited. Please wait and try again")

	case strings.Contains(errStr, "connection"),
		strings.Contains(errStr, "network"),
		strings.Contains(errStr, "no such host"):
		return fmt.Errorf("Network error. Check your internet connection")

	default:
		return fmt.Errorf("gh error: %w", err)
	}
}
