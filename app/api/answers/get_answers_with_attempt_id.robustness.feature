Feature: Get item answers with attempt_id - robustness
Background:
  Given the database has the following table 'groups':
    | id | name        | text_id | grade | type      |
    | 11 | jdoe        |         | -2    | UserSelf  |
    | 12 | jdoe-admin  |         | -2    | UserAdmin |
    | 13 | Group B     |         | -2    | Class     |
    | 21 | guest       |         | -2    | UserSelf  |
    | 22 | guest-admin |         | -2    | UserAdmin |
  And the database has the following table 'users':
    | login | temp_user | group_id | owned_group_id |
    | jdoe  | 0         | 11       | 12             |
    | guest | 0         | 21       | 22             |
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
    | id  | type     | teams_editable | no_score | unlocked_item_ids |
    | 190 | Category | false          | false    | 1234,2345         |
    | 200 | Category | false          | false    | 1234,2345         |
    | 210 | Category | false          | false    | 1234,2345         |
  And the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       |
    | 13       | 190     | none                     |
    | 13       | 200     | content_with_descendants |
    | 13       | 210     | info                     |
  And the database has the following table 'groups_attempts':
    | id  | group_id | item_id | order |
    | 100 | 13       | 190     | 1     |
    | 110 | 13       | 210     | 2     |
    | 120 | 13       | 200     | 0     |

  Scenario: Should fail when the user has only info access to the item
    Given I am the user with id "11"
    When I send a GET request to "/answers?attempt_id=110"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/answers?attempt_id=110"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when the user doesn't have access to the item
    Given I am the user with id "11"
    When I send a GET request to "/answers?attempt_id=100"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the attempt doesn't exist
    Given I am the user with id "11"
    When I send a GET request to "/answers?attempt_id=400"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the authenticated user is not a member of the group and not an owner of the group attached to the attempt
    Given I am the user with id "21"
    When I send a GET request to "/answers?attempt_id=100"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when 'sort' is wrong
    Given I am the user with id "11"
    When I send a GET request to "/answers?attempt_id=120&sort=name"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "name""
