Feature: Get root activities for a participant group
  Background:
    Given the database has the following table 'groups':
      | id | name      | text_id | grade | type  | root_activity_id | created_at          |
      | 1  | all       |         | -2    | Base  | 220              | 2019-01-30 08:26:49 |
      | 11 | jdoe      |         | -2    | User  | null             | 2019-01-30 08:26:48 |
      | 12 | Group A   |         | -2    | Class | 200              | 2019-01-30 08:26:49 |
      | 13 | Group B   |         | -2    | Team  | 220              | 2019-01-30 08:26:46 |
      | 14 | info_root |         | -2    | User  | null             | 2019-01-30 08:26:45 |
      | 16 | info_mid  |         | -2    | User  | null             | 2019-01-30 08:26:44 |
      | 18 | french    |         | -2    | User  | null             | 2019-01-30 08:26:43 |
      | 19 | Team      |         | -2    | Team  | 210              | 2019-01-30 08:26:42 |
      | 20 | Group C   |         | -2    | Club  | 230              | 2019-01-30 08:26:42 |
      | 21 | Group D   |         | -2    | Club  | 240              | 2019-01-30 08:26:42 |
      | 22 | Group E   |         | -2    | Club  | 250              | 2019-01-30 08:26:42 |
      | 23 | Group F   |         | -2    | Club  | 260              | 2019-01-30 08:26:42 |
      | 24 | Group G   |         | -2    | Club  | 270              | 2019-01-30 08:26:42 |
      | 25 | Group H   |         | -2    | Club  | 280              | 2019-01-30 08:26:42 |
      | 26 | Group K   |         | -2    | Club  | 290              | 2019-01-30 08:26:42 |
      | 27 | Group Z   |         | -2    | Club  | 300              | 2019-01-30 08:26:42 |
      | 29 | Class     |         | -2    | Class | 280              | 2019-01-30 08:26:42 |
      | 30 | manager   |         | -2    | User  | 280              | 2019-01-30 08:26:42 |
    And the database has the following table 'languages':
      | tag |
      | fr  |
    And the database has the following table 'users':
      | login     | temp_user | group_id | default_language |
      | jdoe      | 0         | 11       |                  |
      | info_root | 0         | 14       |                  |
      | info_mid  | 0         | 16       |                  |
      | fr_user   | 0         | 18       | fr               |
      | manager   | 0         | 30       | fr               |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 1               | 12             |
      | 1               | 13             |
      | 1               | 14             |
      | 1               | 16             |
      | 1               | 18             |
      | 1               | 19             |
      | 12              | 11             |
      | 12              | 18             |
      | 12              | 22             |
      | 12              | 23             |
      | 13              | 11             |
      | 19              | 11             |
      | 22              | 26             |
      | 24              | 26             |
      | 25              | 26             |
      | 26              | 27             |
      | 29              | 30             |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | manager_id | group_id | can_watch_members |
      | 12         | 1        | true              |
      | 12         | 19       | true              |
      | 29         | 22       | true              |
      | 30         | 24       | true              |
    And the database has the following table 'items':
      | id  | type    | default_language_tag | no_score | requires_explicit_entry | entry_participant_type |
      | 200 | Course  | en                   | false    | false                   | User                   |
      | 210 | Chapter | en                   | false    | false                   | User                   |
      | 220 | Chapter | en                   | false    | false                   | User                   |
      | 230 | Chapter | en                   | true     | true                    | Team                   |
      | 211 | Task    | en                   | false    | false                   | User                   |
      | 231 | Task    | en                   | false    | false                   | User                   |
      | 232 | Task    | en                   | false    | false                   | User                   |
      | 240 | Task    | en                   | false    | false                   | User                   |
      | 250 | Task    | en                   | false    | false                   | User                   |
      | 260 | Task    | en                   | false    | false                   | User                   |
      | 270 | Task    | en                   | false    | false                   | User                   |
      | 280 | Task    | en                   | false    | false                   | User                   |
      | 290 | Task    | en                   | false    | false                   | User                   |
      | 300 | Task    | en                   | false    | false                   | User                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 1        | 230     | none                     | none                     | result              | none               | false              |
      | 11       | 200     | solution                 | solution_with_grant      | none                | none               | true               |
      | 11       | 210     | content                  | none                     | none                | none               | false              |
      | 11       | 230     | info                     | none                     | none                | none               | false              |
      | 12       | 210     | content_with_descendants | none                     | answer_with_grant   | none               | false              |
      | 12       | 211     | content_with_descendants | none                     | none                | all_with_grant     | false              |
      | 12       | 220     | content_with_descendants | none                     | none                | none               | false              |
      | 12       | 230     | content_with_descendants | none                     | none                | none               | false              |
      | 12       | 231     | content_with_descendants | none                     | none                | none               | false              |
      | 12       | 232     | content_with_descendants | none                     | none                | none               | false              |
      | 13       | 200     | content_with_descendants | none                     | none                | none               | false              |
      | 13       | 210     | content_with_descendants | none                     | none                | none               | false              |
      | 13       | 220     | content_with_descendants | none                     | none                | none               | false              |
      | 13       | 230     | content_with_descendants | none                     | none                | none               | false              |
      | 13       | 211     | content_with_descendants | none                     | none                | none               | false              |
      | 13       | 231     | content_with_descendants | none                     | none                | none               | false              |
      | 13       | 232     | content_with_descendants | none                     | none                | none               | false              |
      | 13       | 250     | content_with_descendants | none                     | none                | none               | false              |
      | 14       | 200     | info                     | none                     | none                | none               | false              |
      | 14       | 210     | content_with_descendants | none                     | none                | none               | false              |
      | 14       | 220     | content_with_descendants | none                     | none                | none               | false              |
      | 14       | 230     | content_with_descendants | none                     | none                | none               | false              |
      | 14       | 211     | content_with_descendants | none                     | none                | none               | false              |
      | 14       | 231     | content_with_descendants | none                     | none                | none               | false              |
      | 14       | 232     | content_with_descendants | none                     | none                | none               | false              |
      | 16       | 200     | content_with_descendants | none                     | none                | none               | false              |
      | 16       | 210     | info                     | none                     | none                | none               | false              |
      | 16       | 220     | content_with_descendants | none                     | none                | none               | false              |
      | 16       | 230     | info                     | none                     | none                | none               | false              |
      | 16       | 211     | content_with_descendants | none                     | none                | none               | false              |
      | 16       | 231     | content_with_descendants | none                     | none                | none               | false              |
      | 16       | 232     | content_with_descendants | none                     | none                | none               | false              |
      | 18       | 200     | content                  | none                     | none                | none               | false              |
      | 18       | 210     | content                  | none                     | none                | none               | false              |
      | 18       | 220     | content                  | none                     | none                | none               | false              |
      | 18       | 230     | content                  | none                     | none                | none               | false              |
      | 18       | 211     | content                  | none                     | none                | none               | false              |
      | 18       | 231     | content                  | none                     | none                | none               | false              |
      | 18       | 232     | content                  | none                     | none                | none               | false              |
      | 19       | 200     | content                  | none                     | none                | none               | false              |
      | 19       | 210     | content                  | none                     | none                | none               | false              |
      | 19       | 211     | content                  | none                     | none                | none               | false              |
      | 30       | 210     | content_with_descendants | none                     | none                | none               | false              |
      | 29       | 220     | content_with_descendants | none                     | none                | none               | false              |
      | 30       | 230     | content_with_descendants | none                     | none                | none               | false              |
      | 29       | 240     | content_with_descendants | none                     | none                | none               | false              |
      | 30       | 250     | content_with_descendants | none                     | none                | none               | false              |
      | 29       | 260     | content_with_descendants | none                     | none                | none               | false              |
      | 30       | 270     | content_with_descendants | none                     | none                | none               | false              |
      | 29       | 280     | content_with_descendants | none                     | none                | none               | false              |
      | 30       | 290     | content_with_descendants | none                     | none                | none               | false              |
      | 29       | 300     | content_with_descendants | none                     | none                | none               | false              |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order | content_view_propagation |
      | 200            | 210           | 3           | none                     |
      | 200            | 220           | 2           | as_info                  |
      | 200            | 230           | 1           | as_content               |
      | 210            | 211           | 1           | none                     |
      | 230            | 231           | 2           | none                     |
      | 230            | 232           | 1           | none                     |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title       |
      | 200     | en           | Category 1  |
      | 210     | en           | Chapter A   |
      | 220     | en           | Chapter B   |
      | 230     | en           | Chapter C   |
      | 211     | en           | Task 1      |
      | 231     | en           | Task 2      |
      | 232     | en           | Task 3      |
      | 200     | fr           | Catégorie 1 |
      | 210     | fr           | Chapitre A  |
      | 230     | fr           | Chapitre C  |
      | 211     | fr           | Tâche 1     |
      | 250     | en           | null        |
      | 270     | en           | null        |
      | 290     | en           | null        |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | root_item_id | parent_attempt_id | ended_at            |
      | 0  | 11             | 2019-01-30 08:26:41 | null         | null              | null                |
      | 1  | 11             | 2019-01-30 08:26:41 | 200          | 0                 | 2019-01-30 09:26:48 |
      | 2  | 11             | 2019-01-30 08:26:41 | 230          | 0                 | null                |
      | 0  | 13             | 2018-01-30 09:26:42 | null         | null              | null                |
      | 0  | 18             | 2018-01-30 09:26:42 | null         | null              | null                |
      | 0  | 19             | 2018-01-30 09:26:42 | null         | null              | null                |
      | 0  | 26             | 2017-01-30 09:26:42 | null         | null              | null                |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | score_computed | submissions | started_at          | validated_at        | latest_activity_at  |
      | 0          | 11             | 200     | 91             | 11          | 2019-01-30 09:26:41 | null                | 2019-01-30 09:36:41 |
      | 0          | 11             | 210     | 92             | 12          | 2019-01-30 09:26:42 | 2019-01-31 09:26:42 | 2019-01-30 09:36:42 |
      | 0          | 11             | 211     | 93             | 13          | 2019-01-30 09:26:43 | null                | 2019-01-30 09:36:43 |
      | 0          | 11             | 220     | 94             | 14          | 2019-01-30 09:26:44 | 2019-01-31 09:26:44 | 2019-01-30 09:36:44 |
      | 0          | 11             | 230     | 95             | 15          | 2019-01-30 09:26:45 | 2019-01-31 09:26:45 | 2019-01-30 09:36:45 |
      | 0          | 11             | 231     | 96             | 16          | 2019-01-30 09:26:46 | 2019-01-31 09:26:46 | 2019-01-30 09:36:46 |
      | 0          | 11             | 232     | 97             | 17          | 2019-01-30 09:26:47 | 2019-01-31 09:26:47 | 2019-01-30 09:36:47 |
      | 1          | 11             | 200     | 90             | 11          | 2019-01-30 09:26:41 | 2019-01-31 09:26:41 | 2019-01-30 09:36:41 |
      | 0          | 13             | 200     | 45             | 2           | 2018-01-30 09:26:42 | null                | 2018-01-30 09:36:42 |
      | 0          | 13             | 210     | 56             | 5           | 2018-01-30 09:26:42 | 2018-01-31 09:26:42 | 2018-01-30 09:36:42 |
      | 0          | 13             | 230     | 78             | 4           | 2018-01-30 09:26:42 | 2018-01-31 09:26:42 | 2018-01-30 09:36:42 |
      | 0          | 14             | 200     | 0              | 2           | 2018-01-30 09:26:42 | null                | 2018-01-30 09:36:42 |
      | 0          | 18             | 200     | 0              | 2           | 2018-01-30 09:26:42 | null                | 2018-01-30 09:36:42 |
      | 0          | 19             | 200     | 0              | 2           | 2018-01-30 09:26:42 | null                | 2018-01-30 09:36:42 |
      | 0          | 19             | 210     | 10             | 2           | 2018-01-30 09:26:42 | 2019-01-31 09:26:45 | 2018-01-30 09:36:42 |
      | 0          | 19             | 220     | 20             | 2           | 2018-01-30 09:26:42 | null                | 2018-01-30 09:36:42 |
      | 2          | 11             | 230     | 94             | 15          | 2019-01-30 09:26:48 | 2019-01-31 09:26:45 | 2019-01-30 09:36:48 |
      | 0          | 26             | 200     | 10             | 3           | 2017-01-30 09:26:42 | null                | 2017-01-30 09:36:42 |

  Scenario: Get root activities for the current user
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships/activities"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "19",
        "name": "Team",
        "type": "Team",
        "activity": {
          "best_score": 92,
          "entry_participant_type": "User",
          "has_visible_children": true,
          "id": "210",
          "no_score": false,
          "permissions": {
            "can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants",
            "can_watch": "answer_with_grant", "is_owner": false
          },
          "requires_explicit_entry": false,
          "results": [
            {
              "attempt_allows_submissions_until": "9999-12-31T23:59:59Z",
              "attempt_id": "0",
              "ended_at": null,
              "latest_activity_at": "2019-01-30T09:36:42Z",
              "score_computed": 92,
              "started_at": "2019-01-30T09:26:42Z",
              "validated": true
            }
          ],
          "string": {"language_tag": "en", "title": "Chapter A"},
          "type": "Chapter"
        }
      },
      {
        "group_id": "1",
        "name": "all",
        "type": "Base",
        "activity": {
          "best_score": 94,
          "entry_participant_type": "User",
          "has_visible_children": false,
          "id": "220",
          "no_score": false,
          "permissions": {
            "can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false
          },
          "requires_explicit_entry": false,
          "results": [
            {
              "attempt_allows_submissions_until": "9999-12-31T23:59:59Z", "attempt_id": "0",
              "ended_at": null, "latest_activity_at": "2019-01-30T09:36:44Z", "score_computed": 94,
              "started_at": "2019-01-30T09:26:44Z", "validated": true
            }
          ],
          "string": {"language_tag": "en", "title": "Chapter B"},
          "type": "Chapter"
        }
      },
      {
        "group_id": "12",
        "name": "Group A",
        "type": "Class",
        "activity": {
          "id": "200",
          "type": "Course",
          "string": {"title": "Category 1", "language_tag": "en"},
          "permissions": {
            "can_view": "solution", "can_grant_view": "solution_with_grant", "can_watch": "none", "can_edit": "none", "is_owner": true
          },
          "requires_explicit_entry": false,
          "entry_participant_type": "User",
          "no_score": false,
          "has_visible_children": true,
          "best_score": 91,
          "results": [
            {
              "attempt_id": "0", "score_computed": 91, "validated": false, "started_at": "2019-01-30T09:26:41Z",
              "latest_activity_at": "2019-01-30T09:36:41Z", "ended_at": null,
              "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
            },
            {
              "attempt_id": "1", "score_computed": 90, "validated": true, "started_at": "2019-01-30T09:26:41Z",
              "latest_activity_at": "2019-01-30T09:36:41Z", "ended_at": "2019-01-30T09:26:48Z",
              "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
            }
          ]
        }
      }
    ]
    """

  Scenario: Should prefer the user's default language for titles
    Given I am the user with id "18"
    When I send a GET request to "/current-user/group-memberships/activities"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "19",
        "name": "Team",
        "type": "Team",
        "activity": {
          "best_score": 0,
          "entry_participant_type": "User",
          "has_visible_children": true,
          "id": "210",
          "no_score": false,
          "permissions": {
            "can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants",
            "can_watch": "answer_with_grant", "is_owner": false
          },
          "requires_explicit_entry": false,
          "results": [],
          "string": {"language_tag": "fr", "title": "Chapitre A"},
          "type": "Chapter"
        }
      },
      {
        "group_id": "1",
        "name": "all",
        "type": "Base",
        "activity": {
          "best_score": 0,
          "entry_participant_type": "User",
          "has_visible_children": false,
          "id": "220",
          "no_score": false,
          "permissions": {
            "can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false
          },
          "requires_explicit_entry": false,
          "results": [],
          "string": {"language_tag": "en", "title": "Chapter B"},
          "type": "Chapter"
        }
      },
      {
        "group_id": "12",
        "name": "Group A",
        "type": "Class",
        "activity": {
          "id": "200",
          "type": "Course",
          "string": {"title": "Catégorie 1", "language_tag": "fr"},
          "permissions": {
            "can_view": "content", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
          },
          "best_score": 0,
          "entry_participant_type": "User",
          "has_visible_children": true,
          "no_score": false,
          "requires_explicit_entry": false,
          "results": [
            {
              "attempt_allows_submissions_until": "9999-12-31T23:59:59Z", "attempt_id": "0",
              "ended_at": null, "latest_activity_at": "2018-01-30T09:36:42Z", "score_computed": 0,
              "started_at": "2018-01-30T09:26:42Z", "validated": false
            }
          ]
        }
      }
    ]
    """

  Scenario: Get root activites for a team
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships/activities?as_team_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "19",
        "name": "Team",
        "type": "Team",
        "activity": {
          "best_score": 56,
          "entry_participant_type": "User",
          "has_visible_children": true,
          "id": "210",
          "no_score": false,
          "permissions": {
            "can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants",
            "can_watch": "none", "is_owner": false
          },
          "requires_explicit_entry": false,
          "results": [
            {
              "attempt_allows_submissions_until": "9999-12-31T23:59:59Z",
              "attempt_id": "0",
              "ended_at": null,
              "latest_activity_at": "2018-01-30T09:36:42Z",
              "score_computed": 56,
              "started_at": "2018-01-30T09:26:42Z",
              "validated": true
            }
          ],
          "string": {"language_tag": "en", "title": "Chapter A"},
          "type": "Chapter"
        }
      },
      {
        "group_id": "13",
        "name": "Group B",
        "type": "Team",
        "activity": {
          "best_score": 0,
          "entry_participant_type": "User",
          "has_visible_children": false,
          "id": "220",
          "no_score": false,
          "permissions": {
            "can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false
          },
          "requires_explicit_entry": false,
          "results": [],
          "string": {"language_tag": "en", "title": "Chapter B"},
          "type": "Chapter"
        }
      },
      {
        "group_id": "1",
        "name": "all",
        "type": "Base",
        "activity": {
          "best_score": 0,
          "entry_participant_type": "User",
          "has_visible_children": false,
          "id": "220",
          "no_score": false,
          "permissions": {
            "can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false
          },
          "requires_explicit_entry": false,
          "results": [],
          "string": {"language_tag": "en", "title": "Chapter B"},
          "type": "Chapter"
        }
      }
    ]
    """

  Scenario: Get root activities for another team
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships/activities?as_team_id=19"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "19",
        "name": "Team",
        "type": "Team",
        "activity": {
          "id": "210",
          "type": "Chapter",
          "string": {"title": "Chapter A", "language_tag": "en"},
          "permissions": {
            "can_view": "content", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
          },
          "best_score": 10,
          "entry_participant_type": "User",
          "requires_explicit_entry": false,
          "has_visible_children": true,
          "no_score": false,
          "results": [
            {
              "attempt_allows_submissions_until": "9999-12-31T23:59:59Z", "attempt_id": "0",
              "ended_at": null, "latest_activity_at": "2018-01-30T09:36:42Z",
              "score_computed": 10, "started_at": "2018-01-30T09:26:42Z", "validated": true
            }
          ]
        }
      }
    ]
    """

  Scenario: Get root activities for a watched group
    Given I am the user with id "30"
    When I send a GET request to "/current-user/group-memberships/activities?watched_group_id=26"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "activity": {
          "best_score": 0,
          "entry_participant_type": "User",
          "has_visible_children": false,
          "id": "250",
          "no_score": false,
          "permissions": {
            "can_edit": "none",
            "can_grant_view": "none",
            "can_view": "content_with_descendants",
            "can_watch": "none",
            "is_owner": false
          },
          "requires_explicit_entry": false,
          "results": [],
          "string": {
            "language_tag": "en",
            "title": null
          },
          "type": "Task"
        },
        "group_id": "22",
        "name": "Group E",
        "type": "Club"
      },
      {
        "activity": {
          "best_score": 0,
          "entry_participant_type": "User",
          "has_visible_children": false,
          "id": "270",
          "no_score": false,
          "permissions": {
            "can_edit": "none",
            "can_grant_view": "none",
            "can_view": "content_with_descendants",
            "can_watch": "none",
            "is_owner": false
          },
          "requires_explicit_entry": false,
          "results": [],
          "string": {
            "language_tag": "en",
            "title": null
          },
          "type": "Task"
        },
        "group_id": "24",
        "name": "Group G",
        "type": "Club"
      },
      {
        "activity": {
          "best_score": 0,
          "entry_participant_type": "User",
          "has_visible_children": false,
          "id": "290",
          "no_score": false,
          "permissions": {
            "can_edit": "none",
            "can_grant_view": "none",
            "can_view": "content_with_descendants",
            "can_watch": "none",
            "is_owner": false
          },
          "requires_explicit_entry": false,
          "results": [],
          "string": {
            "language_tag": "en",
            "title": null
          },
          "type": "Task"
        },
        "group_id": "26",
        "name": "Group K",
        "type": "Club"
      },
      {
        "activity": {
          "best_score": 0,
          "entry_participant_type": "User",
          "has_visible_children": false,
          "id": "220",
          "no_score": false,
          "permissions": {
            "can_edit": "none",
            "can_grant_view": "none",
            "can_view": "content_with_descendants",
            "can_watch": "none",
            "is_owner": false
          },
          "requires_explicit_entry": false,
          "results": [],
          "string": {
            "language_tag": "en",
            "title": "Chapter B"
          },
          "type": "Chapter"
        },
        "group_id": "1",
        "name": "all",
        "type": "Base"
      }
    ]
    """
