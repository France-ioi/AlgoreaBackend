Feature: Display the current progress of a group on a subset of items (groupGroupProgress)
  Background:
    Given the database has the following table 'groups':
      | id | type    | name           |
      | 1  | Base    | Root 1         |
      | 3  | Base    | Root 2         |
      | 11 | Class   | Our Class      |
      | 12 | Class   | Other Class    |
      | 13 | Class   | Special Class  |
      | 14 | Team    | Super Team     |
      | 15 | Team    | Our Team       |
      | 16 | Team    | First Team     |
      | 17 | Other   | A custom group |
      | 18 | Club    | Our Club       |
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
      | group_id | manager_id |
      | 1        | 21         |
      | 3        | 21         |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 1               | 11             |
      | 1               | 14             | # direct child of group_id with type = 'Team' (ignored)
      | 1               | 17             |
      | 1               | 51             | # direct child of group_id with type = 'User' (ignored)
      | 3               | 13             |
      | 11              | 14             |
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
      | 17              | 14             |
      | 17              | 18             |
      | 17              | 59             |
      | 20              | 21             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 1                 | 1              |
      | 1                 | 11             |
      | 1                 | 12             |
      | 1                 | 14             |
      | 1                 | 17             |
      | 1                 | 18             |
      | 1                 | 51             |
      | 1                 | 53             |
      | 1                 | 55             |
      | 1                 | 59             |
      | 3                 | 3              |
      | 3                 | 13             |
      | 3                 | 15             |
      | 3                 | 61             |
      | 3                 | 63             |
      | 3                 | 65             |
      | 3                 | 69             |
      | 11                | 11             |
      | 11                | 14             |
      | 11                | 17             |
      | 11                | 18             |
      | 11                | 51             |
      | 11                | 53             |
      | 11                | 55             |
      | 11                | 59             |
      | 12                | 12             |
      | 13                | 13             |
      | 13                | 15             |
      | 13                | 61             |
      | 13                | 63             |
      | 13                | 65             |
      | 13                | 69             |
      | 14                | 14             |
      | 14                | 51             |
      | 14                | 53             |
      | 14                | 55             |
      | 15                | 15             |
      | 15                | 61             |
      | 15                | 63             |
      | 15                | 65             |
      | 16                | 16             |
      | 16                | 63             |
      | 16                | 65             |
      | 16                | 67             |
      | 17                | 14             |
      | 17                | 17             |
      | 17                | 18             |
      | 17                | 51             |
      | 17                | 53             |
      | 17                | 55             |
      | 17                | 59             |
      | 20                | 20             |
      | 20                | 21             |
      | 21                | 21             |
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
    And the database has the following table 'attempts':
      | id | participant_id | created_at          |
      | 0  | 14             | 2017-05-29 06:38:38 |
      | 1  | 14             | 2017-05-29 06:38:38 |
      | 2  | 14             | 2017-05-29 06:38:38 |
      | 3  | 14             | 2017-05-29 06:38:38 |
      | 0  | 15             | 2017-04-29 06:38:38 |
      | 0  | 59             | 2019-01-01 00:00:00 |
      | 4  | 14             | 2017-05-29 06:38:38 |
      | 1  | 15             | 2017-03-29 06:38:38 |
      | 5  | 14             | 2017-05-29 06:38:38 |
      | 2  | 15             | 2017-03-29 06:38:38 |
      | 6  | 14             | 2017-05-29 06:38:38 |
      | 3  | 15             | 2017-03-29 06:38:38 |
      | 7  | 14             | 2017-05-29 06:38:38 |
      | 4  | 15             | 2017-03-29 06:38:38 |
      | 8  | 14             | 2017-05-29 06:38:38 |
      | 5  | 15             | 2017-03-29 06:38:38 |
      | 9  | 14             | 2017-05-29 06:38:38 |
      | 8  | 15             | 2017-03-29 06:38:38 |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          | score_computed | score_obtained_at   | hints_cached | submissions | validated_at        |
      | 0          | 14             | 211     | 2017-05-29 06:38:38 | 0              | 2017-05-29 06:38:38 | 100          | 100         | null                |
      | 1          | 14             | 211     | 2017-05-29 06:38:38 | 40             | 2017-05-29 06:38:38 | 2            | 3           | null                |
      | 2          | 14             | 211     | 2017-05-29 06:38:38 | 50             | 2017-05-29 06:38:38 | 3            | 4           | 2017-05-30 06:38:38 | # hints_cached & submissions for 14,211 come from this line
      | 3          | 14             | 211     | 2017-05-29 06:38:38 | 50             | 2017-05-30 06:38:38 | 10           | 20          | null                |
      | 0          | 15             | 211     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 0          | 15             | 212     | 2017-03-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 0          | 59             | 212     | 2019-01-01 00:00:00 | 10             | null                | 0            | 0           | null                |
      | 4          | 14             | 211     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 1          | 15             | 211     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 1          | 15             | 212     | 2017-03-29 06:38:38 | 100            | null                | 0            | 0           | 2017-05-30 06:38:38 |
      | 5          | 14             | 211     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 2          | 15             | 211     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 2          | 15             | 212     | 2017-03-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 6          | 14             | 211     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 3          | 15             | 211     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 3          | 15             | 212     | 2017-03-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 7          | 14             | 211     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 4          | 15             | 211     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 4          | 15             | 212     | 2017-03-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 8          | 14             | 211     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 5          | 15             | 211     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 5          | 15             | 212     | 2017-03-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 9          | 14             | 211     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 8          | 15             | 212     | 2017-03-29 06:38:38 | 0              | null                | 0            | 0           | null                |
      | 8          | 15             | 211     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                |

  Scenario: Get progress of groups
    Given I am the user with id "21"
    # here we fixate avg_time_spent even if it depends on NOW()
    And the DB time now is "2019-05-30 20:19:05"
    When I send a GET request to "/groups/1/group-progress?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "average_score": 25,
        "avg_hints_requested": 1.5,
        "avg_submissions": 2,
        "avg_time_spent": 43200,
        "group_id": "17",
        "item_id": "211",
        "validation_rate": 0.5
      },
      {
        "average_score": 5,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 6473372.5,
        "group_id": "17",
        "item_id": "212",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "213",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "214",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "215",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "221",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "222",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "223",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "224",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "225",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "311",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "312",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "313",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "314",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "315",
        "validation_rate": 0
      },


      {
        "average_score": 25,
        "avg_hints_requested": 1.5,
        "avg_submissions": 2,
        "avg_time_spent": 43200,
        "group_id": "11",
        "item_id": "211",
        "validation_rate": 0.5
      },
      {
        "average_score": 5,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 6473372.5,
        "group_id": "11",
        "item_id": "212",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "213",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "214",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "215",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "221",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "222",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "223",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "224",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "225",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "311",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "312",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "313",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "314",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "315",
        "validation_rate": 0
      }
    ]
    """

  Scenario: Get progress of the first group
    Given I am the user with id "21"
    # here we fixate avg_time_spent even if it depends on NOW()
    And the DB time now is "2019-05-30 20:19:05"
    When I send a GET request to "/groups/1/group-progress?parent_item_ids=210,220,310&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "average_score": 25,
        "avg_hints_requested": 1.5,
        "avg_submissions": 2,
        "avg_time_spent": 43200,
        "group_id": "17",
        "item_id": "211",
        "validation_rate": 0.5
      },
      {
        "average_score": 5,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 6473372.5,
        "group_id": "17",
        "item_id": "212",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "213",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "214",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "215",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "221",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "222",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "223",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "224",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "225",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "311",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "312",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "313",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "314",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "315",
        "validation_rate": 0
      }
    ]
    """

  Scenario: Get progress of groups skipping the first row
    Given I am the user with id "21"
    # here we fixate avg_time_spent even if it depends on NOW()
    And the DB time now is "2019-05-30 20:19:05"
    When I send a GET request to "/groups/1/group-progress?parent_item_ids=210,220,310&from.name=A%20custom%20group&from.id=17"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "average_score": 25,
        "avg_hints_requested": 1.5,
        "avg_submissions": 2,
        "avg_time_spent": 43200,
        "group_id": "11",
        "item_id": "211",
        "validation_rate": 0.5
      },
      {
        "average_score": 5,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 6473372.5,
        "group_id": "11",
        "item_id": "212",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "213",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "214",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "215",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "221",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "222",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "223",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "224",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "225",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "311",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "312",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "313",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "314",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "315",
        "validation_rate": 0
      }
    ]
    """

  Scenario: No visible items
    Given I am the user with id "21"
    When I send a GET request to "/groups/1/group-progress?parent_item_ids=1010"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: No groups
    Given I am the user with id "21"
    # here we fixate avg_time_spent even if it depends on NOW()
    And the DB time now is "2019-05-30 20:19:05"
    When I send a GET request to "/groups/13/group-progress?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
