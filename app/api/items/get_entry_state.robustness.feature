Feature: Get entry state (itemGetEntryState) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name   | type |
      | 10 | Team 1 | Team |
      | 11 | Team 2 | Team |
      | 21 | owner  | User |
      | 31 | john   | User |
      | 41 | jane   | User |
      | 51 | jack   | User |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | john  | 31       | John        | Doe       |
      | jane  | 41       | Jane        | null      |
      | jack  | 51       | Jack        | Daniel    |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 10              | 31             |
      | 10              | 41             |
      | 10              | 51             |
      | 11              | 31             |
      | 11              | 41             |
      | 11              | 51             |
    And the groups ancestors are computed

  Scenario: Wrong item_id
    Given I am the user with id "31"
    When I send a GET request to "/items/abc/entry-state"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Wrong as_team_id
    Given I am the user with id "31"
    When I send a GET request to "/items/50/entry-state?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: The item is not visible to the current user (can_view = none)
    Given I am the user with id "21"
    And the database has the following table 'items':
      | id | requires_explicit_entry | default_language_tag |
      | 50 | 1                       | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 21       | 50      | none               |
    When I send a GET request to "/items/50/entry-state"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item is not visible to the current user (no permissions)
    Given I am the user with id "21"
    And the database has the following table 'items':
      | id | requires_explicit_entry | default_language_tag |
      | 50 | 1                       | fr                   |
    When I send a GET request to "/items/50/entry-state"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item is visible, but it's not a contest
    Given the database has the following table 'items':
      | id | default_language_tag |
      | 50 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 21       | 50      | info               |
    And I am the user with id "31"
    When I send a GET request to "/items/50/entry-state"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: as_team_id is given while the item's entry_participant_type = User
    Given the database has the following table 'items':
      | id | requires_explicit_entry | entry_participant_type | default_language_tag |
      | 50 | 1                       | User                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 10       | 50      | content                  |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/50/entry-state?as_team_id=10"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: as_team_id is not given while the item's entry_participant_type = Team
    Given the database has the following table 'items':
      | id | requires_explicit_entry | entry_participant_type | default_language_tag |
      | 50 | 1                       | Team                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 10       | 50      | content                  |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/50/entry-state"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The current user is not a member of as_team_id while the item's entry_participant_type = 'Team'
    Given the database has the following table 'items':
      | id | requires_explicit_entry | entry_participant_type | default_language_tag |
      | 60 | 1                       | Team                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "21"
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
