package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"ai-tui/agent"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	headerHeight = 2
	footerHeight = 1
	minWidth     = 80
	minHeight    = 24
)

const (
	searchLimit          = 100
	issuePrefixWidth     = 10
	issuePadding         = 2
	issueMinTitleWidth   = 10
	dialogWidth          = 50
	dialogHeight         = 12
	confirmTitleTruncate = 30
)

const (
	commandDialogWidth    = 40
	commandDialogHeight   = 14
	commandPromptTemplate = "/%s %%d"
)

const (
	newIssueDialogWidth  = 50
	newIssueDialogHeight = 15
	opencodeSecurePath   = "/home/simon/repos/dotfiles/opencode/.config/opencode/opencode-secure"
	opencodeIssuePrompt  = "--model opencode/minimax-m2.5-free --prompt \"/issue\""
)

const (
	phaseDialogWidth  = 40
	phaseDialogHeight = 12
)

var (
	commandNames   = []string{"Skriv tester", "Implementera", "Refactor", "Dokumentera", "Skapa PR"}
	commandAliases = []string{"/tdd", "/implement", "/refactor", "/docs", "/pr"}
)

var phaseLabels = []string{"tester", "implementation", "refactor", "docs", "user_test", "pr"}

var phaseDescriptions = map[string]string{
	"tester":         "Issue är i testfas",
	"implementation": "Issue är i implementationsfas",
	"refactor":       "Issue är i refaktoringsfas",
	"docs":           "Issue är i dokumentationsfas",
	"user_test":      "Issue är i användartestfas",
	"pr":             "Issue är i PR-fas",
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	headerBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("205")).
			Padding(0, 2).
			Bold(true)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Padding(0, 2)

	footerBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	keyHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212")).
				Padding(1, 2, 0, 2)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Padding(0, 2)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				Padding(0, 2)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("141"))

	phaseLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	helpModalStyle = lipgloss.NewStyle().
			Width(60).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("236")).
			Padding(1)

	helpTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Padding(0, 0, 1, 0)

	helpItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	confirmDialogStyle = lipgloss.NewStyle().
				Width(50).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("205")).
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("236")).
				Padding(1)

	confirmDialogTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				Padding(0, 0, 1, 0)

	confirmDialogOptionStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("252"))

	confirmDialogHighlightStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("205")).
					Bold(true)

	commandDialogStyle = lipgloss.NewStyle().
				Width(commandDialogWidth).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("205")).
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("236")).
				Padding(1)

	commandDialogTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				Padding(0, 0, 1, 0)

	commandDialogItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	commandDialogSelectedStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("205")).
					Bold(true)

	commandDialogHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(1, 2)

	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

const (
	specialPathStart    = "start"
	specialPathStdio    = "--stdio"
	specialPathWildcard = "**"
)

var spinners = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

var browserCommands = []string{"xdg-open", "gnome-open", "firefox", "chromium-browser", "google-chrome"}

type model struct {
	agents            []agent.Agent
	issues            []issue
	loading           bool
	err               error
	repo              string
	spinner           int
	currentTab        int
	showHelp          bool
	helpSearch        string
	helpMatches       []string
	width             int
	height            int
	ready             bool
	filterActive      bool
	selectedIssue     int
	issueURL          string
	showConfirmDialog bool
	showCommandDialog bool
	selectedCommand   int
	// New Issue Dialog (Issue #27)
	showNewIssueDialog    bool
	newIssueRepos         []string
	newIssueFilteredRepos []string
	newIssueSelectedRepo  int
	newIssueDialogMode    string
	newIssueErrorMessage  string
	newIssueFilterText    string
	newIssueTitle         string

	// Phase Dialog (Issue #30)
	showPhaseDialog bool
	selectedPhase   int
}

const (
	tabIssues = iota
	tabAgents
	numTabs = 2
)

var tabNames = []string{"Issues", "Agents"}

var allCommands = []struct {
	key   string
	label string
	desc  string
}{
	{"1-2", "tab", "Switch tabs"},
	{"tab", "next", "Next tab"},
	{"shift+tab", "prev", "Previous tab"},
	{"r", "refresh", "Refresh data"},
	{"a", "active", "Toggle active filter"},
	{"j", "down", "Next issue (vim)"},
	{"k", "up", "Previous issue (vim)"},
	{"o", "open", "Open issue in browser"},
	{"q", "quit", "Exit application"},
	{"?", "help", "Show help"},
	{"esc", "close", "Close help"},
}

func (m *model) filterHelpCommands() {
	if m.helpSearch == "" {
		m.helpMatches = nil
		for _, cmd := range allCommands {
			m.helpMatches = append(m.helpMatches, fmt.Sprintf("%-12s %s", cmd.key+":", cmd.desc))
		}
		return
	}
	searchLower := strings.ToLower(m.helpSearch)
	m.helpMatches = nil
	for _, cmd := range allCommands {
		if strings.Contains(strings.ToLower(cmd.key), searchLower) ||
			strings.Contains(strings.ToLower(cmd.label), searchLower) ||
			strings.Contains(strings.ToLower(cmd.desc), searchLower) {
			m.helpMatches = append(m.helpMatches, fmt.Sprintf("%-12s %s", cmd.key+":", cmd.desc))
		}
	}
	if len(m.helpMatches) == 0 {
		m.helpMatches = []string{"(no matches)"}
	}
}

type issue struct {
	Number int
	Title  string
	State  string
	Labels []string
	Repo   string
}

// groupIssuesByRepo groups issues by repository name (without owner prefix)
// Issues are sorted by repo name, then by issue number for deterministic display
func groupIssuesByRepo(issues []issue) map[string][]issue {
	grouped := make(map[string][]issue)
	for _, i := range issues {
		repoName := i.Repo
		if idx := strings.Index(repoName, "/"); idx > 0 {
			repoName = repoName[idx+1:]
		}
		grouped[repoName] = append(grouped[repoName], i)
	}

	for repoName := range grouped {
		sort.SliceStable(grouped[repoName], func(i, j int) bool {
			return grouped[repoName][i].Number < grouped[repoName][j].Number
		})
	}

	return grouped
}

