Feature: Invite users
  Background:
    Given the database has the following table 'users':
      | ID  | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName |
      | 1   | owner  | 21          | 22           | Jean-Michel | Blanquer  |
      | 10  | john   | 101         | 111          | John        | Doe       |
      | 11  | jane   | 102         | 112          | Jane        | Doe       |
      | 12  | Jane   | 103         | 113          | Jane        | Smith     |
    And the database has the following table 'groups':
      | ID  |
      | 13  |
      | 21  |
      | 22  |
      | 101 |
      | 102 |
      | 103 |
      | 111 |
      | 112 |
      | 113 |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate          |
      | 15 | 22            | 13           | direct             | null                 |

  Scenario: Accept requests
    Given I am the user with ID "1"
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
    And the table "groups_groups" should be:
      | idGroupParent | idGroupChild | sType              | sRole  | idUserInviting | iChildOrder = 0 | (sStatusDate IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, sStatusDate, NOW())) < 3) |
      | 13            | 21           | invitationSent     | member | 1              | 0               | 1                                                                                  |
      | 13            | 101          | invitationSent     | member | 1              | 0               | 1                                                                                  |
      | 13            | 102          | invitationSent     | member | 1              | 0               | 1                                                                                  |
      | 22            | 13           | direct             | member | null           | 1               | 0                                                                                  |
    And the table "groups_groups" should be:
      | iChildOrder |
      | 0           |
      | 1           |
      | 2           |
      | 3           |
