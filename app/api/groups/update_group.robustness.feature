Feature: Update a group (groupEdit) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name    | grade | description     | created_at          | type     | redirect_path                          | opened | free_access | code       | code_lifetime | code_expires_at     | open_contest |
      | 11 | Group A | -3    | Group A is here | 2019-02-06 09:26:40 | Class    | 182529188317717510/1672978871462145361 | true   | true        | ybqybxnlyo | 01:00:00      | 2017-10-13 05:39:48 | true         |
      | 13 | Group B | -2    | Group B is here | 2019-03-06 09:26:40 | Class    | 182529188317717610/1672978871462145461 | true   | true        | ybabbxnlyo | 01:00:00      | 2017-10-14 05:39:48 | true         |
      | 21 | owner   | -4    | owner           | 2019-04-06 09:26:40 | UserSelf | null                                   | false  | false       | null       | null          | null                | false        |
      | 31 | user    | -4    | owner           | 2019-04-06 09:26:40 | UserSelf | null                                   | false  | false       | null       | null          | null                | false        |
    And the database has the following table 'users':
      | login | temp_user | group_id | first_name  | last_name |
      | owner | 0         | 21       | Jean-Michel | Blanquer  |
      | user  | 0         | 31       | John        | Doe       |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 13                | 11             | 0       |
      | 13                | 13             | 1       |
      | 15                | 15             | 1       |
      | 21                | 21             | 1       |
      | 31                | 31             | 1       |

  Scenario: Should fail if the user is not a manager of the group
    Given I am the user with id "31"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {}
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: Should fail if the user is not found
    Given I am the user with id "404"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {}
    """
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: User is a manager of the group, but required fields are not filled in correctly
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "free_access": 15,
      "name": 123,
      "grade": "grade",
      "description": 14.5,
      "opened": "true",
      "code_lifetime": 1234,
      "code_expires_at": "the end",
      "open_contest": 12,
      "redirect_path": "some path"
    }
    """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "error_text": "Invalid input data",
      "errors": {
        "description": ["expected type 'string', got unconvertible type 'float64'"],
        "free_access": ["expected type 'bool', got unconvertible type 'float64'"],
        "grade": ["expected type 'int32', got unconvertible type 'string'"],
        "name": ["expected type 'string', got unconvertible type 'float64'"],
        "open_contest": ["expected type 'bool', got unconvertible type 'float64'"],
        "opened": ["expected type 'bool', got unconvertible type 'string'"],
        "code_expires_at": ["decoding error: parsing time \"the end\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"the end\" as \"2006\""],
        "code_lifetime": ["expected type 'string', got unconvertible type 'float64'"],
        "redirect_path": ["invalid redirect path"]
      },
      "message": "Bad Request",
      "success": false
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: User is a manager of the group, but no fields provided
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: The group id is not a number
    Given I am the user with id "21"
    When I send a PUT request to "/groups/1_3" with the following body:
    """
    {
    }
    """
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
