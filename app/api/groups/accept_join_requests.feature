Feature: Accept group requests
  Background:
    Given the database has the following table 'groups':
      | id  | type    | team_item_id | require_personal_info_access_approval |
      | 11  | Class   | null         | none                                  |
      | 13  | Team    | 1234         | none                                  |
      | 14  | Friends | null         | view                                  |
      | 21  | User    | null         | none                                  |
      | 31  | User    | null         | none                                  |
      | 111 | User    | null         | none                                  |
      | 121 | User    | null         | none                                  |
      | 122 | User    | null         | none                                  |
      | 123 | User    | null         | none                                  |
      | 131 | User    | null         | none                                  |
      | 141 | User    | null         | none                                  |
      | 151 | User    | null         | none                                  |
      | 161 | User    | null         | none                                  |
      | 444 | Team    | 1234         | none                                  |
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
      | parent_group_id | child_group_id | personal_info_view_approved_at | lock_membership_approved_at | watch_approved_at |
      | 13              | 111            | null                           | null                        | null              |
      | 13              | 121            | null                           | null                        | null              |
      | 13              | 123            | null                           | null                        | null              |
      | 13              | 151            | null                           | null                        | null              |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         | personal_info_view_approved | lock_membership_approved | watch_approved | at                  |
      | 13       | 21        | invitation   | 0                           | 0                        | 0              | 2019-06-01 00:00:00 |
      | 13       | 31        | join_request | 1                           | 0                        | 0              | 2019-06-02 00:00:00 |
      | 13       | 141       | join_request | 0                           | 1                        | 1              | 2019-06-03 00:00:00 |
      | 13       | 161       | join_request | 0                           | 0                        | 0              | 2019-06-04 00:00:00 |
      | 14       | 11        | invitation   | 0                           | 0                        | 0              | 2019-06-05 00:00:00 |
      | 14       | 21        | join_request | 0                           | 0                        | 0              | 2019-06-06 00:00:00 |
    And the database has the following table 'items':
      | id | default_language_tag |
      | 20 | fr                   |
      | 30 | fr                   |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 20               | 30            |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 20             | 30            | 1           |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 13       | 20      | content            |
      | 31       | 30      | content            |
    And the database has the following table 'attempts':
      | group_id | item_id | order | result_propagation_state |
      | 31       | 30      | 1     | done                     |

  Scenario: Accept requests
    Given I am the user with id "21"
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 13       | 21         | memberships |
    When I send a POST request to "/groups/13/join-requests/accept?group_ids=31,141,21,11,13,122,151"
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
    And the table "groups_groups" should stay unchanged but the row with parent_group_id "13"
    And the table "groups_groups" at parent_group_id "13" should be:
      | parent_group_id | child_group_id | personal_info_view_approved_at | lock_membership_approved_at | watch_approved_at   |
      | 13              | 31             | 2019-06-02 00:00:00            | null                        | null                |
      | 13              | 111            | null                           | null                        | null                |
      | 13              | 121            | null                           | null                        | null                |
      | 13              | 123            | null                           | null                        | null                |
      | 13              | 141            | null                           | 2019-06-03 00:00:00         | 2019-06-03 00:00:00 |
      | 13              | 151            | null                           | null                        | null                |
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
    And the table "attempts" should be:
      | group_id | item_id | result_propagation_state |
      | 31       | 20      | done                     |
      | 31       | 30      | done                     |

  Scenario: Accept requests for a team while skipping members of other teams with the same team_item_id
    Given I am the user with id "21"
    And the database table 'groups_groups' has also the following rows:
      | parent_group_id | child_group_id |
      | 444             | 31             |
      | 444             | 141            |
      | 444             | 161            |
    And the database table 'groups_ancestors' has also the following rows:
      | ancestor_group_id | child_group_id |
      | 444               | 31             |
      | 444               | 141            |
      | 444               | 161            |
      | 444               | 444            |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage            |
      | 13       | 21         | memberships_and_group |
    When I send a POST request to "/groups/13/join-requests/accept?group_ids=31,141,21,11,13,122,151,161"
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
          "122": "invalid",
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
    And the table "attempts" should stay unchanged

  Scenario: Checks approvals if required
    Given I am the user with id "21"
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage            |
      | 14       | 21         | memberships_and_group |
    When I send a POST request to "/groups/14/join-requests/accept?group_ids=21"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "data": {
          "21": "approvals_missing"
        },
        "message": "updated",
        "success": true
      }
      """
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should stay unchanged
