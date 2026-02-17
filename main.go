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

var (
	commandNames   = []string{"Skriv tester", "Implementera", "Refactor", "Dokumentera", "Skapa PR"}
	commandAliases = []string{"/tdd", "/implement", "/refactor", "/docs", "/pr"}
)

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
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, m.refresh
		case "a":
			m.filterActive = !m.filterActive
			return m, nil
		case "?":
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
				m.newIssueSelectedRepo = 0
			}
			return m, nil
		case "tab":
			m.currentTab = (m.currentTab + 1) % numTabs
			return m, nil
		case "shift+tab":
			m.currentTab = (m.currentTab - 1 + numTabs) % numTabs
			return m, nil
		case "j":
			m.moveToNextIssue()
			return m, nil
		case "k":
			m.moveToPreviousIssue()
			return m, nil
		case "o":
			return m, m.openSelectedIssueInBrowser()
		case "d":
			m.showCloseIssueDialog()
			return m, nil
		case "y", "enter":
			if m.showCommandDialog {
				m.executeSelectedCommand()
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
				m.showNewIssueDialog = false
				m.newIssueFilterText = ""
				return m, nil
			}
			if m.currentTab == tabIssues {
				m.openNewIssueDialog()
			}
			m.showConfirmDialog = false
			return m, nil
		case "up":
			if m.showCommandDialog && m.selectedCommand > 0 {
				m.selectedCommand--
			}
			if m.showNewIssueDialog && m.newIssueDialogMode == "repo-select" && m.newIssueSelectedRepo > 0 {
				m.newIssueSelectedRepo--
			}
			return m, nil
		case "down":
			if m.showCommandDialog && m.selectedCommand < len(commandNames)-1 {
				m.selectedCommand++
			}
			if m.showNewIssueDialog && m.newIssueDialogMode == "repo-select" && m.newIssueSelectedRepo < len(m.newIssueFilteredRepos)-1 {
				m.newIssueSelectedRepo++
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
		if len(msg.String()) == 1 {
			key := msg.String()
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
		if m.selectedIssue < len(m.issues)-1 {
			m.selectedIssue++
		}
	}
}

func (m *model) moveToPreviousIssue() {
	if m.currentTab == tabIssues && len(m.issues) > 0 {
		if m.selectedIssue > 0 {
			m.selectedIssue--
		}
	}
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
	command := commandAliases[m.selectedCommand]
	issueNum := issue.Number

	cmd := exec.Command("tmux", "new-window", "-a", "-t", "0", "-n", fmt.Sprintf("opencode %s %d", command, issueNum),
		"bash", "-c", fmt.Sprintf("opencode --prompt '%s %d'; echo; echo 'Press Enter to close...'; read", command, issueNum))

	err := cmd.Run()
	if err != nil {
		m.err = fmt.Errorf("failed to execute command: %w", err)
	}

	err = closeGitHubIssue(issue.Repo, issueNum)
	if err != nil {
		m.err = fmt.Errorf("failed to close issue: %w", err)
	}

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

	if m.showConfirmDialog {
		return m.renderConfirmDialog(s.String())
	}

	if m.showCommandDialog {
		return m.renderCommandDialog(s.String())
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

	// Render New Issue Dialog on top of content
	if m.showNewIssueDialog {
		dialog := m.renderNewIssueDialog(width, height)
		s.WriteString("\n")
		s.WriteString(dialog)
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

		globalIndex := 0

		for _, repoName := range repoNames {
			issues := grouped[repoName]
			s.WriteString(itemStyle.Render(fmt.Sprintf("  📁 %s", repoName)))
			s.WriteString("\n")

			for _, i := range issues {
				labelsWidth := calculateLabelsWidth(i.Labels)
				labels := ""
				if len(i.Labels) > 0 {
					labels = " " + labelStyle.Render(fmt.Sprintf("[%s]", strings.Join(i.Labels, ", ")))
				}
				maxTitleWidth := calculateMaxTitleWidth(m.width, labelsWidth)

				prefix := "    "
				currentStyle := itemStyle
				if globalIndex == m.selectedIssue {
					prefix = "  > "
					currentStyle = selectedItemStyle
				}

				s.WriteString(currentStyle.Render(fmt.Sprintf("%s#%d %s%s", prefix, i.Number, truncate(i.Title, maxTitleWidth), labels)))
				s.WriteString("\n")

				globalIndex++
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

	// Execute tmux command to open new window with opencode-secure
	cmd := exec.Command("tmux", "new-window", "-d", "-n", "opencode-issue")
	if err := cmd.Run(); err != nil {
		m.newIssueDialogMode = "error"
		m.newIssueErrorMessage = "Failed to create tmux window"
		return
	}

	// Send the command to the new window
	fullCommand := fmt.Sprintf("%s %s", opencodeSecurePath, opencodeIssuePrompt)
	cmd = exec.Command("tmux", "send-keys", "-t", "opencode-issue", fullCommand, "Enter")
	if err := cmd.Run(); err != nil {
		m.newIssueDialogMode = "error"
		m.newIssueErrorMessage = "Failed to run opencode-secure"
		return
	}

	// Close the dialog
	m.showNewIssueDialog = false
	m.newIssueFilterText = ""
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
			content += style.Render(fmt.Sprintf("%s%s", prefix, repo)) + "\n"
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

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
		dialog.Render(content))
}
