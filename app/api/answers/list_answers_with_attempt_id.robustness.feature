Feature: List answers by attempt_id - robustness
Background:
  Given the database has the following table 'groups':
    | id | name    | grade | type  |
    | 11 | jdoe    | -2    | User  |
    | 13 | Group B | -2    | Class |
    | 21 | guest   | -2    | User  |
  And the database has the following table 'users':
    | login | temp_user | group_id |
    | jdoe  | 0         | 11       |
    | guest | 0         | 21       |
  And the database has the following table 'groups_groups':
    | parent_group_id | child_group_id |
    | 13              | 11             |
  And the groups ancestors are computed
  And the database has the following table 'items':
    | id  | type    | no_score | default_language_tag |
    | 190 | Chapter | false    | fr                   |
    | 200 | Chapter | false    | fr                   |
    | 210 | Chapter | false    | fr                   |
  And the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       |
    | 13       | 190     | none                     |
    | 13       | 200     | content_with_descendants |
    | 13       | 210     | info                     |
    | 21       | 190     | content_with_descendants |
  And the database has the following table 'attempts':
    | id | participant_id |
    | 1  | 13             |
  And the database has the following table 'results':
    | attempt_id | participant_id | item_id |
    | 1          | 13             | 190     |
    | 1          | 13             | 210     |
    | 1          | 13             | 200     |

  Scenario: Should fail when the user has only info access to the item
    Given I am the user with id "11"
    When I send a GET request to "/items/210/answers?attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/items/210/answers?attempt_id=1"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when the user doesn't have access to the item
    Given I am the user with id "11"
    When I send a GET request to "/items/190/answers?attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the attempt doesn't exist
    Given I am the user with id "11"
    When I send a GET request to "/items/190/answers?attempt_id=400"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the authenticated user is not a member of the group and not a manager of the group attached to the attempt
    Given I am the user with id "21"
    When I send a GET request to "/items/190/answers?attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when 'sort' is wrong
    Given I am the user with id "11"
    When I send a GET request to "/items/200/answers?attempt_id=1&sort=name"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "name""
