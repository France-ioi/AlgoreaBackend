Feature: Get item children
  Background:
    Given the database has the following table 'groups':
      | id | name       | grade | type  |
      | 11 | jdoe       | -2    | User  |
      | 13 | Group B    | -2    | Team  |
      | 14 | nosolution | -2    | User  |
      | 15 | Group C    | -2    | Class |
      | 17 | fr         | -2    | User  |
      | 22 | info       | -2    | User  |
      | 26 | team       | -2    | Team  |
    And the database has the following table 'users':
      | login      | temp_user | group_id | default_language |
      | jdoe       | 0         | 11       |                  |
      | nosolution | 0         | 14       |                  |
      | fr         | 0         | 17       | fr               |
      | info       | 0         | 22       |                  |
    And the database has the following table 'items':
      | id  | type    | default_language_tag | no_score | display_details_in_parent | validation_type | requires_explicit_entry | allows_multiple_attempts | entry_participant_type | duration | title_bar_visible | read_only | full_screen | show_user_infos | url            | uses_api | hints_allowed |
      | 200 | Course  | en                   | true     | true                      | All             | true                    | true                     | Team                   | 10:20:30 | true              | true      | forceYes    | true            | http://someurl | true     | true          |
      | 210 | Chapter | en                   | true     | true                      | All             | false                   | true                     | User                   | 10:20:31 | true              | true      | forceYes    | true            | null           | true     | true          |
      | 220 | Chapter | en                   | true     | true                      | All             | false                   | true                     | Team                   | 10:20:32 | true              | true      | forceYes    | true            | null           | true     | true          |
      | 230 | Chapter | en                   | true     | true                      | All             | false                   | true                     | Team                   | 10:20:32 | true              | true      | forceYes    | true            | null           | true     | true          |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title       | image_url                  | subtitle     | description   | edu_comment    |
      | 200     | en           | Category 1  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment   |
      | 210     | en           | Chapter A   | http://example.com/my1.jpg | Subtitle 1   | Description 1 | Some comment   |
      | 220     | en           | Chapter B   | http://example.com/my2.jpg | Subtitle 2   | Description 2 | Some comment   |
      | 230     | en           | Chapter C   | http://example.com/my2.jpg | Subtitle 2   | Description 2 | Some comment   |
      | 200     | fr           | Catégorie 1 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Un commentaire |
      | 210     | fr           | Chapitre A  | http://example.com/mf1.jpg | Sous-titre 1 | texte 1       | Un commentaire |
      | 220     | fr           | Chapitre B  | http://example.com/mf2.jpg | Sous-titre 2 | texte 2       | Un commentaire |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 11             |
      | 13              | 17             |
      | 15              | 14             |
      | 15              | 17             |
      | 26              | 11             |
      | 26              | 22             |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | manager_id | group_id | can_watch_members |
      | 22         | 15       | true              |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order | category  | score_weight | content_view_propagation | upper_view_levels_propagation | grant_view_propagation | watch_propagation | edit_propagation |
      | 200            | 210           | 2           | Discovery | 1            | none                     | use_content_view_propagation  | true                   | false             | true             |
      | 200            | 220           | 1           | Discovery | 2            | as_info                  | as_content_with_descendants   | false                  | true              | false            |
      | 200            | 230           | 3           | Discovery | 2            | as_info                  | as_content_with_descendants   | false                  | true              | false            |
    And the database has the following table 'item_dependencies':
      | item_id | dependent_item_id | grant_content_view |
      | 210     | 200               | false              |
      | 210     | 210               | true               |
      | 200     | 220               | false              |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_edit_generated | can_watch_generated | is_owner_generated |
      | 11       | 200     | solution                 | enter                    | children           | result              | true               |
      | 11       | 210     | solution                 | none                     | none               | none                | true               |
      | 11       | 220     | solution                 | none                     | none               | none                | false              |
      | 13       | 200     | solution                 | none                     | none               | none                | false              |
      | 13       | 210     | solution                 | none                     | none               | none                | false              |
      | 13       | 220     | solution                 | none                     | none               | none                | false              |
      | 15       | 210     | content_with_descendants | none                     | none               | none                | false              |
      | 17       | 200     | solution                 | none                     | none               | none                | false              |
      | 17       | 210     | solution                 | none                     | none               | none                | false              |
      | 17       | 220     | solution                 | none                     | none               | none                | false              |
      | 22       | 200     | solution                 | none                     | none               | none                | false              |
      | 22       | 210     | info                     | none                     | none               | result              | false              |
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
      | 0  | 17             | 2019-05-30 10:00:00 |
      | 0  | 22             | 2019-05-30 10:00:00 |
      | 1  | 11             | 2019-05-30 11:00:00 |
      | 1  | 13             | 2019-05-30 11:00:00 |
      | 1  | 17             | 2019-05-30 10:00:00 |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          | latest_activity_at  | score_computed | validated_at        |
      | 0          | 11             | 200     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 11.1           | null                |
      | 0          | 11             | 210     | null                | 2018-05-30 11:00:01 | 12.2           | null                |
      | 0          | 11             | 220     | 2019-05-30 11:00:00 | 2019-05-30 11:00:02 | 13.3           | null                |
      | 0          | 13             | 200     | 2019-05-30 11:00:00 | 2019-05-30 11:00:03 | 0.0            | null                |
      | 0          | 13             | 210     | 2019-05-30 11:00:00 | 2019-05-30 11:00:03 | 14.4           | null                |
      | 0          | 13             | 220     | null                | 2018-05-30 11:00:02 | 15.5           | null                |
      | 0          | 17             | 200     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 0.0            | null                |
      | 0          | 17             | 210     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 10.0           | 2019-05-30 11:00:01 |
      | 0          | 22             | 200     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 0.0            | null                |
      | 0          | 26             | 200     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 0.0            | null                |
      | 1          | 11             | 200     | 2019-05-30 12:00:00 | 2019-05-30 12:00:01 | 21.1           | null                |
      | 1          | 11             | 210     | null                | 2018-05-30 12:00:01 | 22.2           | null                |
      | 1          | 11             | 220     | 2019-05-30 12:00:00 | 2019-05-30 12:00:02 | 3.3            | null                |
      | 1          | 13             | 210     | 2019-05-30 12:00:00 | 2019-05-30 12:00:03 | 24.4           | null                |
      | 1          | 13             | 220     | null                | 2018-05-30 12:00:02 | 5.5            | null                |
      | 1          | 17             | 210     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 20.0           | 2019-05-30 11:00:01 |
      | 0          | 22             | 230     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 20.0           | 2019-05-30 11:00:01 |

  Scenario: Full access on all items (as user)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/children?attempt_id=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "220",
        "order": 1,
        "category": "Discovery",
        "score_weight": 2,
        "content_view_propagation": "as_info",
        "upper_view_levels_propagation": "as_content_with_descendants",
        "grant_view_propagation": false,
        "watch_propagation": true,
        "edit_propagation": false,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "Team",
        "duration": "10:20:32",
        "no_score": true,
        "default_language_tag": "en",
        "requires_explicit_entry": false,
        "best_score": 13.3,
        "grants_access_to_items": false,
        "string": {
          "language_tag": "en",
          "title": "Chapter B",
          "image_url": "http://example.com/my2.jpg",
          "subtitle": "Subtitle 2"
        },
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "solution",
          "can_watch": "none",
          "is_owner": false
        },
        "results": [
          {
            "attempt_allows_submissions_until": "9999-12-31T23:59:59Z",
            "attempt_id": "1",
            "ended_at": null,
            "latest_activity_at": "2019-05-30T12:00:02Z",
            "score_computed": 3.3,
            "started_at": "2019-05-30T12:00:00Z",
            "validated": false
          }
        ]
      },
      {
        "id": "210",
        "order": 2,
        "category": "Discovery",
        "score_weight": 1,
        "content_view_propagation": "none",
        "upper_view_levels_propagation": "use_content_view_propagation",
        "grant_view_propagation": true,
        "watch_propagation": false,
        "edit_propagation": true,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "User",
        "duration": "10:20:31",
        "no_score": true,
        "default_language_tag": "en",
        "requires_explicit_entry": false,
        "best_score": 22.2,
        "grants_access_to_items": true,
        "string": {
          "language_tag": "en",
          "title": "Chapter A",
          "image_url": "http://example.com/my1.jpg",
          "subtitle": "Subtitle 1"
        },
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "solution",
          "can_watch": "none",
          "is_owner": true
        },
        "results": [
          {
            "attempt_allows_submissions_until": "9999-12-31T23:59:59Z",
            "attempt_id": "1",
            "ended_at": null,
            "latest_activity_at": "2018-05-30T12:00:01Z",
            "score_computed": 22.2,
            "started_at": null,
            "validated": false
          }
        ]
      }
    ]
    """

  Scenario: Full access on all items (with user language, as user)
    Given I am the user with id "17"
    When I send a GET request to "/items/200/children?attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "220",
        "order": 1,
        "category": "Discovery",
        "score_weight": 2,
        "content_view_propagation": "as_info",
        "upper_view_levels_propagation": "as_content_with_descendants",
        "grant_view_propagation": false,
        "watch_propagation": true,
        "edit_propagation": false,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "Team",
        "duration": "10:20:32",
        "no_score": true,
        "default_language_tag": "en",
        "requires_explicit_entry": false,
        "best_score": 0,
        "grants_access_to_items": false,
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "solution",
          "can_watch": "none",
          "is_owner": false
        },
        "results": [],
        "string": {
          "language_tag": "fr",
          "title": "Chapitre B",
          "image_url": "http://example.com/mf2.jpg",
          "subtitle": "Sous-titre 2"
        }
      },
      {
        "id": "210",
        "order": 2,
        "category": "Discovery",
        "score_weight": 1,
        "content_view_propagation": "none",
        "upper_view_levels_propagation": "use_content_view_propagation",
        "grant_view_propagation": true,
        "watch_propagation": false,
        "edit_propagation": true,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "User",
        "duration": "10:20:31",
        "no_score": true,
        "default_language_tag": "en",
        "requires_explicit_entry": false,
        "best_score": 20,
        "grants_access_to_items": true,
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "solution",
          "can_watch": "none",
          "is_owner": false
        },
        "results": [
          {
            "attempt_allows_submissions_until": "9999-12-31T23:59:59Z",
            "attempt_id": "0",
            "ended_at": null,
            "latest_activity_at": "2019-05-30T11:00:01Z",
            "score_computed": 10,
            "started_at": "2019-05-30T11:00:00Z",
            "validated": true
          }
        ],
        "string": {
          "language_tag": "fr",
          "title": "Chapitre A",
          "image_url": "http://example.com/mf1.jpg",
          "subtitle": "Sous-titre 1"
        }
      }
    ]
    """

  Scenario: Info access on children (as user)
    Given I am the user with id "22"
    When I send a GET request to "/items/200/children?attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "220",
        "order": 1,
        "category": "Discovery",
        "score_weight": 2,
        "content_view_propagation": "as_info",
        "upper_view_levels_propagation": "as_content_with_descendants",
        "grant_view_propagation": false,
        "watch_propagation": true,
        "edit_propagation": false,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "Team",
        "duration": "10:20:32",
        "no_score": true,
        "default_language_tag": "en",
        "best_score": 0,
        "grants_access_to_items": false,
        "requires_explicit_entry": false,
        "results": [],
        "string": {
          "language_tag": "en",
          "title": "Chapter B",
          "image_url": "http://example.com/my2.jpg"
        },
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "info",
          "can_watch": "none",
          "is_owner": false
        }
      },
      {
        "id": "210",
        "order": 2,
        "category": "Discovery",
        "score_weight": 1,
        "content_view_propagation": "none",
        "upper_view_levels_propagation": "use_content_view_propagation",
        "grant_view_propagation": true,
        "watch_propagation": false,
        "edit_propagation": true,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "User",
        "duration": "10:20:31",
        "no_score": true,
        "default_language_tag": "en",
        "best_score": 0,
        "grants_access_to_items": true,
        "requires_explicit_entry": false,
        "results": [],
        "string": {
          "language_tag": "en",
          "title": "Chapter A",
          "image_url": "http://example.com/my1.jpg"
        },
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "info",
          "can_watch": "result",
          "is_owner": false
        }
      }
    ]
    """

  Scenario: Full access on all items (as team)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/children?as_team_id=13&attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "220",
        "order": 1,
        "category": "Discovery",
        "score_weight": 2,
        "content_view_propagation": "as_info",
        "upper_view_levels_propagation": "as_content_with_descendants",
        "grant_view_propagation": false,
        "watch_propagation": true,
        "edit_propagation": false,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "Team",
        "duration": "10:20:32",
        "no_score": true,
        "default_language_tag": "en",
        "requires_explicit_entry": false,
        "best_score": 15.5,
        "grants_access_to_items": false,
        "string": {
          "language_tag": "en",
          "title": "Chapter B",
          "image_url": "http://example.com/my2.jpg",
          "subtitle": "Subtitle 2"
        },
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "solution",
          "can_watch": "none",
          "is_owner": false
        },
        "results": [
          {
            "attempt_allows_submissions_until": "9999-12-31T23:59:59Z",
            "attempt_id": "0",
            "ended_at": null,
            "latest_activity_at": "2018-05-30T11:00:02Z",
            "score_computed": 15.5,
            "started_at": null,
            "validated": false
          }
        ]
      },
      {
        "id": "210",
        "order": 2,
        "category": "Discovery",
        "score_weight": 1,
        "content_view_propagation": "none",
        "upper_view_levels_propagation": "use_content_view_propagation",
        "grant_view_propagation": true,
        "watch_propagation": false,
        "edit_propagation": true,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "User",
        "duration": "10:20:31",
        "no_score": true,
        "default_language_tag": "en",
        "requires_explicit_entry": false,
        "best_score": 24.4,
        "grants_access_to_items": true,
        "string": {
          "language_tag": "en",
          "title": "Chapter A",
          "image_url": "http://example.com/my1.jpg",
          "subtitle": "Subtitle 1"
        },
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "solution",
          "can_watch": "none",
          "is_owner": false
        },
        "results": [
          {
            "attempt_allows_submissions_until": "9999-12-31T23:59:59Z",
            "attempt_id": "0",
            "ended_at": null,
            "latest_activity_at": "2019-05-30T11:00:03Z",
            "score_computed": 14.4,
            "started_at": "2019-05-30T11:00:00Z",
            "validated": false
          }
        ]
      }
    ]
    """

  Scenario: Info access on children (as team)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/children?as_team_id=26&attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "220",
        "order": 1,
        "category": "Discovery",
        "score_weight": 2,
        "content_view_propagation": "as_info",
        "upper_view_levels_propagation": "as_content_with_descendants",
        "grant_view_propagation": false,
        "watch_propagation": true,
        "edit_propagation": false,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "Team",
        "duration": "10:20:32",
        "no_score": true,
        "default_language_tag": "en",
        "requires_explicit_entry": false,
        "best_score": 0,
        "grants_access_to_items": false,
        "string": {
          "language_tag": "en",
          "title": "Chapter B",
          "image_url": "http://example.com/my2.jpg"
        },
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "info",
          "can_watch": "none",
          "is_owner": false
        },
        "results": []
      },
      {
        "id": "210",
        "order": 2,
        "category": "Discovery",
        "score_weight": 1,
        "content_view_propagation": "none",
        "upper_view_levels_propagation": "use_content_view_propagation",
        "grant_view_propagation": true,
        "watch_propagation": false,
        "edit_propagation": true,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "User",
        "duration": "10:20:31",
        "no_score": true,
        "default_language_tag": "en",
        "best_score": 0,
        "grants_access_to_items": true,
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
        },
        "requires_explicit_entry": false,
        "results": []
      }
    ]
    """

  Scenario: Info access on children (as user) with watched_group_id
    Given I am the user with id "22"
    When I send a GET request to "/items/200/children?attempt_id=0&watched_group_id=15"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "220",
        "order": 1,
        "category": "Discovery",
        "score_weight": 2,
        "content_view_propagation": "as_info",
        "upper_view_levels_propagation": "as_content_with_descendants",
        "grant_view_propagation": false,
        "watch_propagation": true,
        "edit_propagation": false,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "Team",
        "duration": "10:20:32",
        "no_score": true,
        "default_language_tag": "en",
        "best_score": 0,
        "grants_access_to_items": false,
        "requires_explicit_entry": false,
        "results": [],
        "string": {
          "language_tag": "en",
          "title": "Chapter B",
          "image_url": "http://example.com/my2.jpg"
        },
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "info",
          "can_watch": "none",
          "is_owner": false
        },
        "watched_group": {
          "can_view": "none"
        }
      },
      {
        "id": "210",
        "order": 2,
        "category": "Discovery",
        "score_weight": 1,
        "content_view_propagation": "none",
        "upper_view_levels_propagation": "use_content_view_propagation",
        "grant_view_propagation": true,
        "watch_propagation": false,
        "edit_propagation": true,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "User",
        "duration": "10:20:31",
        "no_score": true,
        "default_language_tag": "en",
        "best_score": 0,
        "grants_access_to_items": true,
        "requires_explicit_entry": false,
        "results": [],
        "string": {
          "language_tag": "en",
          "title": "Chapter A",
          "image_url": "http://example.com/my1.jpg"
        },
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "info",
          "can_watch": "result",
          "is_owner": false
        },
        "watched_group": {
          "can_view": "content_with_descendants",
          "all_validated": false,
          "avg_score": 10
        }
      }
    ]
    """

  Scenario: Info access on children (as user) with watched_group_id, show invisible items
    Given I am the user with id "22"
    When I send a GET request to "/items/200/children?attempt_id=0&watched_group_id=15&show_invisible_items=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "220",
        "order": 1,
        "category": "Discovery",
        "score_weight": 2,
        "content_view_propagation": "as_info",
        "upper_view_levels_propagation": "as_content_with_descendants",
        "grant_view_propagation": false,
        "watch_propagation": true,
        "edit_propagation": false,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "Team",
        "duration": "10:20:32",
        "no_score": true,
        "default_language_tag": "en",
        "best_score": 0,
        "grants_access_to_items": false,
        "requires_explicit_entry": false,
        "results": [],
        "string": {
          "language_tag": "en",
          "title": "Chapter B",
          "image_url": "http://example.com/my2.jpg"
        },
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "info",
          "can_watch": "none",
          "is_owner": false
        },
        "watched_group": {
          "can_view": "none"
        }
      },
      {
        "id": "210",
        "order": 2,
        "category": "Discovery",
        "score_weight": 1,
        "content_view_propagation": "none",
        "upper_view_levels_propagation": "use_content_view_propagation",
        "grant_view_propagation": true,
        "watch_propagation": false,
        "edit_propagation": true,
        "type": "Chapter",
        "display_details_in_parent": true,
        "validation_type": "All",
        "allows_multiple_attempts": true,
        "entry_participant_type": "User",
        "duration": "10:20:31",
        "no_score": true,
        "default_language_tag": "en",
        "best_score": 0,
        "grants_access_to_items": true,
        "requires_explicit_entry": false,
        "results": [],
        "string": {
          "language_tag": "en",
          "title": "Chapter A",
          "image_url": "http://example.com/my1.jpg"
        },
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "info",
          "can_watch": "result",
          "is_owner": false
        },
        "watched_group": {
          "can_view": "content_with_descendants",
          "all_validated": false,
          "avg_score": 10
        }
      },
      {
        "category": "Discovery",
        "content_view_propagation": "as_info",
        "edit_propagation": false,
        "grant_view_propagation": false,
        "id": "230",
        "order": 3,
        "permissions": {
          "can_edit": "none",
          "can_grant_view": "none",
          "can_view": "none",
          "can_watch": "none",
          "is_owner": false
        },
        "score_weight": 2,
        "upper_view_levels_propagation": "as_content_with_descendants",
        "watch_propagation": true,
        "watched_group": {
          "can_view": "none"
        }
      }
    ]
    """
