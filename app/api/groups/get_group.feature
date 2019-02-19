Feature: Get group by groupID (groupView)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | sDefaultLanguage |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | fr               |
    And the database has the following table 'groups':
      | ID | sName      | iGrade | sDescription    | sDateCreated         | sType     | sRedirectPath                          | bOpened | bFreeAccess | sPassword  | sPasswordTimer | sPasswordEnd        | bOpenContest |
      | 11 | Group A    | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00       | 2017-10-13 05:39:48 | true         |
      | 13 | Group B    | -2     | Group B is here | 2019-03-06T09:26:40Z | Class     | 182529188317717610/1672978871462145461 | true    | true        | ybabbxnlyo | 01:00:00       | 2017-10-14 05:39:48 | true         |
      | 14 | Group C    | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | true        | null       | null           | null                | false        |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 75 | 22              | 13           | 0       | 0        |

  Scenario: The user is an owner of the group
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "name": "Group B",
      "description": "Group B is here",
      "type": "Class",
      "date_created": "2019-03-06T09:26:40Z",
      "free_access": true,
      "grade": -2,
      "name": "Group B",
      "open_contest": true,
      "opened": true,
      "password": "ybabbxnlyo",
      "password_end": "2017-10-14T05:39:48Z",
      "password_timer": "01:00:00",
      "redirect_path": "182529188317717610/1672978871462145461"
    }
    """
