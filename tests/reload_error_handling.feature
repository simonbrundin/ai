Feature: Reload Error Handling
  As a user
  I want to see clear error messages when reload fails
  So that I understand why data is missing without nested UI artifacts

  Background:
    Given the application is running

  Scenario: Reload fails for one repository
    Given I have multiple GitHub repositories
    When I trigger a reload
    And fetching issues from "simonbrundin/crossplane-tutorial" fails
    Then I should see an error message in the main view
    And the error should not render as nested UI (box within box)

  Scenario: Reload fails for multiple repositories
    Given I have multiple GitHub repositories
    When I trigger a reload
    And fetching issues from multiple repos fails:
      | repo                         |
      | simonbrundin/crossplane     |
      | simonbrundin/nonexistent    |
    Then I should see an error message listing all failed repos
    And the error should be displayed inline in the main content area

  Scenario: Reload succeeds after previous failures
    Given a previous reload failed for "simonbrundin/crossplane-tutorial"
    When I trigger a new reload
    And all repos fetch successfully
    Then I should see the normal content without error messages
    And previous error messages should be cleared

  Scenario: Partial failure - some repos succeed, some fail
    Given I have multiple GitHub repositories
    When I trigger a reload
    And "simonbrundin/ai" fetches successfully
    But "simonbrundin/crossplane-tutorial" fails
    Then I should see issues from "simonbrundin/ai"
    And I should see a warning about "simonbrundin/crossplane-tutorial" failure
    And the warning should be inline, not nested

  Scenario: Error message should not create nested borders
    Given I trigger a reload that fails
    When the error is displayed
    Then the TUI should have a single content border
    And the error message should be rendered within the main content area
    And there should be no box rendered inside another box
