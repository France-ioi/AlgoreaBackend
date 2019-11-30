Feature: Display the current progress of a group on a subset of items (groupGroupProgress)
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
      | 3        | 21         |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 1               | 11             |
      | 1               | 14             | # direct child of group_id with type = 'Team' (ignored)
      | 1               | 17             |
      | 1               | 51             | # direct child of group_id with type = 'UserSelf' (ignored)
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
      | ancestor_group_id | child_group_id | is_self |
      | 1                 | 1              | 1       |
      | 1                 | 11             | 0       |
      | 1                 | 12             | 0       |
      | 1                 | 14             | 0       |
      | 1                 | 17             | 0       |
      | 1                 | 18             | 0       |
      | 1                 | 51             | 0       |
      | 1                 | 53             | 0       |
      | 1                 | 55             | 0       |
      | 1                 | 59             | 0       |
      | 3                 | 3              | 1       |
      | 3                 | 13             | 0       |
      | 3                 | 15             | 0       |
      | 3                 | 61             | 0       |
      | 3                 | 63             | 0       |
      | 3                 | 65             | 0       |
      | 3                 | 69             | 0       |
      | 11                | 11             | 1       |
      | 11                | 14             | 0       |
      | 11                | 17             | 0       |
      | 11                | 18             | 0       |
      | 11                | 51             | 0       |
      | 11                | 53             | 0       |
      | 11                | 55             | 0       |
      | 11                | 59             | 0       |
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
      | 17                | 14             | 0       |
      | 17                | 17             | 1       |
      | 17                | 18             | 0       |
      | 17                | 51             | 0       |
      | 17                | 53             | 0       |
      | 17                | 55             | 0       |
      | 17                | 59             | 0       |
      | 20                | 20             | 1       |
      | 20                | 21             | 0       |
      | 21                | 21             | 1       |
    And the database has the following table 'items':
      | id  | type     |
      | 200 | Category |
      | 210 | Chapter  |
      | 211 | Task     |
      | 212 | Task     |
      | 213 | Task     |
      | 214 | Task     |
      | 215 | Task     |
      | 216 | Task     |
      | 217 | Task     |
      | 218 | Task     |
      | 219 | Task     |
      | 220 | Chapter  |
      | 221 | Task     |
      | 222 | Task     |
      | 223 | Task     |
      | 224 | Task     |
      | 225 | Task     |
      | 226 | Task     |
      | 227 | Task     |
      | 228 | Task     |
      | 229 | Task     |
      | 300 | Category |
      | 310 | Chapter  |
      | 311 | Task     |
      | 312 | Task     |
      | 313 | Task     |
      | 314 | Task     |
      | 315 | Task     |
      | 316 | Task     |
      | 317 | Task     |
      | 318 | Task     |
      | 319 | Task     |
      | 400 | Category |
      | 410 | Chapter  |
      | 411 | Task     |
      | 412 | Task     |
      | 413 | Task     |
      | 414 | Task     |
      | 415 | Task     |
      | 416 | Task     |
      | 417 | Task     |
      | 418 | Task     |
      | 419 | Task     |
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
    And the database has the following table 'groups_attempts':
      | group_id | item_id | order | started_at          | score | best_answer_at      | hints_cached | submissions | validated_at        |
      | 14       | 211     | 0     | 2017-05-29 06:38:38 | 0     | 2017-05-29 06:38:38 | 100          | 100         | null                |
      | 14       | 211     | 1     | 2017-05-29 06:38:38 | 40    | 2017-05-29 06:38:38 | 2            | 3           | null                |
      | 14       | 211     | 2     | 2017-05-29 06:38:38 | 50    | 2017-05-29 06:38:38 | 3            | 4           | 2017-05-30 06:38:38 | # hints_cached & submissions for 14,211 come from this line
      | 14       | 211     | 3     | 2017-05-29 06:38:38 | 50    | 2017-05-30 06:38:38 | 10           | 20          | null                |
      | 15       | 211     | 0     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 212     | 0     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 59       | 212     | 0     | 2019-01-01 00:00:00 | 10    | null                | 0            | 0           | null                |
      | 14       | 211     | 4     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 211     | 1     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 212     | 1     | 2017-03-29 06:38:38 | 100   | null                | 0            | 0           | 2017-05-30 06:38:38 |
      | 14       | 211     | 5     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 211     | 2     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 212     | 2     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 14       | 211     | 6     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 211     | 3     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 212     | 3     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 14       | 211     | 7     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 211     | 4     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 212     | 4     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 14       | 211     | 8     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 211     | 5     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 212     | 5     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 14       | 211     | 9     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 211     | 6     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0           | null                |
      | 15       | 212     | 6     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0           | null                |

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
