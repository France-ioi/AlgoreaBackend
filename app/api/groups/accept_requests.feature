Feature: Accept group requests
  Background:
    Given the database has the following table 'users':
      | id | login | group_self_id | group_owned_id | first_name  | last_name | grade |
      | 1  | owner | 21            | 22             | Jean-Michel | Blanquer  | 3     |
    And the database has the following table 'groups':
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
      | 31                | 31             | 1       |
      | 111               | 111            | 1       |
      | 121               | 121            | 1       |
      | 122               | 122            | 1       |
      | 123               | 123            | 1       |
      | 151               | 151            | 1       |
      | 161               | 161            | 1       |
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
      | 15 | 22              | 13             | direct             | null                      |
      | 16 | 13              | 151            | joinedByCode       | null                      |
      | 17 | 13              | 161            | requestSent        | null                      |

  Scenario: Accept requests
    Given I am the user with id "1"
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
    And the table "groups_groups" should stay unchanged but the row with id "3,14"
    And the table "groups_groups" at id "3,14" should be:
      | id | group_parent_id | group_child_id | type            | (status_date IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, status_date, NOW())) < 3) |
      | 3  | 13              | 31             | requestAccepted | 1                                                                                  |
      | 14 | 13              | 141            | requestAccepted | 1                                                                                  |
    And the table "groups_ancestors" should be:
      | group_ancestor_id | group_child_id | is_self |
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
    Given I am the user with id "1"
    And the database table 'groups_groups' has also the following rows:
      | id | group_parent_id | group_child_id | type               | status_date |
      | 18 | 444             | 31             | joinedByCode       | null        |
      | 19 | 444             | 141            | invitationAccepted | null        |
      | 20 | 444             | 161            | requestAccepted    | null        |
    And the database table 'groups_ancestors' has also the following rows:
      | group_ancestor_id | group_child_id | is_self |
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
    And the table "groups_ancestors" should stay unchanged
