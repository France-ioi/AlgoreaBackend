Feature: User accepts an invitation to join a group
  Background:
    Given the database has the following table 'groups':
      | id | require_personal_info_access_approval |
      | 11 | none                                  |
      | 14 | none                                  |
      | 15 | view                                  |
      | 21 | none                                  |
      | 22 | none                                  |
    Given the database has the following table 'users':
      | group_id |
      | 21       |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 15                | 15             | 1       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 7  | 14              | 21             |
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
      | 11       | 20      | content            |
      | 15       | 20      | info               |
      | 21       | 30      | content            |
    And the database has the following table 'attempts':
      | group_id | item_id | order | result_propagation_state |
      | 21       | 30      | 1     | done                     |

  Scenario: Successfully accept an invitation
    Given I am the user with id "21"
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type       |
      | 11       | 21        | invitation |
    When I send a POST request to "/current-user/group-invitations/11/accept"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "updated",
      "data": {"changed": true}
    }
    """
    And the table "groups_groups" should stay unchanged but the row with parent_group_id "11"
    And the table "groups_groups" at parent_group_id "11" should be:
      | parent_group_id | child_group_id | personal_info_view_approved_at | lock_membership_approved_at | watch_approved_at |
      | 11              | 21             | null                           | null                        | null              |
    And the table "group_pending_requests" should be empty
    And the table "group_membership_changes" should be:
      | group_id | member_id | action              | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 11       | 21        | invitation_accepted | 21           | 1                                         |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 11                | 21             | 0       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 15                | 15             | 1       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
    And the table "attempts" should be:
      | group_id | item_id | result_propagation_state |
      | 21       | 20      | done                     |
      | 21       | 30      | done                     |

  Scenario: Successfully accept an invitation into a group that requires approvals
    Given I am the user with id "21"
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type       |
      | 15       | 21        | invitation |
    When I send a POST request to "/current-user/group-invitations/15/accept?approvals=personal_info_view,lock_membership,watch"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "updated",
      "data": {"changed": true}
    }
    """
    And the table "groups_groups" should stay unchanged but the row with parent_group_id "15"
    And the table "groups_groups" at parent_group_id "15" should be:
      | parent_group_id | child_group_id | ABS(TIMESTAMPDIFF(SECOND, personal_info_view_approved_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, lock_membership_approved_at, NOW())) < 3 | ABS(TIMESTAMPDIFF(SECOND, watch_approved_at, NOW())) < 3 |
      | 15              | 21             | 1                                                                     | 1                                                                  | 1                                                        |
    And the table "group_pending_requests" should be empty
    And the table "group_membership_changes" should be:
      | group_id | member_id | action              | initiator_id | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 15       | 21        | invitation_accepted | 21           | 1                                         |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 15                | 15             | 1       |
      | 15                | 21             | 0       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
    And the table "attempts" should be:
      | group_id | item_id | result_propagation_state |
      | 21       | 20      | done                     |
      | 21       | 30      | done                     |
