Feature: Withdraw group invitations
  Background:
    Given the database has the following table 'groups':
      | id  |
      | 11  |
      | 13  |
      | 14  |
      | 21  |
      | 22  |
      | 31  |
      | 111 |
      | 121 |
      | 122 |
      | 123 |
      | 131 |
      | 141 |
      | 151 |
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
      | 22                | 13             |
      | 22                | 22             |
      | 22                | 111            |
      | 22                | 121            |
      | 22                | 123            |
      | 22                | 151            |
      | 31                | 31             |
      | 111               | 111            |
      | 121               | 121            |
      | 122               | 122            |
      | 123               | 123            |
      | 131               | 131            |
      | 141               | 141            |
      | 151               | 151            |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 111            |
      | 13              | 121            |
      | 13              | 123            |
      | 13              | 151            |
      | 22              | 13             |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         | at                        |
      | 13       | 21        | join_request | {{relativeTime("-170h")}} |
      | 13       | 31        | invitation   | {{relativeTime("-168h")}} |
      | 14       | 11        | join_request | {{relativeTime("-167h")}} |
      | 14       | 21        | invitation   | {{relativeTime("-166h")}} |
      | 13       | 141       | invitation   | {{relativeTime("-165h")}} |

  Scenario Outline: Withdraw invitations
    Given I am the user with id "21"
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage   |
      | 13       | 21         | <can_manage> |
    When I send a POST request to "/groups/13/invitations/withdraw?group_ids=31,141,21,11,13,22,151"
    And the response body should be, in JSON:
    """
    {
      "data": {
        "141": "success",
        "31": "success",
        "11": "invalid",
        "13": "invalid",
        "21": "invalid",
        "22": "invalid",
        "151": "invalid"
      },
      "message": "updated",
      "success": true
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should be:
      | group_id | member_id | type         | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 13       | 21        | join_request | 0                                         |
      | 14       | 11        | join_request | 0                                         |
      | 14       | 21        | invitation   | 0                                         |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action               | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 13       | 31        | invitation_withdrawn | 21           | 1                                         |
      | 13       | 141       | invitation_withdrawn | 21           | 1                                         |
    And the table "groups_ancestors" should stay unchanged
  Examples:
    | can_manage            |
    | memberships           |
    | memberships_and_group |
