Feature: Get qualification state (contestGetQualificationState) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name   | type     | team_item_id |
      | 10 | Team 1 | Team     | 50           |
      | 11 | Team 2 | Team     | 60           |
      | 21 | owner  | UserSelf | null         |
      | 31 | john   | UserSelf | null         |
      | 41 | jane   | UserSelf | null         |
      | 51 | jack   | UserSelf | null         |
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
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 10                | 10             | 1       |
      | 10                | 31             | 0       |
      | 10                | 41             | 0       |
      | 10                | 51             | 0       |
      | 11                | 11             | 1       |
      | 11                | 31             | 0       |
      | 11                | 41             | 0       |
      | 11                | 51             | 0       |
      | 21                | 21             | 1       |
      | 31                | 31             | 1       |
      | 41                | 41             | 1       |
      | 51                | 51             | 1       |

  Scenario: Wrong item_id
    Given I am the user with id "31"
    When I send a GET request to "/contests/abc/groups/31/qualification-state"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Wrong group_id
    Given I am the user with id "31"
    When I send a GET request to "/contests/50/groups/abc/qualification-state"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: The item is not visible to the current user (can_view = none)
    Given I am the user with id "21"
    And the database has the following table 'items':
      | id | duration | default_language_tag |
      | 50 | 00:00:01 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 21       | 50      | none               |
    When I send a GET request to "/contests/50/groups/21/qualification-state"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item is not visible to the current user (no permissions)
    Given I am the user with id "21"
    And the database has the following table 'items':
      | id | duration | default_language_tag |
      | 50 | 00:00:01 | fr                   |
    When I send a GET request to "/contests/50/groups/21/qualification-state"
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
    When I send a GET request to "/contests/50/groups/31/qualification-state"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: group_id is not a self group of the current user while the item's allows_multiple_attempts = false
    Given the database has the following table 'items':
      | id | duration | allows_multiple_attempts | default_language_tag |
      | 50 | 00:00:00 | false                    | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 10       | 50      | content                  |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/contests/50/groups/21/qualification-state"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: group_id is not a team related to the item while the item's allows_multiple_attempts = true
    Given the database has the following table 'items':
      | id | duration | allows_multiple_attempts | default_language_tag |
      | 60 | 00:00:00 | true                     | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/contests/60/groups/10/qualification-state"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: group_id is a user self group while the item's allows_multiple_attempts = true
    Given the database has the following table 'items':
      | id | duration | allows_multiple_attempts | default_language_tag |
      | 60 | 00:00:00 | true                     | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/contests/60/groups/31/qualification-state"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The current user is not a member of group_id while the item's allows_multiple_attempts = true
    Given the database has the following table 'items':
      | id | duration | allows_multiple_attempts | default_language_tag |
      | 60 | 00:00:00 | true                     | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "21"
    When I send a GET request to "/contests/60/groups/11/qualification-state"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
