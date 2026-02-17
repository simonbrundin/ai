package steps

import (
	"fmt"
	"time"

	"ai-tui/agent"
	"github.com/cucumber/godog"
)

// ActiveFilterState holds the state for active window filter scenarios
type ActiveFilterState struct {
	Agents         []agent.Agent
	FilteredAgents []agent.Agent
	FilterMode     string
	StartTime      time.Time
}

// InitializeActiveFilterScenario sets up the active window filter step definitions
func InitializeActiveFilterScenario(ctx *godog.ScenarioContext) {
	state := &ActiveFilterState{FilterMode: "all"}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &ActiveFilterState{FilterMode: "all"}
	})

	// Background steps (no-op - setup context only)
	ctx.Step(`^some windows are active with running commands$`, func() error {
		return nil
	})

	ctx.Step(`^some windows are idle$`, func() error {
		return nil
	})

	// Given steps
	ctx.Step(`^I have multiple agent windows$`, func() error {
		state.Agents = []agent.Agent{}
		return nil
	})

	ctx.Step(`^I have the following agents:$`, func(table *godog.Table) error {
		state.Agents = []agent.Agent{}
		for i, row := range table.Rows {
			if i == 0 {
				continue
			}
			isActive := row.Cells[2].Value == "true"
			agentEntry := agent.Agent{
				Name:       row.Cells[0].Value,
				WorkingDir: row.Cells[1].Value,
				IsActive:   isActive,
			}
			state.Agents = append(state.Agents, agentEntry)
		}
		return nil
	})

	ctx.Step(`^no agents are running$`, func() error {
		state.Agents = []agent.Agent{}
		return nil
	})

	ctx.Step(`^I have (\d+) agent windows$`, func(count int) error {
		state.Agents = make([]agent.Agent, count)
		for i := 0; i < count; i++ {
			state.Agents[i] = agent.Agent{
				Name:       "OpenCode",
				WorkingDir: fmt.Sprintf("/home/user/project%d", i+1),
				IsActive:   false,
			}
		}
		return nil
	})

	ctx.Step(`^(\d+) of them are active$`, func(activeCount int) error {
		for i := 0; i < activeCount && i < len(state.Agents); i++ {
			state.Agents[i].IsActive = true
		}
		return nil
	})

	// When steps
	ctx.Step(`^I filter to show only active windows$`, func() error {
		state.StartTime = time.Now()
		state.FilteredAgents = agent.FilterActive(state.Agents, true)
		state.FilterMode = "active_only"
		return nil
	})

	ctx.Step(`^I show all windows$`, func() error {
		state.FilteredAgents = state.Agents
		state.FilterMode = "all"
		return nil
	})

	// Then steps
	ctx.Step(`^I should see (\d+) active window[s]?$`, func(count int) error {
		if state.FilterMode != "active_only" {
			return fmt.Errorf("expected filter mode to be 'active_only', got '%s'", state.FilterMode)
		}
		if len(state.FilteredAgents) != count {
			return fmt.Errorf("expected %d active windows, got %d", count, len(state.FilteredAgents))
		}
		return nil
	})

	ctx.Step(`^I should see (\d+) window[s]?$`, func(count int) error {
		if state.FilterMode != "all" {
			return fmt.Errorf("expected filter mode to be 'all', got '%s'", state.FilterMode)
		}
		if len(state.FilteredAgents) != count {
			return fmt.Errorf("expected %d windows, got %d", count, len(state.FilteredAgents))
		}
		return nil
	})

	ctx.Step(`^I should see an indication that no windows are active$`, func() error {
		if len(state.FilteredAgents) != 0 {
			return fmt.Errorf("expected 0 filtered agents, got %d", len(state.FilteredAgents))
		}
		return nil
	})

	ctx.Step(`^the filter operation should complete quickly$`, func() error {
		elapsed := time.Since(state.StartTime)
		// Should complete in less than 100ms for 100 agents
		if elapsed > 100*time.Millisecond {
			return fmt.Errorf("filter operation took too long: %v", elapsed)
		}
		return nil
	})
}
