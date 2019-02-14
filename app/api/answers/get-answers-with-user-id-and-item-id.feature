Feature: Get answers with (item_id, user_id) pair
Background:
  Given the database has the following table 'users':
    | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName | sLastName |
    | 1  | jdoe   | 0        | 11          | 12           | John       | Doe       |
    | 2  | guest  | 0        | 404         | 404          |            |           |
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
  And the database has the following table 'items':
    | ID  | sType    | bTeamsEditable | bNoScore | idItemUnlocked | bTransparentFolder | iVersion |
    | 190 | Category | false          | false    | 1234,2345      | true               | 0        |
    | 200 | Category | false          | false    | 1234,2345      | true               | 0        |
    | 210 | Category | false          | false    | 1234,2345      | true               | 0        |
  And the database has the following table 'groups_items':
    | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
    | 42 | 13      | 190    | null            | false             | false                | false               | 0             | 0        |
    | 43 | 13      | 200    | null            | true              | true                 | true                | 0             | 0        |
    | 44 | 13      | 210    | null            | false             | false                | true                | 0             | 0        |
  And the database has the following table 'users_answers':
    | ID | idUser | idItem | idAttempt | sName | sType      | sState | sAnswer | sLangProg | sSubmissionDate     | iScore | bValidated |
    | 1  | 1      | 200    |           | name  | Submission | null   | answer  | lang      | 2017-05-29 06:38:38 | 100    | true       |

  Scenario: Full access on the item+user pair
    Given I am the user with ID "1"
    When I send a GET request to "/answers?item_id=200&user_id=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "answers": [
        {
          "id": 1,
          "lang_prog": "lang",
          "name": "name",
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
    }
    """
