Feature: Get group invitations for the current user
  Background:
    Given the database has the following table 'groups':
      | id | type    | name                          | description                   |
      | 1  | Class   | Our Class                     | Our class group               |
      | 2  | Team    | Our Team                      | Our team group                |
      | 3  | Club    | Our Club                      | Our club group                |
      | 4  | Friends | Our Friends                   | Group for our friends         |
      | 5  | Other   | Other people                  | Group for other people        |
      | 6  | Class   | Another Class                 | Another class group           |
      | 7  | Team    | Another Team                  | Another team group            |
      | 8  | Club    | Another Club                  | Another club group            |
      | 9  | Friends | Some other friends            | Another friends group         |
      | 10 | Other   | Secret group                  | Our secret group              |
      | 11 | Club    | Secret club                   | Our secret club               |
      | 12 | User    | user self                     |                               |
      | 13 | User    | another user                  |                               |
      | 21 | User    | owner self                    |                               |
      | 33 | Class   | Other group with invitation   | Other group with invitation   |
      | 34 | Class   | Other group with invitation 2 | Other group with invitation 2 |
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
    And the database has the following table 'group_membership_changes':
      | group_id | member_id | action                | at                        | initiator_id |
      | 1        | 21        | invitation_created    | {{relativeTime("-169h")}} | 12           |
      | 2        | 21        | invitation_refused    | {{relativeTime("-168h")}} | 21           |
      | 3        | 21        | join_request_created  | {{relativeTime("-167h")}} | 21           |
      | 33       | 21        | invitation_created    | {{relativeTime("-167h")}} | 13           |
      | 4        | 21        | join_request_refused  | {{relativeTime("-166h")}} | 12           |
      | 34       | 21        | invitation_created    | {{relativeTime("-166h")}} | 12           |
      | 5        | 21        | invitation_accepted   | {{relativeTime("-165h")}} | 12           |
      | 6        | 21        | join_request_accepted | {{relativeTime("-164h")}} | 12           |
      | 7        | 21        | removed               | {{relativeTime("-163h")}} | 21           |
      | 8        | 21        | left                  | {{relativeTime("-162h")}} | 21           |
      | 9        | 21        | added_directly        | {{relativeTime("-161h")}} | 12           |
      | 1        | 12        | invitation_created    | {{relativeTime("-170h")}} | 12           |
      | 10       | 21        | joined_by_code        | {{relativeTime("-180h")}} | null         |
      | 11       | 21        | joined_by_badge       | {{relativeTime("-190h")}} | null         |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         |
      | 1        | 21        | invitation   |
      | 33       | 21        | invitation   |
      | 34       | 21        | invitation   |
      | 3        | 21        | join_request |
      | 1        | 12        | invitation   |

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
          "type": "Class"
        },
        "at": "{{timeToRFC(db("group_membership_changes[6][at]"))}}"
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
          "type": "Class"
        },
        "at": "{{timeToRFC(db("group_membership_changes[4][at]"))}}"
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
          "type": "Class"
        },
        "at": "{{timeToRFC(db("group_membership_changes[1][at]"))}}"
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
          "type": "Class"
        },
        "at": "{{timeToRFC(db("group_membership_changes[6][at]"))}}"
      }
    ]
    """

  Scenario: Request the second row
    Given I am the user with id "21"
    And the template constant "from_at" is "{{timeToRFC(db("group_membership_changes[4][at]"))}}"
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
          "type": "Class"
        },
        "at": "{{timeToRFC(db("group_membership_changes[4][at]"))}}"
      }
    ]
    """
