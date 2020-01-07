Feature: Make a user a group manager (groupManagerCreate)

  Background:
    Given the database has the following table 'groups':
      | id | name  | type     |
      | 1  | Group | Class    |
      | 2  | Team  | Team     |
      | 21 | owner | UserSelf |
      | 22 | john  | UserSelf |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | john  | 22       | John        | Doe       |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 1                 | 1              | 1       |
      | 1                 | 2              | 0       |
      | 2                 | 2              | 1       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
    And the database has the following table 'group_managers':
      | manager_id | group_id | can_manage            |
      | 21         | 1        | memberships_and_group |

  Scenario: Add a group manager with default permissions
    Given I am the user with id "21"
    When I send a POST request to "/groups/2/managers/22" with the following body:
      """
      {
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "group_managers" should be:
      | manager_id | group_id | can_manage            | can_grant_group_access | can_watch_members |
      | 21         | 1        | memberships_and_group | 0                      | 0                 |
      | 22         | 2        | none                  | 0                      | 0                 |

  Scenario Outline: Add a group manager with custom permissions
    Given I am the user with id "21"
    When I send a POST request to "/groups/2/managers/22" with the following body:
      """
      {
        "can_manage": "<can_manage>",
        "can_grant_group_access": <can_grant_group_access>,
        "can_watch_members": <can_watch_members>
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "group_managers" should be:
      | manager_id | group_id | can_manage            | can_grant_group_access   | can_watch_members   |
      | 21         | 1        | memberships_and_group | 0                        | 0                   |
      | 22         | 2        | <can_manage>          | <can_grant_group_access> | <can_watch_members> |
  Examples:
    | can_manage            | can_grant_group_access | can_watch_members |
    | none                  | true                   | false             |
    | memberships           | false                  | true              |
    | memberships_and_group | true                   | true              |