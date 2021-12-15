Feature: Update an attempt result
  Background:
    Given the database has the following table 'groups':
      | id | name    | type  |
      | 11 | jdoe    | User  |
      | 13 | Group B | Class |
      | 21 | other   | User  |
      | 23 | Group C | Team  |
      | 24 | jane    | User  |
      | 25 | Group D | Club  |
    And the database has the following table 'users':
      | login | group_id | first_name | last_name |
      | jdoe  | 11       | John       | Doe       |
      | other | 21       | George     | Bush      |
      | jane  | 24       | Jane       | Joe       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 13              | 11             | null                           |
      | 13              | 21             | null                           |
      | 23              | 21             | 2019-05-30 11:00:00            |
      | 23              | 24             | null                           |
      | 23              | 31             | null                           |
      | 25              | 24             | 2019-05-30 11:00:00            |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id  | allows_multiple_attempts | default_language_tag |
      | 200 | 0                        | fr                   |
      | 210 | 1                        | fr                   |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | creator_id | parent_attempt_id | root_item_id | ended_at            |
      | 1  | 11             | 2018-05-29 05:38:38 | 21         | 0                 | 210          | 2018-05-29 05:38:38 |
      | 2  | 11             | 2018-05-29 05:38:38 | 11         | 1                 | 200          | 2018-05-29 05:38:38 |
      | 0  | 11             | 2018-05-29 05:38:38 | null       | null              | null         | null                |
      | 0  | 23             | 2019-05-29 05:38:38 | 11         | null              | null         | null                |
      | 1  | 23             | 2019-05-29 05:38:38 | 24         | 0                 | 210          | null                |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | score_computed | validated_at        | started_at          | latest_activity_at  | help_requested |
      | 0          | 11             | 200     | 99             | null                | 2018-05-29 06:38:38 | 2018-05-29 06:38:39 | false          |
      | 0          | 23             | 210     | 99             | 2018-05-29 08:00:00 | 2019-05-29 06:38:38 | 2019-05-29 06:38:39 | true           |
      | 1          | 11             | 200     | 100            | 2018-05-29 07:00:00 | 2018-05-29 06:38:38 | 2018-05-29 06:38:39 | false          |
      | 1          | 11             | 210     | 99             | 2018-05-29 08:00:00 | 2019-05-29 06:38:38 | 2019-05-29 06:38:39 | false          |
      | 1          | 23             | 210     | 99             | 2018-05-29 08:00:00 | 2019-05-29 06:38:38 | 2019-05-29 06:38:39 | false          |
      | 2          | 11             | 200     | 100            | 2018-05-29 07:00:00 | 2018-05-29 06:38:38 | 2018-05-29 06:38:39 | false          |
      | 2          | 11             | 210     | 99             | 2018-05-29 08:00:00 | 2019-05-29 06:38:38 | 2019-05-29 06:38:39 | true           |

  Scenario: A user updates its own result properties (false -> true)
    Given I am the user with id "11"
    When I send a PUT request to "/items/200/attempts/1" with the following body:
      """
      {
        "help_requested": true
      }
      """
    Then the response should be "updated"
    And the table "results" should stay unchanged but the row with participant_id "11"
    And the table "results" at participant_id "11" should be:
      | attempt_id | participant_id | item_id | score_computed | validated_at        | started_at          | latest_activity_at  | help_requested |
      | 0          | 11             | 200     | 99             | null                | 2018-05-29 06:38:38 | 2018-05-29 06:38:39 | false          |
      | 1          | 11             | 200     | 100            | 2018-05-29 07:00:00 | 2018-05-29 06:38:38 | 2018-05-29 06:38:39 | true           |
      | 1          | 11             | 210     | 99             | 2018-05-29 08:00:00 | 2019-05-29 06:38:38 | 2019-05-29 06:38:39 | false          |
      | 2          | 11             | 200     | 100            | 2018-05-29 07:00:00 | 2018-05-29 06:38:38 | 2018-05-29 06:38:39 | false          |
      | 2          | 11             | 210     | 99             | 2018-05-29 08:00:00 | 2019-05-29 06:38:38 | 2019-05-29 06:38:39 | true           |

  Scenario: A user updates its own result properties (true -> false)
    Given I am the user with id "11"
    When I send a PUT request to "/items/210/attempts/2" with the following body:
      """
      {
        "help_requested": false
      }
      """
    Then the response should be "updated"
    And the table "results" should stay unchanged but the row with participant_id "11"
    And the table "results" at participant_id "11" should be:
      | attempt_id | participant_id | item_id | score_computed | validated_at        | started_at          | latest_activity_at  | help_requested |
      | 0          | 11             | 200     | 99             | null                | 2018-05-29 06:38:38 | 2018-05-29 06:38:39 | false          |
      | 1          | 11             | 200     | 100            | 2018-05-29 07:00:00 | 2018-05-29 06:38:38 | 2018-05-29 06:38:39 | false          |
      | 1          | 11             | 210     | 99             | 2018-05-29 08:00:00 | 2019-05-29 06:38:38 | 2019-05-29 06:38:39 | false          |
      | 2          | 11             | 200     | 100            | 2018-05-29 07:00:00 | 2018-05-29 06:38:38 | 2018-05-29 06:38:39 | false          |
      | 2          | 11             | 210     | 99             | 2018-05-29 08:00:00 | 2019-05-29 06:38:38 | 2019-05-29 06:38:39 | false          |

  Scenario: A user updates result properties of his team
    Given I am the user with id "21"
    When I send a PUT request to "/items/210/attempts/1?as_team_id=23" with the following body:
      """
      {
        "help_requested": true
      }
      """
    Then the response should be "updated"
    And the table "results" should stay unchanged but the row with participant_id "23"
    And the table "results" at participant_id "23" should be:
      | attempt_id | participant_id | item_id | score_computed | validated_at        | started_at          | latest_activity_at  | help_requested |
      | 0          | 23             | 210     | 99             | 2018-05-29 08:00:00 | 2019-05-29 06:38:38 | 2019-05-29 06:38:39 | true           |
      | 1          | 23             | 210     | 99             | 2018-05-29 08:00:00 | 2019-05-29 06:38:38 | 2019-05-29 06:38:39 | true           |

  Scenario: Keeps missing properties unchanged
    Given I am the user with id "11"
    When I send a PUT request to "/items/210/attempts/2" with the following body:
      """
      {}
      """
    Then the response should be "updated"
    And the table "results" should stay unchanged
