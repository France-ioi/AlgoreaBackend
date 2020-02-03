Feature: Get answers with (item_id, author_id) pair
Background:
  Given the database has the following table 'groups':
    | id | name    | text_id | grade | type     |
    | 11 | jdoe    |         | -2    | UserSelf |
    | 13 | Group B |         | -2    | Class    |
    | 21 | other   |         | -2    | UserSelf |
    | 23 | Group C |         | -2    | Class    |
    | 25 | jane    |         | -2    | UserSelf |
  And the database has the following table 'users':
    | login | temp_user | group_id | first_name | last_name |
    | jdoe  | 0         | 11       | John       | Doe       |
    | other | 0         | 21       | George     | Bush      |
    | jane  | 0         | 25       | Jane       | Doe       |
  And the database has the following table 'group_managers':
    | group_id | manager_id |
    | 13       | 21         |
  And the database has the following table 'groups_groups':
    | parent_group_id | child_group_id | personal_info_view_approved_at |
    | 13              | 11             | 2019-05-30 11:00:00            |
    | 13              | 25             | null                           |
  And the database has the following table 'groups_ancestors':
    | ancestor_group_id | child_group_id |
    | 11                | 11             |
    | 13                | 11             |
    | 13                | 13             |
    | 13                | 25             |
    | 21                | 21             |
    | 23                | 21             |
    | 23                | 23             |
  And the database has the following table 'items':
    | id  | type    | teams_editable | no_score | default_language_tag |
    | 190 | Chapter | false          | false    | fr                   |
    | 200 | Chapter | false          | false    | fr                   |
    | 210 | Chapter | false          | false    | fr                   |
  And the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       |
    | 13       | 190     | none                     |
    | 13       | 200     | content_with_descendants |
    | 13       | 210     | content                  |
    | 23       | 190     | none                     |
    | 23       | 200     | content_with_descendants |
    | 23       | 210     | info                     |
  And the database has the following table 'attempts':
    | id | group_id | item_id | order |
    | 1  | 11       | 200     | 1     |
    | 2  | 11       | 200     | 2     |
    | 3  | 11       | 210     | 1     |
    | 4  | 13       | 200     | 1     |
  And the database has the following table 'answers':
    | id | author_id | attempt_id | type       | state   | created_at          |
    | 1  | 11        | 1          | Submission | Current | 2017-05-29 06:37:38 |
    | 2  | 11        | 2          | Submission | Current | 2017-05-29 06:38:38 |
    | 3  | 11        | 3          | Submission | Current | 2017-05-29 06:39:38 |
    | 4  | 25        | 4          | Submission | Current | 2017-05-29 06:39:38 |
  And the database has the following table 'gradings':
    | answer_id | score | graded_at           |
    | 1         | 100   | 2018-05-29 06:38:38 |
    | 2         | 100   | 2019-05-29 06:38:38 |
    | 3         | 100   | 2019-05-29 06:38:38 |

  Scenario: Full access on the item+user_group pair (same user)
    Given I am the user with id "11"
    When I send a GET request to "/answers?item_id=200&author_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "score": 100,
        "created_at": "2017-05-29T06:38:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        }
      },
      {
        "id": "1",
        "score": 100,
        "created_at": "2017-05-29T06:37:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        }
      }
    ]
    """

  Scenario: Full access on the item+user_group pair (different user)
    Given I am the user with id "21"
    When I send a GET request to "/answers?item_id=200&author_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "score": 100,
        "created_at": "2017-05-29T06:38:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        }
      },
      {
        "id": "1",
        "score": 100,
        "created_at": "2017-05-29T06:37:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        }
      }
    ]
    """

  Scenario: Full access on the item+user_group pair (different user, no approval to view personal info)
    Given I am the user with id "21"
    When I send a GET request to "/answers?item_id=200&author_id=25"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "4",
        "score": null,
        "created_at": "2017-05-29T06:39:38Z",
        "type": "Submission",
        "user": {
          "login": "jane",
          "first_name": null,
          "last_name": null
        }
      }
    ]
    """

  Scenario: 'Content' access on the item+user_group pair (same user)
    Given I am the user with id "11"
    When I send a GET request to "/answers?item_id=210&author_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "3",
        "score": 100,
        "created_at": "2017-05-29T06:39:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        }
      }
    ]
    """

  Scenario: Full access on the item+user_group pair (same user) [with limit]
    Given I am the user with id "11"
    When I send a GET request to "/answers?item_id=200&author_id=11&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "score": 100,
        "created_at": "2017-05-29T06:38:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        }
      }
    ]
    """

  Scenario: Full access on the item+user_group pair (same user) [with limit and reversed order]
    Given I am the user with id "11"
    When I send a GET request to "/answers?item_id=200&author_id=11&limit=1&sort=created_at,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "1",
        "score": 100,
        "created_at": "2017-05-29T06:37:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        }
      }
    ]
    """

  Scenario: Start from the second row
    Given I am the user with id "21"
    When I send a GET request to "/answers?item_id=200&author_id=11&from.created_at=2017-05-29T06:38:38Z&from.id=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "1",
        "score": 100,
        "created_at": "2017-05-29T06:37:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        }
      }
    ]
    """
