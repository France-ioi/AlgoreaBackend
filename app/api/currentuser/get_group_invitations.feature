Feature: Get group invitations for the current user
  Background:
    Given the database has the following table 'groups':
      | id | type    | name                          | description                   | require_personal_info_access_approval | require_lock_membership_approval_until | require_watch_approval |
      | 1  | Class   | Our Class                     | Our class group               | none                                  | {{relativeTimeDB("+96h")}}             | true                   |
      | 2  | Team    | Our Team                      | Our team group                | none                                  | null                                   | false                  |
      | 3  | Club    | Our Club                      | Our club group                | none                                  | null                                   | false                  |
      | 4  | Friends | Our Friends                   | Group for our friends         | none                                  | null                                   | false                  |
      | 5  | Other   | Other people                  | Group for other people        | none                                  | null                                   | false                  |
      | 6  | Class   | Another Class                 | Another class group           | none                                  | null                                   | false                  |
      | 7  | Team    | Another Team                  | Another team group            | none                                  | null                                   | false                  |
      | 8  | Club    | Another Club                  | Another club group            | none                                  | null                                   | false                  |
      | 9  | Friends | Some other friends            | Another friends group         | none                                  | null                                   | false                  |
      | 10 | Other   | Secret group                  | Our secret group              | none                                  | null                                   | false                  |
      | 11 | Club    | Secret club                   | Our secret club               | none                                  | null                                   | false                  |
      | 12 | User    | user self                     |                               | none                                  | null                                   | false                  |
      | 13 | User    | another user                  |                               | none                                  | null                                   | false                  |
      | 21 | User    | owner self                    |                               | none                                  | null                                   | false                  |
      | 33 | Class   | Other group with invitation   | Other group with invitation   | view                                  | null                                   | false                  |
      | 34 | Class   | Other group with invitation 2 | Other group with invitation 2 | edit                                  | null                                   | false                  |
      | 35 | Class   | Group with broken change log  | Group with broken change log  | edit                                  | null                                   | false                  |
      | 36 | Class   | Group without inviting user   | Group without inviting user   | edit                                  | null                                   | false                  |
    And the database has the following table 'users':
      | login       | temp_user | group_id | first_name  | last_name | grade |
      | owner       | 0         | 21       | Jean-Michel | Blanquer  | 3     |
      | user        | 0         | 12       | John        | Doe       | 1     |
      | anotheruser | 0         | 13       | Another     | User      | 1     |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 5               | 21             |
      | 6               | 21             |
      | 9               | 21             |
      | 10              | 21             |
    And the time now is "2020-01-01T01:00:00.001Z"
    And the DB time now is "2020-01-01 01:00:00.001"
    And the database has the following table 'group_membership_changes':
      | group_id | member_id | action                | at                            | initiator_id |
      | 1        | 21        | invitation_created    | {{relativeTimeDBMs("-169h")}} | 12           |
      | 2        | 21        | invitation_refused    | {{relativeTimeDBMs("-168h")}} | 21           |
      | 3        | 21        | join_request_created  | {{relativeTimeDBMs("-167h")}} | 21           |
      | 33       | 21        | invitation_created    | {{relativeTimeDBMs("-167h")}} | 13           |
      | 4        | 21        | join_request_refused  | {{relativeTimeDBMs("-166h")}} | 12           |
      | 34       | 21        | invitation_created    | {{relativeTimeDBMs("-190h")}} | 13           |
      | 34       | 21        | invitation_refused    | {{relativeTimeDBMs("-180h")}} | 21           |
      | 34       | 21        | invitation_created    | {{relativeTimeDBMs("-166h")}} | 12           |
      | 35       | 21        | invitation_accepted   | {{relativeTimeDBMs("-186h")}} | 12           |
      | 36       | 21        | invitation_created    | {{relativeTimeDBMs("-200h")}} | null         |
      | 5        | 21        | invitation_accepted   | {{relativeTimeDBMs("-165h")}} | 12           |
      | 6        | 21        | join_request_accepted | {{relativeTimeDBMs("-164h")}} | 12           |
      | 7        | 21        | removed               | {{relativeTimeDBMs("-163h")}} | 21           |
      | 8        | 21        | left                  | {{relativeTimeDBMs("-162h")}} | 21           |
      | 9        | 21        | added_directly        | {{relativeTimeDBMs("-161h")}} | 12           |
      | 1        | 12        | invitation_created    | {{relativeTimeDBMs("-170h")}} | 12           |
      | 10       | 21        | joined_by_code        | {{relativeTimeDBMs("-180h")}} | null         |
      | 11       | 21        | joined_by_badge       | {{relativeTimeDBMs("-190h")}} | null         |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         | at                            |
      | 1        | 21        | invitation   | {{currentTimeDBMs()}}         |
      | 33       | 21        | invitation   | {{currentTimeDBMs()}}         |
      | 34       | 21        | invitation   | {{currentTimeDBMs()}}         |
      | 35       | 21        | invitation   | {{relativeTimeDBMs("-200h")}} |
      | 36       | 21        | invitation   | {{currentTimeDBMs()}}         |
      | 3        | 21        | join_request | {{currentTimeDBMs()}}         |
      | 1        | 12        | invitation   | {{currentTimeDBMs()}}         |

  Scenario: Show all invitations
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-invitations"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "34",
        "inviting_user": {
          "id": "12",
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "group": {
          "id": "34",
          "name": "Other group with invitation 2",
          "description": "Other group with invitation 2",
          "type": "Class",
          "require_personal_info_access_approval": "edit",
          "require_lock_membership_approval_until": null,
          "require_watch_approval": false
        },
        "at": "{{timeDBMsToRFC(db("group_membership_changes[8][at]"))}}"
      },
      {
        "group_id": "33",
        "inviting_user": {
          "id": "13",
          "first_name": "Another",
          "last_name": "User",
          "login": "anotheruser"
        },
        "group": {
          "id": "33",
          "name": "Other group with invitation",
          "description": "Other group with invitation",
          "type": "Class",
          "require_personal_info_access_approval": "view",
          "require_lock_membership_approval_until": null,
          "require_watch_approval": false
        },
        "at": "{{timeDBMsToRFC(db("group_membership_changes[4][at]"))}}"
      },
      {
        "group_id": "1",
        "inviting_user": {
          "id": "12",
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "group": {
          "id": "1",
          "name": "Our Class",
          "description": "Our class group",
          "type": "Class",
          "require_personal_info_access_approval": "none",
          "require_lock_membership_approval_until": "{{timeDBToRFC(db("groups[1][require_lock_membership_approval_until]"))}}",
          "require_watch_approval": true
        },
        "at": "{{timeDBMsToRFC(db("group_membership_changes[1][at]"))}}"
      },
      {
        "group_id": "35",
        "inviting_user": null,
        "group": {
          "id": "35",
          "name": "Group with broken change log",
          "description": "Group with broken change log",
          "type": "Class",
          "require_personal_info_access_approval": "edit",
          "require_lock_membership_approval_until": null,
          "require_watch_approval": false
        },
        "at": "{{timeDBMsToRFC(db("group_pending_requests[4][at]"))}}"
      },
      {
        "group_id": "36",
        "inviting_user": null,
        "group": {
          "id": "36",
          "name": "Group without inviting user",
          "description": "Group without inviting user",
          "type": "Class",
          "require_personal_info_access_approval": "edit",
          "require_lock_membership_approval_until": null,
          "require_watch_approval": false
        },
        "at": "{{timeDBMsToRFC(db("group_membership_changes[10][at]"))}}"
      }
    ]
    """

  Scenario: Request the first row
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-invitations?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "34",
        "inviting_user": {
          "id": "12",
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "group": {
          "id": "34",
          "name": "Other group with invitation 2",
          "description": "Other group with invitation 2",
          "type": "Class",
          "require_personal_info_access_approval": "edit",
          "require_lock_membership_approval_until": null,
          "require_watch_approval": false
        },
        "at": "{{timeDBMsToRFC(db("group_membership_changes[8][at]"))}}"
      }
    ]
    """

  Scenario: Request the second row
    Given I am the user with id "21"
    And the template constant "from_at" is "{{timeDBMsToRFC(db("group_membership_changes[4][at]"))}}"
    When I send a GET request to "/current-user/group-invitations?limit=1&from.group_id=4&from.at={{from_at}}"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "33",
        "inviting_user": {
          "id": "13",
          "first_name": "Another",
          "last_name": "User",
          "login": "anotheruser"
        },
        "group": {
          "id": "33",
          "name": "Other group with invitation",
          "description": "Other group with invitation",
          "type": "Class",
          "require_personal_info_access_approval": "view",
          "require_lock_membership_approval_until": null,
          "require_watch_approval": false
        },
        "at": "{{timeDBMsToRFC(db("group_membership_changes[4][at]"))}}"
      }
    ]
    """
