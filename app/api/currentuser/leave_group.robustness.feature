Feature: User leaves a group - robustness
  Background:
    Given the database has the following table 'users':
      | ID | idGroupSelf | idGroupOwned | sLogin |
      | 1  | 21          | 22           | john   |
      | 2  | null        | null         | guest  |
    And the database has the following table 'groups':
      | ID |
      | 11 |
      | 14 |
      | 21 |
      | 22 |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate          |
      | 1  | 11            | 21           | requestSent        | 2017-04-29T06:38:38Z |
      | 2  | 14            | 21           | direct             | 2017-03-29T06:38:38Z |

  Scenario: User tries to leave a group (s)he is not a member of
    Given I am the user with ID "1"
    When I send a DELETE request to "/current-user/group-memberships/11"
    Then the response code should be 404
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Not Found",
      "error_text": "No such relation"
    }
    """
    And the table "groups_groups" should stay unchanged

  Scenario: Fails when the group ID is wrong
    Given I am the user with ID "1"
    When I send a DELETE request to "/current-user/group-memberships/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups_groups" should stay unchanged

  Scenario: Fails when the user's idGroupSelf is NULL
    Given I am the user with ID "2"
    When I send a DELETE request to "/current-user/group-memberships/14"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged

  Scenario: Fails if the user doesn't exist
    Given I am the user with ID "4"
    When I send a DELETE request to "/current-user/group-memberships/14"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

