Feature: Display the current progress of users on a subset of items (groupUserProgress)
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
      | group_id | manager_id | can_watch_members |
      | 1        | 21         | true              |
      | 19       | 4          | true              |
      | 51       | 4          | true              |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 1               | 11             |
      | 1               | 67             |
      | 3               | 13             |
      | 4               | 21             |
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
      | 300 | Course  | fr                   |
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
      | 400 | Chapter | fr                   |
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
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       | can_watch_generated |
      | 21       | 210     | none                     | result              |
      | 21       | 211     | info                     | none                |
      | 20       | 212     | content                  | none                |
      | 21       | 213     | content_with_descendants | none                |
      | 20       | 214     | info                     | none                |
      | 21       | 215     | content                  | none                |
      | 20       | 216     | none                     | none                |
      | 21       | 217     | none                     | none                |
      | 20       | 218     | none                     | none                |
      | 21       | 219     | none                     | none                |
      | 20       | 220     | none                     | answer              |
      | 20       | 221     | info                     | none                |
      | 21       | 222     | content                  | none                |
      | 20       | 223     | content_with_descendants | none                |
      | 21       | 224     | info                     | none                |
      | 20       | 225     | content                  | none                |
      | 21       | 226     | none                     | none                |
      | 20       | 227     | none                     | none                |
      | 21       | 228     | none                     | none                |
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
      | 4        | 1010    | none                     | answer_with_grant   |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          |
      | 0  | 14             | 2017-05-29 06:38:38 |
      | 1  | 14             | 2017-05-29 06:38:38 |
      | 2  | 14             | 2017-05-29 06:38:38 |
      | 3  | 14             | 2017-05-29 06:38:38 |
      | 0  | 15             | 2017-03-29 06:38:38 |
      | 0  | 16             | 2018-12-01 00:00:00 |
      | 1  | 67             | 2019-01-01 00:00:00 |
      | 0  | 67             | 2017-05-29 06:38:38 |
      | 4  | 14             | 2017-05-29 06:38:38 |
      | 1  | 15             | 2017-03-29 06:38:38 |
      | 5  | 14             | 2017-05-29 06:38:38 |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          | score_computed | score_obtained_at   | hints_cached | submissions | validated_at        | latest_activity_at  |
      | 0          | 14             | 210     | 2017-05-29 05:38:38 | 50             | 2017-05-29 06:38:38 | 115          | 127         | null                | 2019-05-29 06:38:38 |
      | 0          | 14             | 211     | 2017-05-29 06:38:38 | 0              | 2017-05-29 06:38:38 | 100          | 100         | 2017-05-30 06:38:38 | 2018-05-30 06:38:38 | # latest_activity_at for 51, 211 comes from this line (the last activity is made by a team)
      | 1          | 14             | 211     | 2017-05-29 06:38:38 | 40             | 2017-05-29 06:38:38 | 2            | 3           | 2017-05-29 06:38:58 | 2018-05-29 06:38:38 | # min(validated_at) for 51, 211 comes from this line (from a team)
      | 2          | 14             | 211     | 2017-05-29 06:38:38 | 50             | 2017-05-29 06:38:38 | 3            | 4           | 2017-05-31 06:58:38 | 2018-05-28 06:38:38 | # hints_cached & submissions for 51, 211 come from this line (the best attempt is made by a team)
      | 3          | 14             | 211     | 2017-05-29 06:38:38 | 50             | 2017-05-30 06:38:38 | 10           | 20          | null                | 2018-05-27 06:38:38 |
      | 0          | 15             | 211     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                | 2018-05-26 06:38:38 |
      | 0          | 15             | 212     | 2017-03-29 06:38:38 | 0              | null                | 0            | 0           | null                | 2018-05-25 06:38:38 |
      | 0          | 16             | 212     | 2018-12-01 00:00:00 | 10             | 2017-05-30 06:38:38 | 0            | 0           | null                | 2019-06-01 00:00:00 | # started_at for 67, 212 & 63, 212 comes from this line (the first attempt is started by a team)
      | 0          | 67             | 212     | 2019-01-01 00:00:00 | 20             | 2017-06-30 06:38:38 | 1            | 2           | null                | 2019-06-01 00:00:00 | # hints_cached & submissions for 67, 212 come from this line (the best attempt is made by a user)
      | 1          | 67             | 212     | 2019-01-01 00:00:00 | 10             | 2017-05-30 06:38:38 | 6            | 7           | null                | 2019-07-01 00:00:00 | # latest_activity_at for 67, 212 comes from this line (the last activity is made by a user)
      | 0          | 67             | 213     | 2018-11-01 00:00:00 | 0              | null                | 0            | 0           | null                | 2018-11-01 00:00:00 | # started_at for 67, 213 comes from this line (the first attempt is started by a user)
      | 0          | 67             | 214     | 2017-05-29 06:38:38 | 15             | 2017-05-29 06:38:48 | 10           | 11          | 2017-05-29 06:38:48 | 2017-05-30 06:38:48 | # min(validated_at) for 67, 214 comes from this line (from a user)
      | 0          | 67             | 215     | 3018-11-01 00:00:00 | 0              | null                | 0            | 0           | null                | 2018-11-01 00:00:00 | # started_at for 67, 213 comes from this line (the first attempt is started by a user)
      | 4          | 14             | 211     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                | 2018-05-24 06:38:38 |
      | 1          | 15             | 211     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                | 2018-05-23 06:38:38 |
      | 1          | 15             | 212     | 2017-03-29 06:38:38 | 100            | null                | 0            | 0           | null                | 2018-05-22 06:38:38 |
      | 5          | 14             | 211     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                | 2018-05-21 06:38:38 |

  Scenario: Get progress of the second and the third users (checks sorting, from.*, and limit)
    Given I am the user with id "21"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30 20:19:05"
    When I send a GET request to "/groups/1/user-progress?parent_item_ids=210&from.id=65&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "67",
        "hints_requested": 0,
        "item_id": "210",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "67",
        "item_id": "211",
        "latest_activity_at": null,
        "score": 0,
        "hints_requested": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "67",
        "hints_requested": 1,
        "item_id": "212",
        "latest_activity_at": "2019-07-01T00:00:00Z",
        "score": 20,
        "submissions": 2,
        "time_spent": 18303545,
        "validated": false
      },
      {
        "group_id": "67",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": "2018-11-01T00:00:00Z",
        "score": 0,
        "submissions": 0,
        "time_spent": 20895545,
        "validated": false
      },
      {
        "group_id": "67",
        "hints_requested": 10,
        "item_id": "214",
        "latest_activity_at": "2017-05-30T06:38:48Z",
        "score": 15,
        "submissions": 11,
        "time_spent": 10,
        "validated": true
      },
      {
        "group_id": "67",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": "2018-11-01T00:00:00Z",
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },

      {
        "group_id": "51",
        "hints_requested": 115,
        "item_id": "210",
        "latest_activity_at": "2019-05-29T06:38:38Z",
        "score": 50,
        "submissions": 127,
        "time_spent": 65889627,
        "validated": false
      },
      {
        "group_id": "51",
        "item_id": "211",
        "latest_activity_at": "2018-05-30T06:38:38Z",
        "score": 50,
        "hints_requested": 3,
        "submissions": 4,
        "time_spent": 20,
        "validated": true
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """


  Scenario: Get progress of the first user for all the visible items (also checks the limit)
    Given I am the user with id "21"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30 20:19:05"
    When I send a GET request to "/groups/1/user-progress?parent_item_ids=210,220,310&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "210",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "item_id": "211",
        "latest_activity_at": null,
        "score": 0,
        "hints_requested": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": "2019-06-01T00:00:00Z",
        "score": 10,
        "submissions": 0,
        "time_spent": 18303545,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "220",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "221",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "222",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "223",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "224",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "225",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "310",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "311",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "312",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "313",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "314",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
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

  Scenario: No users
    Given I am the user with id "21"
    When I send a GET request to "/groups/17/user-progress?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: No visible child items
    Given I am the user with id "21"
    When I send a GET request to "/groups/1/user-progress?parent_item_ids=1010"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "1010",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "65",
        "hints_requested": 0,
        "item_id": "1010",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "67",
        "hints_requested": 0,
        "item_id": "1010",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "1010",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "53",
        "hints_requested": 0,
        "item_id": "1010",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "55",
        "hints_requested": 0,
        "item_id": "1010",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "59",
        "hints_requested": 0,
        "item_id": "1010",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """

  Scenario: The input group_id is a user
    Given I am the user with id "21"
    When I send a GET request to "/groups/51/user-progress?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: The input group_id is a team
    Given I am the user with id "21"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30 20:19:05"
    When I send a GET request to "/groups/14/user-progress?parent_item_ids=210"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "51",
        "hints_requested": 115,
        "item_id": "210",
        "latest_activity_at": "2019-05-29T06:38:38Z",
        "score": 50,
        "submissions": 127,
        "time_spent": 65889627,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 3,
        "item_id": "211",
        "latest_activity_at": "2018-05-30T06:38:38Z",
        "score": 50,
        "submissions": 4,
        "time_spent": 20,
        "validated": true
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },

      {
        "group_id": "53",
        "hints_requested": 115,
        "item_id": "210",
        "latest_activity_at": "2019-05-29T06:38:38Z",
        "score": 50,
        "submissions": 127,
        "time_spent": 65889627,
        "validated": false
      },
      {
        "group_id": "53",
        "hints_requested": 3,
        "item_id": "211",
        "latest_activity_at": "2018-05-30T06:38:38Z",
        "score": 50,
        "submissions": 4,
        "time_spent": 20,
        "validated": true
      },
      {
        "group_id": "53",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "53",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "53",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "53",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },

      {
        "group_id": "55",
        "hints_requested": 115,
        "item_id": "210",
        "latest_activity_at": "2019-05-29T06:38:38Z",
        "score": 50,
        "submissions": 127,
        "time_spent": 65889627,
        "validated": false
      },
      {
        "group_id": "55",
        "hints_requested": 3,
        "item_id": "211",
        "latest_activity_at": "2018-05-30T06:38:38Z",
        "score": 50,
        "submissions": 4,
        "time_spent": 20,
        "validated": true
      },
      {
        "group_id": "55",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "55",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "55",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "55",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """

  Scenario: Users are direct members of the input group_id which is not a team
    Given I am the user with id "21"
    When I send a GET request to "/groups/19/user-progress?parent_item_ids=210"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "69",
        "hints_requested": 0,
        "item_id": "210",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "69",
        "item_id": "211",
        "latest_activity_at": null,
        "score": 0,
        "hints_requested": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "69",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "69",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "69",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "69",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """

  Scenario: No parent item ids given
    Given I am the user with id "21"
    When I send a GET request to "/groups/1/user-progress?parent_item_ids="
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
