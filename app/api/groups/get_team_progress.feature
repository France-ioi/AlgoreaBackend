Feature: Display the current progress of teams on a subset of items (groupTeamProgress)
  Background:
    Given the database has the following table 'groups':
      | id | type     | name           |
      | 1  | Base     | Root 1         |
      | 3  | Base     | Root 2         |
      | 11 | Class    | Our Class      |
      | 12 | Class    | Other Class    |
      | 13 | Class    | Special Class  |
      | 14 | Team     | Super Team     |
      | 15 | Team     | Our Team       |
      | 16 | Team     | First Team     |
      | 17 | Other    | A custom group |
      | 18 | Club     | Our Club       |
      | 20 | Friends  | My Friends     |
      | 21 | UserSelf | owner          |
      | 51 | UserSelf | johna          |
      | 53 | UserSelf | johnb          |
      | 55 | UserSelf | johnc          |
      | 57 | UserSelf | johnd          |
      | 59 | UserSelf | johne          |
      | 61 | UserSelf | janea          |
      | 63 | UserSelf | janeb          |
      | 65 | UserSelf | janec          |
      | 67 | UserSelf | janed          |
      | 69 | UserSelf | janee          |
    And the database has the following table 'users':
      | login | group_id |
      | owner | 21       |
      | johna | 51       |
      | johnb | 53       |
      | johnc | 55       |
      | johnd | 57       |
      | johne | 59       |
      | janea | 61       |
      | janeb | 63       |
      | janec | 65       |
      | janed | 67       |
      | janee | 69       |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 1        | 21         |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 1               | 11             |
      | 3               | 13             |
      | 11              | 14             |
      | 11              | 16             |
      | 11              | 17             |
      | 11              | 18             |
      | 11              | 59             |
      | 13              | 15             |
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
      | 20              | 21             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 1                 | 1              | 1       |
      | 1                 | 11             | 0       |
      | 1                 | 12             | 0       |
      | 1                 | 14             | 0       |
      | 1                 | 16             | 0       |
      | 1                 | 17             | 0       |
      | 1                 | 18             | 0       |
      | 1                 | 51             | 0       |
      | 1                 | 53             | 0       |
      | 1                 | 55             | 0       |
      | 1                 | 59             | 0       |
      | 1                 | 63             | 0       |
      | 1                 | 65             | 0       |
      | 1                 | 67             | 0       |
      | 3                 | 3              | 1       |
      | 3                 | 13             | 0       |
      | 3                 | 15             | 0       |
      | 3                 | 61             | 0       |
      | 3                 | 63             | 0       |
      | 3                 | 65             | 0       |
      | 3                 | 69             | 0       |
      | 11                | 11             | 1       |
      | 11                | 14             | 0       |
      | 11                | 16             | 0       |
      | 11                | 17             | 0       |
      | 11                | 18             | 0       |
      | 11                | 51             | 0       |
      | 11                | 53             | 0       |
      | 11                | 55             | 0       |
      | 11                | 59             | 0       |
      | 11                | 63             | 0       |
      | 11                | 65             | 0       |
      | 11                | 67             | 0       |
      | 12                | 12             | 1       |
      | 13                | 13             | 1       |
      | 13                | 15             | 0       |
      | 13                | 61             | 0       |
      | 13                | 63             | 0       |
      | 13                | 65             | 0       |
      | 13                | 69             | 0       |
      | 14                | 14             | 1       |
      | 14                | 51             | 0       |
      | 14                | 53             | 0       |
      | 14                | 55             | 0       |
      | 15                | 15             | 1       |
      | 15                | 61             | 0       |
      | 15                | 63             | 0       |
      | 15                | 65             | 0       |
      | 16                | 16             | 1       |
      | 16                | 63             | 0       |
      | 16                | 65             | 0       |
      | 16                | 67             | 0       |
      | 20                | 20             | 1       |
      | 20                | 21             | 0       |
      | 21                | 21             | 1       |
    And the database has the following table 'items':
      | id  | type    | default_language_tag |
      | 200 | Chapter | fr                   |
      | 210 | Chapter | fr                   |
      | 211 | Task    | fr                   |
      | 212 | Task    | fr                   |
      | 213 | Task    | fr                   |
      | 214 | Task    | fr                   |
      | 215 | Task    | fr                   |
      | 216 | Task    | fr                   |
      | 217 | Task    | fr                   |
      | 218 | Task    | fr                   |
      | 219 | Task    | fr                   |
      | 220 | Chapter | fr                   |
      | 221 | Task    | fr                   |
      | 222 | Task    | fr                   |
      | 223 | Task    | fr                   |
      | 224 | Task    | fr                   |
      | 225 | Task    | fr                   |
      | 226 | Task    | fr                   |
      | 227 | Task    | fr                   |
      | 228 | Task    | fr                   |
      | 229 | Task    | fr                   |
      | 300 | Chapter | fr                   |
      | 310 | Chapter | fr                   |
      | 311 | Task    | fr                   |
      | 312 | Task    | fr                   |
      | 313 | Task    | fr                   |
      | 314 | Task    | fr                   |
      | 315 | Task    | fr                   |
      | 316 | Task    | fr                   |
      | 317 | Task    | fr                   |
      | 318 | Task    | fr                   |
      | 319 | Task    | fr                   |
      | 400 | Course  | fr                   |
      | 410 | Chapter | fr                   |
      | 411 | Task    | fr                   |
      | 412 | Task    | fr                   |
      | 413 | Task    | fr                   |
      | 414 | Task    | fr                   |
      | 415 | Task    | fr                   |
      | 416 | Task    | fr                   |
      | 417 | Task    | fr                   |
      | 418 | Task    | fr                   |
      | 419 | Task    | fr                   |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 200            | 210           | 0           |
      | 200            | 220           | 1           |
      | 210            | 211           | 0           |
      | 210            | 212           | 1           |
      | 210            | 213           | 2           |
      | 210            | 214           | 3           |
      | 210            | 215           | 4           |
      | 210            | 216           | 5           |
      | 210            | 217           | 6           |
      | 210            | 218           | 7           |
      | 210            | 219           | 8           |
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
      | 310            | 311           | 1           |
      | 310            | 312           | 2           |
      | 310            | 313           | 3           |
      | 310            | 314           | 4           |
      | 310            | 315           | 5           |
      | 310            | 316           | 6           |
      | 310            | 317           | 7           |
      | 310            | 318           | 8           |
      | 310            | 319           | 9           |
      | 400            | 410           | 0           |
      | 410            | 411           | 1           |
      | 410            | 412           | 2           |
      | 410            | 413           | 3           |
      | 410            | 414           | 4           |
      | 410            | 415           | 5           |
      | 410            | 416           | 6           |
      | 410            | 417           | 7           |
      | 410            | 418           | 8           |
      | 410            | 419           | 9           |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 21       | 211     | info                     |
      | 20       | 212     | content                  |
      | 21       | 213     | content_with_descendants |
      | 20       | 214     | info                     |
      | 21       | 215     | content                  |
      | 20       | 216     | none                     |
      | 21       | 217     | none                     |
      | 20       | 218     | none                     |
      | 21       | 219     | none                     |
      | 20       | 221     | info                     |
      | 21       | 222     | content                  |
      | 20       | 223     | content_with_descendants |
      | 21       | 224     | info                     |
      | 20       | 225     | content                  |
      | 21       | 226     | none                     |
      | 20       | 227     | none                     |
      | 21       | 228     | none                     |
      | 20       | 229     | none                     |
      | 21       | 311     | info                     |
      | 20       | 312     | content                  |
      | 21       | 313     | content_with_descendants |
      | 20       | 314     | info                     |
      | 21       | 315     | content                  |
      | 20       | 316     | none                     |
      | 21       | 317     | none                     |
      | 20       | 318     | none                     |
      | 21       | 319     | none                     |
      | 20       | 411     | info                     |
      | 21       | 412     | content                  |
      | 20       | 413     | content_with_descendants |
      | 21       | 414     | info                     |
      | 20       | 415     | content                  |
      | 21       | 416     | none                     |
      | 20       | 417     | none                     |
      | 21       | 418     | none                     |
      | 20       | 419     | none                     |
    And the database has the following table 'answers':
      | id | created_at          | author_id | attempt_id |
      | 1  | 2017-05-29 06:38:38 | 51        | 1          |
      | 2  | 2017-05-29 06:38:38 | 53        | 2          |
      | 3  | 2017-05-29 06:38:38 | 55        | 3          |
      | 4  | 2017-05-30 06:38:38 | 57        | 4          |
    And the database has the following table 'attempts':
      | group_id | item_id | order | started_at          | score_computed | score_obtained_at   | hints_cached | submissions | validated_at        | latest_activity_at  |
      | 14       | 211     | 0     | 2017-05-29 06:38:38 | 0              | 2017-05-29 06:38:38 | 100          | 100         | null                | 2018-05-30 06:38:38 |
      | 14       | 211     | 1     | 2017-05-29 06:38:38 | 40             | 2017-05-29 06:38:38 | 2            | 3           | null                | null                |
      | 14       | 211     | 2     | 2017-05-29 06:38:38 | 50             | 2017-05-29 06:38:38 | 3            | 4           | 2017-05-30 06:38:38 | null                | # hints_cached & submissions for 14,211 come from this line
      | 14       | 211     | 3     | 2017-05-29 06:38:38 | 50             | 2017-05-30 06:38:38 | 10           | 20          | 2017-05-30 06:38:38 | null                |
      | 15       | 211     | 0     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                | null                |
      | 15       | 212     | 0     | 2017-03-29 06:38:38 | 0              | null                | 0            | 0           | null                | null                |
      | 16       | 212     | 0     | 2019-01-01 00:00:00 | 10             | null                | 0            | 0           | null                | 2019-06-01 00:00:00 |
      | 14       | 211     | 4     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                | null                |
      | 15       | 211     | 1     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                | null                |
      | 15       | 212     | 1     | 2017-03-29 06:38:38 | 100            | null                | 0            | 0           | null                | null                |
      | 14       | 211     | 5     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                | null                |

  Scenario: Get progress of teams
    Given I am the user with id "21"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30 20:19:05"
    When I send a GET request to "/groups/1/team-progress?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "16",
        "item_id": "211",
        "latest_activity_at": null,
        "score": 0,
        "hints_requested": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": "2019-06-01T00:00:00Z",
        "score": 10,
        "submissions": 0,
        "time_spent": 15625145,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "221",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "222",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "223",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "224",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "225",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "311",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "312",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "313",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "314",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "315",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },


      {
        "group_id": "14",
        "item_id": "211",
        "latest_activity_at": "2018-05-30T06:38:38Z",
        "score": 50,
        "hints_requested": 3,
        "submissions": 4,
        "time_spent": 86400,
        "validated": true
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "221",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "222",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "223",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "224",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "225",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "311",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "312",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "313",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "314",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "315",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """

  Scenario: Get progress of the first team
    Given I am the user with id "21"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30 20:19:05"
    When I send a GET request to "/groups/1/team-progress?parent_item_ids=210,220,310&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "16",
        "item_id": "211",
        "latest_activity_at": null,
        "score": 0,
        "hints_requested": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": "2019-06-01T00:00:00Z",
        "score": 10,
        "submissions": 0,
        "time_spent": 15625145,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "221",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "222",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "223",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "224",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "225",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "311",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "312",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "313",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "314",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "16",
        "hints_requested": 0,
        "item_id": "315",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """

  Scenario: Get progress of teams skipping the first row
    Given I am the user with id "21"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30 20:19:05"
    When I send a GET request to "/groups/1/team-progress?parent_item_ids=210,220,310&from.name=First%20Team&from.id=16"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "14",
        "item_id": "211",
        "latest_activity_at": "2018-05-30T06:38:38Z",
        "score": 50,
        "hints_requested": 3,
        "submissions": 4,
        "time_spent": 86400,
        "validated": true
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "221",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "222",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "223",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "224",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "225",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "311",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "312",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "313",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "314",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "14",
        "hints_requested": 0,
        "item_id": "315",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """

  Scenario: No teams
    Given I am the user with id "21"
    When I send a GET request to "/groups/16/team-progress?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
