Feature: Remove members from a group (groupRemoveMembers)
  Background:
    Given the database has the following table 'groups':
      | id  |
      | 13  |
      | 14  |
      | 21  |
      | 22  |
      | 31  |
      | 32  |
      | 41  |
      | 42  |
      | 51  |
      | 52  |
      | 61  |
      | 62  |
      | 71  |
      | 72  |
      | 81  |
      | 82  |
      | 91  |
      | 92  |
      | 101 |
      | 102 |
      | 111 |
      | 112 |
      | 121 |
      | 122 |
      | 132 |
    And the database has the following table 'users':
      | login  | group_id | owned_group_id |
      | owner  | 21       | 22             |
      | john   | 31       | 32             |
      | jane   | 41       | 42             |
      | jack   | 51       | 52             |
      | james  | 61       | 62             |
      | jacob  | 71       | 72             |
      | janis  | 81       | 82             |
      | jeff   | 91       | 92             |
      | jenna  | 101      | 102            |
      | jannet | 111      | 112            |
      | judith | 121      | 122            |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 13                | 13             | 1       |
      | 13                | 51             | 0       |
      | 13                | 61             | 0       |
      | 13                | 91             | 0       |
      | 13                | 111            | 0       |
      | 14                | 14             | 1       |
      | 14                | 41             | 0       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |
      | 22                | 51             | 0       |
      | 22                | 61             | 0       |
      | 22                | 91             | 0       |
      | 22                | 111            | 0       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
      | 41                | 41             | 1       |
      | 42                | 42             | 1       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
      | 61                | 61             | 1       |
      | 62                | 62             | 1       |
      | 71                | 71             | 1       |
      | 72                | 72             | 1       |
      | 81                | 81             | 1       |
      | 82                | 82             | 1       |
      | 91                | 91             | 1       |
      | 92                | 92             | 1       |
      | 101               | 101            | 1       |
      | 102               | 102            | 1       |
      | 111               | 111            | 1       |
      | 112               | 112            | 1       |
      | 121               | 121            | 1       |
      | 122               | 122            | 1       |
      | 132               | 132            | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | type               | type_changed_at           |
      | 1  | 13              | 21             | invitationSent     | {{relativeTime("-170h")}} |
      | 2  | 13              | 31             | invitationRefused  | {{relativeTime("-169h")}} |
      | 3  | 13              | 41             | requestSent        | {{relativeTime("-168h")}} |
      | 6  | 14              | 41             | invitationAccepted | null                      |
      | 7  | 14              | 51             | requestSent        | null                      |
      | 9  | 13              | 51             | invitationAccepted | 2017-05-29 06:38:38       |
      | 10 | 13              | 61             | requestAccepted    | null                      |
      | 11 | 13              | 71             | removed            | null                      |
      | 12 | 13              | 81             | left               | null                      |
      | 13 | 13              | 91             | direct             | null                      |
      | 14 | 13              | 101            | requestSent        | null                      |
      | 15 | 13              | 111            | joinedByCode       | null                      |
      | 16 | 22              | 13             | direct             | null                      |

  Scenario: Remove members
    Given I am the user with group_id "21"
    When I send a DELETE request to "/groups/13/members?user_group_ids=31,41,51,61,71,81,91,101,111,121,131,404"
    And the response body should be, in JSON:
    """
    {
      "data": {
        "31":  "invalid",
        "41":  "invalid",
        "51":  "success",
        "61":  "success",
        "71":  "unchanged",
        "81":  "invalid",
        "91":  "invalid",
        "101": "invalid",
        "111": "success",
        "121": "invalid",
        "131": "not_found",
        "404": "not_found"
      },
      "message": "deleted",
      "success": true
    }
    """
    And the table "groups_groups" should be:
      | id | parent_group_id | child_group_id | type               | (type_changed_at IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, type_changed_at, NOW())) < 3) |
      | 1  | 13              | 21             | invitationSent     | 0                                                                                          |
      | 2  | 13              | 31             | invitationRefused  | 0                                                                                          |
      | 3  | 13              | 41             | requestSent        | 0                                                                                          |
      | 6  | 14              | 41             | invitationAccepted | 0                                                                                          |
      | 7  | 14              | 51             | requestSent        | 0                                                                                          |
      | 9  | 13              | 51             | removed            | 1                                                                                          |
      | 10 | 13              | 61             | removed            | 1                                                                                          |
      | 11 | 13              | 71             | removed            | 0                                                                                          |
      | 12 | 13              | 81             | left               | 0                                                                                          |
      | 13 | 13              | 91             | direct             | 0                                                                                          |
      | 14 | 13              | 101            | requestSent        | 0                                                                                          |
      | 15 | 13              | 111            | removed            | 1                                                                                          |
      | 16 | 22              | 13             | direct             | 0                                                                                          |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 13                | 13             | 1       |
      | 13                | 91             | 0       |
      | 14                | 14             | 1       |
      | 14                | 41             | 0       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |
      | 22                | 91             | 0       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
      | 41                | 41             | 1       |
      | 42                | 42             | 1       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
      | 61                | 61             | 1       |
      | 62                | 62             | 1       |
      | 71                | 71             | 1       |
      | 72                | 72             | 1       |
      | 81                | 81             | 1       |
      | 82                | 82             | 1       |
      | 91                | 91             | 1       |
      | 92                | 92             | 1       |
      | 101               | 101            | 1       |
      | 102               | 102            | 1       |
      | 111               | 111            | 1       |
      | 112               | 112            | 1       |
      | 121               | 121            | 1       |
      | 122               | 122            | 1       |
      | 132               | 132            | 1       |
