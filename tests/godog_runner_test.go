package tests

import (
	"testing"

	"ai-tui/tests/steps"
	"github.com/cucumber/godog"
)

func TestGodog(t *testing.T) {
	opts := &godog.Options{
		Format:   "pretty",
		Paths:    []string{"."},
		Strict:   true,
		NoColors: false,
	}

	t.Run("github_issues", func(t *testing.T) {
		suite := godog.TestSuite{
			Name:                 "github features",
			TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {},
			ScenarioInitializer: func(ctx *godog.ScenarioContext) {
				steps.InitializeGitHubScenario(ctx)
				steps.InitializeAgentScenario(ctx)
			},
			Options: opts,
		}

		status := suite.Run()
		if status != 0 {
			t.Errorf("godog tests failed with status: %d", status)
		}
	})

	t.Run("tui_navigation", func(t *testing.T) {
		suite := godog.TestSuite{
			Name:                 "tui navigation features",
			TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {},
			ScenarioInitializer: func(ctx *godog.ScenarioContext) {
				steps.InitializeTUIScenario(ctx)
			},
			Options: opts,
		}

		status := suite.Run()
		if status != 0 {
			t.Errorf("godog tests failed with status: %d", status)
		}
	})

	t.Run("active_filter", func(t *testing.T) {
		suite := godog.TestSuite{
			Name:                 "active filter features",
			TestSuiteInitializer: func(ctx *godog.TestSuiteContext) {},
			ScenarioInitializer: func(ctx *godog.ScenarioContext) {
				steps.InitializeActiveFilterScenario(ctx)
			},
			Options: opts,
		}

		status := suite.Run()
		if status != 0 {
			t.Errorf("godog tests failed with status: %d", status)
		}
	})
}
