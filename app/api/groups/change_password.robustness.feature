Feature: Change the password of the given group - robustness

  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | sDefaultLanguage |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | fr               |
      | 2  | user   | 0        | 11          | 12           | John        | Doe       | en               |
      | 3  | jane   | 0        | 31          | 32           | Jane        | Doe       | en               |
    And the database has the following table 'groups':
      | ID | sName   | iGrade | sDescription    | sDateCreated         | sType     | sPassword  | sPasswordTimer | sPasswordEnd         |
      | 11 | Group A | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | ybqybxnlyo | 01:00:00       | 2017-10-13T05:39:48Z |
      | 13 | Group B | -2     | Group B is here | 2019-03-06T09:26:40Z | Class     | 3456789abc | 01:00:00       | 2017-10-14T05:39:48Z |
      | 14 | Group C | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null       | null           | null                 |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 75 | 22              | 13           | 0       | 0        |
      | 76 | 13              | 11           | 0       | 0        |
      | 77 | 22              | 11           | 0       | 0        |
      | 78 | 21              | 21           | 1       | 0        |

  Scenario: User is not an admin of the group
    Given I am the user with ID "2"
    And the generated group password is "newpassword"
    When I send a POST request to "/groups/13/password"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups" should stay unchanged

  Scenario: User does not exist
    Given I am the user with ID "404"
    And the generated group password is "newpassword"
    When I send a POST request to "/groups/13/password"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups" should stay unchanged

  Scenario: User is an admin of the group, but the generated password is not unique
    Given I am the user with ID "1"
    And the generated group passwords are "ybqybxnlyo","newpassword"
    When I send a POST request to "/groups/13/password"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"password":"newpassword"}
    """
    And the table "groups" should stay unchanged but the row with ID "13"
    And the table "groups" at ID "13" should be:
      | ID | sName   | iGrade | sDescription    | sDateCreated         | sType | sPassword   | sPasswordTimer | sPasswordEnd         |
      | 13 | Group B | -2     | Group B is here | 2019-03-06T09:26:40Z | Class | newpassword | 01:00:00       | 2017-10-14T05:39:48Z |

  Scenario: User is an admin of the group, but the generated password is not unique 3 times in a row
    Given I am the user with ID "1"
    And the generated group passwords are "ybqybxnlyo","ybqybxnlyo","ybqybxnlyo"
    When I send a POST request to "/groups/13/password"
    Then the response code should be 500
    And the response error message should contain "The password generator is broken"
    And the table "groups" should stay unchanged

  Scenario: The group ID is not a number
    Given I am the user with ID "1"
    When I send a POST request to "/groups/1_3/password"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
