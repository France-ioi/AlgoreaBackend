Feature: Get group memberships history for the current user
  Background:
    Given the database has the following table "groups":
      | id | type    | name               |
      | 1  | Class   | Our Class          |
      | 2  | Team    | Our Team           |
      | 3  | Club    | Our Club           |
      | 4  | Friends | Our Friends        |
      | 5  | Other   | Other people       |
      | 6  | Class   | Another Class      |
      | 7  | Team    | Another Team       |
      | 8  | Club    | Another Club       |
      | 9  | Friends | Some other friends |
      | 11 | User    | user               |
      | 13 | User    | jane               |
      | 21 | User    | owner              |
    And the database has the following table "users":
      | login | group_id | first_name  | last_name | grade | notifications_read_at |
      | owner | 21       | Jean-Michel | Blanquer  | 3     | 2017-06-29 06:38:38   |
      | user  | 11       | John        | Doe       | 1     | null                  |
      | jane  | 13       | Jane        | Doe       | 2     | 2019-06-29 06:38:38   |
    And the database has the following table "group_membership_changes":
      | group_id | member_id | action                | at                      |
      | 1        | 21        | invitation_created    | 2017-02-28 06:38:38.001 |
      | 2        | 21        | invitation_refused    | 2017-03-29 06:38:38.001 |
      | 3        | 21        | join_request_created  | 2017-04-29 06:38:38.001 |
      | 4        | 21        | join_request_refused  | 2017-05-29 06:38:38.001 |
      | 5        | 21        | invitation_accepted   | 2017-06-29 06:38:38.001 |
      | 6        | 21        | join_request_accepted | 2017-07-29 06:38:38.001 |
      | 7        | 21        | removed               | 2017-08-29 06:38:38.001 |
      | 8        | 21        | left                  | 2017-09-29 06:38:38.001 |
      | 9        | 21        | added_directly        | 2017-10-29 06:38:38.001 |
      | 1        | 11        | invitation_created    | 2016-02-28 06:38:38.001 |
      | 2        | 11        | invitation_refused    | 2016-03-29 06:38:38.001 |
      | 3        | 11        | join_request_created  | 2016-04-29 06:38:38.001 |
      | 4        | 11        | join_request_refused  | 2016-05-29 06:38:38.001 |
      | 5        | 11        | invitation_accepted   | 2016-06-29 06:38:38.001 |
      | 6        | 11        | join_request_accepted | 2016-07-29 06:38:38.001 |
      | 7        | 11        | removed               | 2016-08-29 06:38:38.001 |
      | 8        | 11        | left                  | 2016-09-29 06:38:38.001 |
      | 9        | 11        | added_directly        | 2016-10-29 06:38:38.001 |
      | 1        | 13        | invitation_created    | 2016-02-28 06:38:38.001 |
      | 2        | 13        | invitation_refused    | 2016-03-29 06:38:38.001 |
      | 3        | 13        | join_request_created  | 2016-04-29 06:38:38.001 |

  Scenario: Show all the history (with notifications_read_at set)
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-memberships-history"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "8",
          "name": "Another Club",
          "type": "Club"
        },
        "at": "2017-09-29T06:38:38.001Z",
        "action": "left"
      },
      {
        "group": {
          "id": "7",
          "name": "Another Team",
          "type": "Team"
        },
        "at": "2017-08-29T06:38:38.001Z",
        "action": "removed"
      },
      {
        "group": {
          "id": "6",
          "name": "Another Class",
          "type": "Class"
        },
        "at": "2017-07-29T06:38:38.001Z",
        "action": "join_request_accepted"
      },
      {
        "group": {
          "id": "5",
          "name": "Other people",
          "type": "Other"
        },
        "at": "2017-06-29T06:38:38.001Z",
        "action": "invitation_accepted"
      }
    ]
    """

  Scenario: Show all the history in reverse order (with notifications_read_at set)
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-memberships-history?sort=at,group_id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "5",
          "name": "Other people",
          "type": "Other"
        },
        "at": "2017-06-29T06:38:38.001Z",
        "action": "invitation_accepted"
      },
      {
        "group": {
          "id": "6",
          "name": "Another Class",
          "type": "Class"
        },
        "at": "2017-07-29T06:38:38.001Z",
        "action": "join_request_accepted"
      },
      {
        "group": {
          "id": "7",
          "name": "Another Team",
          "type": "Team"
        },
        "at": "2017-08-29T06:38:38.001Z",
        "action": "removed"
      },
      {
        "group": {
          "id": "8",
          "name": "Another Club",
          "type": "Club"
        },
        "at": "2017-09-29T06:38:38.001Z",
        "action": "left"
      }
    ]
    """

  Scenario: Show all the history (without notifications_read_at set)
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships-history"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "8",
          "name": "Another Club",
          "type": "Club"
        },
        "at": "2016-09-29T06:38:38.001Z",
        "action": "left"
      },
      {
        "group": {
          "id": "7",
          "name": "Another Team",
          "type": "Team"
        },
        "at": "2016-08-29T06:38:38.001Z",
        "action": "removed"
      },
      {
        "group": {
          "id": "6",
          "name": "Another Class",
          "type": "Class"
        },
        "at": "2016-07-29T06:38:38.001Z",
        "action": "join_request_accepted"
      },
      {
        "group": {
          "id": "5",
          "name": "Other people",
          "type": "Other"
        },
        "at": "2016-06-29T06:38:38.001Z",
        "action": "invitation_accepted"
      },
      {
        "group": {
          "id": "4",
          "name": "Our Friends",
          "type": "Friends"
        },
        "at": "2016-05-29T06:38:38.001Z",
        "action": "join_request_refused"
      },
      {
        "group": {
          "id": "3",
          "name": "Our Club",
          "type": "Club"
        },
        "at": "2016-04-29T06:38:38.001Z",
        "action": "join_request_created"
      },
      {
        "group": {
          "id": "2",
          "name": "Our Team",
          "type": "Team"
        },
        "at": "2016-03-29T06:38:38.001Z",
        "action": "invitation_refused"
      },
      {
        "group": {
          "id": "1",
          "name": "Our Class",
          "type": "Class"
        },
        "at": "2016-02-28T06:38:38.001Z",
        "action": "invitation_created"
      }
    ]
    """

  Scenario: Request the first row
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-memberships-history?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "8",
          "name": "Another Club",
          "type": "Club"
        },
        "at": "2017-09-29T06:38:38.001Z",
        "action": "left"
      }
    ]
    """

  Scenario: Request the first row starting from some date
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-memberships-history?limit=1&from.at=2017-07-29T06:38:38.001Z&from.group_id=7"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "5",
          "name": "Other people",
          "type": "Other"
        },
        "at": "2017-06-29T06:38:38.001Z",
        "action": "invitation_accepted"
      }
    ]
    """

  Scenario: No new notifications
    Given I am the user with id "13"
    When I send a GET request to "/current-user/group-memberships-history"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
