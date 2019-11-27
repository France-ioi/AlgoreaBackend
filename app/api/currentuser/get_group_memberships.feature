Feature: Get group memberships for the current user
  Background:
    Given the database has the following table 'groups':
      | id | type      | name               | description            |
      | 1  | Class     | Our Class          | Our class group        |
      | 2  | Team      | Our Team           | Our team group         |
      | 3  | Club      | Our Club           | Our club group         |
      | 4  | Friends   | Our Friends        | Group for our friends  |
      | 5  | Other     | Other people       | Group for other people |
      | 6  | Class     | Another Class      | Another class group    |
      | 7  | Team      | Another Team       | Another team group     |
      | 8  | Club      | Another Club       | Another club group     |
      | 9  | Friends   | Some other friends | Another friends group  |
      | 11 | UserSelf  | user self          |                        |
      | 12 | UserAdmin | user admin         |                        |
      | 21 | UserSelf  | owner self         |                        |
      | 22 | UserAdmin | owner admin        |                        |
    And the database has the following table 'users':
      | login | temp_user | group_id | owned_group_id | first_name  | last_name | grade |
      | owner | 0         | 21       | 22             | Jean-Michel | Blanquer  | 3     |
      | user  | 0         | 11       | 12             | John        | Doe       | 1     |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id |
      | 6  | 5               | 21             |
      | 7  | 6               | 21             |
      | 10 | 9               | 21             |
      | 11 | 1               | 22             |
    And the database has the following table 'group_membership_changes':
      | group_id | member_id | action                | at                  |
      | 1        | 21        | invitation_created    | 2017-02-28 06:38:38 |
      | 2        | 21        | invitation_refused    | 2017-03-29 06:38:38 |
      | 3        | 21        | join_request_created  | 2017-04-29 06:38:38 |
      | 4        | 21        | join_request_refused  | 2017-05-29 06:38:38 |
      | 5        | 21        | invitation_accepted   | 2017-06-29 06:38:38 |
      | 6        | 21        | join_request_accepted | 2017-07-29 06:38:38 |
      | 7        | 21        | removed               | 2017-08-29 06:38:38 |
      | 8        | 21        | left                  | 2017-09-29 06:38:38 |
      | 1        | 22        | added_directly        | 2017-11-29 06:38:38 |

  Scenario: Show all memberships
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-memberships"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "7",
        "group": {
          "id": "6",
          "name": "Another Class",
          "description": "Another class group",
          "type": "Class"
        },
        "member_since": "2017-07-29T06:38:38Z",
        "action": "join_request_accepted"
      },
      {
        "id": "6",
        "group": {
          "id": "5",
          "name": "Other people",
          "description": "Group for other people",
          "type": "Other"
        },
        "member_since": "2017-06-29T06:38:38Z",
        "action": "invitation_accepted"
      },
      {
        "id": "10",
        "group": {
          "id": "9",
          "name": "Some other friends",
          "description": "Another friends group",
          "type": "Friends"
        },
        "action": "added_directly",
        "member_since": null
      }
    ]
    """

  Scenario: Request the first row
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-memberships?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "7",
        "group": {
          "id": "6",
          "name": "Another Class",
          "description": "Another class group",
          "type": "Class"
        },
        "member_since": "2017-07-29T06:38:38Z",
        "action": "join_request_accepted"
      }
    ]
    """

