Feature: Get group children (groupChildrenView)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | sDefaultLanguage |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | fr               |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 75 | 22              | 13           | 0       | 0        |
      | 76 | 13              | 11           | 0       | 0        |
      | 77 | 22              | 14           | 0       | 0        |
    And the database has the following table 'groups':
      | ID | sName       | iGrade | sType     | bOpened | bFreeAccess | sPassword  |
      | 11 | Group A     | -3     | Class     | true    | true        | ybqybxnlyo |
      | 13 | Group B     | -2     | Class     | true    | true        | ybabbxnlyo |
      | 14 | Group C     | -4     | UserAdmin | true    | false       | null       |
      | 21 | user-admin  | -2     | UserAdmin | true    | false       | null       |
      | 22 | C's Child   | -4     | UserAdmin | true    | false       | null       |
      | 23 | Our Class   | -3     | Class     | true    | false       | null       |
      | 24 | Root        | -2     | Root      | true    | false       | 3456789abc |
      | 25 | Our Team    | -1     | Team      | true    | false       | 456789abcd |
      | 26 | Our Club    |  0     | Club      | true    | false       | null       |
      | 27 | Our Friends |  0     | Friends   | true    | false       | 56789abcde |
      | 28 | Other       |  0     | Other     | true    | false       | null       |
      | 29 | UserSelf    |  0     | UserSelf  | true    | false       | null       |
      | 30 | RootSelf    |  0     | RootSelf  | true    | false       | null       |
      | 31 | RootAdmin   |  0     | RootAdmin | true    | false       | null       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild |
      | 75 | 13            | 21           |
      | 77 | 13            | 23           |
      | 78 | 13            | 24           |
      | 79 | 14            | 22           |
      | 80 | 13            | 25           |
      | 81 | 13            | 26           |
      | 82 | 13            | 27           |
      | 83 | 13            | 28           |
      | 84 | 13            | 29           |
      | 85 | 13            | 30           |
      | 86 | 13            | 31           |

  Scenario: User is an owner of the parent group, rows are sorted by name by default, UserSelf is skipped
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": 28, "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true},
      {"id": 23, "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true},
      {"id": 26, "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true},
      {"id": 27, "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "password": "56789abcde"},
      {"id": 25, "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "password": "456789abcd"},
      {"id": 24, "name": "Root", "type": "Root", "free_access": false, "grade": -2, "opened": true, "password": "3456789abc"},
      {"id": 31, "name": "RootAdmin", "type": "RootAdmin", "free_access": false, "grade": 0, "opened": true},
      {"id": 30, "name": "RootSelf", "type": "RootSelf", "free_access": false, "grade": 0, "opened": true},
      {"id": 21, "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by grade, UserSelf is skipped
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children?sort=grade"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": 23, "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true},
      {"id": 21, "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true},
      {"id": 24, "name": "Root", "type": "Root", "free_access": false, "grade": -2, "opened": true, "password": "3456789abc"},
      {"id": 25, "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "password": "456789abcd"},
      {"id": 26, "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true},
      {"id": 27, "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "password": "56789abcde"},
      {"id": 28, "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true},
      {"id": 30, "name": "RootSelf", "type": "RootSelf", "free_access": false, "grade": 0, "opened": true},
      {"id": 31, "name": "RootAdmin", "type": "RootAdmin", "free_access": false, "grade": 0, "opened": true}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by type, UserSelf is skipped
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children?sort=type"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": 24, "name": "Root", "type": "Root", "free_access": false, "grade": -2, "opened": true, "password": "3456789abc"},
      {"id": 23, "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true},
      {"id": 25, "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "password": "456789abcd"},
      {"id": 26, "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true},
      {"id": 27, "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "password": "56789abcde"},
      {"id": 28, "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true},
      {"id": 21, "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true},
      {"id": 30, "name": "RootSelf", "type": "RootSelf", "free_access": false, "grade": 0, "opened": true},
      {"id": 31, "name": "RootAdmin", "type": "RootAdmin", "free_access": false, "grade": 0, "opened": true}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by name by default, limit applied
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": 28, "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true}
    ]
    """

  Scenario: User is an owner of the parent group, paging applied
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/children?from.name=RootSelf&from.id=30"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": 21, "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true}
    ]
    """
