Feature: Get permissions for a group - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name          | type  |
      | 21 | owner         | User  |
      | 23 | user          | User  |
      | 25 | some class    | Class |
      | 26 | another class | Class |
      | 27 | third class   | Class |
      | 31 | admin         | User  |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | user  | 23       | John        | Doe       |
      | admin | 31       | Allie       | Grater    |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_grant_group_access |
      | 23       | 21         | 1                      |
      | 25       | 21         | 0                      |
      | 25       | 31         | 1                      |
      | 26       | 21         | 1                      |
      | 26       | 31         | 1                      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 25              | 23             |
      | 25              | 31             |
      | 26              | 23             |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id  | default_language_tag |
      | 100 | fr                   |
      | 101 | fr                   |
      | 102 | fr                   |
      | 103 | fr                   |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | content_view_propagation | child_order |
      | 100            | 101           | as_info                  | 0           |
      | 101            | 102           | as_content               | 0           |
      | 102            | 103           | as_content               | 0           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 100              | 101           |
      | 100              | 102           |
      | 100              | 103           |
      | 101              | 102           |
      | 101              | 103           |
      | 102              | 103           |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 100     | solution           | solution_with_grant      | none                | none               | true               |
      | 21       | 101     | none               | none                     | none                | none               | false              |
      | 21       | 102     | none               | solution                 | none                | none               | false              |
      | 21       | 103     | none               | solution                 | none                | all_with_grant     | false              |
      | 25       | 100     | content            | none                     | answer_with_grant   | none               | false              |
      | 25       | 101     | info               | none                     | answer              | all                | false              |
      | 31       | 102     | none               | content_with_descendants | none                | none               | false              |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_view | can_grant_view           | can_enter_until     | is_owner | source_group_id | latest_update_at    |
      | 21       | 100     | none     | none                     | 9999-12-31 23:59:59 | 1        | 23              | 2019-05-30 11:00:00 |
      | 21       | 102     | none     | solution                 | 9999-12-31 23:59:59 | 1        | 23              | 2019-05-30 11:00:00 |
      | 23       | 101     | none     | none                     | 2018-05-30 11:00:00 | 0        | 26              | 2019-05-30 11:00:00 |
      | 25       | 100     | content  | none                     | 9999-12-31 23:59:59 | 0        | 23              | 2019-05-30 11:00:00 |
      | 25       | 101     | info     | none                     | 9999-12-31 23:59:59 | 0        | 23              | 2019-05-30 11:00:00 |
      | 31       | 102     | none     | content_with_descendants | 9999-12-31 23:59:59 | 0        | 31              | 2019-05-30 11:00:00 |

  Scenario: Invalid source_group_id
    Given I am the user with id "21"
    When I send a GET request to "/groups/abc/permissions/23/102"
    Then the response code should be 400
    And the response error message should contain "Wrong value for source_group_id (should be int64)"

  Scenario: Invalid group_id
    Given I am the user with id "21"
    When I send a GET request to "/groups/25/permissions/abc/102"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Invalid item_id
    Given I am the user with id "21"
    When I send a GET request to "/groups/25/permissions/23/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: The user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: The item doesn't exist
    Given I am the user with id "21"
    When I send a GET request to "/groups/25/permissions/23/404"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is not a manager of the source_group_id
    Given I am the user with id "21"
    When I send a GET request to "/groups/27/permissions/27/102"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is a manager of the source_group_id, but he doesn't have 'can_grant_group_access' permission
    Given I am the user with id "21"
    When I send a GET request to "/groups/25/permissions/25/102"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: source_group_id is not an ancestor of group_id
    Given I am the user with id "31"
    When I send a GET request to "/groups/25/permissions/26/102"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: source_group_id is a user group
    Given I am the user with id "31"
    When I send a GET request to "/groups/31/permissions/31/102"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The source group doesn't exist
    Given I am the user with id "21"
    When I send a GET request to "/groups/404/permissions/21/102"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The group doesn't exist
    Given I am the user with id "21"
    When I send a GET request to "/groups/25/permissions/404/102"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: can_grant_view = none and can_watch < answer_with_grant and can_edit < all_with_grant for the current user
    Given I am the user with id "31"
    When I send a GET request to "/groups/26/permissions/23/101"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
