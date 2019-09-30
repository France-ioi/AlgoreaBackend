Feature: Create a group (groupCreate) - robustness

  Background:
    Given the database has the following table 'users':
      | id | login  | temp_user | self_group_id | owned_group_id |
      | 1  | owner  | 0         | 21            | 22             |
      | 2  | tmp12  | 1         | 31            | 32             |
      | 3  | noself | 0         | null          | 42             |
      | 4  | john   | 0         | 51            | 52             |
      | 5  | weird  | 0         | 61            | null           |
    And the database has the following table 'groups':
      | id | name         | type      | team_item_id |
      | 21 | owner        | UserSelf  | null         |
      | 22 | owner-admin  | UserAdmin | null         |
      | 31 | tmp12        | UserSelf  | null         |
      | 32 | tmp12-admin  | UserAdmin | null         |
      | 42 | noself-admin | UserAdmin | null         |
      | 51 | john         | UserSelf  | null         |
      | 52 | john-admin   | UserAdmin | null         |
      | 61 | weird        | UserSelf  | null         |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
      | 42                | 42             | 1       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
      | 61                | 61             | 1       |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_full_access_since | cached_partial_access_since | cached_grayed_access_since | creator_user_id |
      | 21       | 10      | 2019-07-16 21:28:47      | null                        | null                       | 1               |
      | 21       | 11      | null                     | 2019-07-16 21:28:47         | null                       | 1               |
      | 21       | 12      | null                     | null                        | 2019-07-16 21:28:47        | 1               |

  Scenario: No name
    Given I am the user with id "1"
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
    Given I am the user with id "1"
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
    Given I am the user with id "1"
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
    Given I am the user with id "1"
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
    | type      |
    |           |
    | Unknown   |
    | UserSelf  |
    | UserAdmin |
    | Base      |
    | RootSelf  |
    | Root      |
    | RootTemp  |

  Scenario: Zero item_id
    Given I am the user with id "1"
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
    Given I am the user with id "1"
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
    Given I am the user with id "2"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "some name", "type": "Class"}
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User with empty self group
    Given I am the user with id "3"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "some name", "type": "Class"}
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: User with empty owned group
    Given I am the user with id "5"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "some name", "type": "Class"}
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: The item is not visible
    Given I am the user with id "4"
    When I send a POST request to "/groups" with the following body:
    """
    {"name": "some name", "type": "Team", "item_id": 10}
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
