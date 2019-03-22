Feature: Get answers with attempt_id
Background:
  Given the database has the following table 'users':
    | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName |
    | 1  | jdoe   | 0        | 11          | 12           | John        | Doe       |
    | 2  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  |
  And the database has the following table 'groups':
    | ID | sName      | sTextId | iGrade | sType     | iVersion |
    | 11 | jdoe       |         | -2     | UserAdmin | 0        |
    | 12 | jdoe-admin |         | -2     | UserAdmin | 0        |
    | 13 | Group B    |         | -2     | Class     | 0        |
  And the database has the following table 'groups_groups':
    | ID | idGroupParent | idGroupChild | iVersion |
    | 61 | 13            | 11           | 0        |
  And the database has the following table 'groups_ancestors':
    | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
    | 71 | 11              | 11           | 1       | 0        |
    | 72 | 12              | 12           | 1       | 0        |
    | 73 | 13              | 13           | 1       | 0        |
    | 74 | 13              | 11           | 0       | 0        |
    | 75 | 22              | 13           | 0       | 0        |
    | 77 | 41              | 21           | 0       | 0        |
  And the database has the following table 'items':
    | ID  | sType    | bTeamsEditable | bNoScore | idItemUnlocked | bTransparentFolder | iVersion |
    | 190 | Category | false          | false    | 1234,2345      | true               | 0        |
    | 200 | Category | false          | false    | 1234,2345      | true               | 0        |
    | 210 | Category | false          | false    | 1234,2345      | true               | 0        |
  And the database has the following table 'groups_items':
    | ID | idGroup | idItem | sCachedFullAccessDate | sCachedPartialAccessDate | sCachedGrayedAccessDate | idUserCreated | iVersion |
    | 42 | 13      | 190    | 3017-05-29T06:38:38Z  | 3017-05-29T06:38:38Z     | 3017-05-29T06:38:38Z    | 0             | 0        |
    | 43 | 13      | 200    | 2017-05-29T06:38:38Z  | 2017-05-29T06:38:38Z     | 2017-05-29T06:38:38Z    | 0             | 0        |
    | 44 | 13      | 210    | 3017-05-29T06:38:38Z  | 3017-05-29T06:38:38Z     | 2017-05-29T06:38:38Z    | 0             | 0        |
    | 45 | 41      | 200    | 2017-05-29T06:38:38Z  | 2017-05-29T06:38:38Z     | 2017-05-29T06:38:38Z    | 0             | 0        |
  And the database has the following table 'users_answers':
    | ID | idUser | idItem | idAttempt | sName            | sType      | sState  | sLangProg | sSubmissionDate     | iScore | bValidated |
    | 1  | 1      | 200    | 100       | My answer        | Submission | Current | python    | 2017-05-29 06:38:38 | 100    | true       |
    | 2  | 1      | 200    | 101       | My second anwser | Submission | Current | python    | 2017-05-29 06:38:38 | 100    | true       |
  And the database has the following table 'groups_attempts':
    | ID  | idGroup | idItem |
    | 100 | 13      | 200    |

  Scenario: Full access on the item and the user is a member of the attempt's group
    Given I am the user with ID "1"
    When I send a GET request to "/answers?attempt_id=100"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": 1,
        "lang_prog": "python",
        "name": "My answer",
        "score": 100,
        "submission_date": "2017-05-29T06:38:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        },
        "validated": true
      }
    ]
    """

  Scenario: Full access on the item and the user is an owner of some attempt's group parent
    Given I am the user with ID "2"
    When I send a GET request to "/answers?attempt_id=100"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": 1,
        "lang_prog": "python",
        "name": "My answer",
        "score": 100,
        "submission_date": "2017-05-29T06:38:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        },
        "validated": true
      }
    ]
    """
