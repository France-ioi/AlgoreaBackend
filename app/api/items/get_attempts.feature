Feature: Get groups attempts for current user and item_id
  Background:
    Given the database has the following table 'users':
      | id | login | self_group_id | owned_group_id | first_name | last_name |
      | 1  | jdoe  | 11            | 12             | John       | Doe       |
      | 2  | other | 21            | 22             | George     | Bush      |
      | 3  | jane  | 31            | 32             | Jane       | Doe       |
    And the database has the following table 'groups':
      | id | name        | type      |
      | 11 | jdoe        | UserSelf  |
      | 12 | jdoe-admin  | UserAdmin |
      | 13 | Group B     | Class     |
      | 21 | other       | UserSelf  |
      | 22 | other-admin | UserAdmin |
      | 23 | Group C     | Class     |
      | 31 | jane        | UserSelf  |
      | 32 | jane-admin  | UserAdmin |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | type               |
      | 61 | 13              | 11             | invitationAccepted |
      | 62 | 13              | 21             | requestAccepted    |
      | 63 | 13              | 31             | joinedByCode       |
      | 64 | 23              | 21             | direct             |
      | 65 | 23              | 31             | direct             |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 71 | 11                | 11             | 1       |
      | 72 | 12                | 12             | 1       |
      | 73 | 13                | 13             | 1       |
      | 74 | 13                | 11             | 0       |
      | 75 | 13                | 21             | 0       |
      | 76 | 13                | 31             | 0       |
      | 77 | 23                | 21             | 0       |
      | 78 | 23                | 23             | 1       |
      | 79 | 23                | 31             | 0       |
      | 80 | 31                | 31             | 1       |
      | 81 | 32                | 32             | 1       |
    And the database has the following table 'items':
      | id  | has_attempts |
      | 200 | 0            |
      | 210 | 1            |
    And the database has the following table 'groups_items':
      | id | group_id | item_id | cached_full_access_since | cached_partial_access_since | creator_user_id |
      | 43 | 13       | 200     | 2017-05-29 06:38:38      | 2017-05-29 06:38:38         | 1               |
      | 46 | 23       | 210     | 2017-05-29 06:38:38      | 2017-05-29 06:38:38         | 1               |
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | score | order | validated | started_at          | creator_user_id |
      | 150 | 11       | 200     | 100   | 1     | true      | 2018-05-29 06:38:38 | 3               |
      | 151 | 11       | 200     | 99    | 0     | false     | 2018-05-29 06:38:38 | null            |
      | 250 | 13       | 210     | 99    | 0     | true      | 2019-05-29 06:38:38 | 1               |

  Scenario: User has access to the item and the users_answers.user_id = authenticated user's id (type='invitationAccepted')
    Given I am the user with id "1"
    When I send a GET request to "/items/200/attempts"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "151",
        "order": 0,
        "score": 99,
        "started_at": "2018-05-29T06:38:38Z",
        "user_creator": null,
        "validated": false
      },
      {
        "id": "150",
        "order": 1,
        "score": 100,
        "started_at": "2018-05-29T06:38:38Z",
        "user_creator": {
          "first_name": "Jane",
          "last_name": "Doe",
          "login": "jane"
        },
        "validated": true
      }
    ]
    """

  Scenario: User has access to the item and the users_answers.user_id = authenticated user's id (with limit)
    Given I am the user with id "1"
    When I send a GET request to "/items/200/attempts?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "151",
        "order": 0,
        "score": 99,
        "started_at": "2018-05-29T06:38:38Z",
        "user_creator": null,
        "validated": false
      }
    ]
    """

  Scenario: User has access to the item and the users_answers.user_id = authenticated user's id (reverse order)
    Given I am the user with id "1"
    When I send a GET request to "/items/200/attempts?sort=-order"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "150",
        "order": 1,
        "score": 100,
        "started_at": "2018-05-29T06:38:38Z",
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
        "started_at": "2018-05-29T06:38:38Z",
        "user_creator": null,
        "validated": false
      }
    ]
    """

  Scenario: User has access to the item and the users_answers.user_id = authenticated user's id (reverse order, start from the second row)
    Given I am the user with id "1"
    When I send a GET request to "/items/200/attempts?sort=-order&from.order=1&from.id=150"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "151",
        "order": 0,
        "score": 99,
        "started_at": "2018-05-29T06:38:38Z",
        "user_creator": null,
        "validated": false
      }
    ]
    """

  Scenario: User has access to the item and the user is a team member of groups_attempts.group_id (items.has_attempts=1, type='requestAccepted')
    Given I am the user with id "2"
    When I send a GET request to "/items/210/attempts"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "250",
        "order": 0,
        "score": 99,
        "started_at": "2019-05-29T06:38:38Z",
        "user_creator": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "jdoe"
        },
        "validated": true
      }
    ]
    """

  Scenario: User has access to the item and the user is a team member of groups_attempts.group_id (items.has_attempts=1, type='joinedByCode')
    Given I am the user with id "3"
    When I send a GET request to "/items/210/attempts"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "250",
        "order": 0,
        "score": 99,
        "started_at": "2019-05-29T06:38:38Z",
        "user_creator": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "jdoe"
        },
        "validated": true
      }
    ]
    """
