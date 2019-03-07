Feature: Change the password of the given group

  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | sDefaultLanguage |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | fr               |
    And the database has the following table 'groups':
      | ID | sName   | iGrade | sDescription    | sDateCreated         | sType     | sRedirectPath                          | bOpened | bFreeAccess | sPassword  | sPasswordTimer | sPasswordEnd         | bOpenContest |
      | 11 | Group A | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00       | 2017-10-13T05:39:48Z | true         |
      | 13 | Group B | -2     | Group B is here | 2019-03-06T09:26:40Z | Class     | 182529188317717610/1672978871462145461 | true    | true        | ybabbxnlyo | 01:00:00       | 2017-10-14T05:39:48Z | true         |
      | 14 | Group C | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | true        | null       | null           | null                 | false        |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 75 | 22              | 13           | 0       | 0        |
      | 76 | 13              | 11           | 0       | 0        |
      | 77 | 22              | 11           | 0       | 0        |
      | 78 | 21              | 21           | 1       | 0        |

  Scenario: User is an admin of the group
    Given I am the user with ID "1"
    And the generated group password is "newpassword"
    When I send a POST request to "/groups/13/change_password"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"password":"newpassword"}
    """
    And the table "groups" should stay unchanged but the row with ID "13"
    And the table "groups" at ID "13" should be:
      | ID | sName   | iGrade | sDescription    | sDateCreated         | sType | sRedirectPath                          | bOpened | bFreeAccess | sPassword   | sPasswordTimer | sPasswordEnd         | bOpenContest |
      | 13 | Group B | -2     | Group B is here | 2019-03-06T09:26:40Z | Class | 182529188317717610/1672978871462145461 | true    | true        | newpassword | 01:00:00       | 2017-10-14T05:39:48Z | true         |
