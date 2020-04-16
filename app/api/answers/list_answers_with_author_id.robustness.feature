Feature: List answers by (item_id, author_id) pair - robustness
Background:
  Given the database has the following table 'groups':
    | id  | name    | text_id | grade | type  |
    | 11  | jdoe    |         | -2    | User  |
    | 13  | Group B |         | -2    | Class |
    | 404 | guest   |         | -2    | Class |
  And the database has the following table 'users':
    | login | temp_user | group_id |
    | jdoe  | 0         | 11       |
    | guest | 0         | 404      |
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

  Scenario: Should fail when the user has only info access to the item
    Given I am the user with id "11"
    When I send a GET request to "/items/210/answers?author_id=11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with id "10"
    When I send a GET request to "/items/210/answers?author_id=11"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when the user_id doesn't exist
    Given I am the user with id "11"
    When I send a GET request to "/items/210/answers?author_id=10"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the user doesn't have access to the item
    Given I am the user with id "11"
    When I send a GET request to "/items/190/answers?author_id=11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the item doesn't exist
    Given I am the user with id "11"
    When I send a GET request to "/items/404/answers?author_id=11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the authenticated user is not a manager of the selfGroup of the input user (via group_managers)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/answers?author_id=2"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
