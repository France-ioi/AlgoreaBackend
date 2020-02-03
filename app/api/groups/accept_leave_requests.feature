Feature: Accept requests to leave a group
  Background:
    Given the database has the following table 'groups':
      | id  | type     | team_item_id |
      | 11  | Class    | null         |
      | 13  | Team     | 1234         |
      | 14  | Friends  | null         |
      | 21  | UserSelf | null         |
      | 31  | UserSelf | null         |
      | 111 | UserSelf | null         |
      | 121 | UserSelf | null         |
      | 122 | UserSelf | null         |
      | 123 | UserSelf | null         |
      | 131 | UserSelf | null         |
      | 141 | UserSelf | null         |
      | 151 | UserSelf | null         |
      | 161 | UserSelf | null         |
      | 444 | Team     | 1234         |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name | grade |
      | owner | 21       | Jean-Michel | Blanquer  | 3     |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | expires_at          |
      | 11                | 11             | 9999-12-31 23:59:59 |
      | 13                | 13             | 9999-12-31 23:59:59 |
      | 13                | 111            | 9999-12-31 23:59:59 |
      | 13                | 121            | 9999-12-31 23:59:59 |
      | 13                | 123            | 9999-12-31 23:59:59 |
      | 13                | 141            | 2019-05-30 11:00:00 |
      | 13                | 151            | 9999-12-31 23:59:59 |
      | 14                | 14             | 9999-12-31 23:59:59 |
      | 21                | 21             | 9999-12-31 23:59:59 |
      | 31                | 31             | 9999-12-31 23:59:59 |
      | 111               | 111            | 9999-12-31 23:59:59 |
      | 121               | 121            | 9999-12-31 23:59:59 |
      | 122               | 122            | 9999-12-31 23:59:59 |
      | 123               | 123            | 9999-12-31 23:59:59 |
      | 141               | 141            | 9999-12-31 23:59:59 |
      | 151               | 151            | 9999-12-31 23:59:59 |
      | 161               | 161            | 9999-12-31 23:59:59 |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | expires_at          |
      | 8  | 13              | 31             | 9999-12-31 23:59:59 |
      | 9  | 13              | 121            | 9999-12-31 23:59:59 |
      | 10 | 13              | 111            | 9999-12-31 23:59:59 |
      | 13 | 13              | 123            | 9999-12-31 23:59:59 |
      | 14 | 13              | 141            | 2019-05-30 11:00:00 |
      | 16 | 13              | 151            | 9999-12-31 23:59:59 |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type          |
      | 13       | 21        | invitation    |
      | 13       | 31        | leave_request |
      | 13       | 141       | leave_request |
      | 13       | 161       | join_request  |
      | 14       | 11        | invitation    |
      | 14       | 21        | join_request  |

  Scenario: Accept requests to leave a group
    Given I am the user with id "21"
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 13       | 21         | memberships |
    When I send a POST request to "/groups/13/leave-requests/accept?group_ids=31,141,21,11,13,122,151"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "data": {
        "141": "invalid",
        "31": "success",
        "11": "invalid",
        "13": "invalid",
        "21": "invalid",
        "122": "invalid",
        "151": "invalid"
      },
      "message": "updated",
      "success": true
    }
    """
    And the table "groups_groups" should stay unchanged but the row with parent_group_id "13"
    And the table "groups_groups" at parent_group_id "13" should be:
      | parent_group_id | child_group_id | expires_at          |
      | 13              | 111            | 9999-12-31 23:59:59 |
      | 13              | 121            | 9999-12-31 23:59:59 |
      | 13              | 123            | 9999-12-31 23:59:59 |
      | 13              | 141            | 2019-05-30 11:00:00 |
      | 13              | 151            | 9999-12-31 23:59:59 |
    And the table "group_pending_requests" should be:
      | group_id | member_id | type          |
      | 13       | 21        | invitation    |
      | 13       | 141       | leave_request |
      | 13       | 161       | join_request  |
      | 14       | 11        | invitation    |
      | 14       | 21        | join_request  |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action                 | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 13       | 31        | leave_request_accepted | 21           | 1                                         |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self | expires_at          |
      | 11                | 11             | 1       | 9999-12-31 23:59:59 |
      | 13                | 13             | 1       | 9999-12-31 23:59:59 |
      | 13                | 111            | 0       | 9999-12-31 23:59:59 |
      | 13                | 121            | 0       | 9999-12-31 23:59:59 |
      | 13                | 123            | 0       | 9999-12-31 23:59:59 |
      | 13                | 141            | 0       | 2019-05-30 11:00:00 |
      | 13                | 151            | 0       | 9999-12-31 23:59:59 |
      | 14                | 14             | 1       | 9999-12-31 23:59:59 |
      | 21                | 21             | 1       | 9999-12-31 23:59:59 |
      | 31                | 31             | 1       | 9999-12-31 23:59:59 |
      | 111               | 111            | 1       | 9999-12-31 23:59:59 |
      | 121               | 121            | 1       | 9999-12-31 23:59:59 |
      | 122               | 122            | 1       | 9999-12-31 23:59:59 |
      | 123               | 123            | 1       | 9999-12-31 23:59:59 |
      | 131               | 131            | 1       | 9999-12-31 23:59:59 |
      | 141               | 141            | 1       | 9999-12-31 23:59:59 |
      | 151               | 151            | 1       | 9999-12-31 23:59:59 |
      | 161               | 161            | 1       | 9999-12-31 23:59:59 |
      | 444               | 444            | 1       | 9999-12-31 23:59:59 |
