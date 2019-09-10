Feature: Remove members from a group (groupRemoveMembers)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 1  | owner  | 21          | 22           | Jean-Michel | Blanquer  | 3      |
      | 2  | user   | 11          | 12           | John        | Doe       | 1      |
    And the database has the following table 'groups':
      | ID  |
      | 11  |
      | 12  |
      | 13  |
      | 21  |
      | 22  |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 12              | 12           | 1       |
      | 13              | 11           | 0       |
      | 13              | 13           | 1       |
      | 13              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 21           | 0       |
      | 22              | 22           | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate               |
      | 1  | 13            | 21           | invitationAccepted | {{relativeTime("-170h")}} |
      | 2  | 13            | 11           | requestAccepted    | {{relativeTime("-169h")}} |
      | 15 | 22            | 13           | direct             | null                      |

  Scenario: Fails when the user is not an owner of the parent group
    Given I am the user with ID "2"
    When I send a DELETE request to "/groups/13/members?user_ids=1,2"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the user doesn't exist
    Given I am the user with ID "404"
    When I send a DELETE request to "/groups/13/members?user_ids=1,2"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the parent group ID is wrong
    Given I am the user with ID "1"
    When I send a DELETE request to "/groups/abc/members?user_ids=1,2"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when user_ids is wrong
    Given I am the user with ID "1"
    When I send a DELETE request to "/groups/13/members?user_ids=1,abc,2"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: 'abc', param: 'user_ids')"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
