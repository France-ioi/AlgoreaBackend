Feature: Get group by groupID (groupView)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned |
      | 1  | owner  | 21          | 22           |
      | 2  | john   | 31          | 32           |
      | 3  | jane   | 41          | 42           |
      | 4  | rick   | 51          | 52           |
      | 5  | ian    | 61          | 62           |
      | 6  | dirk   | 71          | 72           |
      | 7  | chuck  | 81          | 82           |
    And the database has the following table 'groups':
      | ID | sName       | iGrade | sDescription    | sDateCreated        | sType     | sRedirectPath                          | bOpened | bFreeAccess | sCode      | sCodeTimer | sCodeEnd            | bOpenContest |
      | 11 | Group A     | -3     | Group A is here | 2019-02-06 09:26:40 | Class     | 182529188317717510/1672978871462145361 | true    | false       | ybqybxnlyo | 01:00:00   | 2017-10-13 05:39:48 | true         |
      | 13 | Group B     | -2     | Group B is here | 2019-03-06 09:26:40 | Class     | 182529188317717610/1672978871462145461 | true    | false       | ybabbxnlyo | 01:00:00   | 2017-10-14 05:39:48 | true         |
      | 14 | Group C     | -4     | Admin Group     | 2019-04-06 09:26:40 | UserAdmin | null                                   | true    | false       | null       | null       | null                | false        |
      | 15 | Group D     | -4     | Other Group     | 2019-04-06 09:26:40 | Other     | null                                   | false   | true        | abcdefghij | null       | null                | false        |
      | 21 | owner       | 0      | null            | 2019-01-06 09:26:40 | UserSelf  | null                                   | false   | false       | null       | null       | null                | false        |
      | 22 | owner-admin | 0      | null            | 2019-01-06 09:26:40 | UserAdmin | null                                   | false   | false       | null       | null       | null                | false        |
      | 31 | john        | 0      | null            | 2019-01-06 09:26:40 | UserSelf  | null                                   | false   | false       | null       | null       | null                | false        |
      | 32 | john-admin  | 0      | null            | 2019-01-06 09:26:40 | UserAdmin | null                                   | false   | false       | null       | null       | null                | false        |
      | 41 | jane        | 0      | null            | 2019-01-06 09:26:40 | UserSelf  | null                                   | false   | false       | null       | null       | null                | false        |
      | 42 | jane-admin  | 0      | null            | 2019-01-06 09:26:40 | UserAdmin | null                                   | false   | false       | null       | null       | null                | false        |
      | 51 | rick        | 0      | null            | 2019-01-06 09:26:40 | UserSelf  | null                                   | false   | false       | null       | null       | null                | false        |
      | 52 | rick-admin  | 0      | null            | 2019-01-06 09:26:40 | UserAdmin | null                                   | false   | false       | null       | null       | null                | false        |
      | 61 | ian         | 0      | null            | 2019-01-06 09:26:40 | UserSelf  | null                                   | false   | false       | null       | null       | null                | false        |
      | 62 | ian-admin   | 0      | null            | 2019-01-06 09:26:40 | UserAdmin | null                                   | false   | false       | null       | null       | null                | false        |
      | 71 | dirk        | 0      | null            | 2019-01-06 09:26:40 | UserSelf  | null                                   | false   | false       | null       | null       | null                | false        |
      | 72 | dirk-admin  | 0      | null            | 2019-01-06 09:26:40 | UserAdmin | null                                   | false   | false       | null       | null       | null                | false        |
      | 81 | chuck       | 0      | null            | 2019-01-06 09:26:40 | UserSelf  | null                                   | false   | false       | null       | null       | null                | false        |
      | 82 | chuck-admin | 0      | null            | 2019-01-06 09:26:40 | UserAdmin | null                                   | false   | false       | null       | null       | null                | false        |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType              |
      | 11            | 31           | invitationAccepted |
      | 13            | 11           | direct             |
      | 13            | 31           | requestRefused     |
      | 13            | 51           | requestAccepted    |
      | 13            | 61           | invitationAccepted |
      | 13            | 71           | direct             |
      | 13            | 81           | joinedByCode       |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 11              | 31           | 0       |
      | 13              | 11           | 0       |
      | 13              | 13           | 1       |
      | 13              | 31           | 0       |
      | 13              | 51           | 0       |
      | 13              | 61           | 0       |
      | 13              | 71           | 0       |
      | 13              | 81           | 0       |
      | 14              | 14           | 1       |
      | 15              | 15           | 1       |
      | 21              | 21           | 1       |
      | 22              | 11           | 0       |
      | 22              | 13           | 0       |
      | 22              | 22           | 1       |
      | 22              | 51           | 0       |
      | 22              | 61           | 0       |
      | 22              | 71           | 0       |
      | 22              | 81           | 0       |
      | 31              | 31           | 1       |
      | 32              | 32           | 1       |
      | 41              | 41           | 1       |
      | 42              | 42           | 1       |
      | 51              | 51           | 1       |
      | 52              | 52           | 1       |
      | 61              | 61           | 1       |
      | 62              | 62           | 1       |
      | 71              | 71           | 1       |
      | 72              | 72           | 1       |
      | 81              | 81           | 1       |
      | 82              | 82           | 1       |

  Scenario: The user is an owner of the group
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "13",
      "name": "Group B",
      "grade": -2,
      "description": "Group B is here",
      "date_created": "2019-03-06T09:26:40Z",
      "type": "Class",
      "redirect_path": "182529188317717610/1672978871462145461",
      "opened": true,
      "free_access": false,
      "code": "ybabbxnlyo",
      "code_timer": "01:00:00",
      "code_end": "2017-10-14T05:39:48Z",
      "open_contest": true,
      "current_user_is_owner": true,
      "current_user_is_member": false
    }
    """

  Scenario: The user is a descendant of the group
    Given I am the user with ID "2"
    When I send a GET request to "/groups/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "13",
      "name": "Group B",
      "grade": -2,
      "description": "Group B is here",
      "date_created": "2019-03-06T09:26:40Z",
      "type": "Class",
      "redirect_path": "182529188317717610/1672978871462145461",
      "opened": true,
      "free_access": false,
      "open_contest": true,
      "current_user_is_owner": false,
      "current_user_is_member": false
    }
    """

  Scenario Outline: The user is a member of the group
    Given I am the user with ID "<user_id>"
    When I send a GET request to "/groups/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "13",
      "name": "Group B",
      "grade": -2,
      "description": "Group B is here",
      "date_created": "2019-03-06T09:26:40Z",
      "type": "Class",
      "redirect_path": "182529188317717610/1672978871462145461",
      "opened": true,
      "free_access": false,
      "open_contest": true,
      "current_user_is_owner": false,
      "current_user_is_member": true
    }
    """
  Examples:
    | user_id |
    | 4       |
    | 5       |
    | 6       |
    | 7       |

  Scenario: The group has bFreeAccess = 1
    Given I am the user with ID "3"
    When I send a GET request to "/groups/15"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "15",
      "name": "Group D",
      "grade": -4,
      "description": "Other Group",
      "date_created": "2019-04-06T09:26:40Z",
      "type": "Other",
      "redirect_path": null,
      "opened": false,
      "free_access": true,
      "open_contest": false,
      "current_user_is_owner": false,
      "current_user_is_member": false
    }
    """
