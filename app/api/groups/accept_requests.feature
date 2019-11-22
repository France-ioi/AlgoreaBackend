Feature: Accept group requests
  Background:
    Given the database has the following table 'groups':
      | id  | type      | team_item_id |
      | 11  | Class     | null         |
      | 13  | Team      | 1234         |
      | 14  | Friends   | null         |
      | 21  | UserSelf  | null         |
      | 22  | UserAdmin | null         |
      | 31  | UserSelf  | null         |
      | 111 | UserSelf  | null         |
      | 121 | UserSelf  | null         |
      | 122 | UserSelf  | null         |
      | 123 | UserSelf  | null         |
      | 131 | UserSelf  | null         |
      | 141 | UserSelf  | null         |
      | 151 | UserSelf  | null         |
      | 161 | UserSelf  | null         |
      | 444 | Team      | 1234         |
    And the database has the following table 'users':
      | login | group_id | owned_group_id | first_name  | last_name | grade |
      | owner | 21       | 22             | Jean-Michel | Blanquer  | 3     |
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
      | 31                | 31             | 1       |
      | 111               | 111            | 1       |
      | 121               | 121            | 1       |
      | 122               | 122            | 1       |
      | 123               | 123            | 1       |
      | 151               | 151            | 1       |
      | 161               | 161            | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 9  | 13              | 121            |
      | 10 | 13              | 111            |
      | 13 | 13              | 123            |
      | 15 | 22              | 13             |
      | 16 | 13              | 151            |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         |
      | 13       | 21        | invitation   |
      | 13       | 31        | join_request |
      | 13       | 141       | join_request |
      | 13       | 161       | join_request |
      | 14       | 11        | invitation   |
      | 14       | 21        | join_request |

  Scenario: Accept requests
    Given I am the user with id "21"
    When I send a POST request to "/groups/13/requests/accept?group_ids=31,141,21,11,13,22,151"
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
        "22": "invalid",
        "151": "invalid"
      },
      "message": "updated",
      "success": true
    }
    """
    And the table "groups_groups" should stay unchanged but the row with parent_group_id "13"
    And the table "groups_groups" at parent_group_id "13" should be:
      | parent_group_id | child_group_id |
      | 13              | 31             |
      | 13              | 111            |
      | 13              | 121            |
      | 13              | 123            |
      | 13              | 141            |
      | 13              | 151            |
    And the table "group_pending_requests" should be:
      | group_id | member_id | type         |
      | 13       | 21        | invitation   |
      | 13       | 161       | join_request |
      | 14       | 11        | invitation   |
      | 14       | 21        | join_request |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action                | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 13       | 31        | join_request_accepted | 21           | 1                                         |
      | 13       | 141       | join_request_accepted | 21           | 1                                         |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 13             | 1       |
      | 13                | 31             | 0       |
      | 13                | 111            | 0       |
      | 13                | 121            | 0       |
      | 13                | 123            | 0       |
      | 13                | 141            | 0       |
      | 13                | 151            | 0       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |
      | 22                | 31             | 0       |
      | 22                | 111            | 0       |
      | 22                | 121            | 0       |
      | 22                | 123            | 0       |
      | 22                | 141            | 0       |
      | 22                | 151            | 0       |
      | 31                | 31             | 1       |
      | 111               | 111            | 1       |
      | 121               | 121            | 1       |
      | 122               | 122            | 1       |
      | 123               | 123            | 1       |
      | 131               | 131            | 1       |
      | 141               | 141            | 1       |
      | 151               | 151            | 1       |
      | 161               | 161            | 1       |
      | 444               | 444            | 1       |

  Scenario: Accept requests for a team while skipping members of other teams with the same team_item_id
    Given I am the user with id "21"
    And the database table 'groups_groups' has also the following rows:
      | id | parent_group_id | child_group_id |
      | 18 | 444             | 31             |
      | 19 | 444             | 141            |
      | 20 | 444             | 161            |
    And the database table 'groups_ancestors' has also the following rows:
      | ancestor_group_id | child_group_id | is_self |
      | 444               | 31             | 0       |
      | 444               | 141            | 0       |
      | 444               | 161            | 0       |
      | 444               | 444            | 1       |
    When I send a POST request to "/groups/13/requests/accept?group_ids=31,141,21,11,13,22,151,161"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "data": {
          "31": "in_another_team",
          "141": "in_another_team",
          "11": "invalid",
          "13": "invalid",
          "21": "invalid",
          "22": "invalid",
          "151": "invalid",
          "161": "in_another_team"
        },
        "message": "updated",
        "success": true
      }
      """
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged
