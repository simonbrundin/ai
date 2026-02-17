# AGENTS.md - Agent Coding Guidelines for ai

## MANDATORY: Use td for Task Management

You must run td usage --new-session at conversation start (or after /clear) to see current work.
Use td usage -q for subsequent reads.

## Overview

This is a Go-based Terminal UI application that monitors OpenCode agents and GitHub issues. It uses Bubble Tea for the TUI framework and Godog for BDD testing.

**Note:** The GitHub repository is named `ai` (not `ai-tui`). Owner: `simonbrundin`.

## Build, Lint & Test Commands

### Build
```bash
go build -o ai-tui .
```

### Run Application
```bash
./ai-tui
```

### Run All Tests
```bash
go test ./...
```

### Run Single Test
```bash
go test -v -run TestName ./tests/...
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Run Godog BDD Tests
```bash
go test -v ./tests/... -run TestGodog
```

### Format Code
```bash
go fmt ./...
```

### Vet (Static Analysis)
```bash
go vet ./...
```

### Tidy Dependencies
```bash
go mod tidy
```

## Code Style Guidelines

### Imports

- Standard library first, then third-party
- Blank line between stdlib and external packages
- Group by: stdlib → external → internal

```go
import (
    "encoding/json"
    "fmt"
    "os"
    "strings"

    "ai-tui/agent"
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)
```

### Formatting

- Use `go fmt` for all code
- Chain methods on separate lines for readability:
```go
style := lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("205"))
```

### Naming Conventions

- **Variables/Functions**: `camelCase`
- **Constants**: `PascalCase` or `snake_case` for config values
- **Types**: `PascalCase`
- **Packages**: `snake_case` (e.g., `ai-tui/agent`)
- **Avoid**: Single-letter names except in loops (`i`, `j`, `k`)

### Error Handling (CRITICAL)

- **NEVER ignore errors with `_`**: Always handle errors from I/O, CLI commands, API calls
- **User-friendly messages**: Format errors so users can understand and act
- **Use fmt.Errorf with %w**: Wrap errors for proper error chains
- **Avoid variable shadowing**: Don't use `err` that shadows package names like `errors`

```go
// GOOD
agents, err := agent.DetectAgents()
if err != nil {
    return fmt.Errorf("agent detection failed: %w", err)
}

// BAD
agents, _ := agent.DetectAgents() // Never do this!
```

### Types & Structs

- Use struct tags for JSON: `json:"fieldName"`
- Keep structs focused (single responsibility)
- Use custom types for domain concepts

```go
type issue struct {
    Number int    `json:"number"`
    Title  string `json:"title"`
    State  string `json:"state"`
}
```

### Constants

- Extract magic strings/numbers to named constants
- Group related constants with `const` blocks

```go
const (
    specialPathStart    = "start"
    specialPathStdio    = "--stdio"
    specialPathWildcard = "**"
)
```

### Functions

- Keep functions small and focused
- Use descriptive names: `fetchAllIssues` not `getIssues`
- Return early when possible

### Testing

- Table-driven tests for CLI commands
- Test edge cases: empty input, errors, rate limits, auth failures
- Use testify for assertions: `assert` and `require`
- BDD tests in `tests/*.feature` files

```go
func TestFetchAllIssues_GHNotInstalled_ReturnsClearError(t *testing.T) {
    // Test implementation
}
```

## Project Structure

```
/home/simon/repos/ai/
├── main.go           # Entry point, TUI model and views
├── agent/
│   └── detector.go  # Agent detection logic
├── tests/
│   ├── *_test.go    # Unit tests
│   ├── godog_runner_test.go
│   └── *.feature    # BDD scenarios
├── go.mod
└── go.sum
```

## Git Conventions

### Commit Messages
Use Conventional Commits:
- `feat:` New feature
- `fix:` Bug fix
- `refactor:` Code improvement
- `test:` Adding tests
- `docs:` Documentation

### Branch Naming
- `feature/<short-description>`
- `bugfix/<issue-number>-<description>`
- `chore/<description>`

## Dependencies

- **github.com/charmbracelet/bubbletea** - TUI framework
- **github.com/charmbracelet/lipgloss** - Terminal styling
- **github.com/cucumber/godog** - BDD testing
- **github.com/stretchr/testify** - Test assertions

## Project-Specific Lessons (ai-tui)

### Error Handling

- Use `runGHCommand()` helper for running gh CLI commands to reduce duplication
- Format error messages to be user-friendly and actionable:
  - "gh CLI not found. Please install GitHub CLI: https://cli.github.com"
  - "GitHub not authenticated. Run 'gh auth login'"
  - "GitHub API rate limited. Please wait and try again"
- Combine multiple errors with clear context: "agent detection failed: ...; github: ..."

### Constants

- Extract magic strings to named constants at package level:
```go
const (
    specialPathStart    = "start"
    specialPathStdio    = "--stdio"
    specialPathWildcard = "**"
)
```

### Testing

- Test both happy path and edge cases (empty input, errors, rate limits, auth failures)
- Document expected vs actual behavior in test comments
- Use environment-aware tests when system state affects test results
- Use TDD: write tests first, verify they fail, then implement
- Godog Background steps require step definitions even if they do nothing

## Issue #7: Active Window Filter Lessons

- Added `IsActive` field to `Agent` struct for tracking active commands
- Created `FilterActive()` function in agent package for filtering logic
- BDD tests use real functions (`agent.FilterActive`) with mock data input
- When filtering is disabled (`activeOnly=false`), return all agents unchanged

### TUI State Management

- Add feature state to model struct for new functionality
- Use toggle keys (e.g., 'a' for active filter) in Update() method
- Display feature state in footer for user awareness

### TUI Rendering Best Practices

- Split large `render*` functions by responsibility:
  - `renderContent()` - orchestrator that chooses view
  - `renderAgentsView()` - agents tab rendering
  - `renderIssuesView()` - issues tab rendering
- Avoid duplicate conditional checks in the same function
- Test rendering logic separately from UI framework

## Issue #11: Deterministic Issue List

- Go map iteration is non-deterministic - never rely on iteration order
- Always sort map keys before iterating for stable display
- Sort issues within each group by a consistent field (number, date, etc.)
- Extract sorting logic to helper functions to avoid duplication:
  ```go
  func sortedRepoKeys(grouped map[string][]issue) []string {
      keys := make([]string, 0, len(grouped))
      for repoName := range grouped {
          keys = append(keys, repoName)
      }
      sort.Strings(keys)
      return keys
  }
  ```
- Test helpers should mirror production logic to catch regressions

## Issue #19: Dynamic Title Truncation

- TUI components should adapt to terminal dimensions (`m.width`, `m.height`)
- Extract magic numbers to named constants for rendering calculations:
  ```go
  const (
      issuePrefixWidth   = 10
      issuePadding       = 2
      issueMinTitleWidth = 10
  )
  ```
- Separate calculation logic from rendering:
  ```go
  func calculateLabelsWidth(labels []string) int
  func calculateMaxTitleWidth(terminalWidth, labelsWidth int) int
  ```
- Minimum bounds ensure usability on narrow terminals
