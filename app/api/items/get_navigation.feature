Feature: Get navigation for an item
  Background:
    Given the database has the following table "groups":
      | id | name      | type  |
      | 1  | all       | Base  |
      | 12 | Group A   | Class |
      | 13 | Group B   | Team  |
      | 19 | Group C   | Team  |
    And the database has the following table "languages":
      | tag |
      | fr  |
    And the database has the following users:
      | group_id | login     | default_language |
      | 11       | jdoe      |                  |
      | 14       | info_root |                  |
      | 16       | info_mid  |                  |
      | 18       | fr_user   | fr               |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 1               | 12             |
      | 1               | 13             |
      | 1               | 14             |
      | 1               | 16             |
      | 1               | 18             |
      | 1               | 19             |
      | 12              | 11             |
      | 13              | 11             |
      | 19              | 11             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | manager_id | group_id | can_watch_members |
      | 12         | 1        | true              |
      | 12         | 19       | true              |
    And the database has the following table "items":
      | id  | type    | default_language_tag | no_score | requires_explicit_entry | entry_participant_type |
      | 200 | Task    | en                   | false    | false                   | User                   |
      | 210 | Chapter | en                   | false    | false                   | User                   |
      | 220 | Chapter | en                   | false    | false                   | User                   |
      | 230 | Chapter | en                   | true     | true                    | Team                   |
      | 211 | Task    | en                   | false    | false                   | User                   |
      | 231 | Task    | en                   | false    | false                   | User                   |
      | 232 | Task    | en                   | false    | false                   | User                   |
      | 250 | Task    | en                   | false    | false                   | User                   |
    And the database has the following table "permissions_generated":
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
      | 18       | 210     | info                     | none                     | none                | none               | false              |
      | 18       | 220     | content                  | none                     | none                | none               | false              |
      | 18       | 230     | content                  | none                     | none                | none               | false              |
      | 18       | 211     | content                  | none                     | none                | none               | false              |
      | 18       | 231     | content                  | none                     | none                | none               | false              |
      | 18       | 232     | content                  | none                     | none                | none               | false              |
      | 19       | 200     | content                  | none                     | none                | none               | false              |
      | 19       | 210     | content                  | none                     | none                | none               | false              |
      | 19       | 211     | content                  | none                     | none                | none               | false              |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order | content_view_propagation |
      | 200            | 210           | 3           | none                     |
      | 200            | 220           | 2           | as_info                  |
      | 200            | 230           | 1           | as_content               |
      | 210            | 211           | 1           | none                     |
      | 230            | 231           | 2           | none                     |
      | 230            | 232           | 1           | none                     |
    And the database has the following table "items_strings":
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
    And the database has the following table "attempts":
      | id | participant_id | created_at          | root_item_id | parent_attempt_id | ended_at            |
      | 0  | 11             | 2019-01-30 08:26:41 | null         | null              | null                |
      | 1  | 11             | 2019-01-30 08:26:41 | 200          | 0                 | null                |
      | 2  | 11             | 2019-01-30 08:26:41 | 230          | 0                 | 2019-01-30 09:26:48 |
      | 0  | 13             | 2018-01-30 09:26:42 | null         | null              | null                |
      | 0  | 19             | 2018-01-30 09:26:42 | null         | null              | null                |
    And the database has the following table "results":
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

  Scenario: Get navigation
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Task",
        "string": {"title": "Category 1", "language_tag": "en"},
        "permissions": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_watch": "none", "can_edit": "none", "is_owner": true
        },
        "attempt_id": "0",
        "children": [
          {
            "id": "230",
            "type": "Chapter",
            "requires_explicit_entry": true,
            "entry_participant_type": "Team",
            "no_score": true,
            "has_visible_children": true,
            "string": {"title": "Chapter C", "language_tag": "en"},
            "best_score": 95,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "result", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 95, "validated": true, "started_at": "2019-01-30T09:26:45Z",
                "latest_activity_at": "2019-01-30T09:36:45Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              },
              {
                "attempt_id": "2", "score_computed": 94, "validated": true, "started_at": "2019-01-30T09:26:48Z",
                "latest_activity_at": "2019-01-30T09:36:48Z", "ended_at": "2019-01-30T09:26:48Z",
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ]
          },
          {
            "id": "220",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": false,
            "string": {"title": "Chapter B", "language_tag": "en"},
            "best_score": 94,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 94, "validated": true, "started_at": "2019-01-30T09:26:44Z",
                "latest_activity_at": "2019-01-30T09:36:44Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ]
          },
          {
            "id": "210",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": true,
            "string": {"title": "Chapter A", "language_tag": "en"},
            "best_score": 92,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "answer_with_grant", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 92, "validated": true, "started_at": "2019-01-30T09:26:42Z",
                "latest_activity_at": "2019-01-30T09:36:42Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ]
          }
        ]
      }
      """

  Scenario: Should only return Skills and compute has_visible children on Skills only when we get navigation on a Skill
    Given the database has the following user:
      | login | group_id |
      | user  | 1000     |
    And the database has the following table "items":
      | id   | type    | default_language_tag | no_score | requires_explicit_entry | entry_participant_type |
      | 1000 | Skill   | en                   | false    | false                   | User                   |
      | 1100 | Skill   | en                   | false    | false                   | User                   |
      | 1110 | Skill   | en                   | false    | false                   | User                   |
      | 1120 | Task    | en                   | false    | false                   | User                   |
      | 1200 | Skill   | en                   | false    | false                   | User                   |
      | 1210 | Skill   | en                   | false    | false                   | User                   |
      | 1220 | Task    | en                   | false    | false                   | User                   |
      | 1300 | Skill   | en                   | false    | false                   | User                   |
      | 1400 | Chapter | en                   | false    | false                   | User                   |
      | 1500 | Task    | en                   | false    | false                   | User                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 1000     | 1000    | solution           | solution_with_grant      | none                | none               | false              |
      | 1000     | 1100    | solution           | solution_with_grant      | none                | none               | false              |
      | 1000     | 1110    | none               | none                     | none                | none               | false              |
      | 1000     | 1120    | solution           | solution_with_grant      | none                | none               | false              |
      | 1000     | 1200    | solution           | solution_with_grant      | none                | none               | false              |
      | 1000     | 1210    | solution           | solution_with_grant      | none                | none               | false              |
      | 1000     | 1220    | none               | none                     | none                | none               | false              |
      | 1000     | 1300    | solution           | solution_with_grant      | none                | none               | false              |
      | 1000     | 1400    | solution           | solution_with_grant      | none                | none               | false              |
      | 1000     | 1500    | solution           | solution_with_grant      | none                | none               | false              |
      And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order | content_view_propagation |
      | 1000           | 1100          | 1           | none                     |
      | 1100           | 1110          | 1           | none                     |
      | 1100           | 1120          | 2           | none                     |
      | 1000           | 1200          | 2           | as_info                  |
      | 1200           | 1210          | 1           | as_info                  |
      | 1200           | 1220          | 2           | as_info                  |
      | 1000           | 1300          | 3           | as_content               |
      | 1000           | 1400          | 4           | as_content               |
      | 1000           | 1500          | 5           | as_content               |
    And the database has the following table "items_strings":
      | item_id | language_tag | title              |
      | 1000    | en           | Parent Skill       |
      | 1100    | en           | Child lvl1 Skill 1 |
      | 1110    | en           | Child lvl2 Skill   |
      | 1120    | en           | Child lvl2 Task    |
      | 1200    | en           | Child lvl1 Skill 2 |
      | 1210    | en           | Child lvl2 Skill 2 |
      | 1220    | en           | Child lvl1 Task 2  |
      | 1300    | en           | Child lvl1 Skill 3 |
      | 1400    | en           | Child lvl1 Chapter |
      | 1500    | en           | Child lvl1 Task    |
    And the database has the following table "attempts":
      | id | participant_id | created_at          | root_item_id | parent_attempt_id | ended_at |
      | 0  | 1000           | 2020-01-01 00:00:00 | null         | null              | null     |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | score_computed | submissions | started_at          | validated_at        | latest_activity_at  |
      | 0          | 1000           | 1000    | 100            | 1           | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 |
      | 0          | 1000           | 1100    | 100            | 1           | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 |
      | 0          | 1000           | 1120    | 100            | 1           | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 |
      | 0          | 1000           | 1200    | 100            | 1           | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 |
      | 0          | 1000           | 1210    | 100            | 1           | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 |
      | 0          | 1000           | 1300    | 100            | 1           | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 |
      | 0          | 1000           | 1400    | 100            | 1           | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 |
      | 0          | 1000           | 1500    | 100            | 1           | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 | 2020-01-01 00:00:00 |
    Given I am the user with id "1000"
    When I send a GET request to "/items/1000/navigation?attempt_id=0"
    Then the response code should be 200
    And the response at $.children[*] should be:
      | string.title       | has_visible_children |
      | Child lvl1 Skill 1 | false                |
      | Child lvl1 Skill 2 | true                 |
      | Child lvl1 Skill 3 | false                |

  Scenario Outline: Get navigation (with child_attempt_id)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?child_attempt_id=<child_attempt_id>"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Task",
        "string": {"title": "Category 1", "language_tag": "en"},
        "attempt_id": "0",
        "permissions": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_watch": "none", "can_edit": "none", "is_owner": true
        },
        "children": [
          {
            "id": "230",
            "type": "Chapter",
            "requires_explicit_entry": true,
            "entry_participant_type": "Team",
            "no_score": true,
            "has_visible_children": true,
            "string": {"title": "Chapter C", "language_tag": "en"},
            "best_score": 95,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "result", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 95, "validated": true, "started_at": "2019-01-30T09:26:45Z",
                "latest_activity_at": "2019-01-30T09:36:45Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              },
              {
                "attempt_id": "2", "score_computed": 94, "validated": true, "started_at": "2019-01-30T09:26:48Z",
                "latest_activity_at": "2019-01-30T09:36:48Z", "ended_at": "2019-01-30T09:26:48Z",
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ]
          },
          {
            "id": "220",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": false,
            "string": {"title": "Chapter B", "language_tag": "en"},
            "best_score": 94,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 94, "validated": true, "started_at": "2019-01-30T09:26:44Z",
                "latest_activity_at": "2019-01-30T09:36:44Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ]
          },
          {
            "id": "210",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": true,
            "string": {"title": "Chapter A", "language_tag": "en"},
            "best_score": 92,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "answer_with_grant", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 92, "validated": true, "started_at": "2019-01-30T09:26:42Z",
                "latest_activity_at": "2019-01-30T09:36:42Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ]
          }
        ]
      }
      """
  Examples:
    | child_attempt_id |
    | 0                |
    | 2                |

  Scenario: Should return only one node if the root item doesn't have children
    Given I am the user with id "11"
    When I send a GET request to "/items/232/navigation?attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "232",
        "type": "Task",
        "string": {"title": "Task 3", "language_tag": "en"},
        "attempt_id": "0",
        "permissions": {
          "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
        },
        "children": []
      }
      """

  Scenario: Should return only one node if the user has only info access to the root item
    Given I am the user with id "14"
    When I send a GET request to "/items/200/navigation?attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Task",
        "string": {"title": "Category 1", "language_tag": "en"},
        "attempt_id": "0",
        "permissions": {
          "can_view": "info", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
        },
        "children": []
      }
      """

  Scenario: Should prefer the user's default language for titles
    Given I am the user with id "18"
    When I send a GET request to "/items/200/navigation?attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Task",
        "string": {"title": "Catégorie 1", "language_tag": "fr"},
        "attempt_id": "0",
        "permissions": {
          "can_view": "content", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
        },
        "children": [
          {
            "id": "230",
            "type": "Chapter",
            "requires_explicit_entry": true,
            "entry_participant_type": "Team",
            "no_score": true,
            "has_visible_children": true,
            "string": {"title": "Chapitre C", "language_tag": "fr"},
            "best_score": 0,
            "permissions": {
              "can_view": "content", "can_grant_view": "none", "can_watch": "result", "can_edit": "none", "is_owner": false
            },
            "results": []
          },
          {
            "id": "220",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": false,
            "string": {"title": "Chapter B", "language_tag": "en"},
            "best_score": 0,
            "permissions": {
              "can_view": "content", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
            },
            "results": []
          },
          {
            "id": "210",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": false,
            "type": "Chapter",
            "string": {"title": "Chapitre A", "language_tag": "fr"},
            "best_score": 0,
            "permissions": {
              "can_view": "info", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
            },
            "results": []
          }
        ]
      }
      """

  Scenario: Get navigation (as team)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?as_team_id=13&attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Task",
        "string": {"title": "Category 1", "language_tag": "en"},
        "attempt_id": "0",
        "permissions": {
          "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
        },
        "children": [
          {
            "id": "230",
            "type": "Chapter",
            "requires_explicit_entry": true,
            "entry_participant_type": "Team",
            "no_score": true,
            "has_visible_children": true,
            "string": {"title": "Chapter C", "language_tag": "en"},
            "best_score": 78,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "result", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 78, "validated": true, "started_at": "2018-01-30T09:26:42Z",
                "latest_activity_at": "2018-01-30T09:36:42Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ]
          },
          {
            "id": "220",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": false,
            "string": {"title": "Chapter B", "language_tag": "en"},
            "best_score": 0,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
            },
            "results": []
          },
          {
            "id": "210",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": true,
            "string": {"title": "Chapter A", "language_tag": "en"},
            "best_score": 56,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 56, "validated": true, "started_at": "2018-01-30T09:26:42Z",
                "latest_activity_at": "2018-01-30T09:36:42Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ]
          }
        ]
      }
      """

  Scenario: Get navigation (as another team)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?as_team_id=19&attempt_id=0"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Task",
        "string": {"title": "Category 1", "language_tag": "en"},
        "attempt_id": "0",
        "permissions": {
          "can_view": "content", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
        },
        "children": [
          {
            "id": "210",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": true,
            "string": {"title": "Chapter A", "language_tag": "en"},
            "best_score": 10,
            "permissions": {
              "can_view": "content", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 10, "validated": true, "started_at": "2018-01-30T09:26:42Z",
                "latest_activity_at": "2018-01-30T09:36:42Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ]
          }
        ]
      }
      """

  Scenario: Get navigation with watched_group_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?attempt_id=0&watched_group_id=19"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Task",
        "string": {"title": "Category 1", "language_tag": "en"},
        "permissions": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_watch": "none", "can_edit": "none", "is_owner": true
        },
        "attempt_id": "0",
        "children": [
          {
            "id": "230",
            "type": "Chapter",
            "requires_explicit_entry": true,
            "entry_participant_type": "Team",
            "no_score": true,
            "has_visible_children": true,
            "string": {"title": "Chapter C", "language_tag": "en"},
            "best_score": 95,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "result", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 95, "validated": true, "started_at": "2019-01-30T09:26:45Z",
                "latest_activity_at": "2019-01-30T09:36:45Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              },
              {
                "attempt_id": "2", "score_computed": 94, "validated": true, "started_at": "2019-01-30T09:26:48Z",
                "latest_activity_at": "2019-01-30T09:36:48Z", "ended_at": "2019-01-30T09:26:48Z",
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ],
            "watched_group": {"can_view": "none", "all_validated": false, "avg_score": 0}
          },
          {
            "id": "220",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": false,
            "string": {"title": "Chapter B", "language_tag": "en"},
            "best_score": 94,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 94, "validated": true, "started_at": "2019-01-30T09:26:44Z",
                "latest_activity_at": "2019-01-30T09:36:44Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ],
            "watched_group": {"can_view": "none"}
          },
          {
            "id": "210",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": true,
            "string": {"title": "Chapter A", "language_tag": "en"},
            "best_score": 92,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "answer_with_grant", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 92, "validated": true, "started_at": "2019-01-30T09:26:42Z",
                "latest_activity_at": "2019-01-30T09:36:42Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ],
            "watched_group": {"can_view": "content", "all_validated": true, "avg_score": 10}
          }
        ]
      }
      """

  Scenario: Get navigation with another watched_group_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/navigation?attempt_id=0&watched_group_id=1"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Task",
        "string": {"title": "Category 1", "language_tag": "en"},
        "permissions": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_watch": "none", "can_edit": "none", "is_owner": true
        },
        "attempt_id": "0",
        "children": [
          {
            "id": "230",
            "type": "Chapter",
            "requires_explicit_entry": true,
            "entry_participant_type": "Team",
            "no_score": true,
            "has_visible_children": true,
            "string": {"title": "Chapter C", "language_tag": "en"},
            "best_score": 95,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "result", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 95, "validated": true, "started_at": "2019-01-30T09:26:45Z",
                "latest_activity_at": "2019-01-30T09:36:45Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              },
              {
                "attempt_id": "2", "score_computed": 94, "validated": true, "started_at": "2019-01-30T09:26:48Z",
                "latest_activity_at": "2019-01-30T09:36:48Z", "ended_at": "2019-01-30T09:26:48Z",
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ],
            "watched_group": {"can_view": "none", "all_validated": false, "avg_score": 28.833334}
          },
          {
            "id": "220",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": false,
            "string": {"title": "Chapter B", "language_tag": "en"},
            "best_score": 94,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 94, "validated": true, "started_at": "2019-01-30T09:26:44Z",
                "latest_activity_at": "2019-01-30T09:36:44Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ],
            "watched_group": {"can_view": "none"}
          },
          {
            "id": "210",
            "type": "Chapter",
            "requires_explicit_entry": false,
            "entry_participant_type": "User",
            "no_score": false,
            "has_visible_children": true,
            "string": {"title": "Chapter A", "language_tag": "en"},
            "best_score": 92,
            "permissions": {
              "can_view": "content_with_descendants", "can_grant_view": "none", "can_watch": "answer_with_grant", "can_edit": "none", "is_owner": false
            },
            "results": [
              {
                "attempt_id": "0", "score_computed": 92, "validated": true, "started_at": "2019-01-30T09:26:42Z",
                "latest_activity_at": "2019-01-30T09:36:42Z", "ended_at": null,
                "attempt_allows_submissions_until": "9999-12-31T23:59:59Z"
              }
            ],
            "watched_group": {"can_view": "none", "all_validated": false, "avg_score": 26.333334}
          }
        ]
      }
      """
