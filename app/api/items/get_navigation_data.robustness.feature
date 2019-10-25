Feature: Get item for tree navigation - robustness
Background:
  Given the database has the following table 'groups':
    | id | name       | text_id | grade | type      |
    | 11 | jdoe       |         | -2    | UserSelf  |
    | 12 | jdoe-admin |         | -2    | UserAdmin |
    | 13 | Group B    |         | -2    | Class     |
  And the database has the following table 'users':
    | login | temp_user | group_id | owned_group_id |
    | jdoe  | 0         | 11       | 12             |
  And the database has the following table 'groups_groups':
    | id | parent_group_id | child_group_id |
    | 61 | 13              | 11             |
  And the database has the following table 'groups_ancestors':
    | id | ancestor_group_id | child_group_id | is_self |
    | 71 | 11                | 11             | 1       |
    | 72 | 12                | 12             | 1       |
    | 73 | 13                | 13             | 1       |
    | 74 | 13                | 11             | 0       |
  And the database has the following table 'items':
    | id  | type     | teams_editable | no_score | unlocked_item_ids | transparent_folder |
    | 190 | Category | false          | false    | 1234,2345         | true               |
    | 200 | Category | false          | false    | 1234,2345         | true               |
  And the database has the following table 'groups_items':
    | id | group_id | item_id | cached_full_access_since | cached_partial_access_since | cached_grayed_access_since |
    | 42 | 13       | 190     | 2037-05-29 06:38:38      | 2037-05-29 06:38:38         | 2037-05-29 06:38:38        |
    | 43 | 13       | 200     | 2017-05-29 06:38:38      | 2017-05-29 06:38:38         | 2017-05-29 06:38:38        |
  And the database has the following table 'items_strings':
    | id | item_id | language_id | title      |
    | 53 | 200     | 1           | Category 1 |

  Scenario: Should fail when the user doesn't have access to the root item
    Given I am the user with group_id "11"
    When I send a GET request to "/items/190/as-nav-tree"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item id"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with group_id "404"
    When I send a GET request to "/items/190/as-nav-tree"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when the root item doesn't exist
    Given I am the user with group_id "11"
    When I send a GET request to "/items/404/as-nav-tree"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on given item id"

  Scenario: Invalid item_id
    Given I am the user with group_id "11"
    When I send a GET request to "/items/abc/as-nav-tree"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
