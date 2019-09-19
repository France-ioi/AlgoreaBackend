Feature: Reject group requests
  Background:
    Given the database has the following table 'users':
      | id | login | group_self_id | group_owned_id | first_name  | last_name | grade |
      | 1  | owner | 21            | 22             | Jean-Michel | Blanquer  | 3     |
    And the database has the following table 'groups':
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
    And the database has the following table 'groups_ancestors':
      | group_ancestor_id | group_child_id | is_self |
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
      | id | group_parent_id | group_child_id | type               | status_date               |
      | 1  | 13              | 21             | invitationSent     | {{relativeTime("-170h")}} |
      | 2  | 13              | 11             | invitationRefused  | {{relativeTime("-169h")}} |
      | 3  | 13              | 31             | requestSent        | {{relativeTime("-168h")}} |
      | 5  | 14              | 11             | invitationSent     | null                      |
      | 6  | 14              | 31             | invitationRefused  | null                      |
      | 7  | 14              | 21             | requestSent        | null                      |
      | 8  | 14              | 22             | requestRefused     | null                      |
      | 9  | 13              | 121            | invitationAccepted | 2017-05-29 06:38:38       |
      | 10 | 13              | 111            | requestAccepted    | null                      |
      | 11 | 13              | 131            | removed            | null                      |
      | 12 | 13              | 122            | left               | null                      |
      | 13 | 13              | 123            | direct             | null                      |
      | 14 | 13              | 141            | requestSent        | null                      |
      | 15 | 13              | 151            | joinedByCode       | null                      |
      | 16 | 22              | 13             | direct             | null                      |

  Scenario: Reject requests
    Given I am the user with id "1"
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
    And the table "groups_groups" should stay unchanged but the row with id "3,14"
    And the table "groups_groups" at id "3,14" should be:
      | id | group_parent_id | group_child_id | type           | (status_date IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, status_date, NOW())) < 3) |
      | 3  | 13              | 31             | requestRefused | 1                                                                                  |
      | 14 | 13              | 141            | requestRefused | 1                                                                                  |
    And the table "groups_ancestors" should stay unchanged
