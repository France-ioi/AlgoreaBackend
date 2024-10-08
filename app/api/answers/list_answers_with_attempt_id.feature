Feature: List answers by attempt_id
Background:
  Given the database has the following table "groups":
    | id | name    | grade | type  |
    | 11 | jdoe    | -2    | User  |
    | 13 | Group B | -2    | Class |
    | 21 | owner   | -2    | User  |
    | 41 | Group C | -2    | Class |
  And the database has the following table "users":
    | login | temp_user | group_id | first_name  | last_name |
    | jdoe  | 0         | 11       | John        | Doe       |
    | owner | 0         | 21       | Jean-Michel | Blanquer  |
  And the database has the following table "group_managers":
    | group_id | manager_id |
    | 13       | 21         |
  And the database has the following table "groups_groups":
    | parent_group_id | child_group_id | personal_info_view_approved_at |
    | 13              | 11             | 2019-05-30 11:00:00            |
    | 41              | 21             | null                           |
  And the groups ancestors are computed
  And the database has the following table "items":
    | id  | type    | no_score | default_language_tag |
    | 190 | Chapter | false    | fr                   |
    | 200 | Chapter | false    | fr                   |
    | 210 | Chapter | false    | fr                   |
  And the database has the following table "permissions_generated":
    | group_id | item_id | can_view_generated       |
    | 13       | 190     | none                     |
    | 13       | 200     | content_with_descendants |
    | 13       | 210     | content                  |
    | 41       | 200     | content_with_descendants |
  And the database has the following table "attempts":
    | id | participant_id |
    | 1  | 11             |
    | 2  | 11             |
  And the database has the following table "results":
    | attempt_id | participant_id | item_id |
    | 1          | 11             | 200     |
    | 2          | 11             | 200     |
    | 1          | 11             | 210     |
  And the database has the following table "answers":
    | id | author_id | participant_id | attempt_id | item_id | type       | state  | created_at          |
    | 1  | 11        | 11             | 1          | 200     | Submission | State1 | 2017-05-29 06:38:38 |
    | 2  | 11        | 11             | 2          | 200     | Submission | State2 | 2017-05-29 06:38:38 |
    | 3  | 11        | 11             | 1          | 210     | Submission | State3 | 2017-05-29 06:38:38 |
  And the database has the following table "gradings":
    | answer_id | score | graded_at           |
    | 1         | 100   | 2018-05-29 06:38:38 |
    | 2         | 100   | 2019-05-29 06:38:38 |
    | 3         | 100   | 2019-05-29 06:38:38 |

  Scenario: Full access on the item and the user is a member of the attempt's group
    Given I am the user with id "11"
    When I send a GET request to "/items/200/answers?attempt_id=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "1",
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

  Scenario: Full access on the item and the user is a manager of attempt's group
    Given I am the user with id "21"
    When I send a GET request to "/items/200/answers?attempt_id=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "1",
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

  Scenario: Full access on the item and the user's self group is the attempts.group_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/answers?attempt_id=2"
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

  Scenario: 'Content' access on the item and the user's self group is the attempts.group_id
    Given I am the user with id "11"
    When I send a GET request to "/items/210/answers?attempt_id=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "3",
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
