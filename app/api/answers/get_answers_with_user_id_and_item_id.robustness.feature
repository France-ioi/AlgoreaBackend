Feature: Get item answers with (item_id, user_id) pair - robustness
Background:
  Given the database has the following table 'users':
    | id | login | temp_user | self_group_id | owned_group_id |
    | 1  | jdoe  | 0         | 11            | 12             |
    | 2  | guest | 0         | 404           | 404            |
  And the database has the following table 'groups':
    | id | name       | text_id | grade | type      |
    | 11 | jdoe       |         | -2    | UserAdmin |
    | 12 | jdoe-admin |         | -2    | UserAdmin |
    | 13 | Group B    |         | -2    | Class     |
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
    | 210 | Category | false          | false    | 1234,2345         | true               |
  And the database has the following table 'groups_items':
    | id | group_id | item_id | cached_full_access_since | cached_partial_access_since | cached_grayed_access_since | creator_user_id |
    | 42 | 13       | 190     | 2037-05-29 06:38:38      | 2037-05-29 06:38:38         | 2037-05-29 06:38:38        | 0               |
    | 43 | 13       | 200     | 2017-05-29 06:38:38      | 2017-05-29 06:38:38         | 2017-05-29 06:38:38        | 0               |
    | 44 | 13       | 210     | 2037-05-29 06:38:38      | 2037-05-29 06:38:38         | 2017-05-29 06:38:38        | 0               |

  Scenario: Should fail when the user has only grayed access to the item
    Given I am the user with id "1"
    When I send a GET request to "/answers?item_id=210&user_id=1"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights on the given item id"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with id "10"
    When I send a GET request to "/answers?item_id=210&user_id=1"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when the user_id doesn't exist
    Given I am the user with id "1"
    When I send a GET request to "/answers?item_id=210&user_id=10"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the user doesn't have access to the item
    Given I am the user with id "1"
    When I send a GET request to "/answers?item_id=190&user_id=1"
    Then the response code should be 404
    And the response error message should contain "Insufficient access rights on the given item id"

  Scenario: Should fail when the item doesn't exist
    Given I am the user with id "1"
    When I send a GET request to "/answers?item_id=404&user_id=1"
    Then the response code should be 404
    And the response error message should contain "Insufficient access rights on the given item id"

  Scenario: Should fail when the authenticated user is not an admin of the selfGroup of the input user (via owned_group_id)
    Given I am the user with id "1"
    When I send a GET request to "/answers?item_id=200&user_id=2"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