// sortedRepoKeys extracts and sorts repo names from a grouped map
func sortedRepoKeys(grouped map[string][]issue) []string {
	keys := make([]string, 0, len(grouped))
	for repoName := range grouped {
		keys = append(keys, repoName)
	}
	sort.Strings(keys)
	return keys
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
	// Handle issue-input mode for all keys not explicitly handled
	if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter", "return", "escape", "esc", "n", "y", "q", "ctrl+c",
				"j", "k", "r", "a", "o", "p", "d", "?", "tab", "shift+tab",
				"up", "down", "backspace":
				// Let these be handled by their specific cases below
			default:
				// For all other keys (including å, ö, ä), add to input
				if len(keyMsg.String()) > 0 {
					m.newIssueTitle += keyMsg.String()
					return m, nil
				}
			}
		}
	}

	switch msg := msg.(type) {
	case time.Time:
		m.spinner = (m.spinner + 1) % len(spinners)
		return m, tick()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "q"
				return m, nil
			}
			return m, tea.Quit
		case "r":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "r"
				return m, nil
			}
			m.loading = true
			return m, m.refresh
		case "a":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "a"
				return m, nil
			}
			m.filterActive = !m.filterActive
			return m, nil
		case "?":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "?"
				return m, nil
			}
			m.showHelp = true
			m.helpSearch = ""
			m.filterHelpCommands()
			return m, nil
		case "escape", "esc":
			if m.showHelp {
				m.showHelp = false
				m.helpSearch = ""
			}
			if m.showConfirmDialog {
				m.showConfirmDialog = false
			}
			if m.showCommandDialog {
				m.showCommandDialog = false
				m.selectedCommand = -1
			}
			if m.showNewIssueDialog {
				m.showNewIssueDialog = false
				m.newIssueDialogMode = ""
				m.newIssueFilterText = ""
				m.newIssueTitle = ""
				m.newIssueSelectedRepo = 0
			}
			if m.showPhaseDialog {
				m.showPhaseDialog = false
				m.selectedPhase = -1
			}
			return m, nil
		case "tab":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "\t"
				return m, nil
			}
			m.currentTab = (m.currentTab + 1) % numTabs
			return m, nil
		case "shift+tab":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "\t"
				return m, nil
			}
			m.currentTab = (m.currentTab - 1 + numTabs) % numTabs
			return m, nil
		case "j":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "j"
				return m, nil
			}
			m.moveToNextIssue()
			return m, nil
		case "k":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "k"
				return m, nil
			}
			m.moveToPreviousIssue()
			return m, nil
		case "o":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "o"
				return m, nil
			}
			return m, m.openSelectedIssueInBrowser()
		case "p":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "p"
				return m, nil
			}
			if m.currentTab == tabIssues && len(m.issues) > 0 && m.selectedIssue >= 0 && m.selectedIssue < len(m.issues) {
				m.openPhaseDialog()
			}
			return m, nil
		case "d":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "d"
				return m, nil
			}
			m.showCloseIssueDialog()
			return m, nil
		case "y":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "y"
				return m, nil
			}
			if m.showNewIssueDialog && m.newIssueDialogMode == "repo-select" {
				m.executeNewIssueSelection()
				return m, nil
			}
			if m.showCommandDialog {
				m.executeSelectedCommand()
				return m, nil
			}
			m.confirmAndCloseIssue()
			return m, nil
		case "enter":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.executeIssueTitleInput()
				return m, nil
			}
			if m.showNewIssueDialog && m.newIssueDialogMode == "repo-select" {
				m.executeNewIssueSelection()
				return m, nil
			}
			if m.currentTab == tabIssues && len(m.issues) > 0 && m.selectedIssue >= 0 && m.selectedIssue < len(m.issues) {
				m.showCommandDialog = true
				m.selectedCommand = 0
				return m, nil
			}
			m.confirmAndCloseIssue()
			return m, nil
		case "n":
			if m.showCommandDialog {
				m.showCommandDialog = false
				m.selectedCommand = -1
				return m, nil
			}
			if m.showNewIssueDialog {
				if m.newIssueDialogMode == "issue-input" {
					m.newIssueTitle += "n"
					return m, nil
				}
				m.showNewIssueDialog = false
				m.newIssueFilterText = ""
				m.newIssueTitle = ""
				return m, nil
			}
			if m.currentTab == tabIssues {
				m.openNewIssueDialog()
			}
			m.showConfirmDialog = false
			return m, nil
		case "up":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "↑"
				return m, nil
			}
			if m.showCommandDialog && m.selectedCommand > 0 {
				m.selectedCommand--
			}
			if m.showNewIssueDialog && m.newIssueDialogMode == "repo-select" && m.newIssueSelectedRepo > 0 {
				m.newIssueSelectedRepo--
			}
			if m.showPhaseDialog && m.selectedPhase > 0 {
				m.selectedPhase--
			}
			return m, nil
		case "down":
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += "↓"
				return m, nil
			}
			if m.showCommandDialog && m.selectedCommand < len(commandNames)-1 {
				m.selectedCommand++
			}
			if m.showNewIssueDialog && m.newIssueDialogMode == "repo-select" && m.newIssueSelectedRepo < len(m.newIssueFilteredRepos)-1 {
				m.newIssueSelectedRepo++
			}
			if m.showPhaseDialog && m.selectedPhase < len(phaseLabels)-1 {
				m.selectedPhase++
			}
			return m, nil
		}
		if m.showHelp {
			if msg.String() == "backspace" {
				if len(m.helpSearch) > 0 {
					m.helpSearch = m.helpSearch[:len(m.helpSearch)-1]
					m.filterHelpCommands()
				}
			} else if len(msg.String()) == 1 {
				m.helpSearch += msg.String()
				m.filterHelpCommands()
			}
			return m, nil
		}
		if m.showCommandDialog && len(msg.String()) == 1 {
			key := msg.String()
			if key >= "1" && key <= "5" {
				m.selectedCommand = int(key[0] - '1')
				m.executeSelectedCommand()
				return m, nil
			}
		}
		// Handle Enter key for new issue dialog BEFORE the single-char check
		// This fixes the bug where Enter was never handled because "enter" has len=5
		if m.showNewIssueDialog && m.newIssueDialogMode == "repo-select" && (msg.String() == "enter" || msg.String() == "return") {
			m.executeNewIssueSelection()
			return m, nil
		}
		if m.showNewIssueDialog && m.newIssueDialogMode == "repo-select" && len(msg.String()) == 1 {
			key := msg.String()
			if key >= "1" && key <= "9" {
				repoNum := int(key[0] - '1')
				if repoNum < len(m.newIssueFilteredRepos) {
					m.newIssueSelectedRepo = repoNum
					m.executeNewIssueSelection()
					return m, nil
				}
			}
			m.newIssueFilterText += key
			m.filterNewIssueRepos()
			return m, nil
		}
		if m.showNewIssueDialog && m.newIssueDialogMode == "repo-select" && msg.String() == "backspace" {
			if len(m.newIssueFilterText) > 0 {
				m.newIssueFilterText = m.newIssueFilterText[:len(m.newIssueFilterText)-1]
				m.filterNewIssueRepos()
			}
			return m, nil
		}
		// Handle issue title input mode
		if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
			if msg.String() == "enter" || msg.String() == "return" {
				m.executeIssueTitleInput()
				return m, nil
			}
			if msg.String() == "escape" || msg.String() == "esc" || msg.String() == "n" {
				m.newIssueDialogMode = "repo-select"
				m.newIssueTitle = ""
				return m, nil
			}
			if msg.String() == "backspace" {
				if len(m.newIssueTitle) > 0 {
					m.newIssueTitle = m.newIssueTitle[:len(m.newIssueTitle)-1]
				}
				return m, nil
			}
			if len(msg.String()) == 1 {
				m.newIssueTitle += msg.String()
				return m, nil
			}
			return m, nil
		}
		if m.showPhaseDialog {
			if msg.String() == "enter" || msg.String() == "return" {
				m.executePhaseSelection()
				return m, nil
			}
			if len(msg.String()) == 1 {
				key := msg.String()
				if key >= "1" && key <= "6" {
					m.selectedPhase = int(key[0] - '1')
					m.executePhaseSelection()
					return m, nil
				}
			}
			if msg.String() == "n" {
				m.showPhaseDialog = false
				m.selectedPhase = -1
				return m, nil
			}
			return m, nil
		}
		if len(msg.String()) >= 1 {
			key := msg.String()
			if m.showNewIssueDialog && m.newIssueDialogMode == "issue-input" {
				m.newIssueTitle += key
				return m, nil
			}
			if key >= "1" && key <= "9" {
				tabNum := int(key[0] - '0')
				if tabNum >= 1 && tabNum <= numTabs {
					m.currentTab = tabNum - 1
				}
			}
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

func (m *model) moveToNextIssue() {
	if m.currentTab == tabIssues && len(m.issues) > 0 {
		grouped := groupIssuesByRepo(m.issues)
		repoNames := sortedRepoKeys(grouped)
		visualOrder := buildVisualOrder(grouped, repoNames)

		currentIdx := -1
		for i, iss := range visualOrder {
			if iss.Number == m.issues[m.selectedIssue].Number && iss.Repo == m.issues[m.selectedIssue].Repo {
				currentIdx = i
				break
			}
		}

		if currentIdx >= 0 && currentIdx < len(visualOrder)-1 {
			nextIssue := visualOrder[currentIdx+1]
			for i, iss := range m.issues {
				if iss.Number == nextIssue.Number && iss.Repo == nextIssue.Repo {
					m.selectedIssue = i
					break
				}
			}
		}
	}
}

func (m *model) moveToPreviousIssue() {
	if m.currentTab == tabIssues && len(m.issues) > 0 {
		grouped := groupIssuesByRepo(m.issues)
		repoNames := sortedRepoKeys(grouped)
		visualOrder := buildVisualOrder(grouped, repoNames)

		currentIdx := -1
		for i, iss := range visualOrder {
			if iss.Number == m.issues[m.selectedIssue].Number && iss.Repo == m.issues[m.selectedIssue].Repo {
				currentIdx = i
				break
			}
		}

		if currentIdx > 0 {
			prevIssue := visualOrder[currentIdx-1]
			for i, iss := range m.issues {
				if iss.Number == prevIssue.Number && iss.Repo == prevIssue.Repo {
					m.selectedIssue = i
					break
				}
			}
		}
	}
}

func buildVisualOrder(grouped map[string][]issue, repoNames []string) []issue {
	var visualOrder []issue
	for _, repoName := range repoNames {
		visualOrder = append(visualOrder, grouped[repoName]...)
	}
	return visualOrder
}

func (m *model) openPhaseDialog() {
	if m.currentTab != tabIssues || len(m.issues) == 0 || m.selectedIssue < 0 || m.selectedIssue >= len(m.issues) {
		return
	}
	m.showPhaseDialog = true
	m.selectedPhase = 0
}

func (m *model) executePhaseSelection() {
	if !m.showPhaseDialog || m.selectedPhase < 0 || m.selectedPhase >= len(phaseLabels) {
		m.showPhaseDialog = false
		m.selectedPhase = -1
		return
	}

	if m.selectedIssue < 0 || m.selectedIssue >= len(m.issues) {
		m.showPhaseDialog = false
		m.selectedPhase = -1
		return
	}

	phaseLabel := phaseLabels[m.selectedPhase]

	issue := &m.issues[m.selectedIssue]
	alreadyHasLabel := false
	for _, l := range issue.Labels {
		if l == phaseLabel {
			alreadyHasLabel = true
			break
		}
	}

	err := addIssueLabel(issue.Repo, issue.Number, phaseLabel)
	if err != nil {
		m.err = fmt.Errorf("failed to add label: %w", err)
	} else if !alreadyHasLabel {
		issue.Labels = append(issue.Labels, phaseLabel)
	}

	m.showPhaseDialog = false
	m.selectedPhase = -1
}

func (m *model) openSelectedIssueInBrowser() tea.Cmd {
	if m.currentTab == tabIssues && len(m.issues) > 0 && m.selectedIssue < len(m.issues) {
		issue := m.issues[m.selectedIssue]
		m.issueURL = fmt.Sprintf("https://github.com/%s/issues/%d", issue.Repo, issue.Number)
		return openBrowser(m.issueURL)
	}
	return nil
}

func (m *model) showCloseIssueDialog() {
	if m.currentTab == tabIssues && len(m.issues) > 0 && m.selectedIssue >= 0 && m.selectedIssue < len(m.issues) {
		m.showConfirmDialog = true
	}
}

func (m *model) confirmAndCloseIssue() {
	if !m.showConfirmDialog || m.selectedIssue < 0 || m.selectedIssue >= len(m.issues) {
		return
	}
	issue := m.issues[m.selectedIssue]
	err := closeGitHubIssue(issue.Repo, issue.Number)
	if err != nil {
		m.err = fmt.Errorf("failed to close issue: %w", err)
	}
	m.showConfirmDialog = false
}

func (m *model) executeSelectedCommand() {
	if !m.showCommandDialog || m.selectedCommand < 0 || m.selectedCommand >= len(commandAliases) {
		return
	}
	if m.selectedIssue < 0 || m.selectedIssue >= len(m.issues) {
		m.showCommandDialog = false
		m.selectedCommand = -1
		return
	}

	issue := m.issues[m.selectedIssue]
	issueNum := issue.Number
	command := commandAliases[m.selectedCommand]

	cmd := exec.Command("tmux", "new-window", "-d", "-n", fmt.Sprintf("opencode-%s-%d", command, issueNum))
	if err := cmd.Run(); err != nil {
		m.err = fmt.Errorf("failed to create tmux window: %w", err)
		m.showCommandDialog = false
		m.selectedCommand = -1
		return
	}

	prompt := fmt.Sprintf("--model opencode/minimax-m2.5-free --prompt \"%s %d\"", command, issueNum)
	fullCommand := fmt.Sprintf("%s %s", opencodeSecurePath, prompt)
	cmd = exec.Command("tmux", "send-keys", "-t", fmt.Sprintf("opencode-%s-%d", command, issueNum), fullCommand, "Enter")
	if err := cmd.Run(); err != nil {
		m.err = fmt.Errorf("failed to run opencode-secure: %w", err)
	}

	windowName := fmt.Sprintf("opencode-%s-%d", command, issueNum)
	selectCmd := exec.Command("bash", "-c", fmt.Sprintf("tmux select-window -t %q", windowName))
	_ = selectCmd.Run()

	m.showCommandDialog = false
	m.selectedCommand = -1
}

func (m *model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.width < minWidth || m.height < minHeight {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			errorStyle.Render(fmt.Sprintf("Terminal too small (%dx%d)\nMinimum: %dx%d", m.width, m.height, minWidth, minHeight)))
	}

	var s strings.Builder

	s.WriteString(m.renderHeader())
	s.WriteString("\n")

	contentHeight := m.height - headerHeight - footerHeight
	s.WriteString(m.renderContent(m.width, contentHeight))
	s.WriteString("\n")

	s.WriteString(m.renderFooter())

	if m.showNewIssueDialog {
		return m.renderNewIssueDialogOverlay(s.String())
	}

	if m.showConfirmDialog {
		return m.renderConfirmDialog(s.String())
	}

	if m.showCommandDialog {
		return m.renderCommandDialog(s.String())
	}

	if m.showPhaseDialog {
		return m.renderPhaseDialog(s.String())
	}

	if m.showHelp {
		return m.renderHelpOverlay(s.String())
	}

	return s.String()
}

