Feature: Check if a group has a path to an item
  Background:
    Given the database has the following table "groups":
      | id | name       | type  |
      | 25 | some class | Class |
    And the database has the following users:
      | group_id | login | first_name  | last_name |
      | 21       | owner | Jean-Michel | Blanquer  |
      | 23       | user  | John        | Doe       |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access |
      | 25       | 21         | 1                      |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 25              | 23             |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id  | default_language_tag |
      | 100 | fr                   |
      | 101 | fr                   |
      | 102 | fr                   |
      | 103 | fr                   |
      | 104 | fr                   |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | content_view_propagation | child_order |
      | 100            | 101           | as_info                  | 0           |
      | 101            | 102           | as_content               | 0           |
      | 102            | 103           | as_content               | 0           |
    And the database has the following table "items_ancestors":
      | ancestor_item_id | child_item_id |
      | 100              | 101           |
      | 100              | 102           |
      | 100              | 103           |
      | 101              | 102           |
      | 101              | 103           |
      | 102              | 103           |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated       |
      | 23       | 100     | content_with_descendants |
      | 23       | 101     | info                     |
      | 23       | 103     | info                     |

  Scenario: The group has a visible parent of the item
    Given I am the user with id "21"
    When I send a GET request to "/groups/23/permissions/102/has-path"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"has_path": true}
    """

  Scenario: The item itself is visible to the group
    Given I am the user with id "21"
    When I send a GET request to "/groups/23/permissions/100/has-path"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"has_path": true}
    """

  Scenario Outline: The item is a root activity or skill of one of the group's ancestors
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name                 | type  | root_activity_id   | root_skill_id   |
      | 40 | Group with root item | Class | <root_activity_id> | <root_skill_id> |
    And the database table "groups_groups" also has the following row:
      | parent_group_id | child_group_id |
      | 40              | 23             |
    And the groups ancestors are computed
    When I send a GET request to "/groups/23/permissions/104/has-path"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"has_path": true}
    """
  Examples:
    | root_activity_id | root_skill_id |
    | 104              | null          |
    | null             | 104           |

  Scenario: The group has no path to the item
    Given I am the user with id "21"
    When I send a GET request to "/groups/23/permissions/104/has-path"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"has_path": false}
    """

  Scenario: A non-existent item returns has_path false
    Given I am the user with id "21"
    When I send a GET request to "/groups/23/permissions/404/has-path"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"has_path": false}
    """

  Scenario: A manager of a grandparent group can call has-path for a descendant group
    Given the database table "groups" also has the following row:
      | id | name              | type  |
      | 30 | grandparent class | Class |
    And the database table "groups_groups" also has the following row:
      | parent_group_id | child_group_id |
      | 30              | 25             |
    And the database table "group_managers" also has the following row:
      | group_id | manager_id | can_grant_group_access |
      | 30       | 21         | 1                      |
    And the groups ancestors are computed
    And I am the user with id "21"
    When I send a GET request to "/groups/23/permissions/102/has-path"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"has_path": true}
    """
