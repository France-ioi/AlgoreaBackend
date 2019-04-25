Feature: Reject group requests - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 1  | owner  | 21          | 22           | Jean-Michel | Blanquer  | 3      |
      | 2  | user   | 11          | 12           | John        | Doe       | 1      |
    And the database has the following table 'groups':
      | ID  |
      | 11  |
      | 13  |
      | 14  |
      | 21  |
      | 22  |
      | 31  |
      | 111 |
      | 121 |
      | 122 |
      | 123 |
      | 131 |
      | 141 |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 13           | 1       |
      | 13              | 111          | 0       |
      | 13              | 121          | 0       |
      | 13              | 123          | 0       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 13           | 0       |
      | 22              | 22           | 1       |
      | 31              | 31           | 1       |
      | 111             | 111          | 1       |
      | 121             | 121          | 1       |
      | 122             | 122          | 1       |
      | 123             | 123          | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate          |
      | 1  | 13            | 21           | invitationSent     | relativeTime(-170h)  |
      | 2  | 13            | 11           | invitationRefused  | relativeTime(-169h)  |
      | 3  | 13            | 31           | requestSent        | relativeTime(-168h)  |
      | 5  | 14            | 11           | invitationSent     | null                 |
      | 6  | 14            | 31           | invitationRefused  | null                 |
      | 7  | 14            | 21           | requestSent        | null                 |
      | 8  | 14            | 22           | requestRefused     | null                 |
      | 9  | 13            | 121          | invitationAccepted | 2017-05-29T06:38:38Z |
      | 10 | 13            | 111          | requestAccepted    | null                 |
      | 11 | 13            | 131          | removed            | null                 |
      | 12 | 13            | 122          | left               | null                 |
      | 13 | 13            | 123          | direct             | null                 |
      | 14 | 13            | 141          | requestSent        | null                 |
      | 15 | 22            | 13           | direct             | null                 |

  Scenario: Fails when the user is not an owner of the parent group
    Given I am the user with ID "2"
    When I send a POST request to "/groups/13/requests/reject?group_ids=31,141,21,11,13,22"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the parent group ID is wrong
    Given I am the user with ID "1"
    When I send a POST request to "/groups/abc/requests/reject?group_ids=31,141,21,11,13,22"
    Then the response code should be 400
    And the response error message should contain "Wrong value for parent_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when group_ids is wrong
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/requests/reject?group_ids=31,abc,11,13,22"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integer given as query arg (value: 'abc', param: 'group_ids')"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
