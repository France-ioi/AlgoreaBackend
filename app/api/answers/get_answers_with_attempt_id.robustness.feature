Feature: Get item answers with attempt_id - robustness
Background:
  Given the database has the following table 'users':
    | id | login | temp_user | group_self_id | group_owned_id | version |
    | 1  | jdoe  | 0         | 11            | 12             | 0       |
    | 2  | guest | 0         | 404           | 404            | 0       |
  And the database has the following table 'groups':
    | id | name       | text_id | grade | type      | version |
    | 11 | jdoe       |         | -2    | UserAdmin | 0       |
    | 12 | jdoe-admin |         | -2    | UserAdmin | 0       |
    | 13 | Group B    |         | -2    | Class     | 0       |
  And the database has the following table 'groups_groups':
    | id | group_parent_id | group_child_id | type               |
    | 61 | 13              | 11             | invitationAccepted |
  And the database has the following table 'groups_ancestors':
    | id | group_ancestor_id | group_child_id | is_self | version |
    | 71 | 11                | 11             | 1       | 0       |
    | 72 | 12                | 12             | 1       | 0       |
    | 73 | 13                | 13             | 1       | 0       |
    | 74 | 13                | 11             | 0       | 0       |
  And the database has the following table 'items':
    | id  | type     | teams_editable | no_score | item_unlocked_id | transparent_folder | version |
    | 190 | Category | false          | false    | 1234,2345        | true               | 0       |
    | 200 | Category | false          | false    | 1234,2345        | true               | 0       |
    | 210 | Category | false          | false    | 1234,2345        | true               | 0       |
  And the database has the following table 'groups_items':
    | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | user_created_id | version |
    | 42 | 13       | 190     | 2037-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    | 43 | 13       | 200     | 2017-05-29 06:38:38     | 2017-05-29 06:38:38        | 2017-05-29 06:38:38       | 0               | 0       |
    | 44 | 13       | 210     | 2037-05-29 06:38:38     | 2037-05-29 06:38:38        | 2017-05-29 06:38:38       | 0               | 0       |
  And the database has the following table 'groups_attempts':
    | id  | group_id | item_id | `order` |
    | 100 | 13       | 190     | 1       |
    | 110 | 13       | 210     | 2       |
    | 120 | 13       | 200     | 0       |

  Scenario: Should fail when the user has only grayed access to the item
    Given I am the user with id "1"
    When I send a GET request to "/answers?attempt_id=110"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/answers?attempt_id=110"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when the user doesn't have access to the item
    Given I am the user with id "1"
    When I send a GET request to "/answers?attempt_id=100"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the attempt doesn't exist
    Given I am the user with id "1"
    When I send a GET request to "/answers?attempt_id=400"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the authenticated user is not a member of the group and not an owner of the group attached to the attempt
    Given I am the user with id "2"
    When I send a GET request to "/answers?attempt_id=100"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when 'sort' is wrong
    Given I am the user with id "1"
    When I send a GET request to "/answers?attempt_id=120&sort=name"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "name""
