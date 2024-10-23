Feature: Remove a group manager (groupManagerDelete)

  Background:
    Given the database has the following table "groups":
      | id | name  | type  |
      | 1  | Group | Class |
      | 2  | Team  | Team  |
    And the database has the following users:
      | group_id | login | first_name  | last_name |
      | 21       | owner | Jean-Michel | Blanquer  |
      | 22       | john  | John        | Doe       |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 1               | 2              |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | manager_id | group_id | can_manage            |
      | 21         | 1        | memberships_and_group |
      | 21         | 2        | none                  |
      | 22         | 1        | none                  |
      | 22         | 2        | memberships           |

  Scenario: The current user has permissions to delete a manager
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/2/managers/22"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted"
    }
    """
    And the table "group_managers" should be:
      | manager_id | group_id | can_manage            | can_grant_group_access | can_watch_members |
      | 21         | 1        | memberships_and_group | 0                      | 0                 |
      | 21         | 2        | none                  | 0                      | 0                 |
      | 22         | 1        | none                  | 0                      | 0                 |

  Scenario: The current user doesn't have permissions to delete a manager, but his id is the manager_id
    Given I am the user with id "22"
    When I send a DELETE request to "/groups/2/managers/22"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted"
    }
    """
    And the table "group_managers" should be:
      | manager_id | group_id | can_manage            | can_grant_group_access | can_watch_members |
      | 21         | 1        | memberships_and_group | 0                      | 0                 |
      | 21         | 2        | none                  | 0                      | 0                 |
      | 22         | 1        | none                  | 0                      | 0                 |
