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
      | 131 |
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
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 13                | 13             | 1       |
      | 13                | 51             | 0       |
      | 13                | 61             | 0       |
      | 13                | 91             | 0       |
      | 13                | 111            | 0       |
      | 13                | 131            | 0       |
      | 14                | 14             | 1       |
      | 14                | 41             | 0       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
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
      | 131               | 131            | 1       |
      | 132               | 132            | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 6  | 14              | 41             |
      | 9  | 13              | 51             |
      | 10 | 13              | 61             |
      | 13 | 13              | 91             |
      | 15 | 13              | 111            |
      | 16 | 13              | 131            |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         |
      | 13       | 21        | invitation   |
      | 13       | 41        | join_request |
      | 13       | 101       | join_request |
      | 14       | 51        | join_request |

  Scenario: Remove members
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/13/members?user_ids=31,41,51,61,71,81,91,101,111,121,131,404"
    And the response body should be, in JSON:
    """
    {
      "data": {
        "31":  "invalid",
        "41":  "invalid",
        "51":  "success",
        "61":  "success",
        "71":  "invalid",
        "81":  "invalid",
        "91":  "success",
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
      | id | parent_group_id | child_group_id |
      | 6  | 14              | 41             |
      | 16 | 13              | 131            |
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be:
      | group_id | member_id | action  | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 13       | 51        | removed | 21           | 1                                         |
      | 13       | 61        | removed | 21           | 1                                         |
      | 13       | 91        | removed | 21           | 1                                         |
      | 13       | 111       | removed | 21           | 1                                         |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 13                | 13             | 1       |
      | 13                | 131            | 0       |
      | 14                | 14             | 1       |
      | 14                | 41             | 0       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
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
      | 131               | 131            | 1       |
      | 132               | 132            | 1       |
