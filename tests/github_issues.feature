Feature: GitHub Issues Display
  As a developer
  I want to view GitHub issues in a TUI
  So that I can track open issues without leaving my terminal

  Background:
    Given the GitHub API is available

  Scenario: Fetch open issues from a repository
    Given I have a GitHub repository "simonbrundin/ai"
    When I request all open issues
    Then I should receive a list of issues
    And each issue should have a number, title, and state

  Scenario: Filter issues by label
    Given I have the following issues:
      | number | title                          | labels        |
      | 1      | feat: Add new feature          | enhancement   |
      | 2      | fix: Bug in login              | bug           |
      | 3      | docs: Update README            | documentation |
    When I filter by label "enhancement"
    Then I should see 1 issue
    And issue #1 should be in the result

  Scenario: Filter issues by search term
    Given I have the following issues:
      | number | title                    |
      | 1      | feat: Add new feature    |
      | 2      | fix: Bug in login        |
      | 3      | feat: Improve performance |
    When I search for "feat"
    Then I should see 2 issues

  Scenario: Handle empty repository
    Given the repository "simonbrundin/empty" has no issues
    When I request all open issues
    Then the result should be empty

  Scenario: Handle rate limiting from GitHub API
    Given GitHub API returns rate limit exceeded
    When I request all open issues
    Then I should get an error with "rate limit"
    And I should see retry information

  Scenario: Handle network errors
    Given the network is unavailable
    When I request all open issues
    Then I should get a connection error
