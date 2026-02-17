Feature: Active Window Filter
  As a developer
  I want to filter windows to show only those with active commands
  So that I can focus on actively running agent processes

  Background:
    Given I have multiple agent windows
    And some windows are active with running commands
    And some windows are idle

  Scenario: Filter to show only active windows
    Given I have the following agents:
      | name     | working_dir         | is_active |
      | OpenCode | /home/user/project1 | true      |
      | OpenCode | /home/user/project2 | false     |
      | OpenCode | /home/user/project3 | true      |
    When I filter to show only active windows
    Then I should see 2 active windows

  Scenario: Show all windows (no filter)
    Given I have the following agents:
      | name     | working_dir         | is_active |
      | OpenCode | /home/user/project1 | true      |
      | OpenCode | /home/user/project2 | false     |
    When I show all windows
    Then I should see 2 windows

  Scenario: No active windows available
    Given I have the following agents:
      | name     | working_dir         | is_active |
      | OpenCode | /home/user/project1 | false     |
      | OpenCode | /home/user/project2 | false     |
    When I filter to show only active windows
    Then I should see 0 active windows
    And I should see an indication that no windows are active

  Scenario: All windows are active
    Given I have the following agents:
      | name     | working_dir         | is_active |
      | OpenCode | /home/user/project1 | true      |
      | OpenCode | /home/user/project2 | true      |
    When I filter to show only active windows
    Then I should see 2 active windows

  Scenario: Toggle between active-only and all windows
    Given I have the following agents:
      | name     | working_dir         | is_active |
      | OpenCode | /home/user/project1 | true      |
      | OpenCode | /home/user/project2 | false     |
    When I filter to show only active windows
    Then I should see 1 active window
    When I show all windows
    Then I should see 2 windows
    When I filter to show only active windows
    Then I should see 1 active window

  Scenario: Filter performance with many windows
    Given I have 100 agent windows
    And 50 of them are active
    When I filter to show only active windows
    Then the filter operation should complete quickly

  Scenario: Empty agent list
    Given no agents are running
    When I filter to show only active windows
    Then I should see 0 active windows
    And I should see an indication that no windows are active
