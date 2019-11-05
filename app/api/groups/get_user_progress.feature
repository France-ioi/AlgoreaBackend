Feature: Display the current progress of users on a subset of items (groupUserProgress)
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
      | 3               | 13             | direct             |
      | 11              | 14             | direct             |
      | 11              | 17             | direct             |
      | 11              | 18             | direct             |
      | 11              | 59             | requestAccepted    |
      | 11              | 63             | direct             |
      | 11              | 65             | requestAccepted    |
      | 13              | 15             | direct             |
      | 13              | 16             | direct             |
      | 13              | 69             | invitationAccepted |
      | 14              | 51             | requestAccepted    |
      | 14              | 53             | joinedByCode       |
      | 14              | 55             | invitationAccepted |
      | 15              | 57             | direct             |
      | 15              | 59             | requestAccepted    |
      | 15              | 61             | invitationAccepted |
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
      | 1                 | 63             | 0       |
      | 1                 | 65             | 0       |
      | 1                 | 67             | 0       |
      | 3                 | 3              | 1       |
      | 3                 | 13             | 0       |
      | 3                 | 15             | 0       |
      | 3                 | 16             | 0       |
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
      | 11                | 63             | 0       |
      | 11                | 65             | 0       |
      | 11                | 67             | 0       |
      | 12                | 12             | 1       |
      | 13                | 13             | 1       |
      | 13                | 15             | 0       |
      | 13                | 16             | 0       |
      | 13                | 51             | 0       |
      | 13                | 53             | 0       |
      | 13                | 55             | 0       |
      | 13                | 59             | 0       |
      | 13                | 61             | 0       |
      | 13                | 63             | 0       |
      | 13                | 65             | 0       |
      | 13                | 67             | 0       |
      | 13                | 69             | 0       |
      | 14                | 14             | 1       |
      | 14                | 51             | 0       |
      | 14                | 53             | 0       |
      | 14                | 55             | 0       |
      | 15                | 15             | 1       |
      | 15                | 51             | 0       |
      | 15                | 53             | 0       |
      | 15                | 55             | 0       |
      | 15                | 59             | 0       |
      | 15                | 61             | 0       |
      | 15                | 63             | 0       |
      | 15                | 65             | 0       |
      | 15                | 69             | 0       |
      | 15                | 67             | 0       |
      | 16                | 16             | 1       |
      | 16                | 63             | 0       |
      | 16                | 65             | 0       |
      | 16                | 67             | 0       |
      | 20                | 20             | 1       |
      | 20                | 21             | 0       |
      | 21                | 21             | 1       |
      | 22                | 1              | 0       |
      | 22                | 11             | 0       |
      | 22                | 12             | 0       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 15             | 0       |
      | 22                | 16             | 0       |
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
      | 22                | 67             | 0       |
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
      | group_id | item_id | order | started_at          | score | best_answer_at      | hints_cached | submissions_attempts | validated | validated_at        | latest_activity_at  |
      | 14       | 211     | 0     | 2017-05-29 06:38:38 | 0     | 2017-05-29 06:38:38 | 100          | 100                  | 0         | 2017-05-30 06:38:38 | 2018-05-30 06:38:38 | # latest_activity_at for 51, 211 comes from this line (the last activity is made by a team)
      | 14       | 211     | 1     | 2017-05-29 06:38:38 | 40    | 2017-05-29 06:38:38 | 2            | 3                    | 1         | 2017-05-29 06:38:58 | null                | # min(validated_at) for 51, 211 comes from this line (from a team)
      | 14       | 211     | 2     | 2017-05-29 06:38:38 | 50    | 2017-05-29 06:38:38 | 3            | 4                    | 1         | 2017-05-31 06:58:38 | null                | # hints_cached & submissions_attempts for 51, 211 come from this line (the best attempt is made by a team)
      | 14       | 211     | 3     | 2017-05-29 06:38:38 | 50    | 2017-05-30 06:38:38 | 10           | 20                   | 1         | null                | null                |
      | 15       | 211     | 0     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                | null                |
      | 15       | 212     | 0     | 2017-03-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                | null                |
      | 16       | 212     | 0     | 2018-12-01 00:00:00 | 10    | 2017-05-30 06:38:38 | 0            | 0                    | 0         | null                | 2019-06-01 00:00:00 | # started_at for 67, 212 & 63, 212 comes from this line (the first attempt is started by a team)
      | 67       | 212     | 0     | 2019-01-01 00:00:00 | 20    | 2017-06-30 06:38:38 | 1            | 2                    | 0         | null                | 2019-06-01 00:00:00 | # hints_cached & submissions_attempts for 67, 212 come from this line (the best attempt is made by a user)
      | 67       | 212     | 1     | 2019-01-01 00:00:00 | 10    | 2017-05-30 06:38:38 | 6            | 7                    | 0         | null                | 2019-07-01 00:00:00 | # latest_activity_at for 67, 212 comes from this line (the last activity is made by a user)
      | 67       | 213     | 0     | 2018-11-01 00:00:00 | 0     | null                | 0            | 0                    | 0         | null                | 2018-11-01 00:00:00 | # started_at for 67, 213 comes from this line (the first attempt is started by a user)
      | 67       | 214     | 0     | 2017-05-29 06:38:38 | 15    | 2017-05-29 06:38:48 | 10           | 11                   | 1         | 2017-05-29 06:38:48 | 2017-05-30 06:38:48 | # min(validated_at) for 67, 214 comes from this line (from a user)
      | 14       | 211     | 4     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                | null                |
      | 15       | 211     | 1     | 2017-04-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                | null                |
      | 15       | 212     | 1     | 2017-03-29 06:38:38 | 100   | null                | 0            | 0                    | 1         | null                | null                |
      | 14       | 211     | 4     | 2017-05-29 06:38:38 | 0     | null                | 0            | 0                    | 0         | null                | null                |

  Scenario: Get progress of the second and the third users (checks sorting, from.*, and limit)
    Given I am the user with group_id "21"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30 20:19:05"
    When I send a GET request to "/groups/1/user-progress?parent_item_ids=210&from.name=janec&from.id=65&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "67",
        "item_id": "211",
        "latest_activity_at": null,
        "score": 0,
        "hints_requested": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "67",
        "hints_requested": 1,
        "item_id": "212",
        "latest_activity_at": "2019-07-01T00:00:00Z",
        "score": 20,
        "submissions_attempts": 2,
        "time_spent": 18303545,
        "validated": false
      },
      {
        "group_id": "67",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": "2018-11-01T00:00:00Z",
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 20895545,
        "validated": false
      },
      {
        "group_id": "67",
        "hints_requested": 10,
        "item_id": "214",
        "latest_activity_at": "2017-05-30T06:38:48Z",
        "score": 15,
        "submissions_attempts": 11,
        "time_spent": 10,
        "validated": true
      },
      {
        "group_id": "67",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },

      {
        "group_id": "51",
        "item_id": "211",
        "latest_activity_at": "2018-05-30T06:38:38Z",
        "score": 50,
        "hints_requested": 3,
        "submissions_attempts": 4,
        "time_spent": 20,
        "validated": true
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """


  Scenario: Get progress of the first user for all the visible items (also checks the limit)
    Given I am the user with group_id "21"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30 20:19:05"
    When I send a GET request to "/groups/1/user-progress?parent_item_ids=210,220,310&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "63",
        "item_id": "211",
        "latest_activity_at": null,
        "score": 0,
        "hints_requested": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": "2019-06-01T00:00:00Z",
        "score": 10,
        "submissions_attempts": 0,
        "time_spent": 18303545,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "221",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "222",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "223",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "224",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "225",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "311",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "312",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "313",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "314",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "63",
        "hints_requested": 0,
        "item_id": "315",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """

  Scenario: No users
    Given I am the user with group_id "21"
    When I send a GET request to "/groups/17/user-progress?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: No visible items
    Given I am the user with group_id "21"
    When I send a GET request to "/groups/1/user-progress?parent_item_ids=1010"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: The input group_id is a user
    Given I am the user with group_id "21"
    When I send a GET request to "/groups/51/user-progress?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: The input group_id is a team
    Given I am the user with group_id "21"
    When I send a GET request to "/groups/14/user-progress?parent_item_ids=210"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "51",
        "hints_requested": 3,
        "item_id": "211",
        "latest_activity_at": "2018-05-30T06:38:38Z",
        "score": 50,
        "submissions_attempts": 4,
        "time_spent": 20,
        "validated": true
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "51",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "53",
        "hints_requested": 3,
        "item_id": "211",
        "latest_activity_at": "2018-05-30T06:38:38Z",
        "score": 50,
        "submissions_attempts": 4,
        "time_spent": 20,
        "validated": true
      },
      {
        "group_id": "53",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "53",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "53",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "53",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "55",
        "hints_requested": 3,
        "item_id": "211",
        "latest_activity_at": "2018-05-30T06:38:38Z",
        "score": 50,
        "submissions_attempts": 4,
        "time_spent": 20,
        "validated": true
      },
      {
        "group_id": "55",
        "hints_requested": 0,
        "item_id": "212",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "55",
        "hints_requested": 0,
        "item_id": "213",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "55",
        "hints_requested": 0,
        "item_id": "214",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      },
      {
        "group_id": "55",
        "hints_requested": 0,
        "item_id": "215",
        "latest_activity_at": null,
        "score": 0,
        "submissions_attempts": 0,
        "time_spent": 0,
        "validated": false
      }
    ]
    """
