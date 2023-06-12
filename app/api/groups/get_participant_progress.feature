Feature: Display the current progress of a participant on children of an item (groupParticipantProgress)
  Background:
    Given the database has the following table 'groups':
      | id | type    | name           |
      | 1  | Base    | Root 1         |
      | 3  | Base    | Root 2         |
      | 4  | Club    | Parent         |
      | 11 | Class   | Our Class      |
      | 12 | Class   | Other Class    |
      | 13 | Class   | Special Class  |
      | 14 | Team    | Super Team     |
      | 15 | Team    | Our Team       |
      | 16 | Team    | First Team     |
      | 17 | Other   | A custom group |
      | 18 | Club    | Our Club       |
      | 19 | Club    | Another Club   |
      | 20 | Friends | My Friends     |
      | 21 | User    | owner          |
      | 22 | User    | owner2         |
      | 51 | User    | johna          |
      | 53 | User    | johnb          |
      | 55 | User    | johnc          |
      | 57 | User    | johnd          |
      | 59 | User    | johne          |
      | 61 | User    | janea          |
      | 63 | User    | janeb          |
      | 65 | User    | janec          |
      | 67 | User    | janed          |
      | 69 | User    | janee          |
    And the database has the following table 'users':
      | login  | group_id | default_language |
      | owner  | 21       | en               |
      | owner2 | 22       | en               |
      | johna  | 51       | fr               |
      | johnb  | 53       | fr               |
      | johnc  | 55       | fr               |
      | johnd  | 57       | fr               |
      | johne  | 59       | fr               |
      | janea  | 61       | fr               |
      | janeb  | 63       | fr               |
      | janec  | 65       | fr               |
      | janed  | 67       | fr               |
      | janee  | 69       | fr               |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 1        | 21         | true              |
      | 1        | 22         | true              |
      | 19       | 4          | true              |
      | 51       | 4          | true              |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 1               | 11             |
      | 1               | 67             |
      | 3               | 13             |
      | 4               | 21             |
      | 4               | 22             |
      | 11              | 14             |
      | 11              | 17             |
      | 11              | 18             |
      | 11              | 59             |
      | 11              | 63             |
      | 11              | 65             |
      | 13              | 15             |
      | 13              | 16             |
      | 13              | 69             |
      | 14              | 51             |
      | 14              | 53             |
      | 14              | 55             |
      | 15              | 57             |
      | 15              | 59             |
      | 15              | 61             |
      | 16              | 63             |
      | 16              | 65             |
      | 16              | 67             |
      | 19              | 69             |
      | 20              | 21             |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id   | type    | default_language_tag | no_score |
      | 200  | Chapter | fr                   | false    |
      | 210  | Chapter | fr                   | false    |
      | 211  | Task    | fr                   | false    |
      | 212  | Task    | fr                   | true     |
      | 213  | Task    | fr                   | false    |
      | 214  | Task    | fr                   | false    |
      | 215  | Task    | fr                   | false    |
      | 216  | Task    | fr                   | false    |
      | 217  | Task    | fr                   | false    |
      | 218  | Task    | fr                   | false    |
      | 219  | Task    | fr                   | false    |
      | 220  | Chapter | fr                   | false    |
      | 221  | Task    | fr                   | false    |
      | 222  | Task    | fr                   | false    |
      | 223  | Task    | fr                   | false    |
      | 224  | Task    | fr                   | false    |
      | 225  | Task    | fr                   | false    |
      | 226  | Task    | fr                   | false    |
      | 227  | Task    | fr                   | false    |
      | 228  | Task    | fr                   | false    |
      | 229  | Task    | fr                   | false    |
      | 300  | Task    | fr                   | false    |
      | 310  | Chapter | fr                   | false    |
      | 311  | Task    | fr                   | false    |
      | 312  | Task    | fr                   | false    |
      | 313  | Task    | fr                   | false    |
      | 314  | Task    | fr                   | false    |
      | 315  | Task    | fr                   | false    |
      | 316  | Task    | fr                   | false    |
      | 317  | Task    | fr                   | false    |
      | 318  | Task    | fr                   | false    |
      | 319  | Task    | fr                   | false    |
      | 400  | Chapter | fr                   | false    |
      | 410  | Chapter | fr                   | false    |
      | 411  | Task    | fr                   | false    |
      | 412  | Task    | fr                   | false    |
      | 413  | Task    | fr                   | false    |
      | 414  | Task    | fr                   | false    |
      | 415  | Task    | fr                   | false    |
      | 416  | Task    | fr                   | false    |
      | 417  | Task    | fr                   | false    |
      | 418  | Task    | fr                   | false    |
      | 419  | Task    | fr                   | false    |
      | 1010 | Chapter | fr                   | false    |
      | 1020 | Chapter | fr                   | false    |
      | 1030 | Task    | fr                   | false    |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title    |
      | 214     | fr           | Tâche 14 |
      | 215     | en           | Task 15  |
      | 215     | fr           | Tâche 15 |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 200            | 210           | 0           |
      | 200            | 220           | 1           |
      | 210            | 211           | 8           |
      | 210            | 212           | 7           |
      | 210            | 213           | 6           |
      | 210            | 214           | 5           |
      | 210            | 215           | 4           |
      | 210            | 216           | 3           |
      | 210            | 217           | 2           |
      | 210            | 218           | 1           |
      | 210            | 219           | 0           |
      | 220            | 221           | 0           |
      | 220            | 222           | 1           |
      | 220            | 223           | 2           |
      | 220            | 224           | 3           |
      | 220            | 225           | 4           |
      | 220            | 226           | 5           |
      | 220            | 227           | 6           |
      | 220            | 228           | 7           |
      | 220            | 229           | 8           |
      | 300            | 310           | 0           |
      | 310            | 311           | 0           |
      | 310            | 312           | 1           |
      | 310            | 313           | 2           |
      | 310            | 314           | 3           |
      | 310            | 315           | 4           |
      | 310            | 316           | 5           |
      | 310            | 317           | 6           |
      | 310            | 318           | 7           |
      | 310            | 319           | 8           |
      | 400            | 410           | 0           |
      | 410            | 411           | 0           |
      | 410            | 412           | 1           |
      | 410            | 413           | 2           |
      | 410            | 414           | 3           |
      | 410            | 415           | 4           |
      | 410            | 416           | 5           |
      | 410            | 417           | 6           |
      | 410            | 418           | 7           |
      | 410            | 419           | 8           |
      | 1020           | 1030          | 0           |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       | can_watch_generated |
      | 4        | 210     | content                  | none                |
      | 21       | 210     | none                     | result              |
      | 22       | 210     | none                     | result              |
      | 21       | 211     | info                     | none                |
      | 22       | 211     | info                     | none                |
      | 20       | 212     | content                  | none                |
      | 21       | 213     | content_with_descendants | none                |
      | 22       | 213     | content_with_descendants | none                |
      | 20       | 214     | info                     | none                |
      | 21       | 215     | content                  | none                |
      | 22       | 215     | content                  | none                |
      | 20       | 216     | none                     | none                |
      | 21       | 217     | none                     | none                |
      | 22       | 217     | none                     | none                |
      | 20       | 218     | none                     | none                |
      | 21       | 219     | none                     | none                |
      | 22       | 219     | none                     | none                |
      | 20       | 220     | none                     | answer              |
      | 20       | 221     | info                     | none                |
      | 21       | 222     | content                  | none                |
      | 22       | 222     | content                  | none                |
      | 20       | 223     | content_with_descendants | none                |
      | 21       | 224     | info                     | none                |
      | 22       | 224     | info                     | none                |
      | 20       | 225     | content                  | none                |
      | 21       | 226     | none                     | none                |
      | 22       | 226     | none                     | none                |
      | 20       | 227     | none                     | none                |
      | 21       | 228     | none                     | none                |
      | 22       | 228     | none                     | none                |
      | 20       | 229     | none                     | none                |
      | 4        | 310     | none                     | none                |
      | 20       | 310     | none                     | result              |
      | 21       | 311     | info                     | none                |
      | 20       | 312     | content                  | none                |
      | 21       | 313     | content_with_descendants | none                |
      | 20       | 314     | info                     | none                |
      | 21       | 315     | content                  | none                |
      | 20       | 316     | none                     | none                |
      | 21       | 317     | none                     | none                |
      | 20       | 318     | none                     | none                |
      | 21       | 319     | none                     | none                |
      | 20       | 411     | info                     | none                |
      | 21       | 412     | content                  | none                |
      | 20       | 413     | content_with_descendants | none                |
      | 21       | 414     | info                     | none                |
      | 20       | 415     | content                  | none                |
      | 21       | 416     | none                     | none                |
      | 20       | 417     | none                     | none                |
      | 21       | 418     | none                     | none                |
      | 20       | 419     | none                     | none                |
      | 14       | 210     | content_with_descendants | none                |
      | 14       | 211     | info                     | none                |
      | 14       | 212     | info                     | none                |
      | 14       | 213     | info                     | none                |
      | 14       | 215     | info                     | none                |
      | 14       | 1010    | content                  | none                |
      | 14       | 1020    | content_with_descendants | none                |
      | 14       | 1030    | info                     | none                |
      | 51       | 210     | content                  | result              |
      | 51       | 211     | info                     | none                |
      | 51       | 212     | content                  | none                |
      | 51       | 213     | content_with_descendants | none                |
      | 51       | 214     | info                     | none                |
      | 51       | 215     | content                  | none                |
      | 51       | 216     | none                     | none                |
      | 51       | 217     | none                     | none                |
      | 51       | 1020    | content                  | result              |
      | 51       | 1030    | content                  | result              |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          |
      | 0  | 14             | 2020-01-01 00:03:00 |
      | 1  | 14             | 2020-01-01 00:03:00 |
      | 2  | 14             | 2020-01-01 00:03:00 |
      | 3  | 14             | 2020-01-01 00:03:00 |
      | 0  | 15             | 2020-01-01 00:01:00 |
      | 0  | 16             | 2020-01-01 00:16:00 |
      | 1  | 67             | 2020-01-01 00:17:00 |
      | 0  | 67             | 2020-01-01 00:03:00 |
      | 4  | 14             | 2020-01-01 00:03:00 |
      | 1  | 15             | 2020-01-01 00:01:00 |
      | 5  | 14             | 2020-01-01 00:03:00 |
      | 0  | 21             | 2020-01-01 00:03:00 |
      | 0  | 51             | 2020-01-01 00:03:00 |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          | score_computed | score_obtained_at   | hints_cached | submissions | validated_at        | latest_activity_at  |
      | 0          | 14             | 211     | 2020-01-01 00:03:00 | 0              | 2020-01-01 00:03:00 | 100          | 100         | 2020-01-01 00:04:00 | 2020-01-01 00:14:01 | # latest_activity_at for 51, 211 comes from this line (the last activity is made by a team)
      | 1          | 14             | 211     | 2020-01-01 00:02:00 | 40             | 2020-01-01 00:03:00 | 2            | 3           | 2020-01-01 00:03:02 | 2020-01-01 00:14:00 | # min(validated_at) for 51, 211 comes from this line (from a team)
      | 2          | 14             | 211     | 2020-01-01 00:03:00 | 50             | 2020-01-01 00:03:00 | 3            | 4           | 2020-01-01 00:04:02 | 2020-01-01 00:13:00 | # hints_cached & submissions for 51, 211 come from this line (the best attempt is made by a team)
      | 3          | 14             | 211     | 2020-01-01 00:03:00 | 50             | 2020-01-01 00:04:00 | 10           | 20          | null                | 2020-01-01 00:12:00 |
      | 0          | 15             | 211     | 2020-01-01 00:02:00 | 0              | null                | 0            | 0           | null                | 2020-01-01 00:11:00 |
      | 0          | 15             | 212     | 2020-01-01 00:01:00 | 0              | null                | 0            | 0           | null                | 2020-01-01 00:10:00 |
      | 0          | 16             | 212     | 2020-01-01 00:16:00 | 10             | 2020-01-01 00:04:00 | 0            | 0           | null                | 2020-01-01 00:18:00 | # started_at for 67, 212 & 63, 212 comes from this line (the first attempt is started by a team)
      | 0          | 67             | 212     | 2020-01-01 00:17:00 | 20             | 2020-01-01 00:05:00 | 1            | 2           | null                | 2020-01-01 00:18:00 | # hints_cached & submissions for 67, 212 come from this line (the best attempt is made by a user)
      | 1          | 67             | 212     | 2020-01-01 00:17:00 | 10             | 2020-01-01 00:04:00 | 6            | 7           | null                | 2021-01-01 00:01:00 | # latest_activity_at for 67, 212 comes from this line (the last activity is made by a user)
      | 0          | 67             | 213     | 2020-01-01 00:15:00 | 0              | null                | 0            | 0           | null                | 2020-01-01 00:15:00 | # started_at for 67, 213 comes from this line (the first attempt is started by a user)
      | 0          | 67             | 214     | 2020-01-01 00:03:00 | 15             | 2020-01-01 00:03:01 | 10           | 11          | 2020-01-01 00:03:01 | 2020-01-01 00:04:01 | # min(validated_at) for 67, 214 comes from this line (from a user)
      | 0          | 67             | 215     | 2021-01-01 00:04:00 | 0              | null                | 0            | 0           | null                | 2020-01-01 00:15:00 | # started_at for 67, 213 comes from this line (the first attempt is started by a user)
      | 4          | 14             | 211     | 2020-01-01 00:03:00 | 0              | null                | 0            | 0           | null                | 2020-01-01 00:09:00 |
      | 1          | 15             | 211     | 2020-01-01 00:02:00 | 0              | null                | 0            | 0           | null                | 2020-01-01 00:08:00 |
      | 1          | 15             | 212     | 2020-01-01 00:01:00 | 100            | null                | 0            | 0           | null                | 2020-01-01 00:07:00 |
      | 5          | 14             | 211     | 2020-01-01 00:03:00 | 0              | null                | 0            | 0           | null                | 2020-01-01 00:06:00 |
      | 0          | 14             | 212     | 2020-01-01 00:03:00 | 0              | 2020-01-01 00:03:00 | 1            | 2           | null                | 2020-01-01 00:14:02 |
      | 0          | 14             | 213     | 2021-01-01 00:03:00 | 0              | null                | 0            | 0           | null                | 2021-01-01 00:03:00 |
      | 0          | 14             | 215     | 2020-01-01 00:03:00 | 0              | 2020-01-01 00:03:00 | 100          | 100         | 2020-01-01 00:04:00 | 2020-01-01 00:14:02 |
      | 0          | 21             | 210     | 2021-01-01 00:02:00 | 0              | 2021-01-01 00:02:00 | 0            | 0           | null                | 2021-01-01 00:02:00 |
      | 0          | 51             | 210     | 2021-01-01 00:02:00 | 0              | 2021-01-01 00:02:00 | 0            | 0           | null                | 2021-01-01 00:02:00 |
      | 0          | 51             | 1010    | 2021-01-01 00:02:00 | 0              | 2021-01-01 00:02:00 | 0            | 0           | null                | 2021-01-01 00:02:00 |

  Scenario: Get progress of a user
    Given I am the user with id "21"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2021-01-01 00:00:00"
    When I send a GET request to "/items/210/participant-progress?watched_group_id=67"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "item": {
        "hints_requested": 0,
        "item_id": "210",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      "children": [
        {
          "hints_requested": 0,
          "item_id": "215",
          "no_score": false,
          "type": "Task",
          "string": {"language_tag": "en", "title": "Task 15"},
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content", "can_watch": "none", "is_owner": false},
          "started_at": "2021-01-01T00:04:00Z",
          "latest_activity_at": "2020-01-01T00:15:00Z",
          "score": 0,
          "submissions": 0,
          "time_spent": 0,
          "validated": false
        },
        {
          "hints_requested": 10,
          "item_id": "214",
          "no_score": false,
          "type": "Task",
          "string": {"language_tag": "fr", "title": "Tâche 14"},
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "info", "can_watch": "none", "is_owner": false},
          "started_at": "2020-01-01T00:03:00Z",
          "latest_activity_at": "2020-01-01T00:04:01Z",
          "score": 15,
          "submissions": 11,
          "time_spent": 1,
          "validated": true
        },
        {
          "hints_requested": 0,
          "item_id": "213",
          "no_score": false,
          "type": "Task",
          "string": {"language_tag": "", "title": null},
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false},
          "started_at": "2020-01-01T00:15:00Z",
          "latest_activity_at": "2020-01-01T00:15:00Z",
          "score": 0,
          "submissions": 0,
          "time_spent": 31621500,
          "validated": false
        },
        {
          "hints_requested": 1,
          "item_id": "212",
          "no_score": true,
          "type": "Task",
          "string": {"language_tag": "", "title": null},
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content", "can_watch": "none", "is_owner": false},
          "started_at": "2020-01-01T00:16:00Z",
          "latest_activity_at": "2021-01-01T00:01:00Z",
          "score": 20,
          "submissions": 2,
          "time_spent": 31621440,
          "validated": false
        },
        {
          "item_id": "211",
          "no_score": false,
          "type": "Task",
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "info", "can_watch": "none", "is_owner": false},
          "string": {"language_tag": "", "title": null},
          "started_at": null,
          "latest_activity_at": null,
          "score": 0,
          "hints_requested": 0,
          "submissions": 0,
          "time_spent": 0,
          "validated": false
        }
      ]
    }
    """

  Scenario: Get progress of a current user
    Given I am the user with id "51"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2021-01-01 00:00:00"
    When I send a GET request to "/items/210/participant-progress"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "item": {
        "hints_requested": 0,
        "item_id": "210",
        "latest_activity_at": "2021-01-01T00:02:00Z",
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      "children": [
        {
          "no_score": false,
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content", "can_watch": "none", "is_owner": false},
          "string": {"language_tag": "fr", "title": "Tâche 15"},
          "type": "Task",
          "hints_requested": 100,
          "item_id": "215",
          "started_at": "2020-01-01T00:03:00Z",
          "latest_activity_at": "2020-01-01T00:14:02Z",
          "score": 0,
          "submissions": 100,
          "time_spent": 60,
          "validated": true
        },
        {
          "no_score": false,
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "info", "can_watch": "none", "is_owner": false},
          "string": {"language_tag": "fr", "title": "Tâche 14"},
          "type": "Task",
          "hints_requested": 0,
          "item_id": "214",
          "started_at": null,
          "latest_activity_at": null,
          "score": 0,
          "submissions": 0,
          "time_spent": 0,
          "validated": false
        },
        {
          "no_score": false,
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false},
          "string": {"language_tag": "", "title": null},
          "type": "Task",
          "hints_requested": 0,
          "item_id": "213",
          "started_at": "2021-01-01T00:03:00Z",
          "latest_activity_at": "2021-01-01T00:03:00Z",
          "score": 0,
          "submissions": 0,
          "time_spent": 0,
          "validated": false
        },
        {
          "no_score": true,
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content", "can_watch": "none", "is_owner": false},
          "string": {"language_tag": "", "title": null},
          "type": "Task",
          "hints_requested": 1,
          "item_id": "212",
          "started_at": "2020-01-01T00:03:00Z",
          "latest_activity_at": "2020-01-01T00:14:02Z",
          "score": 0,
          "submissions": 2,
          "time_spent": 31622220,
          "validated": false
        },
        {
          "item_id": "211",
          "no_score": false,
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "info", "can_watch": "none", "is_owner": false},
          "string": {"language_tag": "", "title": null},
          "type": "Task",
          "started_at": "2020-01-01T00:02:00Z",
          "latest_activity_at": "2020-01-01T00:14:01Z",
          "score": 50,
          "hints_requested": 3,
          "submissions": 4,
          "time_spent": 62,
          "validated": true
        }
      ]
    }
    """

  Scenario: Get progress of a current user's team
    Given I am the user with id "51"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2021-01-01 00:00:00"
    When I send a GET request to "/items/210/participant-progress?as_team_id=14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "item": {
        "hints_requested": 0,
        "item_id": "210",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      "children": [
        {
          "no_score": false,
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content", "can_watch": "none", "is_owner": false},
          "string": {"language_tag": "fr", "title": "Tâche 15"},
          "type": "Task",
          "hints_requested": 100,
          "item_id": "215",
          "started_at": "2020-01-01T00:03:00Z",
          "latest_activity_at": "2020-01-01T00:14:02Z",
          "score": 0,
          "submissions": 100,
          "time_spent": 60,
          "validated": true
        },
        {
          "no_score": false,
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content_with_descendants", "can_watch": "none", "is_owner": false},
          "string": {"language_tag": "", "title": null},
          "type": "Task",
          "hints_requested": 0,
          "item_id": "213",
          "started_at": "2021-01-01T00:03:00Z",
          "latest_activity_at": "2021-01-01T00:03:00Z",
          "score": 0,
          "submissions": 0,
          "time_spent": 0,
          "validated": false
        },
        {
          "no_score": true,
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "content", "can_watch": "none", "is_owner": false},
          "string": {"language_tag": "", "title": null},
          "type": "Task",
          "hints_requested": 1,
          "item_id": "212",
          "started_at": null,
          "started_at": "2020-01-01T00:03:00Z",
          "latest_activity_at": "2020-01-01T00:14:02Z",
          "score": 0,
          "submissions": 2,
          "time_spent": 31622220,
          "validated": false
        },
        {
          "item_id": "211",
          "no_score": false,
          "current_user_permissions": {"can_edit": "none", "can_grant_view": "none", "can_view": "info", "can_watch": "none", "is_owner": false},
          "string": {"language_tag": "", "title": null},
          "type": "Task",
          "started_at": null,
          "started_at": "2020-01-01T00:02:00Z",
          "latest_activity_at": "2020-01-01T00:14:01Z",
          "score": 50,
          "hints_requested": 3,
          "submissions": 4,
          "time_spent": 62,
          "validated": true
        }
      ]
    }
    """

  Scenario: No visible child items but the children key should be present because the current user have a started result on the requested item
    Given I am the user with id "51"
    When I send a GET request to "/items/1010/participant-progress?as_team_id=14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "item": {
        "hints_requested": 0,
        "item_id": "1010",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      "children": []
    }
    """

  Scenario: Should not return the children when the current user doesn't have a started result on the requested item with as_team_id
    Given I am the user with id "51"
    When I send a GET request to "/items/1020/participant-progress?as_team_id=14"
    Then the response code should be 200
    And the response at $.item.item_id should be "1020"
    And the response should not be defined at $.children

  Scenario: Should not return the children when the current user doesn't have a started result on the requested item with watched_group_id
    Given I am the user with id "22"
    When I send a GET request to "/items/210/participant-progress?watched_group_id=67"
    Then the response code should be 200
    And the response at $.item.item_id should be "210"
    And the response should not be defined at $.children
