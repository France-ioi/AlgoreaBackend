Feature: Get group children (groupChildrenView)
  Background:
    Given the database has the following table 'groups':
      | id | name          | grade | type      | opened | free_access | code       |
      | 11 | Group A       | -3    | Class     | true   | true        | ybqybxnlyo |
      | 13 | Group B       | -2    | Class     | true   | true        | ybabbxnlyo |
      | 14 | Group C       | -4    | UserAdmin | true   | false       | null       |
      | 21 | user-admin    | -2    | UserAdmin | true   | false       | null       |
      | 22 | C's Child     | -4    | UserAdmin | true   | false       | null       |
      | 23 | Our Class     | -3    | Class     | true   | false       | null       |
      | 24 | Root          | -2    | Base      | true   | false       | 3456789abc |
      | 25 | Our Team      | -1    | Team      | true   | false       | 456789abcd |
      | 26 | Our Club      | 0     | Club      | true   | false       | null       |
      | 27 | Our Friends   | 0     | Friends   | true   | false       | 56789abcde |
      | 28 | Other         | 0     | Other     | true   | false       | null       |
      | 29 | UserSelf      | 0     | UserSelf  | true   | false       | null       |
      | 30 | RootSelf      | 0     | Base      | true   | false       | null       |
      | 31 | RootAdmin     | 0     | Base      | true   | false       | null       |
      | 42 | Their Class   | -3    | Class     | true   | false       | null       |
      | 43 | Other Root    | -2    | Base      | true   | false       | 3567894abc |
      | 44 | Other Team    | -1    | Team      | true   | false       | 678934abcd |
      | 45 | Their Club    | 0     | Club      | true   | false       | null       |
      | 46 | Their Friends | 0     | Friends   | true   | false       | 98765abcde |
      | 47 | Other         | 0     | Other     | true   | false       | null       |
      | 51 | john          | 0     | UserSelf  | false  | false       | null       |
      | 52 | john-admin    | 0     | UserAdmin | false  | false       | null       |
      | 53 | jane          | 0     | UserSelf  | false  | false       | null       |
      | 54 | jane-admin    | 0     | UserAdmin | false  | false       | null       |
      | 90 | Sub-Class     | 0     | Team      | false  | false       | null       |
    And the database has the following table 'users':
      | login | group_id | owned_group_id | first_name  | last_name |
      | owner | 21       | 22             | Jean-Michel | Blanquer  |
      | john  | 51       | 52             | John        | Doe       |
      | jane  | 53       | 54             | Jane        | Doe       |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 11             | 0       |
      | 13                | 13             | 1       |
      | 13                | 21             | 0       |
      | 13                | 23             | 0       |
      | 13                | 24             | 0       |
      | 13                | 25             | 0       |
      | 13                | 26             | 0       |
      | 13                | 27             | 0       |
      | 13                | 28             | 0       |
      | 13                | 29             | 0       |
      | 13                | 30             | 0       |
      | 13                | 31             | 0       |
      | 13                | 51             | 0       |
      | 13                | 53             | 0       |
      | 13                | 90             | 0       |
      | 14                | 14             | 1       |
      | 14                | 22             | 0       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 21             | 0       |
      | 22                | 22             | 1       |
      | 22                | 23             | 0       |
      | 22                | 24             | 0       |
      | 22                | 25             | 0       |
      | 22                | 26             | 0       |
      | 22                | 27             | 0       |
      | 22                | 28             | 0       |
      | 22                | 29             | 0       |
      | 22                | 30             | 0       |
      | 22                | 31             | 0       |
      | 22                | 51             | 0       |
      | 22                | 53             | 0       |
      | 22                | 90             | 0       |
      | 23                | 23             | 1       |
      | 23                | 51             | 0       |
      | 23                | 53             | 0       |
      | 23                | 90             | 0       |
      | 24                | 24             | 1       |
      | 25                | 25             | 1       |
      | 25                | 53             | 0       |
      | 26                | 26             | 1       |
      | 27                | 27             | 1       |
      | 27                | 53             | 0       |
      | 28                | 28             | 1       |
      | 29                | 29             | 1       |
      | 30                | 30             | 1       |
      | 31                | 31             | 1       |
      | 42                | 42             | 1       |
      | 43                | 43             | 1       |
      | 44                | 44             | 1       |
      | 45                | 45             | 1       |
      | 46                | 46             | 1       |
      | 47                | 47             | 1       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
      | 53                | 53             | 1       |
      | 54                | 54             | 1       |
      | 90                | 51             | 0       |
      | 90                | 53             | 0       |
      | 90                | 90             | 1       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type               |
      | 13              | 21             | invitationAccepted |
      | 13              | 23             | direct             |
      | 13              | 24             | direct             |
      | 14              | 22             | direct             |
      | 13              | 25             | direct             |
      | 13              | 26             | direct             |
      | 13              | 27             | direct             |
      | 13              | 28             | direct             |
      | 13              | 29             | requestAccepted    |
      | 13              | 30             | direct             |
      | 13              | 31             | direct             |
      | 13              | 42             | invitationSent     |
      | 13              | 43             | requestSent        |
      | 13              | 44             | invitationRefused  |
      | 13              | 45             | requestRefused     |
      | 13              | 46             | left               |
      | 13              | 47             | removed            |
      | 23              | 51             | invitationAccepted |
      | 23              | 90             | direct             |
      | 25              | 53             | joinedByCode       |
      | 27              | 53             | invitationAccepted |
      | 90              | 51             | requestAccepted    |

  Scenario: User is an owner of the parent group, rows are sorted by name by default, UserSelf is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "code": null, "user_count": 2},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "code": "56789abcde", "user_count": 1},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "code": "456789abcd", "user_count": 1},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by name by default, all the types are by default
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "code": null, "user_count": 2},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "code": "56789abcde", "user_count": 1},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "code": "456789abcd", "user_count": 1},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "code": null, "user_count": 0},
      {"id": "29", "name": "UserSelf", "type": "UserSelf", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by name by default, all the types are included explicitly
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?types_include=Base,Class,Team,Club,Friends,Other,UserSelf,UserAdmin"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "code": null, "user_count": 2},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "code": "56789abcde", "user_count": 1},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "code": "456789abcd", "user_count": 1},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "code": null, "user_count": 0},
      {"id": "29", "name": "UserSelf", "type": "UserSelf", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by name by default, some types are excluded
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?types_exclude=Class,Team,Club,Friends"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "code": null, "user_count": 0},
      {"id": "29", "name": "UserSelf", "type": "UserSelf", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by grade, UserSelf is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?sort=grade&types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "code": null, "user_count": 2},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "code": null, "user_count": 0},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "code": "456789abcd", "user_count": 1},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "code": "56789abcde", "user_count": 1},
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by type, UserSelf is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?sort=type&types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "23", "name": "Our Class", "type": "Class", "free_access": false, "grade": -3, "opened": true, "code": null, "user_count": 2},
      {"id": "25", "name": "Our Team", "type": "Team", "free_access": false, "grade": -1, "opened": true, "code": "456789abcd", "user_count": 1},
      {"id": "26", "name": "Our Club", "type": "Club", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "27", "name": "Our Friends", "type": "Friends", "free_access": false, "grade": 0, "opened": true, "code": "56789abcde", "user_count": 1},
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "code": null, "user_count": 0},
      {"id": "24", "name": "Root", "type": "Base", "free_access": false, "grade": -2, "opened": true, "code": "3456789abc", "user_count": 0},
      {"id": "30", "name": "RootSelf", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0},
      {"id": "31", "name": "RootAdmin", "type": "Base", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, rows are sorted by name by default, limit applied
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "28", "name": "Other", "type": "Other", "free_access": false, "grade": 0, "opened": true, "code": null, "user_count": 0}
    ]
    """

  Scenario: User is an owner of the parent group, paging applied, UserSelf is skipped
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?from.name=RootSelf&from.id=30&types_exclude=UserSelf"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "21", "name": "user-admin", "type": "UserAdmin", "free_access": false, "grade": -2, "opened": true, "code": null, "user_count": 0}
    ]
    """
