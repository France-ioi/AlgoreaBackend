Feature: Get item view information

  Background:
    Given the database has the following table 'groups':
      | id | name       | text_id | grade | type  |
      | 11 | jdoe       |         | -2    | User  |
      | 13 | Group B    |         | -2    | Team  |
      | 14 | nosolution |         | -2    | User  |
      | 15 | Group C    |         | -2    | Class |
      | 17 | fr         |         | -2    | User  |
      | 22 | info       |         | -2    | User  |
      | 26 | team       |         | -2    | Team  |
    And the database has the following table 'users':
      | login      | temp_user | group_id | default_language |
      | jdoe       | 0         | 11       |                  |
      | nosolution | 0         | 14       |                  |
      | fr         | 0         | 17       | fr               |
      | info       | 0         | 22       |                  |
    And the database has the following table 'items':
      | id  | type    | default_language_tag | no_score | text_id  | display_details_in_parent | validation_type | requires_explicit_entry | entry_min_admitted_members_ratio | entry_frozen_teams | entry_max_team_size | allows_multiple_attempts | entry_participant_type | duration | prompt_to_join_group_by_code | title_bar_visible | read_only | full_screen | show_user_infos | url            | uses_api | hints_allowed |
      | 200 | Course  | en                   | true     | Task_30c | true                      | All             | true                    | All                              | false              | 10                  | true                     | Team                   | 10:20:30 | true                         | true              | true      | forceYes    | true            | http://someurl | true     | true          |
      | 210 | Chapter | en                   | true     | null     | true                      | All             | false                   | All                              | false              | 10                  | true                     | User                   | 10:20:31 | true                         | true              | true      | forceYes    | true            | null           | true     | true          |
      | 220 | Chapter | en                   | true     | Task_30e | true                      | All             | false                   | All                              | false              | 10                  | true                     | Team                   | 10:20:32 | true                         | true              | true      | forceYes    | true            | null           | true     | true          |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title       | image_url                  | subtitle     | description   | edu_comment    |
      | 200     | en           | Category 1  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 210     | en           | Chapter A   | http://example.com/my1.jpg | Subtitle 1   | Description 1 | Some comment   |
      | 220     | en           | Chapter B   | http://example.com/my2.jpg | Subtitle 2   | Description 2 | Some comment   |
      | 200     | fr           | Catégorie 1 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
      | 210     | fr           | Chapitre A  | http://example.com/mf1.jpg | Sous-titre 1 | texte 1       | Un commentaire |
      | 220     | fr           | Chapitre B  | http://example.com/mf2.jpg | Sous-titre 2 | texte 2       | Un commentaire |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 11             |
      | 13              | 17             |
      | 15              | 14             |
      | 26              | 11             |
      | 26              | 22             |
    And the groups ancestors are computed
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order | category  | content_view_propagation |
      | 200            | 210           | 2           | Discovery | as_info                  |
      | 200            | 220           | 1           | Discovery | as_info                  |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_edit_generated | can_watch_generated | is_owner_generated |
      | 11       | 200     | solution                 | enter                    | children           | result              | true               |
      | 11       | 210     | solution                 | none                     | none               | none                | false              |
      | 11       | 220     | solution                 | none                     | none               | none                | false              |
      | 13       | 200     | solution                 | none                     | none               | none                | false              |
      | 13       | 210     | solution                 | none                     | none               | none                | false              |
      | 13       | 220     | solution                 | none                     | none               | none                | false              |
      | 15       | 210     | content_with_descendants | none                     | none               | none                | false              |
      | 17       | 200     | solution                 | none                     | none               | none                | false              |
      | 17       | 210     | solution                 | none                     | none               | none                | false              |
      | 17       | 220     | solution                 | none                     | none               | none                | false              |
      | 22       | 200     | solution                 | none                     | none               | none                | false              |
      | 22       | 210     | info                     | none                     | none               | none                | false              |
      | 22       | 220     | info                     | none                     | none               | none                | false              |
      | 26       | 200     | solution                 | none                     | none               | none                | false              |
      | 26       | 210     | info                     | none                     | none               | none                | false              |
      | 26       | 220     | info                     | none                     | none               | none                | false              |
    And the database has the following table 'languages':
      | tag |
      | fr  |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          |
      | 0  | 11             | 2019-05-30 10:00:00 |
      | 0  | 13             | 2019-05-30 10:00:00 |
      | 1  | 13             | 2019-05-30 10:00:00 |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          | score_computed |
      | 0          | 11             | 200     | 2019-05-30 11:00:00 | 0              |
      | 0          | 11             | 210     | null                | 10             |
      | 0          | 11             | 220     | 2019-05-30 11:00:00 | 0              |
      | 0          | 13             | 200     | 2019-05-30 11:00:00 | 1              |
      | 0          | 13             | 210     | 2019-05-30 11:00:00 | 0              |
      | 0          | 13             | 220     | null                | 0              |
      | 1          | 13             | 200     | 2019-05-30 11:00:00 | 2              |

  Scenario: Full access on the item (as user)
    Given I am the user with id "11"
    When I send a GET request to "/items/200"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "200",
      "type": "Course",
      "display_details_in_parent": true,
      "validation_type": "All",
      "requires_explicit_entry": true,
      "entry_min_admitted_members_ratio": "All",
      "entry_frozen_teams": false,
      "entry_max_team_size": 10,
      "entering_time_max": "9999-12-31T23:59:59Z",
      "entering_time_min": "1000-01-01T00:00:00Z",
      "allows_multiple_attempts": true,
      "entry_participant_type": "Team",
      "duration": "10:20:30",
      "no_score": true,
      "text_id": "Task_30c",
      "default_language_tag": "en",
      "supported_language_tags": ["en", "fr"],
      "prompt_to_join_group_by_code": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,
      "url": "http://someurl",
      "uses_api": true,
      "hints_allowed": true,

      "best_score": 0,

      "string": {
        "language_tag": "en",
        "title": "Category 1",
        "image_url": "http://example.com/my0.jpg",
        "subtitle": "Subtitle 0",
        "description": "Description 0",
        "edu_comment": "Some comment"
      },

      "permissions": {
        "can_edit": "children",
        "can_grant_view": "enter",
        "can_view": "solution",
        "can_watch": "result",
        "is_owner": true
      }
    }
    """

  Scenario: Chapter (full access, as user)
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
      "requires_explicit_entry": false,
      "entry_min_admitted_members_ratio": "All",
      "entry_frozen_teams": false,
      "entry_max_team_size": 10,
      "entering_time_max": "9999-12-31T23:59:59Z",
      "entering_time_min": "1000-01-01T00:00:00Z",
      "allows_multiple_attempts": true,
      "entry_participant_type": "User",
      "duration": "10:20:31",
      "no_score": true,
      "text_id": null,
      "default_language_tag": "en",
      "supported_language_tags": ["en", "fr"],
      "prompt_to_join_group_by_code": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,

      "best_score": 10,

      "string": {
        "language_tag": "en",
        "title": "Chapter A",
        "image_url": "http://example.com/my1.jpg",
        "subtitle": "Subtitle 1",
        "description": "Description 1",
        "edu_comment": "Some comment"
      },

      "permissions": {
        "can_edit": "none",
        "can_grant_view": "none",
        "can_view": "solution",
        "can_watch": "none",
        "is_owner": false
      }
    }
    """

  Scenario: Chapter (without solution access, as user)
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
      "entry_min_admitted_members_ratio": "All",
      "requires_explicit_entry": false,
      "entry_frozen_teams": false,
      "entry_max_team_size": 10,
      "entering_time_max": "9999-12-31T23:59:59Z",
      "entering_time_min": "1000-01-01T00:00:00Z",
      "allows_multiple_attempts": true,
      "entry_participant_type": "User",
      "duration": "10:20:31",
      "no_score": true,
      "text_id": null,
      "default_language_tag": "en",
      "supported_language_tags": ["en", "fr"],
      "prompt_to_join_group_by_code": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,

      "best_score": 0,

      "string": {
        "language_tag": "en",
        "title": "Chapter A",
        "image_url": "http://example.com/my1.jpg",
        "subtitle": "Subtitle 1",
        "description": "Description 1"
      },

      "permissions": {
        "can_edit": "none",
        "can_grant_view": "none",
        "can_view": "content_with_descendants",
        "can_watch": "none",
        "is_owner": false
      }
    }
    """

  Scenario: Full access on the item (with user language, as user)
    Given I am the user with id "17"
    When I send a GET request to "/items/200"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "200",
      "type": "Course",
      "display_details_in_parent": true,
      "validation_type": "All",
      "requires_explicit_entry": true,
      "entry_min_admitted_members_ratio": "All",
      "entry_frozen_teams": false,
      "entry_max_team_size": 10,
      "entering_time_max": "9999-12-31T23:59:59Z",
      "entering_time_min": "1000-01-01T00:00:00Z",
      "allows_multiple_attempts": true,
      "entry_participant_type": "Team",
      "duration": "10:20:30",
      "no_score": true,
      "text_id": "Task_30c",
      "default_language_tag": "en",
      "supported_language_tags": ["en", "fr"],
      "prompt_to_join_group_by_code": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,
      "url": "http://someurl",
      "uses_api": true,
      "hints_allowed": true,

      "best_score": 0,

      "string": {
        "language_tag": "fr",
        "title": "Catégorie 1",
        "image_url": "http://example.com/mf0.jpg",
        "subtitle": "Sous-titre 0",
        "description": "texte 0",
        "edu_comment": "Un commentaire"
      },

      "permissions": {
        "can_edit": "none",
        "can_grant_view": "none",
        "can_view": "solution",
        "can_watch": "none",
        "is_owner": false
      }
    }
    """

  Scenario: Full access on the item (as team)
    Given I am the user with id "11"
    When I send a GET request to "/items/200?as_team_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "200",
      "type": "Course",
      "display_details_in_parent": true,
      "validation_type": "All",
      "requires_explicit_entry": true,
      "entry_min_admitted_members_ratio": "All",
      "entry_frozen_teams": false,
      "entry_max_team_size": 10,
      "entering_time_max": "9999-12-31T23:59:59Z",
      "entering_time_min": "1000-01-01T00:00:00Z",
      "allows_multiple_attempts": true,
      "entry_participant_type": "Team",
      "duration": "10:20:30",
      "no_score": true,
      "text_id": "Task_30c",
      "default_language_tag": "en",
      "supported_language_tags": ["en", "fr"],
      "prompt_to_join_group_by_code": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,
      "url": "http://someurl",
      "uses_api": true,
      "hints_allowed": true,

      "best_score": 2,

      "string": {
        "language_tag": "en",
        "title": "Category 1",
        "image_url": "http://example.com/my0.jpg",
        "subtitle": "Subtitle 0",
        "description": "Description 0",
        "edu_comment": "Some comment"
      },

      "permissions": {
        "can_edit": "none",
        "can_grant_view": "none",
        "can_view": "solution",
        "can_watch": "none",
        "is_owner": false
      }
    }
    """

  Scenario: Chapter (info access, as user)
    Given I am the user with id "22"
    When I send a GET request to "/items/210"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "210",
      "type": "Chapter",
      "display_details_in_parent": true,
      "validation_type": "All",
      "requires_explicit_entry": false,
      "entry_min_admitted_members_ratio": "All",
      "entry_frozen_teams": false,
      "entry_max_team_size": 10,
      "entering_time_max": "9999-12-31T23:59:59Z",
      "entering_time_min": "1000-01-01T00:00:00Z",
      "allows_multiple_attempts": true,
      "entry_participant_type": "User",
      "duration": "10:20:31",
      "no_score": true,
      "text_id": null,
      "default_language_tag": "en",
      "supported_language_tags": ["en", "fr"],
      "prompt_to_join_group_by_code": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,

      "best_score": 0,

      "string": {
        "language_tag": "en",
        "title": "Chapter A",
        "image_url": "http://example.com/my1.jpg"
      },

      "permissions": {
        "can_edit": "none",
        "can_grant_view": "none",
        "can_view": "info",
        "can_watch": "none",
        "is_owner": false
      }
    }
    """

  Scenario: Full access on the item (as user), language_tag is given
    Given I am the user with id "11"
    When I send a GET request to "/items/200?language_tag=fr"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "200",
      "type": "Course",
      "display_details_in_parent": true,
      "validation_type": "All",
      "requires_explicit_entry": true,
      "entry_min_admitted_members_ratio": "All",
      "entry_frozen_teams": false,
      "entry_max_team_size": 10,
      "entering_time_max": "9999-12-31T23:59:59Z",
      "entering_time_min": "1000-01-01T00:00:00Z",
      "allows_multiple_attempts": true,
      "entry_participant_type": "Team",
      "duration": "10:20:30",
      "no_score": true,
      "text_id": "Task_30c",
      "default_language_tag": "en",
      "supported_language_tags": ["en", "fr"],
      "prompt_to_join_group_by_code": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_user_infos": true,
      "url": "http://someurl",
      "uses_api": true,
      "hints_allowed": true,

      "best_score": 0,

      "string": {
        "language_tag": "fr",
        "title": "Catégorie 1",
        "image_url": "http://example.com/mf0.jpg",
        "subtitle": "Sous-titre 0",
        "description": "texte 0",
        "edu_comment": "Un commentaire"
      },

      "permissions": {
        "can_edit": "children",
        "can_grant_view": "enter",
        "can_view": "solution",
        "can_watch": "result",
        "is_owner": true
      }
    }
    """
