Feature: Invite users
  Background:
    Given the database has the following table 'groups':
      | id  | type     | team_item_id | require_personal_info_access_approval |
      | 13  | Team     | 1234         | none                                  |
      | 21  | UserSelf | null         | none                                  |
      | 101 | UserSelf | null         | none                                  |
      | 102 | UserSelf | null         | none                                  |
      | 103 | UserSelf | null         | none                                  |
      | 444 | Team     | 1234         | none                                  |
      | 555 | Class    | null         | view                                  |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | john  | 101      | John        | Doe       |
      | jane  | 102      | Jane        | Doe       |
      | Jane  | 103      | Jane        | Smith     |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 13                | 13             |
      | 21                | 21             |
      | 101               | 101            |
      | 102               | 102            |
      | 103               | 103            |
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
      | 555      | 20      | content            |
      | 101      | 30      | content            |
    And the database has the following table 'attempts':
      | group_id | item_id | order | result_propagation_state |
      | 101      | 30      | 1     | done                     |

  Scenario: Successfully invite users
    Given I am the user with id "21"
    And the database table 'groups_ancestors' has also the following rows:
      | ancestor_group_id | child_group_id |
      | 444               | 444            |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage            |
      | 13       | 21         | memberships_and_group |
    When I send a POST request to "/groups/13/invitations" with the following body:
      """
      {
        "logins": ["john", "jane", "owner", "barack"]
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "data": {
          "john": "success",
          "jane": "success",
          "owner": "success",
          "barack": "not_found"
        },
        "message": "created",
        "success": true
      }
      """
    And the table "groups_groups" should be empty
    And the table "group_pending_requests" should be:
      | group_id | member_id | type       | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 13       | 21        | invitation | 1                                         |
      | 13       | 101       | invitation | 1                                         |
      | 13       | 102       | invitation | 1                                         |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action             | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 13       | 21        | invitation_created | 21           | 1                                         |
      | 13       | 101       | invitation_created | 21           | 1                                         |
      | 13       | 102       | invitation_created | 21           | 1                                         |
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Successfully invite users into a team skipping those who are members of other teams with the same team_item_id
    Given I am the user with id "21"
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 13       | 21         | memberships |
    And the database table 'groups_groups' has also the following rows:
      | parent_group_id | child_group_id |
      | 444             | 21             |
      | 444             | 101            |
      | 444             | 102            |
    And the database table 'groups_ancestors' has also the following rows:
      | ancestor_group_id | child_group_id |
      | 444               | 21             |
      | 444               | 101            |
      | 444               | 102            |
      | 444               | 444            |
    When I send a POST request to "/groups/13/invitations" with the following body:
      """
      {
        "logins": ["john", "jane", "owner", "barack"]
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "data": {
          "john": "in_another_team",
          "jane": "in_another_team",
          "owner": "in_another_team",
          "barack": "not_found"
        },
        "message": "created",
        "success": true
      }
      """
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should be empty
    And the table "group_membership_changes" should be empty
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should stay unchanged

  Scenario: Convert join-requests into invitations or make them accepted depending on approvals
    Given I am the user with id "21"
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage            |
      | 555      | 21         | memberships_and_group |
    And the database table 'groups_ancestors' has also the following rows:
      | ancestor_group_id | child_group_id |
      | 555               | 555            |
    And the database table 'group_pending_requests' has also the following rows:
      | group_id | member_id | type         | personal_info_view_approved | at                  |
      | 555      | 101       | join_request | 1                           | 2019-05-30 11:00:00 |
      | 555      | 102       | join_request | 0                           | 2019-05-30 11:00:00 |
    When I send a POST request to "/groups/555/invitations" with the following body:
      """
      {
        "logins": ["john", "jane"]
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "data": {
          "john": "success",
          "jane": "success"
        },
        "message": "created",
        "success": true
      }
      """
    And the table "group_pending_requests" should be:
      | group_id | member_id | type       | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 555      | 102       | invitation | 1                                         |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 555             | 101            | 2019-05-30 11:00:00            |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action                | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 555      | 101       | join_request_accepted | 21           | 1                                         |
      | 555      | 102       | invitation_created    | 21           | 1                                         |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 13                | 13             | 1       |
      | 21                | 21             | 1       |
      | 101               | 101            | 1       |
      | 102               | 102            | 1       |
      | 103               | 103            | 1       |
      | 444               | 444            | 1       |
      | 555               | 101            | 0       |
      | 555               | 555            | 1       |
    And the table "attempts" should be:
      | group_id | item_id | result_propagation_state |
      | 101      | 20      | done                     |
      | 101      | 30      | done                     |
