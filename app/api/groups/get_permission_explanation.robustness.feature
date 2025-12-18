Feature: Explain permissions - robustness
  Background:
    Given the database has the following table "groups":
      | id | name          | type  |
      | 25 | some class    | Class |
      | 26 | some team     | Team  |
    And the database has the following users:
      | group_id | login | first_name  | last_name |
      | 21       | owner | Jean-Michel | Blanquer  |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id  | default_language_tag |
      | 100 | fr                   |

  Scenario: Invalid group_id
    Given I am the user with id "21"
    When I send a GET request to "/groups/abc/permissions/102/explain"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Invalid item_id
    Given I am the user with id "21"
    When I send a GET request to "/groups/25/permissions/abc/explain"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: The user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: The user has can_watch_members on group_id, but doesn't have can_watch on the item
    Given I am the user with id "21"
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 25       | 21         | true              |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_watch_generated |
      | 21       | 100     | none                |
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user has can_watch >= result on the item, but doesn't have can_watch_members on group_id
    Given I am the user with id "21"
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_watch_generated |
      | 21       | 100     | result              |
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is able to grant permissions on the item, but is not able to grant permissions to the group
    Given I am the user with id "21"
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_grant_view_generated |
      | 21       | 100     | enter                    |
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is able to grant permissions to the group, but is not able to grant permissions on the item
    Given I am the user with id "21"
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access |
      | 25       | 21         | true                   |
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is not a descendant of the group
    Given I am the user with id "21"
    When I send a GET request to "/groups/26/permissions/100/explain"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is a manager of the group, but can_manage=none
    Given I am the user with id "21"
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage |
      | 25       | 21         | none       |
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
