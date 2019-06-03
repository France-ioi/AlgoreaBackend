Feature: Update the 'current' answer
  Background:
    Given the database has the following table 'users':
      | ID  | sLogin | idGroupSelf |
      | 10  | john   | 101         |
    And the database has the following table 'groups':
      | ID  |
      | 101 |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 101             | 101          | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate |
      | 15 | 22            | 13           | direct             | null        |
    And the database has the following table 'items':
      | ID |
      | 50 |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate |
      | 101     | 50     | 2017-05-29T06:38:38Z     |
    And the database has the following table 'users_answers':
      | ID  | idUser | idItem | idAttempt |
      | 100 | 10     | 50     | 200       |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem |
      | 200 | 101     | 50     |

  Scenario: Missing attempt_id
    Given I am the user with ID "10"
    When I send a PUT request to "/answers/current" with the following body:
      """
      {
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for 'attempt_id': must be given and not null"
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: Missing answer
    Given I am the user with ID "10"
    When I send a PUT request to "/answers/current" with the following body:
      """
      {
        "attempt_id": "100",
        "state": "some state"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for 'answer': must be given and not null"
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: Missing state
    Given I am the user with ID "10"
    When I send a PUT request to "/answers/current" with the following body:
      """
      {
        "attempt_id": "100",
        "answer": "print 1"
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for 'state': must be given and not nul"
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: User not found
    Given I am the user with ID "404"
    When I send a PUT request to "/answers/current" with the following body:
      """
      {
        "attempt_id": "100",
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged

  Scenario: No access
    Given I am the user with ID "10"
    When I send a PUT request to "/answers/current" with the following body:
      """
      {
        "attempt_id": "300",
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "users_answers" should stay unchanged
