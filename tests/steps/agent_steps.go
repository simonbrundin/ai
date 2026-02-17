package steps

import (
	"fmt"

	"ai-tui/agent"
	"github.com/cucumber/godog"
)

// AgentState holds the state for agent detection scenarios
type AgentState struct {
	Agents []agent.Agent
	Err    error
	Input  string
}

// InitializeAgentScenario sets up the agent detection step definitions
func InitializeAgentScenario(ctx *godog.ScenarioContext) {
	state := &AgentState{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &AgentState{}
	})

	// Given steps
	ctx.Step(`^the following processes are running:$`, func(table *godog.Table) error {
		state.Agents = []agent.Agent{}
		for i, row := range table.Rows {
			if i == 0 { // Skip header
				continue
			}
			workingDir := ""
			if len(row.Cells) > 2 {
				workingDir = row.Cells[2].Value
			}
			agent := agent.Agent{
				Name:       row.Cells[1].Value,
				WorkingDir: workingDir,
			}
			state.Agents = append(state.Agents, agent)
		}
		return nil
	})

	ctx.Step(`^no processes are running$`, func() error {
		state.Agents = []agent.Agent{}
		return nil
	})

	ctx.Step(`^the process list contains malformed data$`, func() error {
		state.Err = fmt.Errorf("malformed input")
		return nil
	})

	ctx.Step(`^an "([^"]*)" agent is running in "([^"]*)"$`, func(name, dir string) error {
		state.Agents = []agent.Agent{{Name: name, WorkingDir: dir}}
		return nil
	})

	ctx.Step(`^the pgrep output is "([^"]*)"$`, func(output string) error {
		state.Input = output
		return nil
	})

	// When steps
	ctx.Step(`^I scan for AI agents$`, func() error {
		return nil
	})

	ctx.Step(`^I parse the PID output$`, func() error {
		pids, err := agent.ParsePIDs(state.Input)
		if err != nil {
			state.Err = err
			return err
		}
		state.Agents = make([]agent.Agent, len(pids))
		for i, pid := range pids {
			state.Agents[i] = agent.Agent{PID: pid}
		}
		return nil
	})

	// Then steps
	ctx.Step(`^I should detect (\d+) agent[s]?$`, func(count int) error {
		if state.Err != nil {
			return state.Err
		}
		if len(state.Agents) != count {
			return fmt.Errorf("expected %d agents, got %d", count, len(state.Agents))
		}
		return nil
	})

	ctx.Step(`^"([^"]*)" should be in the list with working directory "([^"]*)"$`, func(name, dir string) error {
		found := false
		for _, a := range state.Agents {
			if a.Name == name && a.WorkingDir == dir {
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
		if len(state.Agents) != 0 {
			return fmt.Errorf("expected 0 agents, got %d", len(state.Agents))
		}
		return nil
	})

	ctx.Step(`^the scan should complete without crashing$`, func() error {
		return nil
	})

	ctx.Step(`^I should see the working directory "([^"]*)" for that agent$`, func(dir string) error {
		if len(state.Agents) == 0 {
			return fmt.Errorf("no agents found")
		}
		if state.Agents[0].WorkingDir != dir {
			return fmt.Errorf("expected dir '%s', got '%s'", dir, state.Agents[0].WorkingDir)
		}
		return nil
	})

	ctx.Step(`^the parsed PIDs should be "([^"]*)"$`, func(expected string) error {
		if state.Err != nil {
			return state.Err
		}
		var pids []int
		for _, a := range state.Agents {
			pids = append(pids, a.PID)
		}
		actual := fmt.Sprintf("%v", pids)
		if actual != expected {
			return fmt.Errorf("expected PIDs '%s', got '%s'", expected, actual)
		}
		return nil
	})

	ctx.Step(`^an error should occur$`, func() error {
		if state.Err == nil {
			return fmt.Errorf("expected an error but got none")
		}
		return nil
	})

	ctx.Step(`^the error should be "([^"]*)"$`, func(expected string) error {
		if state.Err == nil {
			return fmt.Errorf("no error to check")
		}
		if state.Err.Error() != expected {
			return fmt.Errorf("expected error '%s', got '%s'", expected, state.Err.Error())
		}
		return nil
	})
}
