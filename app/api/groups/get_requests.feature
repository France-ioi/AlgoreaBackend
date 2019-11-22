Feature: Get requests for group_id
  Background:
    Given the database has the following users:
      | login | temp_user | group_id | owned_group_id | first_name  | last_name | grade |
      | owner | 0         | 21       | 22             | Jean-Michel | Blanquer  | 3     |
      | user  | 0         | 11       | 12             | John        | Doe       | 1     |
      | jane  | 0         | 31       | 32             | Jane        | Doe       | 2     |
    And the database has the following table 'groups':
      | id  |
      | 13  |
      | 14  |
      | 111 |
      | 121 |
      | 122 |
      | 123 |
      | 124 |
      | 131 |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 75 | 22                | 13             | 0       |
      | 76 | 13                | 11             | 0       |
      | 77 | 22                | 11             | 0       |
      | 78 | 21                | 21             | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 1  | 13              | 21             |
      | 2  | 13              | 11             |
      | 3  | 13              | 31             |
      | 4  | 13              | 22             |
      | 5  | 14              | 11             |
      | 6  | 14              | 31             |
      | 7  | 14              | 21             |
      | 8  | 14              | 22             |
      | 9  | 13              | 121            |
      | 10 | 13              | 111            |
      | 11 | 13              | 131            |
      | 12 | 13              | 122            |
      | 13 | 13              | 123            |
      | 14 | 13              | 124            |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         | at                        |
      | 13       | 21        | invitation   | {{relativeTime("-170h")}} |
      | 13       | 31        | join_request | {{relativeTime("-168h")}} |
      | 14       | 11        | invitation   | 2017-05-28 06:38:38       |
      | 14       | 21        | join_request | 2017-05-27 06:38:38       |
    And the database has the following table 'group_membership_changes':
      | group_id | member_id | action                | at                        | initiator_id |
      | 13       | 21        | invitation_created    | {{relativeTime("-170h")}} | 11           |
      | 13       | 11        | invitation_refused    | {{relativeTime("-169h")}} | null         |
      | 13       | 31        | join_request_created  | {{relativeTime("-168h")}} | 21           |
      | 13       | 22        | join_request_refused  | {{relativeTime("-167h")}} | 11           |
      | 14       | 11        | invitation_created    | 2017-05-28 06:38:38       | 11           |
      | 14       | 31        | invitation_refused    | 2017-05-26 06:38:38       | 31           |
      | 14       | 21        | join_request_created  | 2017-05-27 06:38:38       | 21           |
      | 14       | 22        | join_request_refused  | 2017-05-24 06:38:38       | 11           |
      | 13       | 121       | invitation_accepted   | 2017-05-29 06:38:38       | 11           |
      | 13       | 111       | join_request_accepted | 2017-05-23 06:38:38       | 31           |
      | 13       | 131       | removed               | 2017-05-22 06:38:38       | 21           |
      | 13       | 122       | left                  | 2017-05-21 06:38:38       | 11           |
      | 13       | 123       | added_directly        | 2017-05-20 06:38:38       | 11           |
      | 13       | 124       | joined_by_code        | 2017-05-19 06:38:38       | 11           |

  Scenario: User is an admin of the group (default sort)
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/requests"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "4",
        "inviting_user": {
          "first_name": "John",
          "group_id": "11",
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": null,
        "type_changed_at": "{{timeToRFC(db("groups_groups[4][type_changed_at]"))}}",
        "type": "requestRefused"
      },
      {
        "id": "3",
        "inviting_user": {
          "first_name": "Jean-Michel",
          "group_id": "21",
          "last_name": "Blanquer",
          "login": "owner"
        },
        "joining_user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "type_changed_at": "{{timeToRFC(db("groups_groups[3][type_changed_at]"))}}",
        "type": "requestSent"
      },
      {
        "id": "2",
        "inviting_user": null,
        "joining_user": {
          "first_name": null,
          "grade": 1,
          "group_id": "11",
          "last_name": null,
          "login": "user"
        },
        "type_changed_at": "{{timeToRFC(db("groups_groups[2][type_changed_at]"))}}",
        "type": "invitationRefused"
      },
      {
        "id": "1",
        "inviting_user": {
          "first_name": "John",
          "group_id": "11",
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {
          "first_name": null,
          "grade": 3,
          "group_id": "21",
          "last_name": null,
          "login": "owner"
        },
        "type_changed_at": "{{timeToRFC(db("groups_groups[1][type_changed_at]"))}}",
        "type": "invitationSent"
      }
    ]
    """

  Scenario: User is an admin of the group (sort by type)
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/requests?sort=type,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "1",
        "inviting_user": {
          "first_name": "John",
          "group_id": "11",
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {
          "first_name": null,
          "grade": 3,
          "group_id": "21",
          "last_name": null,
          "login": "owner"
        },
        "type_changed_at": "{{timeToRFC(db("groups_groups[1][type_changed_at]"))}}",
        "type": "invitationSent"
      },
      {
        "id": "3",
        "inviting_user": {
          "first_name": "Jean-Michel",
          "group_id": "21",
          "last_name": "Blanquer",
          "login": "owner"
        },
        "joining_user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "type_changed_at": "{{timeToRFC(db("groups_groups[3][type_changed_at]"))}}",
        "type": "requestSent"
      },
      {
        "id": "2",
        "inviting_user": null,
        "joining_user": {
          "first_name": null,
          "grade": 1,
          "group_id": "11",
          "last_name": null,
          "login": "user"
        },
        "type_changed_at": "{{timeToRFC(db("groups_groups[2][type_changed_at]"))}}",
        "type": "invitationRefused"
      },
      {
        "id": "4",
        "inviting_user": {
          "first_name": "John",
          "group_id": "11",
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": null,
        "type_changed_at": "{{timeToRFC(db("groups_groups[4][type_changed_at]"))}}",
        "type": "requestRefused"
      }
    ]
    """

  Scenario: User is an admin of the group (sort by joining user's login)
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/requests?sort=joining_user.login,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "4",
        "inviting_user": {
          "first_name": "John",
          "group_id": "11",
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": null,
        "type_changed_at": "{{timeToRFC(db("groups_groups[4][type_changed_at]"))}}",
        "type": "requestRefused"
      },
      {
        "id": "3",
        "inviting_user": {
          "first_name": "Jean-Michel",
          "group_id": "21",
          "last_name": "Blanquer",
          "login": "owner"
        },
        "joining_user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "type_changed_at": "{{timeToRFC(db("groups_groups[3][type_changed_at]"))}}",
        "type": "requestSent"
      },
      {
        "id": "1",
        "inviting_user": {
          "first_name": "John",
          "group_id": "11",
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {
          "first_name": null,
          "grade": 3,
          "group_id": "21",
          "last_name": null,
          "login": "owner"
        },
        "type_changed_at": "{{timeToRFC(db("groups_groups[1][type_changed_at]"))}}",
        "type": "invitationSent"
      },
      {
        "id": "2",
        "inviting_user": null,
        "joining_user": {
          "first_name": null,
          "grade": 1,
          "group_id": "11",
          "last_name": null,
          "login": "user"
        },
        "type_changed_at": "{{timeToRFC(db("groups_groups[2][type_changed_at]"))}}",
        "type": "invitationRefused"
      }
    ]
    """

  Scenario: User is an admin of the group; request the first row
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/requests?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "4",
        "inviting_user": {
          "first_name": "John",
          "group_id": "11",
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": null,
        "type_changed_at": "{{timeToRFC(db("groups_groups[4][type_changed_at]"))}}",
        "type": "requestRefused"
      }
    ]
    """

  Scenario: User is an admin of the group; filter out old rejections
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/requests?rejections_within_weeks=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "4",
        "inviting_user": {
          "first_name": "John",
          "group_id": "11",
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": null,
        "type_changed_at": "{{timeToRFC(db("groups_groups[4][type_changed_at]"))}}",
        "type": "requestRefused"
      },
      {
        "id": "3",
        "inviting_user": {
          "first_name": "Jean-Michel",
          "group_id": "21",
          "last_name": "Blanquer",
          "login": "owner"
        },
        "joining_user": {
          "first_name": "Jane",
          "grade": 2,
          "group_id": "31",
          "last_name": "Doe",
          "login": "jane"
        },
        "type_changed_at": "{{timeToRFC(db("groups_groups[3][type_changed_at]"))}}",
        "type": "requestSent"
      },
      {
        "id": "1",
        "inviting_user": {
          "first_name": "John",
          "group_id": "11",
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {
          "first_name": null,
          "grade": 3,
          "group_id": "21",
          "last_name": null,
          "login": "owner"
        },
        "type_changed_at": "{{timeToRFC(db("groups_groups[1][type_changed_at]"))}}",
        "type": "invitationSent"
      }
    ]
    """
