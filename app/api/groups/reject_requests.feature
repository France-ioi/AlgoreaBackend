Feature: Reject group requests
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
      | login | group_id | owned_group_id | first_name  | last_name | grade |
      | owner | 21       | 22             | Jean-Michel | Blanquer  | 3     |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 13             | 1       |
      | 13                | 111            | 0       |
      | 13                | 121            | 0       |
      | 13                | 123            | 0       |
      | 13                | 151            | 0       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |
      | 22                | 111            | 0       |
      | 22                | 121            | 0       |
      | 22                | 123            | 0       |
      | 22                | 151            | 0       |
      | 31                | 31             | 1       |
      | 111               | 111            | 1       |
      | 121               | 121            | 1       |
      | 122               | 122            | 1       |
      | 123               | 123            | 1       |
      | 131               | 131            | 1       |
      | 141               | 141            | 1       |
      | 151               | 151            | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 9  | 13              | 121            |
      | 10 | 13              | 111            |
      | 13 | 13              | 123            |
      | 15 | 13              | 151            |
      | 16 | 22              | 13             |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         | at                        |
      | 13       | 21        | invitation   | {{relativeTime("-170h")}} |
      | 13       | 31        | join_request | {{relativeTime("-168h")}} |
      | 14       | 11        | invitation   | {{relativeTime("-167h")}} |
      | 14       | 21        | join_request | {{relativeTime("-166h")}} |
      | 13       | 141       | join_request | {{relativeTime("-165h")}} |

  Scenario: Reject requests
    Given I am the user with id "21"
    When I send a POST request to "/groups/13/requests/reject?group_ids=31,141,21,11,13,22,151"
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
      | 13       | 21        | invitation   | 0                                         |
      | 14       | 11        | invitation   | 0                                         |
      | 14       | 21        | join_request | 0                                         |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action               | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 13       | 31        | join_request_refused | 21           | 1                                         |
      | 13       | 141       | join_request_refused | 21           | 1                                         |
    And the table "groups_ancestors" should stay unchanged
