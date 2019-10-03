Feature: Get item view information - robustness
Background:
  Given the database has the following table 'users':
    | id | login  | temp_user | self_group_id | owned_group_id |
    | 1  | jdoe   | 0         | 11            | 12             |
    | 2  | guest  | 0         | 404           | 404            |
    | 3  | grayed | 0         | 14            | 15             |
  And the database has the following table 'groups':
    | id | name       | text_id | grade | type      |
    | 11 | jdoe       |         | -2    | UserAdmin |
    | 12 | jdoe-admin |         | -2    | UserAdmin |
    | 13 | Group B    |         | -2    | Class     |
    | 15 | gra-admin  |         | -2    | UserAdmin |
    | 14 | grayed     |         | -2    | Class     |
    | 16 | Group C    |         | -2    | Class     |
  And the database has the following table 'groups_groups':
    | id | parent_group_id | child_group_id |
    | 61 | 13              | 11             |
    | 62 | 16              | 14             |
  And the database has the following table 'groups_ancestors':
    | id | ancestor_group_id | child_group_id | is_self |
    | 71 | 11                | 11             | 1       |
    | 72 | 12                | 12             | 1       |
    | 73 | 13                | 13             | 1       |
    | 74 | 13                | 11             | 0       |
    | 75 | 16                | 14             | 0       |
  And the database has the following table 'items':
    | id  | type     | teams_editable | no_score | unlocked_item_ids | transparent_folder |
    | 190 | Category | false          | false    | 1234,2345         | true               |
    | 200 | Category | false          | false    | 1234,2345         | true               |
  And the database has the following table 'groups_items':
    | id | group_id | item_id | cached_full_access_since | cached_partial_access_since | cached_grayed_access_since | creator_user_id |
    | 42 | 13       | 190     | 2037-05-29 06:38:38      | 2037-05-29 06:38:38         | 2037-05-29 06:38:38        | 0               |
    | 43 | 13       | 200     | 2017-05-29 06:38:38      | 2017-05-29 06:38:38         | 2017-05-29 06:38:38        | 0               |
    | 44 | 16       | 190     | 2037-05-29 06:38:38      | 2037-05-29 06:38:38         | 2017-05-29 06:38:38        | 0               |
    | 45 | 16       | 200     | 2017-05-29 06:38:38      | 2017-05-29 06:38:38         | 2017-05-29 06:38:38        | 0               |
  And the database has the following table 'items_strings':
    | id | item_id | language_id | title      |
    | 53 | 200     | 1           | Category 1 |
  And the database has the following table 'users_items':
    | id | user_id | item_id | score | submissions_attempts | validated | finished | key_obtained | started_at          | finished_at         | validated_at        |
    | 1  | 1       | 200     | 12345 | 10                   | true      | true     | true         | 2019-01-30 09:26:41 | 2019-02-01 09:26:41 | 2019-01-31 09:26:41 |

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

