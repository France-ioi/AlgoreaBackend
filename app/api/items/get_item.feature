Feature: Get item view information

  Background:
    Given the database has the following table 'groups':
      | id | name             | text_id | grade | type      |
      | 11 | jdoe             |         | -2    | UserSelf  |
      | 12 | jdoe-admin       |         | -2    | UserAdmin |
      | 13 | Group B          |         | -2    | Class     |
      | 14 | nosolution       |         | -2    | UserSelf  |
      | 15 | Group C          |         | -2    | Class     |
      | 16 | nosolution-admin |         | -2    | UserAdmin |
      | 17 | fr               |         | -2    | UserSelf  |
      | 21 | fr-admin         |         | -2    | UserAdmin |
      | 22 | info             |         | -2    | UserSelf  |
      | 26 | info-admin       |         | -2    | UserAdmin |
    And the database has the following table 'users':
      | login      | temp_user | group_id | owned_group_id | default_language |
      | jdoe       | 0         | 11       | 12             |                  |
      | nosolution | 0         | 14       | 16             |                  |
      | fr         | 0         | 17       | 21             | fr               |
      | info       | 0         | 22       | 26             |                  |
    And the database has the following table 'items':
      | id  | type     | no_score | unlocked_item_ids | display_details_in_parent | validation_type | score_min_unlock | contest_entering_condition | teams_editable | contest_max_team_size | has_attempts | duration | group_code_enter | title_bar_visible | read_only | full_screen | show_user_infos | contest_phase | url            | uses_api | hints_allowed |
      | 200 | Category | true     | 1234,2345         | true                      | All             | 100              | All                        | true           | 10                    | true         | 10:20:30 | true             | true              | true      | forceYes    | true            | Running       | http://someurl | true     | true          |
      | 210 | Chapter  | true     | 1234,2345         | true                      | All             | 100              | All                        | true           | 10                    | true         | 10:20:31 | true             | true              | true      | forceYes    | true            | Running       | null           | true     | true          |
      | 220 | Chapter  | true     | 1234,2345         | true                      | All             | 100              | All                        | true           | 10                    | true         | 10:20:32 | true             | true              | true      | forceYes    | true            | Running       | null           | true     | true          |
    And the database has the following table 'items_strings':
      | id | item_id | language_id | title       | image_url                  | subtitle     | description   | edu_comment    |
      | 53 | 200     | 1           | Category 1  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 54 | 210     | 1           | Chapter A   | http://example.com/my1.jpg | Subtitle 1   | Description 1 | Some comment   |
      | 55 | 220     | 1           | Chapter B   | http://example.com/my2.jpg | Subtitle 2   | Description 2 | Some comment   |
      | 63 | 200     | 2           | Catégorie 1 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
      | 64 | 210     | 2           | Chapitre A  | http://example.com/mf1.jpg | Sous-titre 1 | texte 1       | Un commentaire |
      | 66 | 220     | 2           | Chapitre B  | http://example.com/mf2.jpg | Sous-titre 2 | texte 2       | Un commentaire |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 71 | 11                | 11             | 1       |
      | 72 | 12                | 12             | 1       |
      | 73 | 13                | 13             | 1       |
      | 74 | 13                | 11             | 0       |
      | 75 | 15                | 14             | 0       |
      | 76 | 13                | 17             | 0       |
      | 77 | 26                | 22             | 0       |
    And the database has the following table 'items_items':
      | id | parent_item_id | child_item_id | child_order | category  | content_view_propagation |
      | 54 | 200            | 210           | 2           | Discovery | as_info                  |
      | 55 | 200            | 220           | 1           | Discovery | as_info                  |
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order | score | submissions_attempts | validated | finished | key_obtained | hints_cached | started_at          | finished_at         | validated_at        |
      | 101 | 11       | 200     | 1     | 12341 | 11                   | true      | true     | true         | 11           | 2019-01-30 09:26:41 | 2019-02-01 09:26:41 | 2019-01-31 09:26:41 |
      | 102 | 11       | 210     | 1     | 12342 | 12                   | true      | true     | true         | 11           | 2019-01-30 09:26:42 | 2019-02-01 09:26:42 | 2019-01-31 09:26:42 |
      | 103 | 11       | 220     | 1     | 12344 | 14                   | true      | true     | true         | 11           | 2019-01-30 09:26:44 | 2019-02-01 09:26:44 | 2019-01-31 09:26:44 |
      | 104 | 14       | 210     | 1     | 12342 | 12                   | true      | true     | true         | 11           | 2019-01-30 09:26:42 | 2019-02-01 09:26:42 | 2019-01-31 09:26:42 |
      | 105 | 17       | 200     | 1     | 12341 | 11                   | true      | true     | true         | 11           | 2019-01-30 09:26:41 | 2019-02-01 09:26:41 | 2019-01-31 09:26:41 |
      | 106 | 17       | 210     | 1     | 12342 | 12                   | true      | true     | true         | 11           | 2019-01-30 09:26:42 | 2019-02-01 09:26:42 | 2019-01-31 09:26:42 |
      | 107 | 17       | 220     | 1     | 12344 | 14                   | true      | true     | true         | 11           | 2019-01-30 09:26:44 | 2019-02-01 09:26:44 | 2019-01-31 09:26:44 |
      | 108 | 22       | 200     | 1     | 12341 | 11                   | true      | true     | true         | 11           | 2019-01-30 09:26:41 | 2019-02-01 09:26:41 | 2019-01-31 09:26:41 |
      | 109 | 22       | 210     | 1     | 12342 | 12                   | true      | true     | true         | 11           | 2019-01-30 09:26:42 | 2019-02-01 09:26:42 | 2019-01-31 09:26:42 |
      | 110 | 22       | 220     | 1     | 12344 | 14                   | true      | true     | true         | 11           | 2019-01-30 09:26:44 | 2019-02-01 09:26:44 | 2019-01-31 09:26:44 |
    And the database has the following table 'users_items':
      | user_id | item_id | active_attempt_id |
      | 11      | 200     | 101               |
      | 11      | 210     | 102               |
      | 11      | 220     | 103               |
      | 14      | 210     | 104               |
      | 17      | 200     | 105               |
      | 17      | 210     | 106               |
      | 22      | 200     | 108               |
      | 22      | 210     | 109               |
      | 22      | 220     | 110               |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 200     | solution                 |
      | 13       | 210     | solution                 |
      | 13       | 220     | solution                 |
      | 15       | 210     | content_with_descendants |
      | 26       | 200     | solution                 |
      | 26       | 210     | info                     |
      | 26       | 220     | info                     |
    And the database has the following table 'languages':
      | id | code |
      | 2  | fr   |

  Scenario: Full access on all items
    Given I am the user with id "11"
    When I send a GET request to "/items/200"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "200",
      "type": "Category",
      "display_details_in_parent": true,
      "validation_type": "All",
      "has_unlocked_items": true,
      "score_min_unlock": 100,
      "contest_entering_condition": "All",
      "teams_editable": true,
      "contest_max_team_size": 10,
      "has_attempts": true,
      "duration": "10:20:30",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,
      "contest_phase": "Running",
      "url": "http://someurl",
      "uses_api": true,
      "hints_allowed": true,

      "string": {
        "language_id": "1",
        "title": "Category 1",
        "image_url": "http://example.com/my0.jpg",
        "subtitle": "Subtitle 0",
        "description": "Description 0",
        "edu_comment": "Some comment"
      },

      "user_active_attempt": {
        "attempt_id": "101",
        "score": 12341,
        "submissions_attempts": 11,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "started_at": "2019-01-30T09:26:41Z",
        "validated_at": "2019-01-31T09:26:41Z",
        "finished_at": "2019-02-01T09:26:41Z"
      },

      "children": [
        {
          "id": "220",
          "order": 1,
          "category": "Discovery",
          "content_view_propagation": "as_info",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "contest_entering_condition": "All",
          "teams_editable": true,
          "contest_max_team_size": 10,
          "has_attempts": true,
          "duration": "10:20:32",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "1",
            "title": "Chapter B",
            "image_url": "http://example.com/my2.jpg",
            "subtitle": "Subtitle 2",
            "description": "Description 2"
          },

          "user_active_attempt": {
            "attempt_id": "103",
            "score": 12344,
            "submissions_attempts": 14,
            "validated": true,
            "finished": true,
            "key_obtained": true,
            "hints_cached": 11,
            "started_at": "2019-01-30T09:26:44Z",
            "validated_at": "2019-01-31T09:26:44Z",
            "finished_at": "2019-02-01T09:26:44Z"
          }
        },
        {
          "id": "210",

          "order": 2,
          "category": "Discovery",
          "content_view_propagation": "as_info",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "contest_entering_condition": "All",
          "teams_editable": true,
          "contest_max_team_size": 10,
          "has_attempts": true,
          "duration": "10:20:31",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "1",
            "title": "Chapter A",
            "image_url": "http://example.com/my1.jpg",
            "subtitle": "Subtitle 1",
            "description": "Description 1"
          },

          "user_active_attempt": {
            "attempt_id": "102",
            "score": 12342,
            "submissions_attempts": 12,
            "validated": true,
            "finished": true,
            "key_obtained": true,
            "hints_cached": 11,
            "started_at": "2019-01-30T09:26:42Z",
            "validated_at": "2019-01-31T09:26:42Z",
            "finished_at": "2019-02-01T09:26:42Z"
          }
        }
      ]
    }
    """

  Scenario: Chapter as a root node (full access)
    Given I am the user with id "11"
    When I send a GET request to "/items/210"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "210",
      "type": "Chapter",
      "display_details_in_parent": true,
      "validation_type": "All",
      "has_unlocked_items": true,
      "score_min_unlock": 100,
      "contest_entering_condition": "All",
      "teams_editable": true,
      "contest_max_team_size": 10,
      "has_attempts": true,
      "duration": "10:20:31",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,
      "contest_phase": "Running",

      "string": {
        "language_id": "1",
        "title": "Chapter A",
        "image_url": "http://example.com/my1.jpg",
        "subtitle": "Subtitle 1",
        "description": "Description 1",
        "edu_comment": "Some comment"
      },

      "user_active_attempt": {
        "attempt_id": "102",
        "score": 12342,
        "submissions_attempts": 12,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "started_at": "2019-01-30T09:26:42Z",
        "validated_at": "2019-01-31T09:26:42Z",
        "finished_at": "2019-02-01T09:26:42Z"
      },

      "children": []
    }
    """

  Scenario: Chapter as a root node (without solution access)
    Given I am the user with id "14"
    When I send a GET request to "/items/210"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "210",
      "type": "Chapter",
      "display_details_in_parent": true,
      "validation_type": "All",
      "has_unlocked_items": true,
      "score_min_unlock": 100,
      "contest_entering_condition": "All",
      "teams_editable": true,
      "contest_max_team_size": 10,
      "has_attempts": true,
      "duration": "10:20:31",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,
      "contest_phase": "Running",

      "string": {
        "language_id": "1",
        "title": "Chapter A",
        "image_url": "http://example.com/my1.jpg",
        "subtitle": "Subtitle 1",
        "description": "Description 1"
      },

      "user_active_attempt": {
        "attempt_id": "104",
        "score": 12342,
        "submissions_attempts": 12,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "started_at": "2019-01-30T09:26:42Z",
        "validated_at": "2019-01-31T09:26:42Z",
        "finished_at": "2019-02-01T09:26:42Z"
      },

      "children": []
    }
    """

  Scenario: Full access on all items (with user language)
    Given I am the user with id "17"
    When I send a GET request to "/items/200"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "200",
      "type": "Category",
      "display_details_in_parent": true,
      "validation_type": "All",
      "has_unlocked_items": true,
      "score_min_unlock": 100,
      "contest_entering_condition": "All",
      "teams_editable": true,
      "contest_max_team_size": 10,
      "has_attempts": true,
      "duration": "10:20:30",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,
      "contest_phase": "Running",
      "url": "http://someurl",
      "uses_api": true,
      "hints_allowed": true,

      "string": {
        "language_id": "2",
        "title": "Catégorie 1",
        "image_url": "http://example.com/mf0.jpg",
        "subtitle": "Sous-titre 0",
        "description": "texte 0",
        "edu_comment": "Un commentaire"
      },

      "user_active_attempt": {
        "attempt_id": "105",
        "score": 12341,
        "submissions_attempts": 11,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "started_at": "2019-01-30T09:26:41Z",
        "validated_at": "2019-01-31T09:26:41Z",
        "finished_at": "2019-02-01T09:26:41Z"
      },

      "children": [
        {
          "id": "220",
          "order": 1,
          "category": "Discovery",
          "content_view_propagation": "as_info",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "contest_entering_condition": "All",
          "teams_editable": true,
          "contest_max_team_size": 10,
          "has_attempts": true,
          "duration": "10:20:32",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "2",
            "title": "Chapitre B",
            "image_url": "http://example.com/mf2.jpg",
            "subtitle": "Sous-titre 2",
            "description": "texte 2"
          },

          "user_active_attempt": null
        },
        {
          "id": "210",

          "order": 2,
          "category": "Discovery",
          "content_view_propagation": "as_info",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "contest_entering_condition": "All",
          "teams_editable": true,
          "contest_max_team_size": 10,
          "has_attempts": true,
          "duration": "10:20:31",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "2",
            "title": "Chapitre A",
            "image_url": "http://example.com/mf1.jpg",
            "subtitle": "Sous-titre 1",
            "description": "texte 1"
          },

          "user_active_attempt": {
            "attempt_id": "106",
            "score": 12342,
            "submissions_attempts": 12,
            "validated": true,
            "finished": true,
            "key_obtained": true,
            "hints_cached": 11,
            "started_at": "2019-01-30T09:26:42Z",
            "validated_at": "2019-01-31T09:26:42Z",
            "finished_at": "2019-02-01T09:26:42Z"
          }
        }
      ]
    }
    """

  Scenario: Info access on children
    Given I am the user with id "22"
    When I send a GET request to "/items/200"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "200",
      "type": "Category",
      "display_details_in_parent": true,
      "validation_type": "All",
      "has_unlocked_items": true,
      "score_min_unlock": 100,
      "contest_entering_condition": "All",
      "teams_editable": true,
      "contest_max_team_size": 10,
      "has_attempts": true,
      "duration": "10:20:30",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,
      "contest_phase": "Running",
      "url": "http://someurl",
      "uses_api": true,
      "hints_allowed": true,

      "string": {
        "language_id": "1",
        "title": "Category 1",
        "image_url": "http://example.com/my0.jpg",
        "subtitle": "Subtitle 0",
        "description": "Description 0",
        "edu_comment": "Some comment"
      },

      "user_active_attempt": {
        "attempt_id": "108",
        "score": 12341,
        "submissions_attempts": 11,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "started_at": "2019-01-30T09:26:41Z",
        "validated_at": "2019-01-31T09:26:41Z",
        "finished_at": "2019-02-01T09:26:41Z"
      },

      "children": [
        {
          "id": "220",
          "order": 1,
          "category": "Discovery",
          "content_view_propagation": "as_info",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "contest_entering_condition": "All",
          "teams_editable": true,
          "contest_max_team_size": 10,
          "has_attempts": true,
          "duration": "10:20:32",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "1",
            "title": "Chapter B",
            "image_url": "http://example.com/my2.jpg"
          },

          "user_active_attempt": null
        },
        {
          "id": "210",

          "order": 2,
          "category": "Discovery",
          "content_view_propagation": "as_info",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "contest_entering_condition": "All",
          "teams_editable": true,
          "contest_max_team_size": 10,
          "has_attempts": true,
          "duration": "10:20:31",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "1",
            "title": "Chapter A",
            "image_url": "http://example.com/my1.jpg"
          },

          "user_active_attempt": null
        }
      ]
    }
    """
