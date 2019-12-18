Feature: Get recent activity for group_id and item_id
  Background:
    Given the database has the following users:
      | login | temp_user | group_id | first_name  | last_name | default_language |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | fr               |
      | user  | 0         | 11       | John        | Doe       | en               |
      | jane  | 0         | 31       | Jane        | Doe       | en               |
    And the database has the following table 'groups':
      | id |
      | 13 |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 75 | 11                | 11             | 1       |
      | 76 | 13                | 11             | 0       |
      | 77 | 13                | 13             | 1       |
      | 78 | 21                | 21             | 1       |
    And the database has the following table 'groups_attempts':
      | id  | item_id | group_id | order |
      | 100 | 200     | 11       | 1     |
      | 101 | 200     | 11       | 2     |
    And the database has the following table 'users_answers':
      | id | user_id | attempt_id | name             | type       | state   | lang_prog | submitted_at        | score | validated |
      | 2  | 11      | 101        | My second anwser | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | true      |
      | 1  | 11      | 100        | My answer        | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | false     |
      | 3  | 11      | 101        | My third anwser  | Submission | Current | python    | 2017-05-30 06:38:38 | 100   | true      |
      | 4  | 11      | 101        | My fourth answer | Saved      | Current | python    | 2017-05-30 06:38:38 | 100   | true      |
      | 5  | 11      | 101        | My fifth answer  | Current    | Current | python    | 2017-05-30 06:38:38 | 100   | true      |
      | 6  | 31      | 101        | My second anwser | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | true      |
      | 7  | 31      | 100        | My answer        | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | false     |
      | 8  | 31      | 101        | My third anwser  | Submission | Current | python    | 2017-05-30 06:38:38 | 100   | true      |
    And the database has the following table 'items':
      | id  | type     | teams_editable | no_score |
      | 200 | Category | false          | false    |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 21       | 200     | info               |
    And the database has the following table 'items_ancestors':
      | id | ancestor_item_id | child_item_id |
      | 1  | 200              | 200           |
    And the database has the following table 'items_strings':
      | id | item_id | language_id | title       | image_url                  | subtitle     | description   | edu_comment    |
      | 53 | 200     | 1           | Category 1  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 63 | 200     | 2           | Catégorie 1 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
    And the database has the following table 'languages':
      | id | code |
      | 2  | fr   |

  Scenario: User is a manager of the group and there are visible descendants of the item
    This spec also checks:
      1) that answers having type!="Submission" are filtered out,
      2) answers ordering,
      3) filtering by users groups
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/recent_activity?item_id=200"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "3",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Category"
        },
        "score": 100,
        "submitted_at": "2017-05-30T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "validated": true
      },
      {
        "id": "1",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Category"
        },
        "score": 100,
        "submitted_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "validated": false
      },
      {
        "id": "2",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Category"
        },
        "score": 100,
        "submitted_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "validated": true
      }
    ]
    """

  Scenario: User is a manager of the group and there are visible descendants of the item; request the first row
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "3",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Category"
        },
        "score": 100,
        "submitted_at": "2017-05-30T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "validated": true
      }
    ]
    """

  Scenario: User is a manager of the group and there are visible descendants of the item; request the second and the third rows
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&from.submitted_at=2017-05-30T06:38:38Z&from.id=3"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "1",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Category"
        },
        "score": 100,
        "submitted_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "validated": false
      },
      {
        "id": "2",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Category"
        },
        "score": 100,
        "submitted_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "validated": true
      }
    ]
    """

  Scenario: User is a manager of the group and there are visible descendants of the item; request the third row
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&from.submitted_at=2017-05-29T06:38:38Z&from.id=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Category"
        },
        "score": 100,
        "submitted_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "validated": true
      }
    ]
    """

  Scenario: User is a manager of the group and there are visible descendants of the item; request validated answers only
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&validated=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "3",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Category"
        },
        "score": 100,
        "submitted_at": "2017-05-30T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "validated": true
      },
      {
        "id": "2",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Category"
        },
        "score": 100,
        "submitted_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "validated": true
      }
    ]
    """
