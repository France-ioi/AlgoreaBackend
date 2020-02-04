Feature: Create a group (groupCreate)

  Background:
    Given the database has the following table 'groups':
      | id | name  | type |
      | 21 | owner | User |
    And the database has the following table 'users':
      | login | temp_user | group_id | first_name  | last_name | allow_subgroups |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | 1               |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 21                | 21             |
    And the database has the following table 'items':
      | id | default_language_tag |
      | 10 | fr                   |
      | 11 | fr                   |
      | 12 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 21       | 10      | content_with_descendants |
      | 21       | 11      | content                  |
      | 21       | 12      | info                     |

  Scenario Outline: Create a group
    Given I am the user with id "21"
    When I send a POST request to "/groups" with the following body:
    """
    {
      "name": "some name",
      "type": "<group_type>"
      <item_spec>
    }
    """
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": {"id":"5577006791947779410"}
    }
    """
    And the table "groups" should stay unchanged but the row with id "5577006791947779410"
    And the table "groups" at id "5577006791947779410" should be:
      | id                  | name      | type         | team_item_id   | TIMESTAMPDIFF(SECOND, NOW(), created_at) < 3 |
      | 5577006791947779410 | some name | <group_type> | <want_item_id> | true                                         |
    And the table "group_managers" should be:
      | manager_id | group_id            | can_manage            | can_grant_group_access | can_watch_members |
      | 21         | 5577006791947779410 | memberships_and_group | 1                      | 1                 |
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should be:
      | ancestor_group_id   | child_group_id      | is_self |
      | 21                  | 21                  | 1       |
      | 5577006791947779410 | 5577006791947779410 | 1       |
  Examples:
    | group_type | item_spec         | want_item_id |
    | Class      |                   | null         |
    | Team       |                   | null         |
    | Team       | , "item_id": "10" | 10           | # full access
    | Team       | , "item_id": "11" | 11           | # content access
    | Team       | , "item_id": "12" | 12           | # info access
    | Club       |                   | null         |
    | Friends    |                   | null         |
    | Other      |                   | null         |
