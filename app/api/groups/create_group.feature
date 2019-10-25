Feature: Create a group (groupCreate)

  Background:
    Given the database has the following table 'groups':
      | id | name        | type      |
      | 21 | owner       | UserSelf  |
      | 22 | owner-admin | UserAdmin |
    And the database has the following table 'users':
      | login | temp_user | group_id | owned_group_id | first_name  | last_name | allow_subgroups |
      | owner | 0         | 21       | 22             | Jean-Michel | Blanquer  | 1               |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_full_access_since | cached_partial_access_since | cached_grayed_access_since |
      | 21       | 10      | 2019-07-16 21:28:47      | null                        | null                       |
      | 21       | 11      | null                     | 2019-07-16 21:28:47         | null                       |
      | 21       | 12      | null                     | null                        | 2019-07-16 21:28:47        |

  Scenario Outline: Create a group
    Given I am the user with group_id "21"
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
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id      | child_order | type   | role  |
      | 22              | 5577006791947779410 | 1           | direct | owner |
    And the table "groups_ancestors" should be:
      | ancestor_group_id   | child_group_id      | is_self |
      | 21                  | 21                  | 1       |
      | 22                  | 22                  | 1       |
      | 22                  | 5577006791947779410 | 0       |
      | 5577006791947779410 | 5577006791947779410 | 1       |
  Examples:
    | group_type | item_spec         | want_item_id |
    | Class      |                   | null         |
    | Team       |                   | null         |
    | Team       | , "item_id": "10" | 10           | # full access
    | Team       | , "item_id": "11" | 11           | # partial access
    | Team       | , "item_id": "12" | 12           | # grayed access
    | Club       |                   | null         |
    | Friends    |                   | null         |
    | Other      |                   | null         |
