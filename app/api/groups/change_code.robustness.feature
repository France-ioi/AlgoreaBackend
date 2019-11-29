Feature: Change the code of the given group - robustness

  Background:
    Given the database has the following table 'groups':
      | id | name        | grade | description     | created_at          | type      | code       | code_lifetime | code_expires_at     |
      | 11 | Group A     | -3    | Group A is here | 2019-02-06 09:26:40 | Class     | ybqybxnlyo | 01:00:00      | 2017-10-13 05:39:48 |
      | 13 | Group B     | -2    | Group B is here | 2019-03-06 09:26:40 | Class     | 3456789abc | 01:00:00      | 2017-10-14 05:39:48 |
      | 14 | Group C     | -4    | Admin Group     | 2019-04-06 09:26:40 | UserAdmin | null       | null          | null                |
      | 21 | owner       | -4    | owner           | 2019-04-06 09:26:40 | UserSelf  | null       | null          | null                |
      | 22 | owner-admin | -4    | owner-admin     | 2019-04-06 09:26:40 | UserAdmin | null       | null          | null                |
      | 31 | jane        | -4    | jane            | 2019-04-06 09:26:40 | UserSelf  | null       | null          | null                |
      | 32 | jane-admin  | -4    | jane-admin      | 2019-04-06 09:26:40 | UserAdmin | null       | null          | null                |
      | 41 | user        | -4    | user            | 2019-04-06 09:26:40 | UserSelf  | null       | null          | null                |
      | 42 | user-admin  | -4    | user-admin      | 2019-04-06 09:26:40 | UserAdmin | null       | null          | null                |
    And the database has the following table 'users':
      | login | temp_user | group_id | owned_group_id | first_name  | last_name | default_language |
      | owner | 0         | 21       | 22             | Jean-Michel | Blanquer  | fr               |
      | user  | 0         | 41       | 42             | John        | Doe       | en               |
      | jane  | 0         | 31       | 32             | Jane        | Doe       | en               |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 75 | 11                | 11             | 1       |
      | 76 | 13                | 11             | 0       |
      | 77 | 13                | 13             | 1       |
      | 78 | 21                | 21             | 1       |

  Scenario: User is not a manager of the group
    Given I am the user with id "41"
    And the generated group code is "newpassword"
    When I send a POST request to "/groups/13/code"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged

  Scenario: User does not exist
    Given I am the user with id "404"
    And the generated group code is "newpassword"
    When I send a POST request to "/groups/13/code"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups" should stay unchanged

  Scenario: User is a manager of the group, but the generated code is not unique
    Given I am the user with id "21"
    And the generated group codes are "ybqybxnlyo","newpassword"
    When I send a POST request to "/groups/13/code"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"code":"newpassword"}
    """
    And the table "groups" should stay unchanged but the row with id "13"
    And the table "groups" at id "13" should be:
      | id | name    | grade | description     | created_at          | type  | code        | code_lifetime | code_expires_at     |
      | 13 | Group B | -2    | Group B is here | 2019-03-06 09:26:40 | Class | newpassword | 01:00:00      | 2017-10-14 05:39:48 |

  Scenario: User is a manager of the group, but the generated code is not unique 3 times in a row
    Given I am the user with id "21"
    And the generated group codes are "ybqybxnlyo","ybqybxnlyo","ybqybxnlyo"
    When I send a POST request to "/groups/13/code"
    Then the response code should be 500
    And the response error message should contain "The code generator is broken"
    And the table "groups" should stay unchanged

  Scenario: The group id is not a number
    Given I am the user with id "21"
    When I send a POST request to "/groups/1_3/code"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
