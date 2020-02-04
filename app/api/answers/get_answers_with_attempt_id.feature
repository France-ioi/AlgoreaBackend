Feature: Get answers with attempt_id
Background:
  Given the database has the following table 'groups':
    | id | name    | text_id | grade | type  |
    | 11 | jdoe    |         | -2    | User  |
    | 13 | Group B |         | -2    | Class |
    | 21 | owner   |         | -2    | User  |
    | 41 | Group C |         | -2    | Class |
  And the database has the following table 'users':
    | login | temp_user | group_id | first_name  | last_name |
    | jdoe  | 0         | 11       | John        | Doe       |
    | owner | 0         | 21       | Jean-Michel | Blanquer  |
  And the database has the following table 'group_managers':
    | group_id | manager_id |
    | 13       | 21         |
  And the database has the following table 'groups_groups':
    | parent_group_id | child_group_id | personal_info_view_approved_at |
    | 13              | 11             | 2019-05-30 11:00:00            |
  And the database has the following table 'groups_ancestors':
    | ancestor_group_id | child_group_id |
    | 11                | 11             |
    | 13                | 13             |
    | 13                | 11             |
    | 21                | 21             |
    | 41                | 21             |
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
    | 41       | 200     | content_with_descendants |
  And the database has the following table 'attempts':
    | id  | group_id | item_id | order |
    | 100 | 11       | 200     | 1     |
    | 101 | 11       | 200     | 2     |
    | 102 | 11       | 210     | 1     |
  And the database has the following table 'answers':
    | id | author_id | attempt_id | type       | state   | created_at          |
    | 1  | 11        | 100        | Submission | Current | 2017-05-29 06:38:38 |
    | 2  | 11        | 101        | Submission | Current | 2017-05-29 06:38:38 |
    | 3  | 11        | 102        | Submission | Current | 2017-05-29 06:38:38 |
  And the database has the following table 'gradings':
    | answer_id | score | graded_at           |
    | 1         | 100   | 2018-05-29 06:38:38 |
    | 2         | 100   | 2019-05-29 06:38:38 |
    | 3         | 100   | 2019-05-29 06:38:38 |

  Scenario: Full access on the item and the user is a member of the attempt's group
    Given I am the user with id "11"
    When I send a GET request to "/answers?attempt_id=100"
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
    When I send a GET request to "/answers?attempt_id=100"
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
    When I send a GET request to "/answers?attempt_id=101"
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
    When I send a GET request to "/answers?attempt_id=102"
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
