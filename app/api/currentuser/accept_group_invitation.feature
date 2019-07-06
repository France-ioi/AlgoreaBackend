Feature: User accepts an invitation to join a group
  Background:
    Given the database has the following table 'users':
      | ID | idGroupSelf | idGroupOwned |
      | 1  | 21          | 22           |
    And the database has the following table 'groups':
      | ID |
      | 11 |
      | 14 |
      | 21 |
      | 22 |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 14              | 14           | 1       |
      | 14              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate          |
      | 1  | 11            | 21           | invitationSent     | 2017-04-29T06:38:38Z |
      | 7  | 14            | 21           | invitationAccepted | 2017-02-21T06:38:38Z |

  Scenario: Successfully accept an invitation
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-invitations/11/accept"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created"
    }
    """
    And the table "groups_groups" should stay unchanged but the row with ID "1"
    And the table "groups_groups" at ID "1" should be:
      | ID | idGroupParent | idGroupChild | sType              | (sStatusDate IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, sStatusDate, NOW())) < 3) |
      | 1  | 11            | 21           | invitationAccepted | 1                                                                                  |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 11              | 21           | 0       |
      | 14              | 14           | 1       |
      | 14              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 22           | 1       |

  Scenario: Accept an already accepted invitation
    Given I am the user with ID "1"
    When I send a POST request to "/current-user/group-invitations/14/accept"
    Then the response code should be 205
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "not changed"
    }
    """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
