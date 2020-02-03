Feature: Update the 'current' answer
  Background:
    Given the database has the following users:
      | login | group_id |
      | john  | 101      |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 101               | 101            |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 22              | 13             |
    And the database has the following table 'items':
      | id | default_language_tag |
      | 50 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 101      | 50      | content            |
    And the database has the following table 'attempts':
      | id  | group_id | item_id | order |
      | 200 | 101      | 50      | 1     |
    And the database has the following table 'answers':
      | id  | author_id | attempt_id | created_at          |
      | 100 | 101       | 200        | 2017-05-29 06:38:38 |

  Scenario: Invalid attempt_id
    Given I am the user with id "101"
    When I send a POST request to "/attempts/abc/answers" with the following body:
      """
      {
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for attempt_id (should be int64)"
    And the table "answers" should stay unchanged

  Scenario: Missing answer
    Given I am the user with id "101"
    When I send a POST request to "/attempts/100/answers" with the following body:
      """
      {
        "state": "some state"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "error_text": "Invalid input data",
        "errors": {
          "answer": ["missing field"]
        },
        "message": "Bad Request",
        "success": false
      }
      """
    And the table "answers" should stay unchanged

  Scenario: Missing state
    Given I am the user with id "101"
    When I send a POST request to "/attempts/100/answers" with the following body:
      """
      {
        "answer": "print 1"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "error_text": "Invalid input data",
        "errors": {
          "state": ["missing field"]
        },
        "message": "Bad Request",
        "success": false
      }
      """
    And the table "answers" should stay unchanged

  Scenario: User not found
    Given I am the user with id "404"
    When I send a POST request to "/attempts/100/answers" with the following body:
      """
      {
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "answers" should stay unchanged

  Scenario: No access
    Given I am the user with id "101"
    When I send a POST request to "/attempts/300/answers" with the following body:
      """
      {
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "answers" should stay unchanged
