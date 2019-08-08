Feature: Get group children (groupChildrenView)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName |
      | 1  | owner  | 21          | 22           | Jean-Michel | Blanquer  |
      | 2  | john   | 51          | 52           | John        | Doe       |
      | 3  | jane   | 53          | 54           | Jane        | Doe       |
    And the database has the following table 'groups':
      | ID | sName         | iGrade | sType     | bOpened | bFreeAccess | sPassword  |
      | 11 | Group A       | -3     | Class     | true    | true        | ybqybxnlyo |
      | 13 | Group B       | -2     | Class     | true    | true        | ybabbxnlyo |
      | 14 | Group C       | -4     | UserAdmin | true    | false       | null       |
      | 21 | user-admin    | -2     | UserAdmin | true    | false       | null       |
      | 22 | C's Child     | -4     | UserAdmin | true    | false       | null       |
      | 23 | Our Class     | -3     | Class     | true    | false       | null       |
      | 24 | Root          | -2     | Base      | true    | false       | 3456789abc |
      | 25 | Our Team      | -1     | Team      | true    | false       | 456789abcd |
      | 26 | Our Club      |  0     | Club      | true    | false       | null       |
      | 27 | Our Friends   |  0     | Friends   | true    | false       | 56789abcde |
      | 28 | Other         |  0     | Other     | true    | false       | null       |
      | 29 | UserSelf      |  0     | UserSelf  | true    | false       | null       |
      | 30 | RootSelf      |  0     | Base      | true    | false       | null       |
      | 31 | RootAdmin     |  0     | Base      | true    | false       | null       |
      | 42 | Their Class   | -3     | Class     | true    | false       | null       |
      | 43 | Other Root    | -2     | Base      | true    | false       | 3456789abc |
      | 44 | Other Team    | -1     | Team      | true    | false       | 456789abcd |
      | 45 | Their Club    |  0     | Club      | true    | false       | null       |
      | 46 | Their Friends |  0     | Friends   | true    | false       | 56789abcde |
      | 47 | Other         |  0     | Other     | true    | false       | null       |
      | 51 | john          |  0     | UserSelf  | false   | false       | null       |
      | 52 | john-admin    |  0     | UserAdmin | false   | false       | null       |
      | 53 | jane          |  0     | UserSelf  | false   | false       | null       |
      | 54 | jane-admin    |  0     | UserAdmin | false   | false       | null       |
      | 90 | Sub-Class     |  0     | Team      | false   | false       | null       |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType              |
      | 13            | 21           | invitationAccepted |
      | 13            | 23           | direct             |
      | 13            | 24           | direct             |
      | 14            | 22           | direct             |
      | 13            | 25           | direct             |
      | 13            | 26           | direct             |
      | 13            | 27           | direct             |
      | 13            | 28           | direct             |
      | 13            | 29           | requestAccepted    |
      | 13            | 30           | direct             |
      | 13            | 31           | direct             |
      | 13            | 42           | invitationSent     |
      | 13            | 43           | requestSent        |
      | 13            | 44           | invitationRefused  |
      | 13            | 45           | requestRefused     |
      | 13            | 46           | left               |
      | 13            | 47           | removed            |
      | 22            | 13           | direct             |
      | 23            | 51           | invitationAccepted |
      | 23            | 90           | direct             |
      | 27            | 53           | invitationAccepted |
      | 90            | 51           | requestAccepted    |

  Scenario: User is an owner of the parent group, rows are sorted by name by default, UserSelf is skipped
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children?types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "password": null, "user_count": 2},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "password": "56789abcde", "user_count": 1},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "password": "456789abcd", "user_count": 0},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "password": "3456789abc", "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "password": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by name by default, all the types are by default
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "password": null, "user_count": 2},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "password": "56789abcde", "user_count": 1},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "password": "456789abcd", "user_count": 0},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "password": "3456789abc", "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "password": null, "user_count": 0},
      {"id": "29", "name": "UserSelf", "type": "UserSelf", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by name by default, all the types are included explicitly
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children?types_include=Base,Class,Team,Club,Friends,Other,UserSelf,UserAdmin"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "password": null, "user_count": 2},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "password": "56789abcde", "user_count": 1},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "password": "456789abcd", "user_count": 0},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "password": "3456789abc", "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "password": null, "user_count": 0},
      {"id": "29", "name": "UserSelf", "type": "UserSelf", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by name by default, some types are excluded
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children?types_exclude=Class,Team,Club,Friends"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "password": "3456789abc", "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "password": null, "user_count": 0},
      {"id": "29", "name": "UserSelf", "type": "UserSelf", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by grade, UserSelf is skipped
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children?sort=grade&types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "password": null, "user_count": 2},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "password": null, "user_count": 0},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "password": "3456789abc", "user_count": 0},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "password": "456789abcd", "user_count": 0},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "password": "56789abcde", "user_count": 1},
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by type, UserSelf is skipped
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children?sort=type&types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "password": null, "user_count": 2},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "password": "456789abcd", "user_count": 0},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "password": "56789abcde", "user_count": 1},
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "password": null, "user_count": 0},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "password": "3456789abc", "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by name by default, limit applied
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "password": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, paging applied, UserSelf is skipped
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children?from.name=RootSelf&from.id=30&types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "password": null, "user_count": 0}
    ]
    """
