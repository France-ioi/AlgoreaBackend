Feature: Create a group (groupCreate) - robustness

  Background:
    Given the database has the following table 'groups':
      | id | name  | type |
      | 21 | owner | User |
      | 31 | tmp12 | User |
      | 51 | john  | User |
    And the database has the following table 'users':
      | login  | temp_user | group_id |
      | owner  | 0         | 21       |
      | tmp12  | 1         | 31       |
      | john   | 0         | 51       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
    And the groups ancestors are computed

  Scenario: No name
    Given I am the user with id "21"
    When I send a POST request to "/groups" with the following body:
    """
    {"type": "Class"}
    """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors": {"name": ["missing field"]}
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Empty name
    Given I am the user with id "21"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "", "type": "Class"}
    """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors": {"name": ["name must be at least 1 character in length"]}
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: No type
    Given I am the user with id "21"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "some name"}
    """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors": {"type": ["missing field"]}
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario Outline: Empty or wrong type
    Given I am the user with id "21"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "some name", "type": "<type>"}
    """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors": {"type": ["type must be one of [Class Team Club Friends Other Session]"]}
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
  Examples:
    | type     |
    |          |
    | Unknown  |
    | User     |
    | Base     |
    | RootSelf |
    | Root     |
    | RootTemp |

  Scenario: Temporary user
    Given I am the user with id "31"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "some name", "type": "Class"}
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
