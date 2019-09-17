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
      | idGroup | idItem | sCachedPartialAccessDate | idUserCreated |
      | 101     | 50     | 2017-05-29 06:38:38      | 10            |
    And the database has the following table 'users_answers':
      | ID  | idUser | idItem | idAttempt | sType      | sSubmissionDate     |
      | 100 | 10     | 50     | 200       | Submission | 2017-05-29 06:38:38 |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | iOrder |
      | 200 | 101     | 50     | 0      |

  Scenario: User is able to create the 'current' answer and users_items.idAttemptActive = request.attempt_id
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 50     | 200             |
    When I send a PUT request to "/answers/current" with the following body:
      """
      {
        "attempt_id": "200",
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | idAttemptActive | sAnswer | sState     |
      | 10     | 50     | 200             | print 1 | some state |
    And the table "users_answers" should be:
      | idUser | idItem | idAttempt | sType      | sAnswer | sState     | ABS(TIMESTAMPDIFF(SECOND, sSubmissionDate, NOW())) < 3 |
      | 10     | 50     | 200       | Submission | null    | null       | 0                                                      |
      | 10     | 50     | 200       | Current    | print 1 | some state | 1                                                      |

  Scenario: User is able to create the 'current' answer and users_items.idAttemptActive != request.attempt_id
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | sAnswer | sState |
      | 10     | 50     | 100             | null    | null   |
    When I send a PUT request to "/answers/current" with the following body:
      """
      {
        "attempt_id": "200",
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should stay unchanged
    And the table "users_answers" should be:
      | idUser | idItem | idAttempt | sType      | sAnswer | sState     | ABS(TIMESTAMPDIFF(SECOND, sSubmissionDate, NOW())) < 3 |
      | 10     | 50     | 200       | Submission | null    | null       | 0                                                      |
      | 10     | 50     | 200       | Current    | print 1 | some state | 1                                                      |

  Scenario: User is able to update the 'current' answer
    Given I am the user with ID "10"
    And the database has the following table 'users_answers':
      | ID  | idUser | idItem | idAttempt | sType   | sSubmissionDate     |
      | 101 | 10     | 50     | 200       | Current | 2017-05-29 06:38:38 |
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 50     | 200             |
    When I send a PUT request to "/answers/current" with the following body:
      """
      {
        "attempt_id": "200",
        "answer": "print 1",
        "state": "some state"
      }
      """
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | idAttemptActive | sAnswer | sState     |
      | 10     | 50     | 200             | print 1 | some state |
    And the table "users_answers" should be:
      | ID  | idUser | idItem | idAttempt | sType      | sAnswer | sState     | ABS(TIMESTAMPDIFF(SECOND, sSubmissionDate, NOW())) < 3 |
      | 100 | 10     | 50     | 200       | Submission | null    | null       | 0                                                      |
      | 101 | 10     | 50     | 200       | Current    | print 1 | some state | 0                                                      |
