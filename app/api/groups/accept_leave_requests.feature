Feature: Accept requests to leave a group
  Background:
    Given the database has the following table 'groups':
      | id  | type    |
      | 11  | Class   |
      | 13  | Team    |
      | 14  | Friends |
      | 21  | User    |
      | 31  | User    |
      | 111 | User    |
      | 121 | User    |
      | 122 | User    |
      | 123 | User    |
      | 131 | User    |
      | 141 | User    |
      | 151 | User    |
      | 161 | User    |
      | 444 | Team    |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name | grade |
      | owner | 21       | Jean-Michel | Blanquer  | 3     |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | expires_at          |
      | 13              | 31             | 9999-12-31 23:59:59 |
      | 13              | 111            | 9999-12-31 23:59:59 |
      | 13              | 121            | 9999-12-31 23:59:59 |
      | 13              | 123            | 9999-12-31 23:59:59 |
      | 13              | 141            | 2019-05-30 11:00:00 |
      | 13              | 151            | 9999-12-31 23:59:59 |
    And the groups ancestors are computed
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
    And the table "groups_ancestors" should stay unchanged
