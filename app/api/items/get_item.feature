Feature: Get item view information

  Background:
    Given the database has the following table 'users':
      | id | login      | temp_user | self_group_id | owned_group_id | default_language | version |
      | 1  | jdoe       | 0         | 11            | 12             |                  | 0       |
      | 2  | nosolution | 0         | 14            | 16             |                  | 0       |
      | 3  | fr         | 0         | 17            | 21             | fr               | 0       |
      | 4  | grayed     | 0         | 22            | 26             |                  | 0       |
    And the database has the following table 'groups':
      | id | name       | text_id | grade | type      | version |
      | 11 | jdoe       |         | -2    | UserAdmin | 0       |
      | 12 | jdoe-admin |         | -2    | UserAdmin | 0       |
      | 13 | Group B    |         | -2    | Class     | 0       |
      | 14 | nosolution |         | -2    | UserAdmin | 0       |
      | 15 | Group C    |         | -2    | Class     | 0       |
      | 22 | grayed     |         | -2    | Class     | 0       |
      | 26 | Group D    |         | -2    | Class     | 0       |
    And the database has the following table 'items':
      | id  | type     | no_score | unlocked_item_ids | access_open_date    | display_details_in_parent | validation_type | score_min_unlock | team_mode | teams_editable | team_max_members | has_attempts | duration | end_contest_date    | group_code_enter | title_bar_visible | read_only | full_screen | show_source | validation_min | show_user_infos | contest_phase | url            | uses_api | hints_allowed |
      | 200 | Category | true     | 1234,2345         | 2019-02-06 09:26:40 | true                      | All             | 100              | All       | true           | 10               | true         | 10:20:30 | 2019-03-06 09:26:40 | true             | true              | true      | forceYes    | true        | 100            | true            | Running       | http://someurl | true     | true          |
      | 210 | Chapter  | true     | 1234,2345         | 2019-02-06 09:26:41 | true                      | All             | 100              | All       | true           | 10               | true         | 10:20:31 | 2019-03-06 09:26:41 | true             | true              | true      | forceYes    | true        | 100            | true            | Running       | null           | true     | true          |
      | 220 | Chapter  | true     | 1234,2345         | 2019-02-06 09:26:42 | true                      | All             | 100              | All       | true           | 10               | true         | 10:20:32 | 2019-03-06 09:26:42 | true             | true              | true      | forceYes    | true        | 100            | true            | Running       | null           | true     | true          |
    And the database has the following table 'items_strings':
      | id | item_id | language_id | title       | image_url                  | subtitle     | description   | edu_comment    | version |
      | 53 | 200     | 1           | Category 1  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   | 0       |
      | 54 | 210     | 1           | Chapter A   | http://example.com/my1.jpg | Subtitle 1   | Description 1 | Some comment   | 0       |
      | 55 | 220     | 1           | Chapter B   | http://example.com/my2.jpg | Subtitle 2   | Description 2 | Some comment   | 0       |
      | 63 | 200     | 2           | Catégorie 1 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire | 0       |
      | 64 | 210     | 2           | Chapitre A  | http://example.com/mf1.jpg | Sous-titre 1 | texte 1       | Un commentaire | 0       |
      | 66 | 220     | 2           | Chapitre B  | http://example.com/mf2.jpg | Sous-titre 2 | texte 2       | Un commentaire | 0       |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self | version |
      | 71 | 11                | 11             | 1       | 0       |
      | 72 | 12                | 12             | 1       | 0       |
      | 73 | 13                | 13             | 1       | 0       |
      | 74 | 13                | 11             | 0       | 0       |
      | 75 | 15                | 14             | 0       | 0       |
      | 76 | 13                | 17             | 0       | 0       |
      | 77 | 26                | 22             | 0       | 0       |
    And the database has the following table 'items_items':
      | id | parent_item_id | child_item_id | child_order | category  | partial_access_propagation | version |
      | 54 | 200            | 210           | 2           | Discovery | AsGrayed                   | 0       |
      | 55 | 200            | 220           | 1           | Discovery | AsGrayed                   | 0       |
    And the database has the following table 'users_items':
      | id | user_id | item_id | active_attempt_id | score | submissions_attempts | validated | finished | key_obtained | hints_cached | start_date          | finish_date         | validation_date     | contest_start_date  | state      | answer      | version |
      | 1  | 1       | 200     | 100               | 12341 | 11                   | true      | true     | true         | 11           | 2019-01-30 09:26:41 | 2019-02-01 09:26:41 | 2019-01-31 09:26:41 | 2019-02-01 06:26:41 | Some state | Some answer | 0       |
      | 2  | 1       | 210     | 100               | 12342 | 12                   | true      | true     | true         | 11           | 2019-01-30 09:26:42 | 2019-02-01 09:26:42 | 2019-01-31 09:26:42 | 2019-02-01 06:26:42 | Some state | null        | 0       |
      | 3  | 1       | 220     | 100               | 12344 | 14                   | true      | true     | true         | 11           | 2019-01-30 09:26:44 | 2019-02-01 09:26:44 | 2019-01-31 09:26:44 | 2019-02-01 06:26:44 | Some state | Some answer | 0       |
      | 4  | 2       | 210     | 100               | 12342 | 12                   | true      | true     | true         | 11           | 2019-01-30 09:26:42 | 2019-02-01 09:26:42 | 2019-01-31 09:26:42 | 2019-02-01 06:26:42 | Some state | null        | 0       |
      | 5  | 3       | 200     | 100               | 12341 | 11                   | true      | true     | true         | 11           | 2019-01-30 09:26:41 | 2019-02-01 09:26:41 | 2019-01-31 09:26:41 | 2019-02-01 06:26:41 | Some state | Some answer | 0       |
      | 6  | 3       | 210     | 100               | 12342 | 12                   | true      | true     | true         | 11           | 2019-01-30 09:26:42 | 2019-02-01 09:26:42 | 2019-01-31 09:26:42 | 2019-02-01 06:26:42 | Some state | null        | 0       |
      | 7  | 3       | 220     | 100               | 12344 | 14                   | true      | true     | true         | 11           | 2019-01-30 09:26:44 | 2019-02-01 09:26:44 | 2019-01-31 09:26:44 | 2019-02-01 06:26:44 | Some state | null        | 0       |
      | 8  | 4       | 200     | 100               | 12341 | 11                   | true      | true     | true         | 11           | 2019-01-30 09:26:41 | 2019-02-01 09:26:41 | 2019-01-31 09:26:41 | 2019-02-01 06:26:41 | Some state | Some answer | 0       |
      | 9  | 4       | 210     | 100               | 12342 | 12                   | true      | true     | true         | 11           | 2019-01-30 09:26:42 | 2019-02-01 09:26:42 | 2019-01-31 09:26:42 | 2019-02-01 06:26:42 | Some state | null        | 0       |
      | 10 | 4       | 220     | 100               | 12344 | 14                   | true      | true     | true         | 11           | 2019-01-30 09:26:44 | 2019-02-01 09:26:44 | 2019-01-31 09:26:44 | 2019-02-01 06:26:44 | Some state | null        | 0       |
    And the database has the following table 'groups_items':
      | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | cached_access_solutions_date | creator_user_id | version |
      | 43 | 13       | 200     | 2017-05-29 06:38:38     | 2017-05-29 06:38:38        | 2017-05-29 06:38:38       | 2017-05-29 06:38:38          | 0               | 0       |
      | 44 | 13       | 210     | 2017-05-29 06:38:38     | 2017-05-29 06:38:38        | 2017-05-29 06:38:38       | 2017-05-29 06:38:38          | 0               | 0       |
      | 45 | 13       | 220     | 2017-05-29 06:38:38     | 2017-05-29 06:38:38        | 2017-05-29 06:38:38       | 2017-05-29 06:38:38          | 0               | 0       |
      | 46 | 15       | 210     | 2017-05-29 06:38:38     | 2017-05-29 06:38:38        | 2017-05-29 06:38:38       | 2037-05-29 06:38:38          | 0               | 0       |
      | 47 | 26       | 200     | 2017-05-29 06:38:38     | 2017-05-29 06:38:38        | 2017-05-29 06:38:38       | 2017-05-29 06:38:38          | 0               | 0       |
      | 48 | 26       | 210     | 2037-05-29 06:38:38     | 2037-05-29 06:38:38        | 2017-05-29 06:38:38       | 2017-05-29 06:38:38          | 0               | 0       |
      | 49 | 26       | 220     | 2037-05-29 06:38:38     | 2037-05-29 06:38:38        | 2017-05-29 06:38:38       | 2017-05-29 06:38:38          | 0               | 0       |
    And the database has the following table 'languages':
      | id | code |
      | 2  | fr   |

  Scenario: Full access on all items
    Given I am the user with id "1"
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
      "team_mode": "All",
      "teams_editable": true,
      "team_max_members": 10,
      "has_attempts": true,
      "access_open_date": "2019-02-06T09:26:40Z",
      "duration": "10:20:30",
      "end_contest_date": "2019-03-06T09:26:40Z",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_source": true,
      "validation_min": 100,
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

      "user": {
        "active_attempt_id": "100",
        "score": 12341,
        "submissions_attempts": 11,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "start_date": "2019-01-30T09:26:41Z",
        "validation_date": "2019-01-31T09:26:41Z",
        "finish_date": "2019-02-01T09:26:41Z",
        "contest_start_date": "2019-02-01T06:26:41Z",

        "state": "Some state",
        "answer": "Some answer"
      },

      "children": [
        {
          "id": "220",
          "order": 1,
          "category": "Discovery",
          "partial_access_propagation": "AsGrayed",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "team_mode": "All",
          "teams_editable": true,
          "team_max_members": 10,
          "has_attempts": true,
          "access_open_date": "2019-02-06T09:26:42Z",
          "duration": "10:20:32",
          "end_contest_date": "2019-03-06T09:26:42Z",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "1",
            "title": "Chapter B",
            "image_url": "http://example.com/my2.jpg",
            "subtitle": "Subtitle 2",
            "description": "Description 2"
          },

          "user": {
            "active_attempt_id": "100",
            "score": 12344,
            "submissions_attempts": 14,
            "validated": true,
            "finished": true,
            "key_obtained": true,
            "hints_cached": 11,
            "start_date": "2019-01-30T09:26:44Z",
            "validation_date": "2019-01-31T09:26:44Z",
            "finish_date": "2019-02-01T09:26:44Z",
            "contest_start_date": "2019-02-01T06:26:44Z"
          }
        },
        {
          "id": "210",

          "order": 2,
          "category": "Discovery",
          "partial_access_propagation": "AsGrayed",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "team_mode": "All",
          "teams_editable": true,
          "team_max_members": 10,
          "has_attempts": true,
          "access_open_date": "2019-02-06T09:26:41Z",
          "duration": "10:20:31",
          "end_contest_date": "2019-03-06T09:26:41Z",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "1",
            "title": "Chapter A",
            "image_url": "http://example.com/my1.jpg",
            "subtitle": "Subtitle 1",
            "description": "Description 1"
          },

          "user": {
            "active_attempt_id": "100",
            "score": 12342,
            "submissions_attempts": 12,
            "validated": true,
            "finished": true,
            "key_obtained": true,
            "hints_cached": 11,
            "start_date": "2019-01-30T09:26:42Z",
            "validation_date": "2019-01-31T09:26:42Z",
            "finish_date": "2019-02-01T09:26:42Z",
            "contest_start_date": "2019-02-01T06:26:42Z"
          }
        }
      ]
    }
    """

  Scenario: Chapter as a root node (full access)
    Given I am the user with id "1"
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
      "team_mode": "All",
      "teams_editable": true,
      "team_max_members": 10,
      "has_attempts": true,
      "access_open_date": "2019-02-06T09:26:41Z",
      "duration": "10:20:31",
      "end_contest_date": "2019-03-06T09:26:41Z",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_source": true,
      "validation_min": 100,
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

      "user": {
        "active_attempt_id": "100",
        "score": 12342,
        "submissions_attempts": 12,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "start_date": "2019-01-30T09:26:42Z",
        "validation_date": "2019-01-31T09:26:42Z",
        "finish_date": "2019-02-01T09:26:42Z",
        "contest_start_date": "2019-02-01T06:26:42Z"
      },

      "children": []
    }
    """

  Scenario: Chapter as a root node (without solution access)
    Given I am the user with id "2"
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
      "team_mode": "All",
      "teams_editable": true,
      "team_max_members": 10,
      "has_attempts": true,
      "access_open_date": "2019-02-06T09:26:41Z",
      "duration": "10:20:31",
      "end_contest_date": "2019-03-06T09:26:41Z",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_source": true,
      "validation_min": 100,
      "show_user_infos": true,
      "contest_phase": "Running",

      "string": {
        "language_id": "1",
        "title": "Chapter A",
        "image_url": "http://example.com/my1.jpg",
        "subtitle": "Subtitle 1",
        "description": "Description 1"
      },

      "user": {
        "active_attempt_id": "100",
        "score": 12342,
        "submissions_attempts": 12,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "start_date": "2019-01-30T09:26:42Z",
        "validation_date": "2019-01-31T09:26:42Z",
        "finish_date": "2019-02-01T09:26:42Z",
        "contest_start_date": "2019-02-01T06:26:42Z"
      },

      "children": []
    }
    """

  Scenario: Full access on all items (with user language)
    Given I am the user with id "3"
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
      "team_mode": "All",
      "teams_editable": true,
      "team_max_members": 10,
      "has_attempts": true,
      "access_open_date": "2019-02-06T09:26:40Z",
      "duration": "10:20:30",
      "end_contest_date": "2019-03-06T09:26:40Z",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_source": true,
      "validation_min": 100,
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

      "user": {
        "active_attempt_id": "100",
        "score": 12341,
        "submissions_attempts": 11,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "start_date": "2019-01-30T09:26:41Z",
        "validation_date": "2019-01-31T09:26:41Z",
        "finish_date": "2019-02-01T09:26:41Z",
        "contest_start_date": "2019-02-01T06:26:41Z",

        "state": "Some state",
        "answer": "Some answer"
      },

      "children": [
        {
          "id": "220",
          "order": 1,
          "category": "Discovery",
          "partial_access_propagation": "AsGrayed",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "team_mode": "All",
          "teams_editable": true,
          "team_max_members": 10,
          "has_attempts": true,
          "access_open_date": "2019-02-06T09:26:42Z",
          "duration": "10:20:32",
          "end_contest_date": "2019-03-06T09:26:42Z",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "2",
            "title": "Chapitre B",
            "image_url": "http://example.com/mf2.jpg",
            "subtitle": "Sous-titre 2",
            "description": "texte 2"
          },

          "user": {
            "active_attempt_id": "100",
            "score": 12344,
            "submissions_attempts": 14,
            "validated": true,
            "finished": true,
            "key_obtained": true,
            "hints_cached": 11,
            "start_date": "2019-01-30T09:26:44Z",
            "validation_date": "2019-01-31T09:26:44Z",
            "finish_date": "2019-02-01T09:26:44Z",
            "contest_start_date": "2019-02-01T06:26:44Z"
          }
        },
        {
          "id": "210",

          "order": 2,
          "category": "Discovery",
          "partial_access_propagation": "AsGrayed",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "team_mode": "All",
          "teams_editable": true,
          "team_max_members": 10,
          "has_attempts": true,
          "access_open_date": "2019-02-06T09:26:41Z",
          "duration": "10:20:31",
          "end_contest_date": "2019-03-06T09:26:41Z",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "2",
            "title": "Chapitre A",
            "image_url": "http://example.com/mf1.jpg",
            "subtitle": "Sous-titre 1",
            "description": "texte 1"
          },

          "user": {
            "active_attempt_id": "100",
            "score": 12342,
            "submissions_attempts": 12,
            "validated": true,
            "finished": true,
            "key_obtained": true,
            "hints_cached": 11,
            "start_date": "2019-01-30T09:26:42Z",
            "validation_date": "2019-01-31T09:26:42Z",
            "finish_date": "2019-02-01T09:26:42Z",
            "contest_start_date": "2019-02-01T06:26:42Z"
          }
        }
      ]
    }
    """

  Scenario: Grayed access on children
    Given I am the user with id "4"
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
      "team_mode": "All",
      "teams_editable": true,
      "team_max_members": 10,
      "has_attempts": true,
      "access_open_date": "2019-02-06T09:26:40Z",
      "duration": "10:20:30",
      "end_contest_date": "2019-03-06T09:26:40Z",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_source": true,
      "validation_min": 100,
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

      "user": {
        "active_attempt_id": "100",
        "score": 12341,
        "submissions_attempts": 11,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "start_date": "2019-01-30T09:26:41Z",
        "validation_date": "2019-01-31T09:26:41Z",
        "finish_date": "2019-02-01T09:26:41Z",
        "contest_start_date": "2019-02-01T06:26:41Z",

        "state": "Some state",
        "answer": "Some answer"
      },

      "children": [
        {
          "id": "220",
          "order": 1,
          "category": "Discovery",
          "partial_access_propagation": "AsGrayed",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "team_mode": "All",
          "teams_editable": true,
          "team_max_members": 10,
          "has_attempts": true,
          "access_open_date": "2019-02-06T09:26:42Z",
          "duration": "10:20:32",
          "end_contest_date": "2019-03-06T09:26:42Z",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "1",
            "title": "Chapter B",
            "image_url": "http://example.com/my2.jpg"
          },

          "user": {
          }
        },
        {
          "id": "210",

          "order": 2,
          "category": "Discovery",
          "partial_access_propagation": "AsGrayed",

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "team_mode": "All",
          "teams_editable": true,
          "team_max_members": 10,
          "has_attempts": true,
          "access_open_date": "2019-02-06T09:26:41Z",
          "duration": "10:20:31",
          "end_contest_date": "2019-03-06T09:26:41Z",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": "1",
            "title": "Chapter A",
            "image_url": "http://example.com/my1.jpg"
          },

          "user": {
          }
        }
      ]
    }
    """
