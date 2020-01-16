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
      | 11       | 31         |
      | 13       | 21         |
      | 31       | 31         |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 13              | 11             | 2019-05-30 11:00:00            |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 11             | 0       |
      | 13                | 13             | 1       |
      | 21                | 21             | 1       |
      | 31                | 31             | 1       |
    And the database has the following table 'attempts':
      | id  | item_id | group_id | order |
      | 100 | 200     | 11       | 1     |
      | 101 | 200     | 11       | 2     |
    And the database has the following table 'answers':
      | id | author_id | attempt_id | type       | state   | created_at          |
      | 2  | 11        | 101        | Submission | Current | 2017-05-29 06:38:38 |
      | 1  | 11        | 100        | Submission | Current | 2017-05-29 06:38:38 |
      | 3  | 11        | 101        | Submission | Current | 2017-05-30 06:38:38 |
      | 4  | 11        | 101        | Saved      | Current | 2017-05-30 06:38:38 |
      | 5  | 11        | 101        | Current    | Current | 2017-05-30 06:38:38 |
      | 6  | 31        | 101        | Submission | Current | 2017-05-29 06:38:38 |
      | 7  | 31        | 100        | Submission | Current | 2017-05-29 06:38:38 |
      | 8  | 31        | 101        | Submission | Current | 2017-05-30 06:38:38 |
    And the database has the following table 'gradings':
      | answer_id | graded_at           | score |
      | 2         | 2017-05-29 06:38:38 | 100   |
      | 1         | 2017-05-29 06:38:38 | 99    |
      | 3         | 2017-05-30 06:38:38 | 100   |
      | 4         | 2017-05-30 06:38:38 | 100   |
      | 5         | 2017-05-30 06:38:38 | 100   |
      | 6         | 2017-05-29 06:38:38 | 100   |
      | 7         | 2017-05-29 06:38:38 | 98    |
      | 8         | 2017-05-30 06:38:38 | 100   |
    And the database has the following table 'items':
      | id  | type    | teams_editable | no_score |
      | 200 | Chapter | false          | false    |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 21       | 200     | info               |
      | 31       | 200     | info               |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 200              | 200           |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title       | image_url                  | subtitle     | description   | edu_comment    |
      | 200     | en           | Category 1  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 200     | fr           | Catégorie 1 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
    And the database has the following table 'languages':
      | tag |
      | fr  |

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
          "type": "Chapter"
        },
        "score": 100,
        "created_at": "2017-05-30T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        }
      },
      {
        "id": "1",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Chapter"
        },
        "score": 99,
        "created_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        }
      },
      {
        "id": "2",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Chapter"
        },
        "score": 100,
        "created_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        }
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
          "type": "Chapter"
        },
        "score": 100,
        "created_at": "2017-05-30T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        }
      }
    ]
    """

  Scenario: User is a manager of the group and there are visible descendants of the item; request the second and the third rows
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&from.created_at=2017-05-30T06:38:38Z&from.id=3"
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
          "type": "Chapter"
        },
        "score": 99,
        "created_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        }
      },
      {
        "id": "2",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Chapter"
        },
        "score": 100,
        "created_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        }
      }
    ]
    """

  Scenario: User is a manager of the group and there are visible descendants of the item; request the third row
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&from.created_at=2017-05-29T06:38:38Z&from.id=1"
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
          "type": "Chapter"
        },
        "score": 100,
        "created_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        }
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
          "type": "Chapter"
        },
        "score": 100,
        "created_at": "2017-05-30T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        }
      },
      {
        "id": "2",
        "item": {
          "id": "200",
          "string": {
            "title": "Catégorie 1"
          },
          "type": "Chapter"
        },
        "score": 100,
        "created_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        }
      }
    ]
    """

  Scenario: User is a manager of the group and there are visible descendants of the item; request unvalidated answers only
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&validated=0"
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
          "type": "Chapter"
        },
        "score": 99,
        "created_at": "2017-05-29T06:38:38Z",
        "user": {
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        }
      }
    ]
    """

  Scenario: User can see their own name
    Given I am the user with id "31"
    When I send a GET request to "/groups/31/recent_activity?item_id=200&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "created_at": "2017-05-30T06:38:38Z",
        "id": "8",
        "item": {
          "id": "200",
          "string": {
            "title": "Category 1"
          },
          "type": "Chapter"
        },
        "score": 100,
        "user": {
          "first_name": "Jane",
          "last_name": "Doe",
          "login": "jane"
        }
      }
    ]
    """

  Scenario: User cannot see names without approval
    Given I am the user with id "31"
    When I send a GET request to "/groups/11/recent_activity?item_id=200&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "created_at": "2017-05-30T06:38:38Z",
        "id": "3",
        "item": {
          "id": "200",
          "string": {
            "title": "Category 1"
          },
          "type": "Chapter"
        },
        "score": 100,
        "user": {
          "first_name": null,
          "last_name": null,
          "login": "user"
        }
      }
    ]
    """
