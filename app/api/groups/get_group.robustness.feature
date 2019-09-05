Feature: Get group by groupID (groupView) - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | sDefaultLanguage |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | fr               |
    And the database has the following table 'groups':
      | ID | sName      | iGrade | sDescription    | sDateCreated         | sType     | sRedirectPath                          | bOpened | bFreeAccess | sCode      | sCodeTimer | sCodeEnd            | bOpenContest |
      | 11 | Group A    | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00   | 2017-10-13 05:39:48 | true         |
      | 13 | Group B    | -2     | Group B is here | 2019-03-06T09:26:40Z | Class     | 182529188317717610/1672978871462145461 | true    | false       | ybabbxnlyo | 01:00:00   | 2017-10-14 05:39:48 | true         |
      | 14 | Group C    | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | true        | null       | null       | null                | false        |

  Scenario: Should fail when the user is neither an owner of the group nor a descendant of the group and bFreeAccess=0
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when the user doesn't exist
    Given I am the user with ID "10"
    When I send a GET request to "/groups/13"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
