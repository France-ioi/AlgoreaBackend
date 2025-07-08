Feature: Get group by groupID (groupView) - robustness
  Background:
    Given the database has the following user:
      | group_id | login | first_name  | last_name | default_language |
      | 21       | owner | Jean-Michel | Blanquer  | fr               |
    And the database has the following table "groups":
      | id | name                | description         | created_at          | type                | root_activity_id    | is_open | is_public | code       | code_lifetime | code_expires_at     | open_activity_when_joining |
      | 11 | Group A             | Group A is here     | 2019-02-06 09:26:40 | Class               | 1672978871462145361 | true    | true      | ybqybxnlyo | 3600          | 2017-10-13 05:39:48 | true                       |
      | 13 | Group B             | Group B is here     | 2019-03-06 09:26:40 | Class               | 1672978871462145461 | true    | false     | ybabbxnlyo | 3600          | 2017-10-14 05:39:48 | true                       |
      | 22 | ContestParticipants | ContestParticipants | 2019-03-06 09:26:40 | ContestParticipants | null                | false   | true      | null       | null          | null                | false                      |

  Scenario: Should fail when group_id is invalid
    Given I am the user with id "21"
    When I send a GET request to "/groups/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Should fail when the group is not visible
    Given I am the user with id "21"
    When I send a GET request to "/groups/13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/groups/13"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should fail when group_id is a user
    Given I am the user with id "21"
    When I send a GET request to "/groups/21"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when group_id is a contest participants group
    Given I am the user with id "21"
    When I send a GET request to "/groups/22"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
