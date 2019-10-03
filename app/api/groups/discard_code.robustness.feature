Feature: Discard the code of the given group - robustness

  Background:
    Given the database has the following table 'users':
      | id | login | temp_user | self_group_id | owned_group_id | first_name  | last_name | default_language |
      | 1  | owner | 0         | 21            | 22             | Jean-Michel | Blanquer  | fr               |
      | 2  | user  | 0         | 11            | 12             | John        | Doe       | en               |
      | 3  | jane  | 0         | 31            | 32             | Jane        | Doe       | en               |
    And the database has the following table 'groups':
      | id | name    | grade | description     | created_at          | type      | code       | code_lifetime | code_expires_at     |
      | 11 | Group A | -3    | Group A is here | 2019-02-06 09:26:40 | Class     | ybqybxnlyo | 01:00:00      | 2017-10-13 05:39:48 |
      | 13 | Group B | -2    | Group B is here | 2019-03-06 09:26:40 | Class     | 3456789abc | 01:00:00      | 2017-10-14 05:39:48 |
      | 14 | Group C | -4    | Admin Group     | 2019-04-06 09:26:40 | UserAdmin | null       | null          | null                |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 75 | 22                | 13             | 0       |
      | 76 | 13                | 11             | 0       |
      | 77 | 22                | 11             | 0       |
      | 78 | 21                | 21             | 1       |

  Scenario: User is not an admin of the group
    Given I am the user with id "2"
    When I send a DELETE request to "/groups/13/code"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a DELETE request to "/groups/13/code"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups" should stay unchanged

  Scenario: The group id is not a number
    Given I am the user with id "1"
    When I send a DELETE request to "/groups/1_3/code"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
