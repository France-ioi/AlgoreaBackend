Feature: User accepts an invitation to join a group
  Background:
    Given the database has the following table 'users':
      | id | group_self_id | group_owned_id |
      | 1  | 21            | 22             |
    And the database has the following table 'groups':
      | id |
      | 11 |
      | 14 |
      | 21 |
      | 22 |
    And the database has the following table 'groups_ancestors':
      | group_ancestor_id | group_child_id | is_self |
      | 11                | 11             | 1       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
    And the database has the following table 'groups_groups':
      | id | group_parent_id | group_child_id | type               | status_date         |
      | 1  | 11              | 21             | invitationSent     | 2017-04-29 06:38:38 |
      | 7  | 14              | 21             | invitationAccepted | 2017-02-21 06:38:38 |

  Scenario: Successfully accept an invitation
    Given I am the user with id "1"
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
    And the table "groups_groups" should stay unchanged but the row with id "1"
    And the table "groups_groups" at id "1" should be:
      | id | group_parent_id | group_child_id | type               | (status_date IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, status_date, NOW())) < 3) |
      | 1  | 11              | 21             | invitationAccepted | 1                                                                                  |
    And the table "groups_ancestors" should be:
      | group_ancestor_id | group_child_id | is_self |
      | 11                | 11             | 1       |
      | 11                | 21             | 0       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |

  Scenario: Accept an already accepted invitation
    Given I am the user with id "1"
    When I send a POST request to "/current-user/group-invitations/14/accept"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "unchanged",
      "data": {"changed": false}
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
