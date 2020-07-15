Feature: List attempts for current user and item_id
  Background:
    Given the database has the following table 'groups':
      | id | name    | type  |
      | 11 | jdoe    | User  |
      | 13 | Group B | Class |
      | 21 | other   | User  |
      | 23 | Group C | Team  |
    And the database has the following table 'users':
      | login | group_id | first_name | last_name |
      | jdoe  | 11       | John       | Doe       |
      | other | 21       | George     | Bush      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 11             |
      | 13              | 21             |
      | 23              | 21             |
      | 23              | 31             |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id  | allows_multiple_attempts | default_language_tag | is_root |
      | 200 | 0                        | fr                   | true    |
      | 210 | 1                        | fr                   | false   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 200     | content                  |
      | 13       | 210     | info                     |
      | 23       | 210     | content_with_descendants |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | creator_id | parent_attempt_id | ended_at            |
      | 1  | 11             | 2018-05-29 05:38:38 | 21         | 0                 | 2018-05-29 05:38:38 |
      | 0  | 11             | 2018-05-29 05:38:38 | null       | null              | null                |
      | 0  | 23             | 2019-05-29 05:38:38 | 11         | null              | null                |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | score_computed | validated_at        | started_at          | latest_activity_at  |
      | 1          | 11             | 200     | 100            | 2018-05-29 07:00:00 | 2018-05-29 06:38:38 | 2018-05-29 06:38:39 |
      | 0          | 11             | 200     | 99             | null                | 2018-05-29 06:38:38 | 2018-05-29 06:38:39 |
      | 0          | 23             | 210     | 99             | 2018-05-29 08:00:00 | 2019-05-29 06:38:38 | 2019-05-29 06:38:39 |

  Scenario: User has access to the item and the attempts.group_id = authenticated user's group_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/attempts?parent_attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "0",
        "created_at": "2018-05-29T05:38:38Z",
        "score_computed": 99,
        "allows_submissions_until": "9999-12-31T23:59:59Z",
        "started_at": "2018-05-29T06:38:38Z",
        "ended_at": null,
        "latest_activity_at": "2018-05-29T06:38:39Z",
        "user_creator": null,
        "validated": false
      },
      {
        "id": "1",
        "created_at": "2018-05-29T05:38:38Z",
        "score_computed": 100,
        "allows_submissions_until": "9999-12-31T23:59:59Z",
        "started_at": "2018-05-29T06:38:38Z",
        "ended_at": "2018-05-29T05:38:38Z",
        "latest_activity_at": "2018-05-29T06:38:39Z",
        "user_creator": {
          "group_id": "21",
          "first_name": "George",
          "last_name": "Bush",
          "login": "other"
        },
        "validated": true
      }
    ]
    """

  Scenario: User has access to the item and the attempts.group_id = authenticated user's group_id (with limit)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/attempts?parent_attempt_id=0&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "0",
        "created_at": "2018-05-29T05:38:38Z",
        "score_computed": 99,
        "allows_submissions_until": "9999-12-31T23:59:59Z",
        "started_at": "2018-05-29T06:38:38Z",
        "ended_at": null,
        "latest_activity_at": "2018-05-29T06:38:39Z",
        "user_creator": null,
        "validated": false
      }
    ]
    """

  Scenario: User has access to the item and the attempts.group_id = authenticated user's group_id (reverse order)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/attempts?parent_attempt_id=0&sort=-id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "1",
        "created_at": "2018-05-29T05:38:38Z",
        "score_computed": 100,
        "allows_submissions_until": "9999-12-31T23:59:59Z",
        "started_at": "2018-05-29T06:38:38Z",
        "ended_at": "2018-05-29T05:38:38Z",
        "latest_activity_at": "2018-05-29T06:38:39Z",
        "user_creator": {
          "group_id": "21",
          "first_name": "George",
          "last_name": "Bush",
          "login": "other"
        },
        "validated": true
      },
      {
        "id": "0",
        "created_at": "2018-05-29T05:38:38Z",
        "score_computed": 99,
        "allows_submissions_until": "9999-12-31T23:59:59Z",
        "started_at": "2018-05-29T06:38:38Z",
        "ended_at": null,
        "latest_activity_at": "2018-05-29T06:38:39Z",
        "user_creator": null,
        "validated": false
      }
    ]
    """

  Scenario: User has access to the item and the attempts.group_id = authenticated user's group_id (reverse order, start from the second row)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/attempts?parent_attempt_id=0&sort=-id&from.id=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "0",
        "created_at": "2018-05-29T05:38:38Z",
        "score_computed": 99,
        "allows_submissions_until": "9999-12-31T23:59:59Z",
        "started_at": "2018-05-29T06:38:38Z",
        "ended_at": null,
        "latest_activity_at": "2018-05-29T06:38:39Z",
        "user_creator": null,
        "validated": false
      }
    ]
    """

  Scenario: Team has access to the item and the attempts.group_id = team's group_id
    Given I am the user with id "21"
    When I send a GET request to "/items/210/attempts?parent_attempt_id=0&as_team_id=23"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "0",
        "created_at": "2019-05-29T05:38:38Z",
        "score_computed": 99,
        "allows_submissions_until": "9999-12-31T23:59:59Z",
        "started_at": "2019-05-29T06:38:38Z",
        "ended_at": null,
        "latest_activity_at": "2019-05-29T06:38:39Z",
        "user_creator": {
          "group_id": "11",
          "first_name": "John",
          "last_name": "Doe",
          "login": "jdoe"
        },
        "validated": true
      }
    ]
    """
