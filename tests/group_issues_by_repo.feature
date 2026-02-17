Feature: Group GitHub Issues by Repository
  As a developer
  I want to see GitHub issues grouped by repository
  So that I can easily identify which repo each issue belongs to

  Background:
    Given the GitHub API is available

  Scenario: Issues are grouped by repository
    Given I have issues from multiple repositories:
      | number | title                  | repo             |
      | 1      | Fix bug in parser      | simonbrundin/ai  |
      | 2      | Add new feature        | simonbrundin/ai  |
      | 3      | Update documentation   | simonbrundin/web |
    When I group issues by repository
    Then there should be 2 repositories displayed
    And repo "ai" should have 2 issues
    And repo "web" should have 1 issue

  Scenario: Repository heading is displayed
    Given I have issues from multiple repositories:
      | number | title       | repo            |
      | 1      | Fix login   | simonbrundin/ai |
    When I group issues by repository
    Then I should see repository heading "ai"

  Scenario: Labels are preserved when grouped
    Given I have issues with labels:
      | number | title       | repo            | labels     |
      | 1      | Fix login   | simonbrundin/ai | bug        |
      | 2      | Add feature | simonbrundin/ai | enhancement |
    When I group issues by repository
    Then issue #1 should have label "bug"
    And issue #2 should have label "enhancement"

  Scenario: Empty repository still shows heading
    Given I have issues from repositories:
      | number | title       | repo            |
      | 1      | Fix login   | simonbrundin/ai |
    And repo "simonbrundin/web" has no issues
    When I group issues by repository
    Then there should be 2 repositories displayed

  Scenario: Single repository with multiple issues
    Given I have issues from one repository:
      | number | title        |
      | 1      | First issue  |
      | 2      | Second issue |
      | 3      | Third issue  |
    And they all belong to "simonbrundin/ai"
    When I group issues by repository
    Then repo "ai" should have 3 issues
