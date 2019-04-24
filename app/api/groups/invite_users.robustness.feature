Feature: Invite users - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName |
      | 1  | owner  | 21          | 22           | Jean-Michel | Blanquer  |
      | 2  | user   | 11          | 12           | John        | Doe       |
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
      | 13              | 13           | 1       |
      | 21              | 21           | 1       |
      | 22              | 13           | 0       |
      | 22              | 22           | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate          |
      | 15 | 22            | 13           | direct             | null                 |

  Scenario: Fails when the user is not an owner of the parent group
    Given I am the user with ID "2"
    When I send a POST request to "/groups/13/invite_users" with the following body:
      """
      {
        "logins": ["john", "jane", "owner", "barack"]
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the parent group ID is wrong
    Given I am the user with ID "1"
    When I send a POST request to "/groups/abc/invite_users" with the following body:
      """
      {
        "logins": ["john", "jane", "owner", "barack"]
      }
      """
    Then the response code should be 400
    And the response error message should contain "Wrong value for parent_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when logins are wrong
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/invite_users" with the following body:
      """
      {
        "logins": [1, 2, 3]
      }
      """
    Then the response code should be 400
    And the response error message should contain "Json: cannot unmarshal number into Go struct field .logins of type string"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when logins are not present
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/invite_users" with the following body:
      """
      {
      }
      """
    Then the response code should be 400
    And the response error message should contain "There should be at least one login listed"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when logins are empty
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/invite_users" with the following body:
      """
      {
        "logins": []
      }
      """
    Then the response code should be 400
    And the response error message should contain "There should be at least one login listed"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

