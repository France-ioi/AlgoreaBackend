Feature: Create a group (groupCreate) - robustness

  Background:
    Given the database has the following table 'groups':
      | id | name  | type | team_item_id |
      | 21 | owner | User | null         |
      | 31 | tmp12 | User | null         |
      | 51 | john  | User | null         |
    And the database has the following table 'users':
      | login  | temp_user | group_id |
      | owner  | 0         | 21       |
      | tmp12  | 1         | 31       |
      | john   | 0         | 51       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 21                | 21             |
      | 31                | 31             |
      | 51                | 51             |
    And the database has the following table 'items':
      | id | default_language_tag |
      | 10 | fr                   |
      | 11 | fr                   |
      | 12 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 21       | 10      | content_with_descendants |
      | 21       | 11      | content                  |
      | 21       | 12      | info                     |
      | 51       | 11      | none                     |

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
      "errors": {"type": ["type must be one of [Class Team Club Friends Other]"]}
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

  Scenario: Zero item_id
    Given I am the user with id "21"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "some name", "type": "Team", "item_id": "0"}
    """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors": {"item_id": ["item_id must be 1 or greater"]}
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: item_id set for non-team group
    Given I am the user with id "21"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "some name", "type": "Class", "item_id": "1"}
    """
    Then the response code should be 400
    And the response error message should contain "Only teams can be created with item_id set"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

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

  Scenario: The item is not visible (no permissions)
    Given I am the user with id "51"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "some name", "type": "Team", "item_id": 10}
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: The item is not visible (can_view = none)
    Given I am the user with id "51"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "some name", "type": "Team", "item_id": 11}
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
