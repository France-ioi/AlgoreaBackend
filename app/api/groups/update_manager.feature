Feature: Update the group manager's permissions (groupManagerEdit)

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

  Scenario: Update without any permissions (permissions should stay unchanged)
    Given I am the user with id "21"
    When I send a PUT request to "/groups/2/managers/22" with the following body:
      """
      {
      }
      """
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "updated"
    }
    """
    And the table "group_managers" should be:
      | manager_id | group_id | can_manage            | can_grant_group_access | can_watch_members |
      | 21         | 1        | memberships_and_group | 0                      | 0                 |
      | 21         | 2        | none                  | 0                      | 0                 |
      | 22         | 1        | none                  | 0                      | 0                 |
      | 22         | 2        | memberships           | 0                      | 0                 |

  Scenario Outline: Update all permissions
    Given I am the user with id "21"
    When I send a PUT request to "/groups/2/managers/22" with the following body:
      """
      {
        "can_manage": "<can_manage>",
        "can_grant_group_access": <can_grant_group_access>,
        "can_watch_members": <can_watch_members>
      }
      """
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "updated"
    }
    """
    And the table "group_managers" should be:
      | manager_id | group_id | can_manage            | can_grant_group_access   | can_watch_members   |
      | 21         | 1        | memberships_and_group | 0                        | 0                   |
      | 21         | 2        | none                  | 0                        | 0                   |
      | 22         | 1        | none                  | 0                        | 0                   |
      | 22         | 2        | <can_manage>          | <can_grant_group_access> | <can_watch_members> |
  Examples:
    | can_manage            | can_grant_group_access | can_watch_members |
    | none                  | true                   | false             |
    | memberships           | false                  | true              |
    | memberships_and_group | true                   | true              |
