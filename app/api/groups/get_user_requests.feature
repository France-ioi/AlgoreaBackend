Feature: Get pending requests for managed groups
  Background:
    Given the database has the following users:
      | login   | temp_user | group_id | first_name  | last_name | grade |
      | owner   | 0         | 21       | Jean-Michel | Blanquer  | 3     |
      | user    | 0         | 11       | John        | Doe       | 1     |
      | jane    | 0         | 31       | Jane        | Doe       | 2     |
      | richard | 0         | 41       | Richard     | Roe       | 2     |
    And the database has the following table 'groups':
      | id  | name       |
      | 1   | Root       |
      | 13  | Class      |
      | 14  | Friends    |
      | 22  | Group      |
      | 23  | Club       |
      | 111 | Subgroup 1 |
      | 121 | Subgroup 2 |
      | 122 | Subgroup 3 |
      | 123 | Subgroup 4 |
      | 124 | Subgroup 5 |
      | 131 | Subgroup 6 |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 1        | 21         | memberships |
      | 13       | 21         | memberships |
      | 13       | 31         | none        |
      | 14       | 31         | memberships |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 1               | 13             | null                           |
      | 1               | 14             | null                           |
      | 1               | 21             | 2019-05-30 11:00:00            |
      | 13              | 21             | null                           |
      | 13              | 11             | null                           |
      | 13              | 31             | 2019-05-30 11:00:00            |
      | 13              | 22             | null                           |
      | 14              | 11             | null                           |
      | 14              | 31             | null                           |
      | 14              | 21             | null                           |
      | 14              | 22             | null                           |
      | 13              | 121            | null                           |
      | 13              | 111            | null                           |
      | 13              | 131            | null                           |
      | 13              | 122            | null                           |
      | 13              | 123            | null                           |
      | 13              | 124            | null                           |
    And the groups ancestors are computed
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type          | at                          | personal_info_view_approved |
      | 13       | 21        | invitation    | {{relativeTimeDB("-170h")}} | true                        |
      | 13       | 31        | join_request  | {{relativeTimeDB("-168h")}} | false                       |
      | 13       | 41        | join_request  | {{relativeTimeDB("-169h")}} | true                        |
      | 13       | 11        | leave_request | {{relativeTimeDB("-171h")}} | false                       |
      | 14       | 11        | invitation    | 2017-05-28 06:38:38         | false                       |
      | 14       | 21        | join_request  | 2017-05-27 06:38:38         | false                       |
      | 14       | 31        | leave_request | 2017-05-27 06:38:38         | false                       |
      | 23       | 21        | join_request  | 2017-05-27 06:38:38         | true                        |

  Scenario: group_id is given (default sort)
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[2][at]"))}}",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Richard",
          "grade": 2,
          "group_id": "41",
          "last_name": "Roe",
          "login": "richard"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[3][at]"))}}",
        "type": "join_request"
      }
    ]
    """

  Scenario: group_id is given, include descendant groups (default sort)
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=1&include_descendant_groups=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[2][at]"))}}",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Richard",
          "grade": 2,
          "group_id": "41",
          "last_name": "Roe",
          "login": "richard"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[3][at]"))}}",
        "type": "join_request"
      },
      {
        "group": {
          "id": "14",
          "name": "Friends"
        },
        "user": {
          "first_name": "Jean-Michel",
          "grade": 3,
          "group_id": "21",
          "last_name": "Blanquer",
          "login": "owner"
        },
        "at": "2017-05-27T06:38:38Z",
        "type": "join_request"
      }
    ]
    """

  Scenario: group_id is given, include descendant groups (sort by group name desc & login)
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=1&include_descendant_groups=1&sort=-group.name,user.login"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "14",
          "name": "Friends"
        },
        "user": {
          "first_name": "Jean-Michel",
          "grade": 3,
          "group_id": "21",
          "last_name": "Blanquer",
          "login": "owner"
        },
        "at": "2017-05-27T06:38:38Z",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[2][at]"))}}",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Richard",
          "grade": 2,
          "group_id": "41",
          "last_name": "Roe",
          "login": "richard"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[3][at]"))}}",
        "type": "join_request"
      }
    ]
    """

  Scenario: group_id is given (sort by joining user's login)
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=13&sort=user.login,user.group_id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[2][at]"))}}",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Richard",
          "grade": 2,
          "group_id": "41",
          "last_name": "Roe",
          "login": "richard"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[3][at]"))}}",
        "type": "join_request"
      }
    ]
    """

  Scenario: group_id is given, include descendant groups (sort by joining user's login desc)
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=1&include_descendant_groups=1&sort=-user.login,user.group_id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Richard",
          "grade": 2,
          "group_id": "41",
          "last_name": "Roe",
          "login": "richard"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[3][at]"))}}",
        "type": "join_request"
      },
      {
        "group": {
          "id": "14",
          "name": "Friends"
        },
        "user": {
          "first_name": "Jean-Michel",
          "grade": 3,
          "group_id": "21",
          "last_name": "Blanquer",
          "login": "owner"
        },
        "at": "2017-05-27T06:38:38Z",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[2][at]"))}}",
        "type": "join_request"
      }
    ]
    """

  Scenario: group_id is given; request the first row
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=13&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[2][at]"))}}",
        "type": "join_request"
      }
    ]
    """

  Scenario: group_id is given, include descendant groups (sort by group name desc & login, start from the second row)
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=1&include_descendant_groups=1&sort=-group.name,user.login&from.group.id=14&from.user.group_id=21"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[2][at]"))}}",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Richard",
          "grade": 2,
          "group_id": "41",
          "last_name": "Roe",
          "login": "richard"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[3][at]"))}}",
        "type": "join_request"
      }
    ]
    """

  Scenario: group_id is not given (sort by group name desc & login)
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?sort=-group.name,user.login"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "14",
          "name": "Friends"
        },
        "user": {
          "first_name": "Jean-Michel",
          "grade": 3,
          "group_id": "21",
          "last_name": "Blanquer",
          "login": "owner"
        },
        "at": "2017-05-27T06:38:38Z",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[2][at]"))}}",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Richard",
          "grade": 2,
          "group_id": "41",
          "last_name": "Roe",
          "login": "richard"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[3][at]"))}}",
        "type": "join_request"
      }
    ]
    """

  Scenario: group_id is not given (sort by group name desc & login, start from the second row)
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?sort=-group.name,user.login&from.group.id=14&from.user.group_id=21"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[2][at]"))}}",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Richard",
          "grade": 2,
          "group_id": "41",
          "last_name": "Roe",
          "login": "richard"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[3][at]"))}}",
        "type": "join_request"
      }
    ]
    """

  Scenario: group_id is not given, another user (sort by group name desc & login)
    Given I am the user with id "31"
    When I send a GET request to "/groups/user-requests?types=join_request,leave_request&sort=-group.name,user.login"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "14",
          "name": "Friends"
        },
        "user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "at": "2017-05-27T06:38:38Z",
        "type": "leave_request"
      },
      {
        "group": {
          "id": "14",
          "name": "Friends"
        },
        "user": {
          "grade": 3,
          "group_id": "21",
          "login": "owner"
        },
        "at": "2017-05-27T06:38:38Z",
        "type": "join_request"
      }
    ]
    """

  Scenario: group_id is given, types=leave_request (default sort)
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=13&types=leave_request"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "grade": 1,
          "group_id": "11",
          "login": "user"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[4][at]"))}}",
        "type": "leave_request"
      }
    ]
    """

  Scenario: group_id is given, types=leave_request,join_request (default sort)
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=13&types=leave_request,join_request"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[2][at]"))}}",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "first_name": "Richard",
          "grade": 2,
          "group_id": "41",
          "last_name": "Roe",
          "login": "richard"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[3][at]"))}}",
        "type": "join_request"
      },
      {
        "group": {
          "id": "13",
          "name": "Class"
        },
        "user": {
          "grade": 1,
          "group_id": "11",
          "login": "user"
        },
        "at": "{{timeDBToRFC(db("group_pending_requests[4][at]"))}}",
        "type": "leave_request"
      }
    ]
    """
