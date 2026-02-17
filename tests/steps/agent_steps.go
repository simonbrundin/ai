package steps

import (
	"fmt"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

// AgentState holds the state for agent detection scenarios
type AgentState struct {
	Agents []Agent
	Err    error
}

// Agent represents a running AI agent
type Agent struct {
	Name       string
	WorkingDir string
	PID        int
}

// InitializeAgentScenario sets up the agent detection step definitions
func InitializeAgentScenario(ctx *godog.ScenarioContext) {
	state := &AgentState{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &AgentState{}
	})

	// Given steps
	ctx.Step(`^the following processes are running:$`, func(table *godog.Table) error {
		state.Agents = []Agent{}
		for i, row := range table.Rows {
			if i == 0 { // Skip header
				continue
			}
			// Handle tables with 2 or 3 columns
			workingDir := ""
			if len(row.Cells) > 2 {
				workingDir = row.Cells[2].Value
			}
			agent := Agent{
				Name:       row.Cells[1].Value,
				WorkingDir: workingDir,
			}
			state.Agents = append(state.Agents, agent)
		}
		return nil
	})

	ctx.Step(`^no processes are running$`, func() error {
		state.Agents = []Agent{}
		return nil
	})

	ctx.Step(`^the process list contains malformed data$`, func() error {
		state.Err = fmt.Errorf("malformed input")
		return nil
	})

	ctx.Step(`^an "([^"]*)" agent is running in "([^"]*)"$`, func(name, dir string) error {
		state.Agents = []Agent{{Name: name, WorkingDir: dir}}
		return nil
	})

	// When steps
	ctx.Step(`^I scan for AI agents$`, func() error {
		state.Agents = filterToAgents(state.Agents)
		return nil
	})

	// Then steps
	ctx.Step(`^I should detect (\d+) agent[s]?$`, func(count int) error {
		if state.Err != nil {
			return state.Err
		}
		assert.Equal(nil, count, len(state.Agents))
		return nil
	})

	ctx.Step(`^"([^"]*)" should be in the list with working directory "([^"]*)"$`, func(name, dir string) error {
		found := false
		for _, agent := range state.Agents {
			if agent.Name == name && agent.WorkingDir == dir {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("agent '%s' with dir '%s' not found", name, dir)
		}
		return nil
	})

	ctx.Step(`^no agents should be detected$`, func() error {
		if state.Err != nil {
			return state.Err
		}
		assert.Equal(nil, 0, len(state.Agents))
		return nil
	})

	ctx.Step(`^the scan should complete without crashing$`, func() error {
		return nil
	})

	ctx.Step(`^I should see the working directory "([^"]*)" for that agent$`, func(dir string) error {
		if len(state.Agents) == 0 {
			return fmt.Errorf("no agents found")
		}
		assert.Equal(nil, dir, state.Agents[0].WorkingDir)
		return nil
	})
}
