Feature: End an attempt (itemAttemptEnd)
  Background:
    Given the database has the following table 'groups':
      | id  | type                |
      | 101 | User                |
      | 102 | Team                |
      | 111 | User                |
      | 201 | ContestParticipants |
      | 202 | ContestParticipants |
      | 203 | ContestParticipants |
    And the database has the following table 'users':
      | login | group_id |
      | john  | 101      |
      | jane  | 111      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | expires_at          |
      | 102             | 101            | 9999-12-31 23:59:59 |
      | 201             | 101            | 9999-12-31 23:59:59 |
      | 201             | 111            | 2019-12-31 23:59:59 |
      | 202             | 101            | 9999-12-31 23:59:59 |
      | 202             | 102            | 9999-12-31 23:59:59 |
      | 202             | 111            | 9999-12-31 23:59:59 |
      | 203             | 101            | 9999-12-31 23:59:59 |
      | 203             | 102            | 2019-12-31 23:59:59 |
      | 203             | 111            | 9999-12-31 23:59:59 |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id | url                                                                     | type    | allows_multiple_attempts | participants_group_id | default_language_tag |
      | 10 | null                                                                    | Chapter | 0                        | 201                   | fr                   |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | 0                        | 202                   | fr                   |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course  | 1                        | 203                   | fr                   |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 10             | 60            | 1           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 101      | 50      | content                  |
      | 102      | 60      | content                  |
      | 111      | 10      | content_with_descendants |
      | 111      | 50      | content_with_descendants |
    And the database has the following table 'attempts':
      | id | participant_id | root_item_id | parent_attempt_id | allows_submissions_until | ended_at            |
      | 1  | 101            | 10           | null              | 9999-12-31 23:59:59      | null                |
      | 1  | 102            | 10           | null              | 9999-12-31 23:59:59      | null                |
      | 1  | 111            | 10           | null              | 9999-12-31 23:59:59      | null                |
      | 2  | 101            | 50           | null              | 9999-12-31 23:59:59      | null                |
      | 2  | 102            | 50           | null              | 9999-12-31 23:59:59      | null                |
      | 2  | 111            | 50           | null              | 9999-12-31 23:59:59      | null                |
      | 3  | 101            | 60           | 1                 | 9999-12-31 23:59:59      | null                |
      | 3  | 102            | 60           | 1                 | 2019-12-31 23:59:59      | 2019-12-31 23:59:59 |
      | 3  | 111            | 60           | 1                 | 9999-12-31 23:59:59      | null                |

  Scenario: User is able to end an attempt for his self group
    Given I am the user with id "111"
    When I send a POST request to "/attempts/1/end"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "updated"
      }
      """
    And the table "attempts" should be:
      | id | participant_id | root_item_id | parent_attempt_id | ABS(TIMESTAMPDIFF(SECOND, allows_submissions_until, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, ended_at, NOW())) < 3 |
      | 1  | 101            | 10           | null              | 0                                                               | null                                            |
      | 1  | 102            | 10           | null              | 0                                                               | null                                            |
      | 1  | 111            | 10           | null              | 1                                                               | 1                                               |
      | 2  | 101            | 50           | null              | 0                                                               | null                                            |
      | 2  | 102            | 50           | null              | 0                                                               | null                                            |
      | 2  | 111            | 50           | null              | 0                                                               | null                                            |
      | 3  | 101            | 60           | 1                 | 0                                                               | null                                            |
      | 3  | 102            | 60           | 1                 | 0                                                               | 0                                               |
      | 3  | 111            | 60           | 1                 | 1                                                               | 1                                               |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | ABS(TIMESTAMPDIFF(SECOND, expires_at, NOW())) < 3 |
      | 102             | 101            | 0                                                 |
      | 201             | 101            | 0                                                 |
      | 201             | 111            | 0                                                 |
      | 202             | 101            | 0                                                 |
      | 202             | 102            | 0                                                 |
      | 202             | 111            | 0                                                 |
      | 203             | 101            | 0                                                 |
      | 203             | 102            | 0                                                 |
      | 203             | 111            | 1                                                 |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | ABS(TIMESTAMPDIFF(SECOND, expires_at, NOW())) < 3 |
      | 101               | 101            | 0                                                 |
      | 102               | 102            | 0                                                 |
      | 111               | 111            | 0                                                 |
      | 201               | 101            | 0                                                 |
      | 201               | 201            | 0                                                 |
      | 202               | 101            | 0                                                 |
      | 202               | 102            | 0                                                 |
      | 202               | 111            | 0                                                 |
      | 202               | 202            | 0                                                 |
      | 203               | 101            | 0                                                 |
      | 203               | 203            | 0                                                 |

  Scenario: User is able to end an attempt as a team
    Given I am the user with id "101"
    When I send a POST request to "/attempts/1/end?as_team_id=102"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "updated"
      }
      """
    And the table "attempts" should be:
      | id | participant_id | root_item_id | parent_attempt_id | ABS(TIMESTAMPDIFF(SECOND, allows_submissions_until, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, ended_at, NOW())) < 3 |
      | 1  | 101            | 10           | null              | 0                                                               | null                                            |
      | 1  | 102            | 10           | null              | 1                                                               | 1                                               |
      | 1  | 111            | 10           | null              | 0                                                               | null                                            |
      | 2  | 101            | 50           | null              | 0                                                               | null                                            |
      | 2  | 102            | 50           | null              | 0                                                               | null                                            |
      | 2  | 111            | 50           | null              | 0                                                               | null                                            |
      | 3  | 101            | 60           | 1                 | 0                                                               | null                                            |
      | 3  | 102            | 60           | 1                 | 0                                                               | 0                                               |
      | 3  | 111            | 60           | 1                 | 0                                                               | null                                            |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | ABS(TIMESTAMPDIFF(SECOND, expires_at, NOW())) < 3 |
      | 102             | 101            | 0                                                 |
      | 201             | 101            | 0                                                 |
      | 201             | 111            | 0                                                 |
      | 202             | 101            | 0                                                 |
      | 202             | 102            | 0                                                 |
      | 202             | 111            | 0                                                 |
      | 203             | 101            | 0                                                 |
      | 203             | 102            | 0                                                 |
      | 203             | 111            | 0                                                 |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | ABS(TIMESTAMPDIFF(SECOND, expires_at, NOW())) < 3 |
      | 101               | 101            | 0                                                 |
      | 102               | 102            | 0                                                 |
      | 111               | 111            | 0                                                 |
      | 201               | 101            | 0                                                 |
      | 201               | 201            | 0                                                 |
      | 202               | 101            | 0                                                 |
      | 202               | 102            | 0                                                 |
      | 202               | 111            | 0                                                 |
      | 202               | 202            | 0                                                 |
      | 203               | 101            | 0                                                 |
      | 203               | 111            | 0                                                 |
      | 203               | 203            | 0                                                 |
