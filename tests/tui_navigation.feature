Feature: TUI Navigation with Tabs and Keyboard Shortcuts
  As a developer using ai-tui
  I want to navigate between tabs using keyboard shortcuts
  So that I can efficiently switch between Agents and Issues views

  Background:
    Given the TUI application is running

  Scenario: Display tabs in header
    Given I have tabs available
    When the TUI renders
    Then I should see "Agents" tab
    And I should see "Issues" tab

  Scenario: Navigate to next tab with Tab key
    Given I am on the "Agents" tab
    When I press "Tab"
    Then I should be on the "Issues" tab

  Scenario: Navigate to previous tab with Shift+Tab
    Given I am on the "Issues" tab
    When I press "Shift+Tab"
    Then I should be on the "Agents" tab

  Scenario: Navigate directly to tab with number keys
    Given I have tabs numbered 1-9
    When I press "1"
    Then I should be on the first tab

  Scenario: Navigate directly to second tab with number key
    Given I have tabs numbered 1-9
    When I press "2"
    Then I should be on the second tab

  Scenario: Display keyboard shortcuts in footer
    When the TUI renders
    Then I should see "r: refresh" in footer
    And I should see "q: quit" in footer

  Scenario: Display help shortcut in footer
    When the TUI renders
    Then I should see "?:" in footer

  Scenario: Open help overlay with question mark
    Given the TUI is displayed
    When I press "?"
    Then a help overlay should open
    And the overlay should be searchable

  Scenario: Close help overlay with Escape
    Given the help overlay is open
    When I press "Escape"
    Then the help overlay should close

  Scenario: Search help commands
    Given the help overlay is open
    When I search for "refresh"
    Then I should see "r: refresh" in results

  Scenario: Tab cycling wraps around
    Given I am on the last tab
    When I press "Tab"
    Then I should be on the first tab

  Scenario: Previous tab cycling wraps around
    Given I am on the first tab
    When I press "Shift+Tab"
    Then I should be on the last tab

  Scenario: Refresh data with r key
    Given the TUI is displayed
    When I press "r"
    Then the data should refresh

  Scenario: Quit application with q key
    Given the TUI is displayed
    When I press "q"
    Then the application should quit

  Scenario: Edge case - pressing number key beyond tab count
    Given I have 2 tabs
    When I press "9"
    Then I should stay on the current tab

  Scenario: Edge case - help overlay shows all commands
    Given the help overlay is open
    Then I should see at least 3 commands listed
