Feature: Reject requests to leave a group
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
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 13                | 13             |
      | 13                | 111            |
      | 13                | 121            |
      | 13                | 123            |
      | 13                | 151            |
      | 14                | 14             |
      | 21                | 21             |
      | 31                | 31             |
      | 111               | 111            |
      | 121               | 121            |
      | 122               | 122            |
      | 123               | 123            |
      | 151               | 151            |
      | 161               | 161            |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 8  | 13              | 31             |
      | 9  | 13              | 121            |
      | 10 | 13              | 111            |
      | 13 | 13              | 123            |
      | 14 | 13              | 141            |
      | 16 | 13              | 151            |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type          |
      | 13       | 21        | invitation    |
      | 13       | 31        | leave_request |
      | 13       | 141       | leave_request |
      | 13       | 161       | join_request  |
      | 14       | 11        | invitation    |
      | 14       | 21        | join_request  |

  Scenario: Reject requests to leave a group
    Given I am the user with id "21"
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 13       | 21         | memberships |
    When I send a POST request to "/groups/13/leave-requests/reject?group_ids=31,141,21,11,13,122,151"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "data": {
        "141": "success",
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
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should be:
      | group_id | member_id | type         |
      | 13       | 21        | invitation   |
      | 13       | 161       | join_request |
      | 14       | 11        | invitation   |
      | 14       | 21        | join_request |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action                | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 13       | 31        | leave_request_refused | 21           | 1                                         |
      | 13       | 141       | leave_request_refused | 21           | 1                                         |
    And the table "groups_ancestors" should stay unchanged
