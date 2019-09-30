Feature: Get answers with (item_id, user_id) pair
Background:
  Given the database has the following table 'users':
    | id | login | temp_user | self_group_id | owned_group_id | first_name | last_name |
    | 1  | jdoe  | 0         | 11            | 12             | John       | Doe       |
    | 2  | other | 0         | 21            | 22             | George     | Bush      |
  And the database has the following table 'groups':
    | id | name       | text_id | grade | type      | version |
    | 11 | jdoe       |         | -2    | UserAdmin | 0       |
    | 12 | jdoe-admin |         | -2    | UserAdmin | 0       |
    | 13 | Group B    |         | -2    | Class     | 0       |
    | 23 | Group C    |         | -2    | Class     | 0       |
  And the database has the following table 'groups_groups':
    | id | parent_group_id | child_group_id | version |
    | 61 | 13              | 11             | 0       |
    | 62 | 22              | 13             | 0       |
  And the database has the following table 'groups_ancestors':
    | id | ancestor_group_id | child_group_id | is_self | version |
    | 71 | 11                | 11             | 1       | 0       |
    | 72 | 12                | 12             | 1       | 0       |
    | 73 | 13                | 13             | 1       | 0       |
    | 74 | 13                | 11             | 0       | 0       |
    | 75 | 22                | 11             | 0       | 0       |
    | 76 | 23                | 21             | 0       | 0       |
  And the database has the following table 'items':
    | id  | type     | teams_editable | no_score | unlocked_item_ids | transparent_folder | version |
    | 190 | Category | false          | false    | 1234,2345         | true               | 0       |
    | 200 | Category | false          | false    | 1234,2345         | true               | 0       |
    | 210 | Category | false          | false    | 1234,2345         | true               | 0       |
  And the database has the following table 'groups_items':
    | id | group_id | item_id | cached_full_access_since | cached_partial_access_since | cached_grayed_access_since | creator_user_id | version |
    | 42 | 13       | 190     | 2037-05-29 06:38:38      | 2037-05-29 06:38:38         | 2037-05-29 06:38:38        | 0               | 0       |
    | 43 | 13       | 200     | 2017-05-29 06:38:38      | 2017-05-29 06:38:38         | 2017-05-29 06:38:38        | 0               | 0       |
    | 44 | 13       | 210     | 2037-05-29 06:38:38      | 2017-05-29 06:38:38         | 2017-05-29 06:38:38        | 0               | 0       |
    | 45 | 23       | 190     | 2037-05-29 06:38:38      | 2037-05-29 06:38:38         | 2037-05-29 06:38:38        | 0               | 0       |
    | 46 | 23       | 200     | 2017-05-29 06:38:38      | 2017-05-29 06:38:38         | 2017-05-29 06:38:38        | 0               | 0       |
    | 47 | 23       | 210     | 2037-05-29 06:38:38      | 2037-05-29 06:38:38         | 2017-05-29 06:38:38        | 0               | 0       |
  And the database has the following table 'users_answers':
    | id | user_id | item_id | attempt_id | name             | type       | state   | lang_prog | submitted_at        | score | validated |
    | 1  | 1       | 200     | 1          | My answer        | Submission | Current | python    | 2017-05-29 06:37:38 | 100   | true      |
    | 2  | 1       | 200     | 2          | My second answer | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | true      |
    | 3  | 1       | 210     | 3          | My third answer  | Submission | Current | python    | 2017-05-29 06:39:38 | 100   | true      |

  Scenario: Full access on the item+user pair (same user)
    Given I am the user with id "1"
    When I send a GET request to "/answers?item_id=200&user_id=1"
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

  Scenario: Full access on the item+user pair (different user)
    Given I am the user with id "2"
    When I send a GET request to "/answers?item_id=200&user_id=1"
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

  Scenario: Partial access on the item+user pair (same user)
    Given I am the user with id "1"
    When I send a GET request to "/answers?item_id=210&user_id=1"
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

  Scenario: Full access on the item+user pair (same user) [with limit]
    Given I am the user with id "1"
    When I send a GET request to "/answers?item_id=200&user_id=1&limit=1"
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

  Scenario: Full access on the item+user pair (same user) [with limit and reversed order]
    Given I am the user with id "1"
    When I send a GET request to "/answers?item_id=200&user_id=1&limit=1&sort=submitted_at"
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
    Given I am the user with id "2"
    When I send a GET request to "/answers?item_id=200&user_id=1&from.submitted_at=2017-05-29T06:38:38Z&from.id=2"
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
