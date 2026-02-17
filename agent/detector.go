package agent

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

const agentName = "OpenCode"
const agentPattern = "opencode"

var ErrNoAgentsFound = errors.New("no agents found")

type Agent struct {
	Name       string
	WorkingDir string
	PID        int
	IsActive   bool
}

func DetectAgents() ([]Agent, error) {
	pids, err := findAgentPIDs()
	if err != nil {
		return nil, err
	}
	if len(pids) == 0 {
		return nil, ErrNoAgentsFound
	}
	return buildAgentList(pids)
}

func findAgentPIDs() ([]int, error) {
	cmd := exec.Command("pgrep", "-f", agentPattern)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return ParsePIDs(string(output))
}

func ParsePIDs(output string) ([]int, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var pids []int
	for _, line := range lines {
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

func buildAgentList(pids []int) ([]Agent, error) {
	seen := make(map[string]bool)
	var agents []Agent

	for _, pid := range pids {
		workingDir, err := getWorkingDir(pid)
		if err != nil {
			continue
		}

		key := workingDir
		if key == "" {
			key = strconv.Itoa(pid)
		}
		if seen[key] {
			continue
		}
		seen[key] = true

		agents = append(agents, Agent{
			Name:       agentName,
			WorkingDir: workingDir,
			PID:        pid,
		})
	}
	return agents, nil
}

func getWorkingDir(pid int) (string, error) {
	cmd := exec.Command("pwdx", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	parts := strings.Fields(string(out))
	if len(parts) < 2 {
		return "", nil
	}
	return strings.TrimSpace(parts[1]), nil
}

func FilterActive(agents []Agent, activeOnly bool) []Agent {
	if !activeOnly {
		return agents
	}
	var filtered []Agent
	for _, a := range agents {
		if a.IsActive {
			filtered = append(filtered, a)
		}
	}
	return filtered
}
