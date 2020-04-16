Feature: Update a group (groupEdit) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name    | grade | description     | created_at          | type  | activity_id         | is_official_session | is_open | is_public | code       | code_lifetime | code_expires_at     | open_contest | frozen_membership |
      | 11 | Group A | -3    | Group A is here | 2019-02-06 09:26:40 | Class | 1672978871462145361 | false               | true    | true      | ybqybxnlyo | 01:00:00      | 2017-10-13 05:39:48 | true         | 0                 |
      | 13 | Group B | -2    | Group B is here | 2019-03-06 09:26:40 | Class | 1672978871462145461 | false               | true    | true      | ybabbxnlyo | 01:00:00      | 2017-10-14 05:39:48 | true         | 1                 |
      | 14 | Group C | -2    | Group C is here | 2019-03-06 09:26:40 | Class | null                | false               | true    | true      | null       | null          | 2017-10-14 05:39:48 | true         | 0                 |
      | 15 | Group D | -2    | Group D is here | 2019-03-06 09:26:40 | Class | null                | true                | true    | true      | null       | null          | 2017-10-14 05:39:48 | true         | 0                 |
      | 21 | owner   | -4    | owner           | 2019-04-06 09:26:40 | User  | null                | false               | false   | false     | null       | null          | null                | false        | 0                 |
      | 31 | user    | -4    | owner           | 2019-04-06 09:26:40 | User  | null                | false               | false   | false     | null       | null          | null                | false        | 0                 |
    And the database has the following table 'users':
      | login | temp_user | group_id | first_name  | last_name |
      | owner | 0         | 21       | Jean-Michel | Blanquer  |
      | user  | 0         | 31       | John        | Doe       |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
      | 14       | 21         |
      | 15       | 21         |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id   | default_language_tag |
      | 123  | fr                   |
      | 5678 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 21       | 123     | info               |
      | 21       | 5678    | none               |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_make_session_official | source_group_id |
      | 21       | 123     | false                     | 13              |

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
      "is_public": 15,
      "name": 123,
      "grade": "grade",
      "description": 14.5,
      "is_open": "true",
      "code_lifetime": 1234,
      "code_expires_at": "the end",
      "open_contest": 12,

      "activity_id": "abc",
      "is_official_session": "abc",
      "require_members_to_join_parent": "abc",
      "organizer": 123,
      "address_line1": 123,
      "address_line2": 123,
      "address_postcode": 123,
      "address_city": 123,
      "address_country": 123,
      "expected_start": "abc"
    }
    """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "error_text": "Invalid input data",
      "errors": {
        "description": ["expected type 'string', got unconvertible type 'float64'"],
        "is_public": ["expected type 'bool', got unconvertible type 'float64'"],
        "grade": ["expected type 'int32', got unconvertible type 'string'"],
        "name": ["expected type 'string', got unconvertible type 'float64'"],
        "open_contest": ["expected type 'bool', got unconvertible type 'float64'"],
        "is_open": ["expected type 'bool', got unconvertible type 'string'"],
        "code_expires_at": ["decoding error: parsing time \"the end\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"the end\" as \"2006\""],
        "code_lifetime": ["expected type 'string', got unconvertible type 'float64'"],
        "activity_id": ["decoding error: strconv.ParseInt: parsing \"abc\": invalid syntax"],
        "is_official_session": ["expected type 'bool', got unconvertible type 'string'"],
        "require_members_to_join_parent": ["expected type 'bool', got unconvertible type 'string'"],
        "organizer": ["expected type 'string', got unconvertible type 'float64'"],
        "address_line1": ["expected type 'string', got unconvertible type 'float64'"],
        "address_line2": ["expected type 'string', got unconvertible type 'float64'"],
        "address_postcode": ["expected type 'string', got unconvertible type 'float64'"],
        "address_city": ["expected type 'string', got unconvertible type 'float64'"],
        "address_country": ["expected type 'string', got unconvertible type 'float64'"],
        "expected_start": ["decoding error: parsing time \"abc\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"abc\" as \"2006\""]
      },
      "message": "Bad Request",
      "success": false
    }
    """
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

  Scenario: The activity does not exist
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "activity_id": "404"
    }
    """
    Then the response code should be 403
    And the response error message should contain "No access to the activity"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: The user cannot view the activity
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "activity_id": "5678"
    }
    """
    Then the response code should be 403
    And the response error message should contain "No access to the activity"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: The user cannot view the activity
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "activity_id": "5678"
    }
    """
    Then the response code should be 403
    And the response error message should contain "No access to the activity"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: is_official_session becomes true & activity_id becomes not null while the user doesn't have the permission
    Given I am the user with id "21"
    When I send a PUT request to "/groups/14" with the following body:
    """
    {
      "activity_id": "123",
      "is_official_session": true
    }
    """
    Then the response code should be 403
    And the response error message should contain "Not enough permissions for attaching the group to the activity as an official session"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: is_official_session becomes true & activity_id is set in the db while the user doesn't have the permission
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "is_official_session": true
    }
    """
    Then the response code should be 403
    And the response error message should contain "Not enough permissions for attaching the group to the activity as an official session"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: is_official_session is true in the db & activity_id becomes not null while the user doesn't have the permission
    Given I am the user with id "21"
    When I send a PUT request to "/groups/15" with the following body:
    """
    {
      "activity_id": "123"
    }
    """
    Then the response code should be 403
    And the response error message should contain "Not enough permissions for attaching the group to the activity as an official session"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: is_official_session becomes true, but activity_id is null in the db
    Given I am the user with id "21"
    When I send a PUT request to "/groups/14" with the following body:
    """
    {
      "is_official_session": true
    }
    """
    Then the response code should be 400
    And the response error message should contain "The activity_id should be set for official sessions"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: is_official_session becomes true, but the new activity_id is null
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "activity_id": null,
      "is_official_session": true
    }
    """
    Then the response code should be 400
    And the response error message should contain "The activity_id should be set for official sessions"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: frozen_membership changes from true to false
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "frozen_membership": false
    }
    """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "error_text": "Invalid input data",
      "errors": {
        "frozen_membership": ["can only be changed from false to true"]
      },
      "message": "Bad Request",
      "success": false
    }
    """
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged
