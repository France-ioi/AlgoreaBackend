Feature: Display the current progress of a group on a subset of items (groupGroupProgress)
  Background:
    Given the database has the following table 'groups':
      | id | type      | name           |
      | 1  | Base      | Root 1         |
      | 3  | Base      | Root 2         |
      | 11 | Class     | Our Class      |
      | 12 | Class     | Other Class    |
      | 13 | Class     | Special Class  |
      | 14 | Team      | Super Team     |
      | 15 | Team      | Our Team       |
      | 16 | Team      | First Team     |
      | 17 | Other     | A custom group |
      | 18 | Club      | Our Club       |
      | 20 | Friends   | My Friends     |
      | 21 | UserSelf  | owner          |
      | 51 | UserSelf  | johna          |
      | 53 | UserSelf  | johnb          |
      | 55 | UserSelf  | johnc          |
      | 57 | UserSelf  | johnd          |
      | 59 | UserSelf  | johne          |
      | 61 | UserSelf  | janea          |
      | 63 | UserSelf  | janeb          |
      | 65 | UserSelf  | janec          |
      | 67 | UserSelf  | janed          |
      | 69 | UserSelf  | janee          |
      | 22 | UserAdmin | owner-admin    |
      | 52 | UserAdmin | johna-admin    |
      | 54 | UserAdmin | johnb-admin    |
      | 56 | UserAdmin | johnc-admin    |
      | 58 | UserAdmin | johnd-admin    |
      | 60 | UserAdmin | johne-admin    |
      | 62 | UserAdmin | janea-admin    |
      | 64 | UserAdmin | janeb-admin    |
      | 66 | UserAdmin | janec-admin    |
      | 68 | UserAdmin | janed-admin    |
      | 70 | UserAdmin | janee-admin    |
    And the database has the following table 'users':
      | login | group_id | owned_group_id |
      | owner | 21       | 22             |
      | johna | 51       | 52             |
      | johnb | 53       | 54             |
      | johnc | 55       | 56             |
      | johnd | 57       | 58             |
      | johne | 59       | 60             |
      | janea | 61       | 62             |
      | janeb | 63       | 64             |
      | janec | 65       | 66             |
      | janed | 67       | 68             |
      | janee | 69       | 70             |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type               |
      | 1               | 11             | direct             |
      | 1               | 14             | direct             | # direct child of group_id with type = 'Team' (ignored)
      | 1               | 17             | direct             |
      | 1               | 51             | direct             | # direct child of group_id with type = 'UserSelf' (ignored)
      | 3               | 13             | direct             |
      | 11              | 14             | direct             |
      | 11              | 17             | direct             |
      | 11              | 18             | direct             |
      | 11              | 59             | requestAccepted    |
      | 13              | 15             | direct             |
      | 13              | 69             | invitationAccepted |
      | 14              | 51             | requestAccepted    |
      | 14              | 53             | joinedByCode       |
      | 14              | 55             | invitationAccepted |
      | 15              | 57             | direct             |
      | 15              | 59             | requestAccepted    |
      | 15              | 61             | joinedByCode       |
      | 15              | 63             | invitationRefused  |
      | 15              | 65             | left               |
      | 15              | 67             | invitationSent     |
      | 15              | 69             | requestSent        |
      | 16              | 51             | invitationRefused  |
      | 16              | 53             | requestRefused     |
      | 16              | 55             | removed            |
      | 16              | 63             | direct             |
      | 16              | 65             | requestAccepted    |
      | 16              | 67             | invitationAccepted |
      | 17              | 14             | direct             |
      | 17              | 18             | direct             |
      | 17              | 59             | requestAccepted    |
      | 20              | 21             | direct             |
      | 22              | 1              | direct             |
      | 22              | 3              | direct             |
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
      | 22                | 1              | 0       |
      | 22                | 3              | 0       |
      | 22                | 11             | 0       |
      | 22                | 12             | 0       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 15             | 0       |
      | 22                | 17             | 0       |
      | 22                | 18             | 0       |
      | 22                | 22             | 1       |
      | 22                | 51             | 0       |
      | 22                | 53             | 0       |
      | 22                | 55             | 0       |
      | 22                | 59             | 0       |
      | 22                | 61             | 0       |
      | 22                | 63             | 0       |
      | 22                | 65             | 0       |
      | 22                | 69             | 0       |
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
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_full_access_since | cached_partial_access_since | cached_grayed_access_since |
      | 21       | 211     | null                     | null                        | 2017-05-29 06:38:38        |
      | 20       | 212     | null                     | 2017-05-29 06:38:38         | null                       |
      | 21       | 213     | 2017-05-29 06:38:38      | null                        | null                       |
      | 20       | 214     | null                     | null                        | 2017-05-29 06:38:38        |
      | 21       | 215     | null                     | 2017-05-29 06:38:38         | null                       |
      | 20       | 216     | null                     | null                        | 2037-05-29 06:38:38        |
      | 21       | 217     | null                     | 2037-05-29 06:38:38         | null                       |
      | 20       | 218     | 2037-05-29 06:38:38      | null                        | null                       |
      | 21       | 219     | null                     | null                        | 2037-05-29 06:38:38        |
      | 20       | 221     | null                     | null                        | 2017-05-29 06:38:38        |
      | 21       | 222     | null                     | 2017-05-29 06:38:38         | null                       |
      | 20       | 223     | 2017-05-29 06:38:38      | null                        | null                       |
      | 21       | 224     | null                     | null                        | 2017-05-29 06:38:38        |
      | 20       | 225     | null                     | 2017-05-29 06:38:38         | null                       |
      | 21       | 226     | null                     | null                        | 2037-05-29 06:38:38        |
      | 20       | 227     | null                     | 2037-05-29 06:38:38         | null                       |
      | 21       | 228     | 2037-05-29 06:38:38      | null                        | null                       |
      | 20       | 229     | null                     | null                        | 2037-05-29 06:38:38        |
      | 21       | 311     | null                     | null                        | 2017-05-29 06:38:38        |
      | 20       | 312     | null                     | 2017-05-29 06:38:38         | null                       |
      | 21       | 313     | 2017-05-29 06:38:38      | null                        | null                       |
      | 20       | 314     | null                     | null                        | 2017-05-29 06:38:38        |
      | 21       | 315     | null                     | 2017-05-29 06:38:38         | null                       |
      | 20       | 316     | null                     | null                        | 2037-05-29 06:38:38        |
      | 21       | 317     | null                     | 2037-05-29 06:38:38         | null                       |
      | 20       | 318     | 2037-05-29 06:38:38      | null                        | null                       |
      | 21       | 319     | null                     | null                        | 2037-05-29 06:38:38        |
      | 20       | 411     | null                     | null                        | 2017-05-29 06:38:38        |
      | 21       | 412     | null                     | 2017-05-29 06:38:38         | null                       |
      | 20       | 413     | 2017-05-29 06:38:38      | null                        | null                       |
      | 21       | 414     | null                     | null                        | 2017-05-29 06:38:38        |
      | 20       | 415     | null                     | 2017-05-29 06:38:38         | null                       |
      | 21       | 416     | null                     | null                        | 2037-05-29 06:38:38        |
      | 20       | 417     | null                     | 2037-05-29 06:38:38         | null                       |
      | 21       | 418     | 2037-05-29 06:38:38      | null                        | null                       |
      | 20       | 419     | null                     | null                        | 2037-05-29 06:38:38        |
    And the database has the following table 'groups_attempts':
      | group_id | item_id | order | started_at          | score | best_answer_at      | hints_cached | submissions_attempts | validated | validated_at        |
      | 14       | 211     | 0     | 2017-05-29 06:38:38 | 0     | 2017-05-29 06:38:38 | 100          | 100                  | 0         | 2017-05-30 06:38:38 |
      | 14       | 211     | 1     | 2017-05-29 06:38:38 | 40    | 2017-05-29 06:38:38 | 2            | 3                    | 0         | null                |
      | 14       | 211     | 2     | 2017-05-29 06:38:38 | 50    | 2017-05-29 06:38:38 | 3            | 4                    | 1         | null                | # hints_cached & submissions_attempts for 14,211 come from this line
      | 14       | 211     | 3     | 2017-05-29 06:38:38 | 50    | 2017-05-30 06:38:38 | 10           | 20                   | 0         | null                |
      | 15       | 211     | 0     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 212     | 0     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 59       | 212     | 0     | 2019-01-01 00:00:00 | 10    | null                | 0            | 0                    | 0         | null                |
      | 14       | 211     | 4     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 211     | 1     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 212     | 1     | 2017-03-29 06:38:38 | 100   | null                | 0            | 0                    | 1         | null                |
      | 14       | 211     | 5     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 211     | 2     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 212     | 2     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 14       | 211     | 6     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 211     | 3     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 212     | 3     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 14       | 211     | 7     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 211     | 4     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 212     | 4     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 14       | 211     | 8     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 211     | 5     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 212     | 5     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 14       | 211     | 9     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 211     | 6     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |
      | 15       | 212     | 6     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                |

  Scenario: Get progress of groups
    Given I am the user with group_id "21"
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
        "avg_submissions_attempts": 2,
        "avg_time_spent": 43200,
        "group_id": "17",
        "item_id": "211",
        "validation_rate": 0.5
      },
      {
        "average_score": 5,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 6473372.5,
        "group_id": "17",
        "item_id": "212",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "213",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "214",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "215",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "221",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "222",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "223",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "224",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "225",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "311",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "312",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "313",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "314",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "315",
        "validation_rate": 0
      },


      {
        "average_score": 25,
        "avg_hints_requested": 1.5,
        "avg_submissions_attempts": 2,
        "avg_time_spent": 43200,
        "group_id": "11",
        "item_id": "211",
        "validation_rate": 0.5
      },
      {
        "average_score": 5,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 6473372.5,
        "group_id": "11",
        "item_id": "212",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "213",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "214",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "215",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "221",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "222",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "223",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "224",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "225",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "311",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "312",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "313",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "314",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "315",
        "validation_rate": 0
      }
    ]
    """

  Scenario: Get progress of the first group
    Given I am the user with group_id "21"
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
        "avg_submissions_attempts": 2,
        "avg_time_spent": 43200,
        "group_id": "17",
        "item_id": "211",
        "validation_rate": 0.5
      },
      {
        "average_score": 5,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 6473372.5,
        "group_id": "17",
        "item_id": "212",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "213",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "214",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "215",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "221",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "222",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "223",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "224",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "225",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "311",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "312",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "313",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "314",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "17",
        "item_id": "315",
        "validation_rate": 0
      }
    ]
    """

  Scenario: Get progress of groups skipping the first row
    Given I am the user with group_id "21"
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
        "avg_submissions_attempts": 2,
        "avg_time_spent": 43200,
        "group_id": "11",
        "item_id": "211",
        "validation_rate": 0.5
      },
      {
        "average_score": 5,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 6473372.5,
        "group_id": "11",
        "item_id": "212",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "213",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "214",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "215",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "221",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "222",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "223",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "224",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "225",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "311",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "312",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "313",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "314",
        "validation_rate": 0
      },
      {
        "average_score": 0,
        "avg_hints_requested": 0,
        "avg_submissions_attempts": 0,
        "avg_time_spent": 0,
        "group_id": "11",
        "item_id": "315",
        "validation_rate": 0
      }
    ]
    """

  Scenario: No visible items
    Given I am the user with group_id "21"
    When I send a GET request to "/groups/1/group-progress?parent_item_ids=1010"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: No groups
    Given I am the user with group_id "21"
    # here we fixate avg_time_spent even if it depends on NOW()
    And the DB time now is "2019-05-30 20:19:05"
    When I send a GET request to "/groups/13/group-progress?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
