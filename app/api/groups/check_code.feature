Feature: Check if the group code is valid
  Background:
    Given the database has the following table 'groups':
      | id | type  | code       | code_expires_at     | code_lifetime | is_public | frozen_membership |
      | 3  | Base  | null       | null                | null          | false     | false             |
      | 11 | Team  | 3456789abc | 2037-05-29 06:38:38 | 01:02:03      | true      | false             |
      | 12 | Team  | abc3456789 | null                | 12:34:56      | true      | false             |
      | 13 | Team  | 456789abcd | 2017-05-29 06:38:38 | 01:02:03      | true      | false             |
      | 14 | Team  | cba9876543 | null                | null          | true      | false             |
      | 15 | Team  | 987654321a | null                | null          | true      | false             |
      | 16 | Class | dcef123492 | null                | null          | false     | false             |
      | 17 | Team  | 75987654ab | null                | null          | false     | false             |
      | 18 | Team  | 5987654abc | null                | null          | true      | false             |
      | 19 | Team  | 87654abcde | null                | null          | true      | true              |
      | 21 | User  | null       | null                | null          | false     | false             |
      | 22 | User  | null       | null                | null          | false     | false             |
    And the database has the following table 'users':
      | group_id | login | temp_user |
      | 21       | john  | false     |
      | 22       | tmp   | true      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 14              | 21             |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id   | default_language_tag |
      | 1234 | fr                   |
    And the database has the following table 'attempts':
      | id | participant_id | root_item_id |
      | 0  | 21             | null         |
      | 2  | 14             | 1234         |
      | 2  | 18             | 1234         |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | result_propagation_state |
      | 0          | 21             | 30      | done                     |

  Scenario: The code is valid for normal user
    Given I am the user with id "21"
    When I send a GET request to "/groups/is-code-valid?code=3456789abc"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": true}
    """

  Scenario: The code is wrong
    Given I am the user with id "21"
    When I send a GET request to "/groups/is-code-valid?code=abcdef"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": false}
    """

  Scenario: The code is expired
    Given I am the user with id "21"
    When I send a GET request to "/groups/is-code-valid?code=456789abcd"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": false}
    """

  Scenario: The group is not a team
    Given I am the user with id "21"
    When I send a GET request to "/groups/is-code-valid?code=dcef123492"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": false}
    """

  Scenario: The user is temporary
    Given I am the user with id "22"
    And the application config is:
      """
      domains:
        -
          domains: [127.0.0.1]
          allUsersGroup: 3
      """
    When I send a GET request to "/groups/is-code-valid?code=3456789abc"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": true}
    """

  Scenario: The user is temporary (custom all-users group)
    Given I am the user with id "22"
    And the database has the following table 'items':
      | id | default_language_tag | allows_multiple_attempts |
      | 2  | fr                   | false                    |
    And the database table 'attempts' has also the following row:
      | participant_id | id | root_item_id |
      | 3              | 1  | 2            |
      | 11             | 1  | 2            |
    And the database has the following table 'results':
      | participant_id | attempt_id | item_id | started_at          |
      | 3              | 1          | 2       | 2019-05-30 11:00:00 |
      | 11             | 1          | 2       | 2019-05-30 11:00:00 |
    And the application config is:
      """
      domains:
        -
          domains: [127.0.0.1]
          allUsersGroup: 3
      """
    When I send a GET request to "/groups/is-code-valid?code=3456789abc"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": false}
    """

  Scenario: A closed team
    Given I am the user with id "21"
    When I send a GET request to "/groups/is-code-valid?code=75987654ab"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": false}
    """

  Scenario: A member of another team participating in same contests
    Given I am the user with id "21"
    When I send a GET request to "/groups/is-code-valid?code=5987654abc"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": false}
    """

  Scenario: The team membership is frozen
    Given I am the user with id "21"
    When I send a GET request to "/groups/is-code-valid?code=87654abcde"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": false}
    """

  Scenario: Joining would break entry conditions for the team
    Given I am the user with id "21"
    And the database has the following table 'items':
      | id | default_language_tag | entry_min_admitted_members_ratio |
      | 2  | fr                   | All                              |
    And the database table 'attempts' has also the following row:
      | participant_id | id | root_item_id |
      | 12             | 1  | 2            |
    And the database has the following table 'results':
      | participant_id | attempt_id | item_id | started_at          |
      | 12             | 1          | 2       | 2019-05-30 11:00:00 |
    When I send a GET request to "/groups/is-code-valid?code=abc3456789"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": false}
    """