func (m *model) renderHeader() string {
	title := titleStyle.Render(" AI Monitor")

	var tabs []string
	for i, name := range tabNames {
		isActive := i == m.currentTab
		tabStr := name
		if isActive {
			tabStr = tabActiveStyle.Render(" " + name + " ")
		} else {
			tabStr = tabInactiveStyle.Render(" " + name + " ")
		}
		tabs = append(tabs, tabStr)
	}
	tabBar := strings.Join(tabs, " ")

	titleWidth := lipgloss.Width(title)
	tabWidth := lipgloss.Width(tabBar)
	spacing := m.width - titleWidth - tabWidth - 1
	if spacing < 1 {
		spacing = 1
	}

	header := title + strings.Repeat(" ", spacing/2) + tabBar
	return headerBarStyle.Width(m.width).Render(header)
}

func (m *model) renderContent(width, height int) string {
	if m.loading {
		spinner := spinners[m.spinner]
		msg := spinner + " Loading..."
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			statusStyle.Render(msg))
	}

	var s strings.Builder

	if m.currentTab == tabAgents {
		s.WriteString(m.renderAgentsView())
	} else {
		s.WriteString(m.renderIssuesView())
	}

	if m.err != nil {
		s.WriteString("\n")
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		s.WriteString("\n")
	}

	content := s.String()
	return lipgloss.NewStyle().Width(width).Height(height).Render(content)
}

func (m *model) renderAgentsView() string {
	var s strings.Builder

	agentsToShow := m.agents
	if m.filterActive {
		agentsToShow = agent.FilterActive(m.agents, true)
		s.WriteString(mutedStyle.Render("  [Filtering: active only]"))
		s.WriteString("\n")
	}

	s.WriteString(sectionTitleStyle.Render("🤖 Running Agents"))
	s.WriteString("\n")

	if len(agentsToShow) == 0 && m.err == nil {
		s.WriteString(itemStyle.Render("  No agents running"))
		s.WriteString("\n")
	} else if len(agentsToShow) > 0 {
		seen := make(map[string]bool)
		for _, a := range agentsToShow {
			key := a.Name + ":" + a.WorkingDir
			if seen[key] {
				continue
			}
			seen[key] = true
			repoName := getRepoName(a.WorkingDir)
			s.WriteString(itemStyle.Render(fmt.Sprintf("  • OpenCode @ %s", repoName)))
			s.WriteString("\n")
		}
	}

	return s.String()
}

func (m *model) renderIssuesView() string {
	var s strings.Builder

	s.WriteString(sectionTitleStyle.Render("📋 GitHub Issues"))
	s.WriteString("\n")

	if len(m.issues) == 0 && m.err == nil {
		s.WriteString(itemStyle.Render("  No issues found"))
		s.WriteString("\n")
	} else if len(m.issues) > 0 {
		grouped := groupIssuesByRepo(m.issues)
		repoNames := sortedRepoKeys(grouped)

		for _, repoName := range repoNames {
			issues := grouped[repoName]
			s.WriteString(itemStyle.Render(fmt.Sprintf("  📁 %s", repoName)))
			s.WriteString("\n")

			for _, i := range issues {
				labelsWidth := calculateLabelsWidth(i.Labels)
				labels := ""
				if len(i.Labels) > 0 {
					var labelParts []string
					for _, l := range i.Labels {
						if isPhaseLabel(l) {
							labelParts = append(labelParts, phaseLabelStyle.Render(l))
						} else {
							labelParts = append(labelParts, labelStyle.Render(l))
						}
					}
					labels = " [" + strings.Join(labelParts, ", ") + "]"
				}
				maxTitleWidth := calculateMaxTitleWidth(m.width, labelsWidth)

				prefix := "    "
				currentStyle := itemStyle
				selectedIssuePtr := -1
				if m.selectedIssue >= 0 && m.selectedIssue < len(m.issues) {
					selectedIssuePtr = m.issues[m.selectedIssue].Number
				}
				if selectedIssuePtr == i.Number && m.issues[m.selectedIssue].Repo == i.Repo {
					prefix = "  > "
					currentStyle = selectedItemStyle
				}

				s.WriteString(currentStyle.Render(fmt.Sprintf("%s#%d %s%s", prefix, i.Number, truncate(i.Title, maxTitleWidth), labels)))
				s.WriteString("\n")
			}
		}
	}

	return s.String()
}

