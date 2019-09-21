Feature: Get item view information - robustness
Background:
  Given the database has the following table 'users':
    | id | login  | temp_user | self_group_id | owned_group_id | version |
    | 1  | jdoe   | 0         | 11            | 12             | 0       |
    | 2  | guest  | 0         | 404           | 404            | 0       |
    | 3  | grayed | 0         | 14            | 15             | 0       |
  And the database has the following table 'groups':
    | id | name       | text_id | grade | type      | version |
    | 11 | jdoe       |         | -2    | UserAdmin | 0       |
    | 12 | jdoe-admin |         | -2    | UserAdmin | 0       |
    | 13 | Group B    |         | -2    | Class     | 0       |
    | 15 | gra-admin  |         | -2    | UserAdmin | 0       |
    | 14 | grayed     |         | -2    | Class     | 0       |
    | 16 | Group C    |         | -2    | Class     | 0       |
  And the database has the following table 'groups_groups':
    | id | parent_group_id | child_group_id | version |
    | 61 | 13              | 11             | 0       |
    | 62 | 16              | 14             | 0       |
  And the database has the following table 'groups_ancestors':
    | id | ancestor_group_id | child_group_id | is_self | version |
    | 71 | 11                | 11             | 1       | 0       |
    | 72 | 12                | 12             | 1       | 0       |
    | 73 | 13                | 13             | 1       | 0       |
    | 74 | 13                | 11             | 0       | 0       |
    | 75 | 16                | 14             | 0       | 0       |
  And the database has the following table 'items':
    | id  | type     | teams_editable | no_score | unlocked_item_ids | transparent_folder | version |
    | 190 | Category | false          | false    | 1234,2345         | true               | 0       |
    | 200 | Category | false          | false    | 1234,2345         | true               | 0       |
  And the database has the following table 'groups_items':
    | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | creator_user_id | version |
    | 42 | 13       | 190     | 2037-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    | 43 | 13       | 200     | 2017-05-29 06:38:38     | 2017-05-29 06:38:38        | 2017-05-29 06:38:38       | 0               | 0       |
    | 44 | 16       | 190     | 2037-05-29 06:38:38     | 2037-05-29 06:38:38        | 2017-05-29 06:38:38       | 0               | 0       |
    | 45 | 16       | 200     | 2017-05-29 06:38:38     | 2017-05-29 06:38:38        | 2017-05-29 06:38:38       | 0               | 0       |
  And the database has the following table 'items_strings':
    | id | item_id | language_id | title      | version |
    | 53 | 200     | 1           | Category 1 | 0       |
  And the database has the following table 'users_items':
    | id | user_id | item_id | score | submissions_attempts | validated | finished | key_obtained | start_date          | finish_date         | validation_date     | version |
    | 1  | 1       | 200     | 12345 | 10                   | true      | true     | true         | 2019-01-30 09:26:41 | 2019-02-01 09:26:41 | 2019-01-31 09:26:41 | 0       |

  Scenario: Should fail when the root item is invalid
    Given I am the user with id "1"
    When I send a GET request to "/items/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Should fail when the user doesn't have access to the root item
    Given I am the user with id "1"
    When I send a GET request to "/items/190"
    Then the response code should be 404
    And the response error message should contain "Insufficient access rights on the given item id"

  Scenario: Should fail when the user doesn't have access to the root item (for a user with a non-existent group)
    Given I am the user with id "2"
    When I send a GET request to "/items/200"
    Then the response code should be 404
    And the response error message should contain "Insufficient access rights on the given item id"

  Scenario: Should fail when the root item doesn't exist
    Given I am the user with id "1"
    When I send a GET request to "/items/404"
    Then the response code should be 404
    And the response error message should contain "Insufficient access rights on the given item id"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with id "4"
    When I send a GET request to "/items/200"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when the user has only grayed access rights to the root item
    Given I am the user with id "3"
    When I send a GET request to "/items/190"
    Then the response code should be 403
    And the response error message should contain "The item is grayed"

