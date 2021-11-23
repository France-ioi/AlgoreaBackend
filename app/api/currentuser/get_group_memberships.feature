Feature: Get group memberships for the current user
  Background:
    Given the database has the following table 'groups':
      | id | type    | name                       | description                | frozen_membership | require_lock_membership_approval_until |
      | 1  | Class   | Our Class                  | Our class group            | false             | null                                   |
      | 2  | Team    | Our Team                   | Our team group             | false             | null                                   |
      | 3  | Club    | Our Club                   | Our club group             | false             | null                                   |
      | 4  | Friends | Our Friends                | Group for our friends      | false             | null                                   |
      | 5  | Other   | Other people               | Group for other people     | false             | 2020-05-30 11:00:00                    |
      | 6  | Class   | Another Class              | Another class group        | false             | 3020-05-30 11:00:00                    |
      | 7  | Team    | Another Team               | Another team group         | false             | null                                   |
      | 8  | Club    | Another Club               | Another club group         | false             | null                                   |
      | 9  | Friends | Some other friends         | Another friends group      | false             | null                                   |
      | 11 | User    | user self                  |                            | false             | null                                   |
      | 21 | User    | owner self                 |                            | false             | null                                   |
      | 30 | Team    | Frozen Team                | Frozen Team                | true              | null                                   |
      | 31 | Team    | Team With Entry Conditions | Team With Entry Conditions | false             | null                                   |
    And the database has the following table 'users':
      | login | temp_user | group_id | first_name  | last_name | grade |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | 3     |
      | user  | 0         | 11       | John        | Doe       | 1     |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | lock_membership_approved_at |
      | 5               | 21             | 2019-05-30 11:00:00         |
      | 6               | 21             | 2019-05-30 11:00:00         |
      | 9               | 21             | 2019-05-30 11:00:00         |
      | 1               | 11             | null                        |
      | 30              | 21             | null                        |
      | 31              | 11             | null                        |
      | 31              | 21             | null                        |
    And the groups ancestors are computed
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
      | 1        | 11        | added_directly        | 2017-11-29 06:38:38 |
    And the database has the following table 'items':
      | id | default_language_tag | entry_min_admitted_members_ratio |
      | 2  | fr                   | All                              |
    And the database table 'attempts' has also the following row:
      | participant_id | id | root_item_id |
      | 31             | 1  | 2            |
    And the database has the following table 'results':
      | participant_id | attempt_id | item_id | started_at          |
      | 31             | 1          | 2       | 2019-05-30 11:00:00 |

  Scenario: Show all memberships
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-memberships"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "6",
          "name": "Another Class",
          "description": "Another class group",
          "type": "Class"
        },
        "member_since": "2017-07-29T06:38:38Z",
        "action": "join_request_accepted",
        "is_membership_locked": true
      },
      {
        "group": {
          "id": "5",
          "name": "Other people",
          "description": "Group for other people",
          "type": "Other"
        },
        "member_since": "2017-06-29T06:38:38Z",
        "action": "invitation_accepted",
        "is_membership_locked": false
      },
      {
        "group": {
          "id": "9",
          "name": "Some other friends",
          "description": "Another friends group",
          "type": "Friends"
        },
        "action": "added_directly",
        "member_since": null,
        "is_membership_locked": false
      },
      {
        "action": "added_directly",
        "group": {
          "description": "Frozen Team",
          "id": "30",
          "name": "Frozen Team",
          "type": "Team"
        },
        "member_since": null,
        "is_membership_locked": false,
        "can_leave_team": "frozen_membership"
      },
      {
        "action": "added_directly",
        "group": {
          "description": "Team With Entry Conditions",
          "id": "31",
          "name": "Team With Entry Conditions",
          "type": "Team"
        },
        "member_since": null,
        "is_membership_locked": false,
        "can_leave_team": "would_break_entry_conditions"
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
        "group": {
          "id": "6",
          "name": "Another Class",
          "description": "Another class group",
          "type": "Class"
        },
        "member_since": "2017-07-29T06:38:38Z",
        "action": "join_request_accepted",
        "is_membership_locked": true
      }
    ]
    """

  Scenario: Request the second row
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-memberships?limit=1&from.id=6"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "5",
          "name": "Other people",
          "description": "Group for other people",
          "type": "Other"
        },
        "member_since": "2017-06-29T06:38:38Z",
        "action": "invitation_accepted",
        "is_membership_locked": false
      }
    ]
    """
