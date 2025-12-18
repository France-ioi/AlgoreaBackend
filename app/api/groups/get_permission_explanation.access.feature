Feature: Explain permissions - access
  Background:
    Given the database has the following table "groups":
      | id | name          | type  |
      | 25 | some class    | Class |
      | 26 | some team     | Team  |
      | 27 | some club     | Club  |
    And the database has the following users:
      | group_id | login | first_name  | last_name |
      | 21       | owner | Jean-Michel | Blanquer  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 26              | 21             |
      | 27              | 21             |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id  | default_language_tag |
      | 100 | fr                   |

  Scenario Outline: The user has can_watch_members on group_id and can_watch on the item
    Given I am the user with id "21"
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 25       | 21         | true              |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_watch_generated   |
      | 21       | 100     | <can_watch_generated> |
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """
  Examples:
    | can_watch_generated |
    | result              |
    | answer              |
    | answer_with_grant   |

  Scenario Outline: The user is able to grant permissions on the item and grant permissions to the group
    Given I am the user with id "21"
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_grant_view_generated   |
      | 21       | 100     | <can_grant_view_generated> |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access |
      | 25       | 21         | true                   |
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """
  Examples:
    | can_grant_view_generated |
    | enter                    |
    | content                  |
    | content_with_descendants |
    | solution                 |
    | solution_with_grant      |

  Scenario: The user is a descendant of the group
    Given I am the user with id "21"
    When I send a GET request to "/groups/27/permissions/100/explain"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: The user is a team member of the group
    Given I am the user with id "21"
    When I send a GET request to "/groups/26/permissions/100/explain"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario Outline: The user is a manager of the group with can_manage>=memberships
    Given I am the user with id "21"
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage   |
      | 25       | 21         | <can_manage> |
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """
  Examples:
    | can_manage            |
    | memberships           |
    | memberships_and_group |
