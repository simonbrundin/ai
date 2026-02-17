Feature: Agent Detection
  As a developer
  I want to see which AI agents are running on my machine
  So that I can monitor active agent processes

  Scenario Outline: Parse valid PID output
    Given the pgrep output is "<input>"
    When I parse the PID output
    Then the parsed PIDs should be "<output>"
    Examples:
      | input        | output                |
      | 12345        | [12345]               |
      | 12345\n12346 | [12345 12346]         |
      | 12345\n12346\n12347 | [12345 12346 12347] |

  Scenario: Parse empty PID output
    Given the pgrep output is ""
    When I parse the PID output
    Then I should detect 0 agents

  Scenario Outline: Parse PID output with empty lines
    Given the pgrep output is "<input>"
    When I parse the PID output
    Then the parsed PIDs should be "<output>"
    Examples:
      | input       | output          |
      | 12345\n    | [12345]         |
      | \n12345    | [12345]         |
      | 12345\n\n12347 | [12345 12347] |

  Scenario Outline: Parse PID output with invalid values
    Given the pgrep output is "<input>"
    When I parse the PID output
    Then the parsed PIDs should be "<output>"
    Examples:
      | input           | output          |
      | abc             | []              |
      | 12345\nabc     | [12345]         |
      | abc\n12345\nxyz | [12345]         |

  Scenario: No agents running
    Given the pgrep output is ""
    When I parse the PID output
    Then no agents should be detected

  Scenario: Handle malformed process output
    Given the process list contains malformed data
    When I scan for AI agents
    Then the scan should complete without crashing