func (m *model) renderFooter() string {
	filterStatus := "a: all"
	if m.filterActive {
		filterStatus = "a: active"
	}
	hints := []string{
		"1-2: tab",
		"r: refresh",
		filterStatus,
		"q: quit",
		"?: help",
	}

	// Add vim navigation hints when on Issues tab
	if m.currentTab == tabIssues && len(m.issues) > 0 {
		hints = append(hints, "j/k: nav")
		hints = append(hints, "o: open")
		hints = append(hints, "d: done")
		hints = append(hints, "n: new")
		hints = append(hints, "p: phase")
	}

	hintStr := hints[0]
	for i := 1; i < len(hints); i++ {
		hintStr += "  " + hints[i]
	}

	footer := keyHintStyle.Render(hintStr)
	return footerBarStyle.Width(m.width).Render(footer)
}

func (m *model) renderHelpOverlay(content string) string {
	var s strings.Builder

	s.WriteString(helpTitleStyle.Render("Keyboard Shortcuts"))
	s.WriteString("\n")

	if m.helpSearch != "" {
		s.WriteString(mutedStyle.Render("Search: "))
		s.WriteString(helpKeyStyle.Render(m.helpSearch))
		s.WriteString("\n\n")
	}

	for _, cmd := range m.helpMatches {
		s.WriteString(helpItemStyle.Render("  " + cmd))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(mutedStyle.Render("  esc: close"))

	helpContent := helpModalStyle.Render(s.String())

	helpWidth := 50
	helpHeight := len(m.helpMatches) + 8
	if helpWidth > m.width-4 {
		helpWidth = m.width - 4
	}
	if helpHeight > m.height-4 {
		helpHeight = m.height - 4
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpContent)
}

func (m *model) renderNewIssueDialogOverlay(content string) string {
	dialog := m.renderNewIssueDialogRaw(m.width, m.height)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

func (m *model) renderConfirmDialog(content string) string {
	var s strings.Builder

	issueNum := 0
	issueTitle := ""
	if m.selectedIssue >= 0 && m.selectedIssue < len(m.issues) {
		issueNum = m.issues[m.selectedIssue].Number
		issueTitle = m.issues[m.selectedIssue].Title
	}

	s.WriteString(confirmDialogTitleStyle.Render("Stäng issue i GitHub"))
	s.WriteString("\n\n")
	s.WriteString(confirmDialogOptionStyle.Render(fmt.Sprintf("  Issue #%d: %s", issueNum, truncate(issueTitle, confirmTitleTruncate))))
	s.WriteString("\n\n")
	s.WriteString(confirmDialogOptionStyle.Render("  Bekräfta?"))
	s.WriteString("\n\n")
	s.WriteString(confirmDialogHighlightStyle.Render("  [Ja] Enter / y"))
	s.WriteString("\n")
	s.WriteString(confirmDialogOptionStyle.Render("  [Nej] n / Esc"))

	confirmContent := confirmDialogStyle.Render(s.String())

	actualDialogWidth := dialogWidth
	actualDialogHeight := dialogHeight
	if actualDialogWidth > m.width-4 {
		actualDialogWidth = m.width - 4
	}
	if actualDialogHeight > m.height-4 {
		actualDialogHeight = m.height - 4
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, confirmContent)
}

func (m *model) renderCommandDialog(content string) string {
	var s strings.Builder

	issueNum := 0
	issueTitle := ""
	if m.selectedIssue >= 0 && m.selectedIssue < len(m.issues) {
		issueNum = m.issues[m.selectedIssue].Number
		issueTitle = m.issues[m.selectedIssue].Title
	}

	s.WriteString(commandDialogTitleStyle.Render("Välj kommando för issue #" + fmt.Sprint(issueNum)))
	s.WriteString("\n\n")
	s.WriteString(commandDialogItemStyle.Render("  " + truncate(issueTitle, 30)))
	s.WriteString("\n\n")

	for i, cmdName := range commandNames {
		if i == m.selectedCommand {
			s.WriteString(commandDialogSelectedStyle.Render(fmt.Sprintf("  > %d. %s ", i+1, cmdName)))
		} else {
			s.WriteString(commandDialogItemStyle.Render(fmt.Sprintf("    %d. %s", i+1, cmdName)))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(commandDialogHintStyle.Render("  Enter: Kör  |  ↑↓: Navigera  |  1-5: Snabbval  |  Esc: Avbryt"))

	commandContent := commandDialogStyle.Render(s.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, commandContent)
}

func (m *model) renderPhaseDialog(content string) string {
	var s strings.Builder

	issueNum := 0
	issueTitle := ""
	if m.selectedIssue >= 0 && m.selectedIssue < len(m.issues) {
		issueNum = m.issues[m.selectedIssue].Number
		issueTitle = m.issues[m.selectedIssue].Title
	}

	s.WriteString(commandDialogTitleStyle.Render("Välj fas för issue #" + fmt.Sprint(issueNum)))
	s.WriteString("\n\n")
	s.WriteString(commandDialogItemStyle.Render("  " + truncate(issueTitle, 30)))
	s.WriteString("\n\n")

	for i, phase := range phaseLabels {
		desc := phaseDescriptions[phase]
		if i == m.selectedPhase {
			s.WriteString(commandDialogSelectedStyle.Render(fmt.Sprintf("  > %d. %s ", i+1, desc)))
		} else {
			s.WriteString(commandDialogItemStyle.Render(fmt.Sprintf("    %d. %s", i+1, desc)))
		}
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(commandDialogHintStyle.Render("  Enter: Välj  |  ↑↓: Navigera  |  1-6: Snabbval  |  Esc: Avbryt"))

	phaseContent := commandDialogStyle.Render(s.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, phaseContent)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func calculateMaxTitleWidth(terminalWidth, labelsWidth int) int {
	available := terminalWidth - issuePrefixWidth - labelsWidth - issuePadding
	if available < issueMinTitleWidth {
		return issueMinTitleWidth
	}
	return available
}

func calculateLabelsWidth(labels []string) int {
	if len(labels) == 0 {
		return 0
	}
	width := len(labels) + 2
	for _, l := range labels {
		width += len(l)
	}
	return width
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
		return refreshComplete{agents: agents, issues: issues, err: fetchErr}
	}

	return refreshComplete{agents: agents, issues: issues, err: nil}
}

func fetchAllIssues() ([]issue, error) {
	out, err := runGHCommand("search", "issues", "--owner", "simonbrundin", "--state", "open", "--limit", fmt.Sprintf("%d", searchLimit), "--json", "number,title,state,repository,labels")
	if err != nil {
		return nil, formatGHError(err)
	}

	var searchResults []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		State  string `json:"state"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Repository struct {
			FullName string `json:"nameWithOwner"`
		} `json:"repository"`
	}
	if err := json.Unmarshal(out, &searchResults); err != nil {
		return nil, err
	}

	var allIssues []issue
	for _, result := range searchResults {
		labelNames := make([]string, len(result.Labels))
		for i, label := range result.Labels {
			labelNames[i] = label.Name
		}
		allIssues = append(allIssues, issue{
			Number: result.Number,
			Title:  result.Title,
			State:  result.State,
			Labels: labelNames,
			Repo:   result.Repository.FullName,
		})
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

func isPhaseLabel(label string) bool {
	for _, p := range phaseLabels {
		if p == label {
			return true
		}
	}
	return false
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

// openBrowser opens a URL in the default browser
func openBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		for _, cmd := range browserCommands {
			err := exec.Command(cmd, url).Run()
			if err == nil {
				return nil
			}
		}
		return nil
	}
}

// closeGitHubIssue closes an issue in GitHub using gh CLI
func closeGitHubIssue(repo string, number int) error {
	cmd := exec.Command("gh", "issue", "close", "--repo", repo, fmt.Sprintf("%d", number))
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Try to provide a helpful error message
		errStr := string(out)
		if strings.Contains(errStr, "already closed") {
			return fmt.Errorf("issue #%d is already closed", number)
		}
		return formatGHError(fmt.Errorf("%s: %s", err.Error(), errStr))
	}
	return nil
}

// addIssueLabel adds a label to an issue in GitHub using gh CLI
func addIssueLabel(repo string, number int, label string) error {
	cmd := exec.Command("gh", "issue", "edit", "--repo", repo, fmt.Sprintf("%d", number), "--add-label", label)
	out, err := cmd.CombinedOutput()
	if err != nil {
		errStr := string(out)
		if strings.Contains(errStr, "not found") {
			return fmt.Errorf("issue #%d not found in %s", number, repo)
		}
		return formatGHError(fmt.Errorf("%s: %s", err.Error(), errStr))
	}
	return nil
}

// =============================================================================
// New Issue Dialog (Issue #27)
// =============================================================================

func (m *model) openNewIssueDialog() {
	m.showNewIssueDialog = true
	m.newIssueDialogMode = "repo-select"
	m.newIssueSelectedRepo = 0
	m.newIssueFilterText = ""
	m.newIssueErrorMessage = ""

	// Fetch user's repos
	repos, err := fetchUserRepos()
	if err != nil {
		m.newIssueDialogMode = "error"
		m.newIssueErrorMessage = err.Error()
		return
	}

	m.newIssueRepos = repos
	m.newIssueFilteredRepos = repos
}

func (m *model) filterNewIssueRepos() {
	if m.newIssueFilterText == "" {
		m.newIssueFilteredRepos = m.newIssueRepos
		if m.newIssueSelectedRepo >= len(m.newIssueFilteredRepos) {
			m.newIssueSelectedRepo = 0
		}
		return
	}

	query := strings.ToLower(m.newIssueFilterText)
	m.newIssueFilteredRepos = nil
	for _, repo := range m.newIssueRepos {
		if fuzzyMatchRepo(repo, query) {
			m.newIssueFilteredRepos = append(m.newIssueFilteredRepos, repo)
		}
	}

	if m.newIssueSelectedRepo >= len(m.newIssueFilteredRepos) {
		m.newIssueSelectedRepo = 0
	}
}

func fuzzyMatchRepo(text, query string) bool {
	textLower := strings.ToLower(text)
	queryIdx := 0
	for _, c := range textLower {
		if queryIdx < len(query) && string(c) == string(query[queryIdx]) {
			queryIdx++
		}
	}
	return queryIdx == len(query)
}

func (m *model) executeNewIssueSelection() {
	if len(m.newIssueFilteredRepos) == 0 {
		m.newIssueDialogMode = "error"
		m.newIssueErrorMessage = "No repository selected"
		return
	}

	// Get the selected repo
	selectedRepo := m.newIssueFilteredRepos[m.newIssueSelectedRepo]

	// Convert GitHub repo name to local path
	localRepoPath := getLocalRepoPath(selectedRepo)

	// Try to find a matching tmuxinator session
	muxProject := findMatchingTmuxinatorSession(selectedRepo)
	sessionName := muxProject
	if sessionName == "" {
		sessionName = strings.ReplaceAll(selectedRepo, "/", "-")
	}

	// Check if tmux session exists
	checkCmd := exec.Command("tmux", "has-session", "-t", sessionName)
	if err := checkCmd.Run(); err != nil {
		// Session doesn't exist
		if muxProject != "" {
			// Use tmuxinator to start the project
			startCmd := exec.Command("tmuxinator", "start", muxProject, "-d")
			if err := startCmd.Run(); err != nil {
				m.newIssueDialogMode = "error"
				m.newIssueErrorMessage = "Failed to start tmuxinator project"
				return
			}
		} else {
			// Create new session manually
			createCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-n", "main")
			if err := createCmd.Run(); nil != err {
				m.newIssueDialogMode = "error"
				m.newIssueErrorMessage = "Failed to create tmux session"
				return
			}
		}
	}

	// Execute tmux command to open new window in the repo's session
	cmd := exec.Command("tmux", "new-window", "-d", "-n", "opencode-issue", "-t", sessionName, "-c", localRepoPath)
	if err := cmd.Run(); err != nil {
		m.newIssueDialogMode = "error"
		m.newIssueErrorMessage = "Failed to create tmux window"
		return
	}

	time.Sleep(500 * time.Millisecond)

	// Build the prompt with /issue (title will be entered in the new tab)
	prompt := fmt.Sprintf("--model opencode/minimax-m2.5-free --prompt \"/issue\"")
	fullCommand := fmt.Sprintf("%s %s", opencodeSecurePath, prompt)

	// Send the command to the new window in the repo's session
	cmd = exec.Command("tmux", "send-keys", "-t", fmt.Sprintf("%s:opencode-issue", sessionName), fullCommand, "Enter")
	if err := cmd.Run(); err != nil {
		m.newIssueDialogMode = "error"
		m.newIssueErrorMessage = "Failed to send command to tmux window"
		return
	}

	// Switch to the new window in the repo's session
	selectCmd := exec.Command("bash", "-c", fmt.Sprintf("tmux select-window -t %s:opencode-issue && tmux switch-client -t %s", sessionName, sessionName))
	if err := selectCmd.Run(); err != nil {
		m.newIssueDialogMode = "error"
		m.newIssueErrorMessage = "Failed to switch to tmux window"
		return
	}

	// Close the dialog
	m.showNewIssueDialog = false
	m.newIssueDialogMode = ""
	m.newIssueTitle = ""
	m.newIssueFilterText = ""
}

func (m *model) executeIssueTitleInput() {
	if m.newIssueTitle == "" {
		m.newIssueDialogMode = "error"
		m.newIssueErrorMessage = "Issue title cannot be empty"
		return
	}

	// Get the selected repo
	selectedRepo := m.newIssueFilteredRepos[m.newIssueSelectedRepo]

	// Convert GitHub repo name to local path
	localRepoPath := getLocalRepoPath(selectedRepo)

	// Try to find a matching tmuxinator session
	muxProject := findMatchingTmuxinatorSession(selectedRepo)
	sessionName := muxProject
	if sessionName == "" {
		sessionName = strings.ReplaceAll(selectedRepo, "/", "-")
	}

	// Check if tmux session exists
	checkCmd := exec.Command("tmux", "has-session", "-t", sessionName)
	if err := checkCmd.Run(); err != nil {
		// Session doesn't exist
		if muxProject != "" {
			// Use tmuxinator to start the project
			startCmd := exec.Command("tmuxinator", "start", muxProject, "-d")
			if err := startCmd.Run(); err != nil {
				m.newIssueDialogMode = "error"
				m.newIssueErrorMessage = "Failed to start tmuxinator project"
				return
			}
		} else {
			// Create new session manually
			createCmd := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-n", "main")
			if err := createCmd.Run(); nil != err {
				m.newIssueDialogMode = "error"
				m.newIssueErrorMessage = "Failed to create tmux session"
				return
			}
		}
	}

	// Execute tmux command to open new window in the repo's session
	cmd := exec.Command("tmux", "new-window", "-d", "-n", "opencode-issue", "-t", sessionName, "-c", localRepoPath)
	if err := cmd.Run(); err != nil {
		m.newIssueDialogMode = "error"
		m.newIssueErrorMessage = "Failed to create tmux window"
		return
	}

	time.Sleep(500 * time.Millisecond)

	// Build the prompt with the issue title
	prompt := fmt.Sprintf("--model opencode/minimax-m2.5-free --prompt \"/issue %s\"", m.newIssueTitle)
	fullCommand := fmt.Sprintf("%s %s", opencodeSecurePath, prompt)

	// Send the command to the new window in the repo's session
	cmd = exec.Command("tmux", "send-keys", "-t", fmt.Sprintf("%s:opencode-issue", sessionName), fullCommand, "Enter")
	_ = cmd.Run()

	// Switch to the new window in the repo's session
	selectCmd := exec.Command("bash", "-c", fmt.Sprintf("tmux select-window -t %s:opencode-issue && tmux switch-client -t %s", sessionName, sessionName))
	_ = selectCmd.Run()

	// Close the dialog
	m.showNewIssueDialog = false
	m.newIssueDialogMode = ""
	m.newIssueTitle = ""
	m.newIssueFilterText = ""
}

func getLocalRepoPath(githubRepo string) string {
	parts := strings.Split(githubRepo, "/")
	if len(parts) != 2 {
		return ""
	}
	repoName := parts[1]
	return "/home/simon/repos/" + repoName
}

func findMatchingTmuxinatorSession(repo string) string {
	homeDir := os.Getenv("HOME")
	configDir := homeDir + "/.config/tmuxinator"

	cmd := exec.Command("bash", "-c", "tmuxinator list")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(out), "\n")[1:]
	for _, line := range lines {
		fields := strings.Fields(line)
		for _, projectName := range fields {
			projectName = strings.TrimSpace(projectName)
			if projectName == "" {
				continue
			}

			configPath := configDir + "/" + projectName + ".yml"
			data, err := os.ReadFile(configPath)
			if err != nil {
				continue
			}

			content := string(data)
			for _, line := range strings.Split(content, "\n") {
				if strings.HasPrefix(line, "root:") {
					rootPath := strings.TrimSpace(strings.TrimPrefix(line, "root:"))
					rootPath = os.ExpandEnv(rootPath)
					rootPath = strings.TrimRight(rootPath, "/")

					parts := strings.Split(rootPath, "/")
					folderName := parts[len(parts)-1]

					repoOwner, repoName, _ := strings.Cut(repo, "/")
					searchName := repoName
					if searchName == "" {
						searchName = repoOwner
					}

					if strings.EqualFold(folderName, searchName) {
						return projectName
					}
				}
			}
		}
	}

	return ""
}

func fetchUserRepos() ([]string, error) {
	cmd := exec.Command("gh", "repo", "list", "--limit", "100", "--json", "nameWithOwner")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, formatGHError(fmt.Errorf("failed to list repos: %w", err))
	}

	// gh repo list --json returns an array of objects
	type repoResult []struct {
		NameWithOwner string `json:"nameWithOwner"`
	}

	var repoData repoResult
	if err := json.Unmarshal(out, &repoData); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}

	repos := make([]string, len(repoData))
	for i, r := range repoData {
		repos[i] = r.NameWithOwner
	}

	return repos, nil
}

func (m *model) renderNewIssueDialog(width, height int) string {
	dialogWidth := newIssueDialogWidth
	dialogHeight := newIssueDialogHeight

	if dialogWidth > width-4 {
		dialogWidth = width - 4
	}
	if dialogHeight > height-4 {
		dialogHeight = height - 4
	}

	var content string

	if m.newIssueDialogMode == "error" {
		content = errorStyle.Render("Error: "+m.newIssueErrorMessage) + "\n\n" +
			mutedStyle.Render("Press any key to close...")
	} else {
		// Repo selection mode
		content = titleStyle.Render("Create New Issue") + "\n\n" +
			mutedStyle.Render("Select repository:") + "\n\n"

		// Show filter text
		if m.newIssueFilterText != "" {
			content += mutedStyle.Render("Filter: ") + m.newIssueFilterText + "\n\n"
		}

		// Show repos
		maxVisible := dialogHeight - 10
		if len(m.newIssueFilteredRepos) > maxVisible {
			m.newIssueFilteredRepos = m.newIssueFilteredRepos[:maxVisible]
		}

		for i, repo := range m.newIssueFilteredRepos {
			prefix := "   "
			style := mutedStyle
			if i == m.newIssueSelectedRepo {
				prefix = " > "
				style = selectedItemStyle
			}
			repoName := repo
			if idx := strings.LastIndex(repo, "/"); idx >= 0 {
				repoName = repo[idx+1:]
			}
			content += style.Render(fmt.Sprintf("%s%s", prefix, repoName)) + "\n"
		}

		if len(m.newIssueFilteredRepos) == 0 {
			content += mutedStyle.Render("No repositories found") + "\n"
		}

		content += "\n" + mutedStyle.Render("↑/↓: navigate  1-9: select  Enter: confirm  n/Esc: close")
	}

	dialog := lipgloss.NewStyle().
		Width(dialogWidth).
		Height(dialogHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(1)

	return dialog.Render(content)
}

func (m *model) renderNewIssueDialogRaw(width, height int) string {
	dialogWidth := newIssueDialogWidth
	dialogHeight := newIssueDialogHeight

	if dialogWidth > width-4 {
		dialogWidth = width - 4
	}
	if dialogHeight > height-4 {
		dialogHeight = height - 4
	}

	var content string

	if m.newIssueDialogMode == "error" {
		content = errorStyle.Render("Error: "+m.newIssueErrorMessage) + "\n\n" +
			mutedStyle.Render("Press any key to close...")
	} else if m.newIssueDialogMode == "issue-input" {
		selectedRepo := m.newIssueFilteredRepos[m.newIssueSelectedRepo]
		repoName := selectedRepo
		if idx := strings.LastIndex(selectedRepo, "/"); idx >= 0 {
			repoName = selectedRepo[idx+1:]
		}

		content = titleStyle.Render("Create New Issue") + "\n\n" +
			mutedStyle.Render("Repository: "+repoName) + "\n\n" +
			mutedStyle.Render("Issue title:") + "\n\n" +
			selectedItemStyle.Render("  > "+m.newIssueTitle+"_") + "\n\n\n" +
			mutedStyle.Render("Enter: Skapa  |  Backspace: Ta bort  |  Esc: Avbryt")
	} else {
		content = titleStyle.Render("Create New Issue") + "\n\n" +
			mutedStyle.Render("Select repository:") + "\n\n"

		if m.newIssueFilterText != "" {
			content += mutedStyle.Render("Filter: ") + m.newIssueFilterText + "\n\n"
		}

		maxVisible := dialogHeight - 10
		if len(m.newIssueFilteredRepos) > maxVisible {
			m.newIssueFilteredRepos = m.newIssueFilteredRepos[:maxVisible]
		}

		for i, repo := range m.newIssueFilteredRepos {
			prefix := "   "
			style := mutedStyle
			if i == m.newIssueSelectedRepo {
				prefix = " > "
				style = selectedItemStyle
			}
			repoName := repo
			if idx := strings.LastIndex(repo, "/"); idx >= 0 {
				repoName = repo[idx+1:]
			}
			content += style.Render(fmt.Sprintf("%s%s", prefix, repoName)) + "\n"
		}

		if len(m.newIssueFilteredRepos) == 0 {
			content += mutedStyle.Render("No repositories found") + "\n"
		}

		content += "\n" + mutedStyle.Render("↑/↓: navigate  1-9: select  Enter: confirm  n/Esc: close")
	}

	dialog := lipgloss.NewStyle().
		Width(dialogWidth).
		Height(dialogHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(1)

	return dialog.Render(content)
}
