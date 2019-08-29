Feature: Get groups attempts for current user and item_id
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName | sLastName |
      | 1  | jdoe   | 11          | 12           | John       | Doe       |
      | 2  | other  | 21          | 22           | George     | Bush      |
      | 3  | jane   | 31          | 32           | Jane       | Doe       |
    And the database has the following table 'groups':
      | ID | sName       | sType     |
      | 11 | jdoe        | UserSelf  |
      | 12 | jdoe-admin  | UserAdmin |
      | 13 | Group B     | Class     |
      | 21 | other       | UserSelf  |
      | 22 | other-admin | UserAdmin |
      | 23 | Group C     | Class     |
      | 31 | jane        | UserSelf  |
      | 32 | jane-admin  | UserAdmin |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              |
      | 61 | 13            | 11           | invitationAccepted |
      | 62 | 13            | 21           | requestAccepted    |
      | 63 | 13            | 31           | joinedByCode       |
      | 64 | 23            | 21           | direct             |
      | 65 | 23            | 31           | direct             |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf |
      | 71 | 11              | 11           | 1       |
      | 72 | 12              | 12           | 1       |
      | 73 | 13              | 13           | 1       |
      | 74 | 13              | 11           | 0       |
      | 75 | 13              | 21           | 0       |
      | 76 | 13              | 31           | 0       |
      | 77 | 23              | 21           | 0       |
      | 78 | 23              | 23           | 1       |
      | 79 | 23              | 31           | 0       |
      | 80 | 31              | 31           | 1       |
      | 81 | 32              | 32           | 1       |
    And the database has the following table 'items':
      | ID  | bHasAttempts |
      | 200 | 0            |
      | 210 | 1            |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | sCachedFullAccessDate | sCachedPartialAccessDate |
      | 43 | 13      | 200    | 2017-05-29T06:38:38Z  | 2017-05-29T06:38:38Z     |
      | 46 | 23      | 210    | 2017-05-29T06:38:38Z  | 2017-05-29T06:38:38Z     |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | iScore | iOrder | bValidated | sStartDate          | idUserCreator |
      | 150 | 11      | 200    | 100    | 1      | true       | 2018-05-29 06:38:38 | 3             |
      | 151 | 11      | 200    | 99     | 0      | false      | 2018-05-29 06:38:38 | null          |
      | 250 | 13      | 210    | 99     | 0      | true       | 2019-05-29 06:38:38 | 1             |

  Scenario: User has access to the item and the users_answers.idUser = authenticated user's ID (sType='invitationAccepted')
    Given I am the user with ID "1"
    When I send a GET request to "/items/200/attempts"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "151",
        "order": 0,
        "score": 99,
        "start_date": "2018-05-29T06:38:38Z",
        "user_creator": null,
        "validated": false
      },
      {
        "id": "150",
        "order": 1,
        "score": 100,
        "start_date": "2018-05-29T06:38:38Z",
        "user_creator": {
          "first_name": "Jane",
          "last_name": "Doe",
          "login": "jane"
        },
        "validated": true
      }
    ]
    """

  Scenario: User has access to the item and the users_answers.idUser = authenticated user's ID (with limit)
    Given I am the user with ID "1"
    When I send a GET request to "/items/200/attempts?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "151",
        "order": 0,
        "score": 99,
        "start_date": "2018-05-29T06:38:38Z",
        "user_creator": null,
        "validated": false
      }
    ]
    """

  Scenario: User has access to the item and the users_answers.idUser = authenticated user's ID (reverse order)
    Given I am the user with ID "1"
    When I send a GET request to "/items/200/attempts?sort=-order"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "150",
        "order": 1,
        "score": 100,
        "start_date": "2018-05-29T06:38:38Z",
        "user_creator": {
          "first_name": "Jane",
          "last_name": "Doe",
          "login": "jane"
        },
        "validated": true
      },
      {
        "id": "151",
        "order": 0,
        "score": 99,
        "start_date": "2018-05-29T06:38:38Z",
        "user_creator": null,
        "validated": false
      }
    ]
    """

  Scenario: User has access to the item and the user is a team member of groups_attempts.idGroup (items.bHasAttempts=1, sType='requestAccepted')
    Given I am the user with ID "2"
    When I send a GET request to "/items/210/attempts"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "250",
        "order": 0,
        "score": 99,
        "start_date": "2019-05-29T06:38:38Z",
        "user_creator": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "jdoe"
        },
        "validated": true
      }
    ]
    """

  Scenario: User has access to the item and the user is a team member of groups_attempts.idGroup (items.bHasAttempts=1, sType='joinedByCode')
    Given I am the user with ID "3"
    When I send a GET request to "/items/210/attempts"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "250",
        "order": 0,
        "score": 99,
        "start_date": "2019-05-29T06:38:38Z",
        "user_creator": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "jdoe"
        },
        "validated": true
      }
    ]
    """
