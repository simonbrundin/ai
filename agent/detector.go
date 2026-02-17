package agent

import (
	"os/exec"
	"strings"
)

var knownAgents = []string{"opencode", "claude", "claude-code", "aider", "devin"}

type Agent struct {
	Name       string
	WorkingDir string
	PID        int
}

func DetectAgents() ([]Agent, error) {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseProcessList(string(output)), nil
}

func parseProcessList(output string) []Agent {
	var agents []Agent
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 11 {
			continue
		}

		cmd := parts[10]
		for _, agent := range knownAgents {
			if strings.Contains(cmd, agent) {
				workingDir := extractWorkingDir(parts)
				agents = append(agents, Agent{
					Name:       agent,
					WorkingDir: workingDir,
				})
				break
			}
		}
	}
	return agents
}

func extractWorkingDir(parts []string) string {
	if len(parts) > 10 {
		return parts[len(parts)-1]
	}
	return ""
}

func FilterKnownAgents(processes []Agent) []Agent {
	var agents []Agent
	for _, p := range processes {
		for _, agent := range knownAgents {
			if p.Name == agent {
				agents = append(agents, p)
				break
			}
		}
	}
	return agents
}
