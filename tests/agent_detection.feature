Feature: Agent Detection
  As a developer
  I want to see which AI agents are running on my machine
  So that I can monitor active agent processes

  Scenario: Detect running agents
    Given the following processes are running:
      | pid   | command    | working_directory          |
      | 12345 | opencode   | /home/simon/repos/project-a |
      | 12346 | claude     | /home/simon/repos/project-b |
      | 12347 | bash       | /home/simon                 |
    When I scan for AI agents
    Then I should detect 2 agents
    And "opencode" should be in the list with working directory "/home/simon/repos/project-a"
    And "claude" should be in the list with working directory "/home/simon/repos/project-b"

  Scenario: No agents running
    Given the following processes are running:
      | pid   | command |
      | 12345 | bash    |
      | 12346 | zsh     |
      | 12347 | vim     |
    When I scan for AI agents
    Then no agents should be detected

  Scenario: Handle empty process list
    Given no processes are running
    When I scan for AI agents
    Then no agents should be detected

  Scenario: Handle malformed process output
    Given the process list contains malformed data
    When I scan for AI agents
    Then the scan should complete without crashing

  Scenario: Extract working directory from agent process
    Given an "opencode" agent is running in "/home/simon/repos/ai"
    When I scan for AI agents
    Then I should see the working directory "/home/simon/repos/ai" for that agent
