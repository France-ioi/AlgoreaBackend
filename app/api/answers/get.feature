Feature: Get user's answer by user_answer_id
Background:
  Given the database has the following table 'users':
    | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName | sLastName |
    | 1  | jdoe   | 0        | 11          | 12           | John       | Doe       |
    | 2  | other  | 0        | 21          | 22           | George     | Bush      |
  And the database has the following table 'groups':
    | ID | sName       | sType     |
    | 11 | jdoe        | UserSelf  |
    | 12 | jdoe-admin  | UserAdmin |
    | 13 | Group B     | Class     |
    | 21 | other       | UserSelf  |
    | 22 | other-admin | UserAdmin |
    | 23 | Group C     | Class     |
  And the database has the following table 'groups_groups':
    | ID | idGroupParent | idGroupChild | sType              |
    | 61 | 13            | 11           | invitationAccepted |
    | 62 | 13            | 21           | requestAccepted    |
    | 63 | 23            | 21           | direct             |
  And the database has the following table 'groups_ancestors':
    | ID | idGroupAncestor | idGroupChild | bIsSelf |
    | 71 | 11              | 11           | 1       |
    | 72 | 12              | 12           | 1       |
    | 73 | 13              | 13           | 1       |
    | 74 | 13              | 11           | 0       |
    | 75 | 13              | 21           | 0       |
    | 76 | 23              | 21           | 0       |
    | 77 | 23              | 23           | 1       |
  And the database has the following table 'items':
    | ID  | bHasAttempts |
    | 200 | 0            |
    | 210 | 1            |
  And the database has the following table 'groups_items':
    | ID | idGroup | idItem | sCachedFullAccessDate | sCachedPartialAccessDate |
    | 43 | 13      | 200    | 2017-05-29T06:38:38Z  | 2017-05-29T06:38:38Z     |
    | 46 | 23      | 210    | 2017-05-29T06:38:38Z  | 2017-05-29T06:38:38Z     |
  And the database has the following table 'users_answers':
    | ID  | idUser | idItem | idAttempt | sType      | sState  | sAnswer  | sLangProg | sSubmissionDate     | iScore | bValidated | sGradingDate        | idUserGrader |
    | 101 | 1      | 200    | 150       | Submission | Current | print(1) | python    | 2017-05-29 06:38:38 | 100    | true       | 2018-05-29 06:38:38 | 123          |
    | 102 | 1      | 210    | 250       | Submission | Current | print(2) | python    | 2017-05-29 06:38:38 | 100    | true       | 2019-05-29 06:38:38 | 456          |
  And the database has the following table 'groups_attempts':
    | ID  | idGroup | idItem |
    | 150 | 11      | 200    |
    | 250 | 13      | 210    |

  Scenario: User has access to the item and the users_answers.idUser = authenticated user's ID
    Given I am the user with ID "1"
    When I send a GET request to "/answers/101"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "101",
      "attempt_id": "150",
      "score": 100.0,
      "answer": "print(1)",
      "state": "Current",
      "submission_date": "2017-05-29T06:38:38Z",
      "type": "Submission",
      "item_id": "200",
      "user_id": "1",
      "user_grader_id": "123",
      "grading_date": "2018-05-29T06:38:38Z",
      "validated": true
    }
    """

  Scenario: User has access to the item and the user is a team member of groups_attempts.idGroup (items.bHasAttempts=1)
    Given I am the user with ID "2"
    When I send a GET request to "/answers/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "102",
      "attempt_id": "250",
      "score": 100,
      "answer": "print(2)",
      "state": "Current",
      "submission_date": "2017-05-29T06:38:38Z",
      "type": "Submission",
      "item_id": "210",
      "user_id": "1",
      "user_grader_id": "456",
      "grading_date": "2019-05-29T06:38:38Z",
      "validated": true
    }
    """
