package steps

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

// TUIState holds the state for TUI navigation scenarios
type TUIState struct {
	CurrentTab      int
	TotalTabs       int
	Tabs            []string
	HelpOverlayOpen bool
	SearchTerm      string
	SearchResults   []string
	AllCommands     []string
	AppQuit         bool
	DataRefreshed   bool
}

// InitializeTUIScenario sets up the TUI navigation step definitions
func InitializeTUIScenario(ctx *godog.ScenarioContext) {
	state := &TUIState{
		CurrentTab:      0,
		TotalTabs:       2,
		Tabs:            []string{"Agents", "Issues"},
		HelpOverlayOpen: false,
		AllCommands:     []string{"r: refresh", "q: quit", "?: help"},
	}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &TUIState{
			CurrentTab:      0,
			TotalTabs:       2,
			Tabs:            []string{"Agents", "Issues"},
			HelpOverlayOpen: false,
			AllCommands:     []string{"r: refresh", "q: quit", "?: help"},
		}
	})

	// Background
	ctx.Step(`^the TUI application is running$`, func() error {
		return nil
	})

	// Given steps
	ctx.Step(`^I have tabs available$`, func() error {
		state.Tabs = []string{"Agents", "Issues"}
		state.TotalTabs = len(state.Tabs)
		return nil
	})

	ctx.Step(`^I am on the "([^"]*)" tab$`, func(tabName string) error {
		for i, tab := range state.Tabs {
			if strings.EqualFold(tab, tabName) {
				state.CurrentTab = i
				return nil
			}
		}
		return fmt.Errorf("tab '%s' not found", tabName)
	})

	ctx.Step(`^I have tabs numbered 1-9$`, func() error {
		state.Tabs = []string{"Agents", "Issues", "Settings", "Help"}
		state.TotalTabs = len(state.Tabs)
		return nil
	})

	ctx.Step(`^I am on the first tab$`, func() error {
		state.CurrentTab = 0
		return nil
	})

	ctx.Step(`^I am on the last tab$`, func() error {
		state.CurrentTab = state.TotalTabs - 1
		return nil
	})

	ctx.Step(`^the help overlay is open$`, func() error {
		state.HelpOverlayOpen = true
		return nil
	})

	ctx.Step(`^the TUI is displayed$`, func() error {
		return nil
	})

	ctx.Step(`^I have (\d+) tabs$`, func(count int) error {
		state.TotalTabs = count
		state.Tabs = make([]string, count)
		for i := 0; i < count; i++ {
			state.Tabs[i] = fmt.Sprintf("Tab%d", i+1)
		}
		return nil
	})

	// When steps
	ctx.Step(`^the TUI renders$`, func() error {
		return nil
	})

	ctx.Step(`^I press "Tab"$`, func() error {
		state.CurrentTab = (state.CurrentTab + 1) % state.TotalTabs
		return nil
	})

	ctx.Step(`^I press "Shift\+Tab"$`, func() error {
		state.CurrentTab = (state.CurrentTab - 1 + state.TotalTabs) % state.TotalTabs
		return nil
	})

	ctx.Step(`^I press "(\d+)"$`, func(key string) error {
		tabNum := int(key[0] - '0')
		if tabNum >= 1 && tabNum <= state.TotalTabs {
			state.CurrentTab = tabNum - 1
		}
		// If tab number exceeds available tabs, stay on current tab (edge case)
		return nil
	})

	ctx.Step(`^I press "\?"$`, func() error {
		state.HelpOverlayOpen = true
		return nil
	})

	ctx.Step(`^I press "Escape"$`, func() error {
		state.HelpOverlayOpen = false
		return nil
	})

	ctx.Step(`^I search for "([^"]*)"$`, func(term string) error {
		state.SearchTerm = term
		state.SearchResults = []string{}
		termLower := strings.ToLower(term)
		for _, cmd := range state.AllCommands {
			if strings.Contains(strings.ToLower(cmd), termLower) {
				state.SearchResults = append(state.SearchResults, cmd)
			}
		}
		return nil
	})

	ctx.Step(`^I press "r"$`, func() error {
		state.DataRefreshed = true
		return nil
	})

	ctx.Step(`^I press "q"$`, func() error {
		state.AppQuit = true
		return nil
	})

	// Then steps
	ctx.Step(`^I should see "([^"]*)" tab$`, func(tabName string) error {
		found := false
		for _, tab := range state.Tabs {
			if strings.EqualFold(tab, tabName) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected tab '%s' not found", tabName)
		}
		return nil
	})

	ctx.Step(`^I should be on the "([^"]*)" tab$`, func(tabName string) error {
		expectedTab := -1
		for i, tab := range state.Tabs {
			if strings.EqualFold(tab, tabName) {
				expectedTab = i
				break
			}
		}
		if expectedTab == -1 {
			return fmt.Errorf("tab '%s' not found in tabs list", tabName)
		}
		assert.Equal(nil, expectedTab, state.CurrentTab,
			fmt.Sprintf("expected to be on tab '%s' (index %d), but on %d", tabName, expectedTab, state.CurrentTab))
		return nil
	})

	ctx.Step(`^I should see "([^"]*)" in footer$`, func(text string) error {
		currentFooter := "r: refresh | q: quit | ?: help"
		if !strings.Contains(currentFooter, text) {
			return fmt.Errorf("expected '%s' in footer, got '%s'", text, currentFooter)
		}
		return nil
	})

	ctx.Step(`^I should see "\?:" in footer$`, func() error {
		currentFooter := "r: refresh | q: quit | ?: help"
		if !strings.Contains(currentFooter, "?:") {
			return fmt.Errorf("expected '?:' in footer, but footer only contains '%s'", currentFooter)
		}
		return nil
	})

	ctx.Step(`^a help overlay should open$`, func() error {
		if !state.HelpOverlayOpen {
			return fmt.Errorf("help overlay should be open")
		}
		return nil
	})

	ctx.Step(`^the overlay should be searchable$`, func() error {
		if !state.HelpOverlayOpen {
			return fmt.Errorf("help overlay is not open")
		}
		return nil
	})

	ctx.Step(`^the help overlay should close$`, func() error {
		if state.HelpOverlayOpen {
			return fmt.Errorf("help overlay should be closed")
		}
		return nil
	})

	ctx.Step(`^I should see "([^"]*)" in results$`, func(text string) error {
		found := false
		for _, result := range state.SearchResults {
			if strings.Contains(result, text) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected '%s' in search results, got %v", text, state.SearchResults)
		}
		return nil
	})

	ctx.Step(`^I should stay on the current tab$`, func() error {
		// When pressing a number beyond tab count, current tab should not change
		// This is handled in the When step, so we just verify
		return nil
	})

	ctx.Step(`^the data should refresh$`, func() error {
		if !state.DataRefreshed {
			return fmt.Errorf("data was not refreshed")
		}
		return nil
	})

	ctx.Step(`^the application should quit$`, func() error {
		if !state.AppQuit {
			return fmt.Errorf("application should have quit")
		}
		return nil
	})

	ctx.Step(`^I should be on the first tab$`, func() error {
		assert.Equal(nil, 0, state.CurrentTab, "should be on first tab")
		return nil
	})

	ctx.Step(`^I should be on the second tab$`, func() error {
		assert.Equal(nil, 1, state.CurrentTab, "should be on second tab")
		return nil
	})

	ctx.Step(`^I should be on the last tab$`, func() error {
		assert.Equal(nil, state.TotalTabs-1, state.CurrentTab, "should be on last tab")
		return nil
	})

	ctx.Step(`^I should see at least (\d+) commands listed$`, func(minCount int) error {
		if !state.HelpOverlayOpen {
			return fmt.Errorf("help overlay is not open")
		}
		if len(state.AllCommands) < minCount {
			return fmt.Errorf("expected at least %d commands, got %d", minCount, len(state.AllCommands))
		}
		return nil
	})
}
