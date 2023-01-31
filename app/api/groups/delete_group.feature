Feature: Delete a group
  Background:
    Given the database has the following table 'groups':
      | id | name              | type  |
      | 11 | Group A           | Class |
      | 13 | Group B           | Class |
      | 14 | Group C           | Class |
      | 15 | Group D           | Class |
      | 21 | Self              | User  |
      | 22 | Group             | Class |
      | 30 | ThreadHelperGroup | Class |
      | 31 | AllUsers          | Base  |
    And the application config is:
      """
      domains:
        -
          domains: [127.0.0.1]
          allUsersGroup: 31
      """
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage            |
      | 14       | 21         | none                  |
      | 15       | 14         | memberships_and_group |
      | 22       | 21         | memberships           |
      | 30       | 21         | memberships_and_group |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | expires_at          |
      | 13              | 11             | 2019-05-30 11:00:00 |
      | 14              | 21             | 9999-12-31 23:59:59 |
      | 15              | 11             | 9999-12-31 23:59:59 |
      | 15              | 13             | 9999-12-31 23:59:59 |
      | 22              | 13             | 9999-12-31 23:59:59 |
      | 22              | 14             | 9999-12-31 23:59:59 |
      | 31              | 21             | 9999-12-31 23:59:59 |
    And the groups ancestors are computed
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type       |
      | 13       | 11        | invitation |
      | 22       | 11        | invitation |
      | 22       | 13        | invitation |
      | 22       | 14        | invitation |
    And the database has the following table 'group_membership_changes':
      | group_id | member_id |
      | 13       | 11        |
      | 22       | 11        |
      | 22       | 13        |
      | 22       | 14        |
    And the database has the following table 'items':
      | id | default_language_tag |
      | 1  | fr                   |
      | 2  | fr                   |
    And the database has the following table 'threads':
      | participant_id | item_id | status                  | helper_group_id |
      | 21             | 1       | waiting_for_participant | 30              |
      | 21             | 2       | waiting_for_participant | 22              |

  Scenario: User deletes a group
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted"
    }
    """
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id |
      | 14              | 21             |
      | 15              | 13             |
      | 22              | 13             |
      | 22              | 14             |
    And the table "group_pending_requests" should be:
      | group_id | member_id |
      | 22       | 13        |
      | 22       | 14        |
    And the table "group_membership_changes" should be:
      | group_id | member_id |
      | 22       | 13        |
      | 22       | 14        |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 13                | 13             | 1       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 15                | 13             | 0       |
      | 15                | 15             | 1       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 21             | 0       |
      | 22                | 22             | 1       |
    And the table "groups" should stay unchanged but the row with id "11"
    And the table "groups" should not contain id "11"

  Scenario: User deletes a group ignoring an expired parent-child relation
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted"
    }
    """
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id |
      | 14              | 21             |
      | 15              | 11             |
      | 22              | 14             |
    And the table "group_membership_changes" should be:
      | group_id | member_id |
      | 22       | 11        |
      | 22       | 14        |
    And the table "group_pending_requests" should be:
      | group_id | member_id |
      | 22       | 11        |
      | 22       | 14        |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 15                | 11             | 0       |
      | 15                | 15             | 1       |
      | 21                | 21             | 1       |
      | 22                | 14             | 0       |
      | 22                | 21             | 0       |
      | 22                | 22             | 1       |
    And the table "groups" should be:
      | id | name    | type  |
      | 11 | Group A | Class |
      | 14 | Group C | Class |
      | 15 | Group D | Class |
      | 21 | Self    | User  |
      | 22 | Group   | Class |
    
  Scenario: User deletes a group that is the helper_group_id of a thread should change the helper group to AllUsers
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/30"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted"
    }
    """
    And the table "threads" should stay unchanged but the row with item_id "1"
    And the table "threads" at item_id "1" should be:
      | participant_id | item_id | status                  | helper_group_id |
      | 21             | 1       | waiting_for_participant | 31               |
