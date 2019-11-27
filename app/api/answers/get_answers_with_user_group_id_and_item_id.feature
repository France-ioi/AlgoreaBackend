Feature: Get answers with (item_id, user_id) pair
Background:
  Given the database has the following table 'groups':
    | id | name       | text_id | grade | type      |
    | 11 | jdoe       |         | -2    | UserSelf  |
    | 12 | jdoe-admin |         | -2    | UserAdmin |
    | 13 | Group B    |         | -2    | Class     |
    | 21 | jdoe       |         | -2    | UserSelf  |
    | 23 | Group C    |         | -2    | Class     |
  And the database has the following table 'users':
    | login | temp_user | group_id | owned_group_id | first_name | last_name |
    | jdoe  | 0         | 11       | 12             | John       | Doe       |
    | other | 0         | 21       | 22             | George     | Bush      |
  And the database has the following table 'group_managers':
    | group_id | manager_id |
    | 11       | 21         |
  And the database has the following table 'groups_groups':
    | id | parent_group_id | child_group_id |
    | 61 | 13              | 11             |
  And the database has the following table 'groups_ancestors':
    | id | ancestor_group_id | child_group_id | is_self |
    | 71 | 11                | 11             | 1       |
    | 72 | 12                | 12             | 1       |
    | 73 | 13                | 13             | 1       |
    | 74 | 13                | 11             | 0       |
    | 76 | 23                | 21             | 0       |
  And the database has the following table 'items':
    | id  | type     | teams_editable | no_score | unlocked_item_ids |
    | 190 | Category | false          | false    | 1234,2345         |
    | 200 | Category | false          | false    | 1234,2345         |
    | 210 | Category | false          | false    | 1234,2345         |
  And the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       |
    | 13       | 190     | none                     |
    | 13       | 200     | content_with_descendants |
    | 13       | 210     | content                  |
    | 23       | 190     | none                     |
    | 23       | 200     | content_with_descendants |
    | 23       | 210     | info                     |
  And the database has the following table 'users_answers':
    | id | user_id | item_id | attempt_id | name             | type       | state   | lang_prog | submitted_at        | score | validated |
    | 1  | 11      | 200     | 1          | My answer        | Submission | Current | python    | 2017-05-29 06:37:38 | 100   | true      |
    | 2  | 11      | 200     | 2          | My second answer | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | true      |
    | 3  | 11      | 210     | 3          | My third answer  | Submission | Current | python    | 2017-05-29 06:39:38 | 100   | true      |

  Scenario: Full access on the item+user_group pair (same user)
    Given I am the user with id "11"
    When I send a GET request to "/answers?item_id=200&user_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "lang_prog": "python",
        "name": "My second answer",
        "score": 100,
        "submitted_at": "2017-05-29T06:38:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        },
        "validated": true
      },
      {
        "id": "1",
        "lang_prog": "python",
        "name": "My answer",
        "score": 100,
        "submitted_at": "2017-05-29T06:37:38Z",
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

  Scenario: Full access on the item+user_group pair (different user)
    Given I am the user with id "21"
    When I send a GET request to "/answers?item_id=200&user_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "lang_prog": "python",
        "name": "My second answer",
        "score": 100,
        "submitted_at": "2017-05-29T06:38:38Z",
        "type": "Submission",
        "user": {
          "login": "jdoe",
          "first_name": "John",
          "last_name": "Doe"
        },
        "validated": true
      },
      {
        "id": "1",
        "lang_prog": "python",
        "name": "My answer",
        "score": 100,
        "submitted_at": "2017-05-29T06:37:38Z",
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

  Scenario: 'Content' access on the item+user_group pair (same user)
    Given I am the user with id "11"
    When I send a GET request to "/answers?item_id=210&user_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "3",
        "lang_prog": "python",
        "name": "My third answer",
        "score": 100,
        "submitted_at": "2017-05-29T06:39:38Z",
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

  Scenario: Full access on the item+user_group pair (same user) [with limit]
    Given I am the user with id "11"
    When I send a GET request to "/answers?item_id=200&user_id=11&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "lang_prog": "python",
        "name": "My second answer",
        "score": 100,
        "submitted_at": "2017-05-29T06:38:38Z",
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

  Scenario: Full access on the item+user_group pair (same user) [with limit and reversed order]
    Given I am the user with id "11"
    When I send a GET request to "/answers?item_id=200&user_id=11&limit=1&sort=submitted_at,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "1",
        "lang_prog": "python",
        "name": "My answer",
        "score": 100,
        "submitted_at": "2017-05-29T06:37:38Z",
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

  Scenario: Start from the second row
    Given I am the user with id "21"
    When I send a GET request to "/answers?item_id=200&user_id=11&from.submitted_at=2017-05-29T06:38:38Z&from.id=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "1",
        "lang_prog": "python",
        "name": "My answer",
        "score": 100,
        "submitted_at": "2017-05-29T06:37:38Z",
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
