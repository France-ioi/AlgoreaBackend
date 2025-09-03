Feature: Get group memberships for the current user
  Background:
    Given the database has the following table "groups":
      | id | type                | name                       | description                | frozen_membership | require_lock_membership_approval_until | require_personal_info_access_approval | require_watch_approval |
      | 1  | Class               | Our Class                  | Our class group            | false             | null                                   | none                                  | false                  |
      | 2  | Team                | Our Team                   | Our team group             | false             | null                                   | none                                  | false                  |
      | 3  | Club                | Our Club                   | Our club group             | false             | null                                   | none                                  | false                  |
      | 4  | Friends             | Our Friends                | Group for our friends      | false             | null                                   | none                                  | false                  |
      | 5  | Other               | Other people               | Group for other people     | false             | 2020-05-30 11:00:00                    | none                                  | false                  |
      | 6  | Class               | Another Class              | Another class group        | false             | 3020-05-30 11:00:00                    | edit                                  | false                  |
      | 7  | Team                | Another Team               | Another team group         | false             | null                                   | none                                  | false                  |
      | 8  | Club                | Another Club               | Another club group         | false             | null                                   | none                                  | false                  |
      | 9  | Friends             | Some other friends         | Another friends group      | false             | null                                   | view                                  | false                  |
      | 30 | Team                | Frozen Team                | Frozen Team                | true              | null                                   | none                                  | false                  |
      | 31 | Team                | Team With Entry Conditions | Team With Entry Conditions | false             | null                                   | none                                  | true                   |
      | 32 | ContestParticipants |                            |                            | false             | null                                   | none                                  | false                  |
    And the database has the following users:
      | group_id | login | first_name  | last_name |
      | 21       | owner | Jean-Michel | Blanquer  |
      | 11       | user  | John        | Doe       |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | lock_membership_approved_at |
      | 5               | 21             | 2019-05-30 11:00:00         |
      | 6               | 21             | 2019-05-30 11:00:00         |
      | 9               | 21             | 2019-05-30 11:00:00         |
      | 1               | 11             | null                        |
      | 30              | 21             | null                        |
      | 31              | 11             | null                        |
      | 31              | 21             | null                        |
      | 32              | 21             | null                        |
    And the groups ancestors are computed
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
      | 1        | 11        | added_directly        | 2017-11-29 06:38:38.001 |
    And the database has the following table "items":
      | id | default_language_tag | entry_min_admitted_members_ratio |
      | 2  | fr                   | All                              |
    And the database table "attempts" also has the following row:
      | participant_id | id | root_item_id |
      | 31             | 1  | 2            |
    And the database has the following table "results":
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
          "type": "Class",
          "require_personal_info_access_approval": "edit",
          "require_watch_approval": false
        },
        "member_since": "2017-07-29T06:38:38.001Z",
        "action": "join_request_accepted",
        "is_membership_locked": true
      },
      {
        "group": {
          "id": "5",
          "name": "Other people",
          "description": "Group for other people",
          "type": "Other",
          "require_personal_info_access_approval": "none",
          "require_watch_approval": false
        },
        "member_since": "2017-06-29T06:38:38.001Z",
        "action": "invitation_accepted",
        "is_membership_locked": false
      },
      {
        "group": {
          "id": "9",
          "name": "Some other friends",
          "description": "Another friends group",
          "type": "Friends",
          "require_personal_info_access_approval": "view",
          "require_watch_approval": false
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
          "type": "Team",
          "require_personal_info_access_approval": "none",
          "require_watch_approval": false
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
          "type": "Team",
          "require_personal_info_access_approval": "none",
          "require_watch_approval": true
        },
        "member_since": null,
        "is_membership_locked": false,
        "can_leave_team": "would_break_entry_conditions"
      }
    ]
    """

  Scenario: Show memberships requiring access to personal info
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-memberships?only_requiring_personal_info_access_approval=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "6",
          "name": "Another Class",
          "description": "Another class group",
          "type": "Class",
          "require_personal_info_access_approval": "edit",
          "require_watch_approval": false
        },
        "member_since": "2017-07-29T06:38:38.001Z",
        "action": "join_request_accepted",
        "is_membership_locked": true
      },
      {
        "group": {
          "id": "9",
          "name": "Some other friends",
          "description": "Another friends group",
          "type": "Friends",
          "require_personal_info_access_approval": "view",
          "require_watch_approval": false
        },
        "action": "added_directly",
        "member_since": null,
        "is_membership_locked": false
      }
    ]
    """

  Scenario: Request the first row of all memberships
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
          "type": "Class",
          "require_personal_info_access_approval": "edit",
          "require_watch_approval": false
        },
        "member_since": "2017-07-29T06:38:38.001Z",
        "action": "join_request_accepted",
        "is_membership_locked": true
      }
    ]
    """

  Scenario: Request the second row of all memberships
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
          "type": "Other",
          "require_personal_info_access_approval": "none",
          "require_watch_approval": false
        },
        "member_since": "2017-06-29T06:38:38.001Z",
        "action": "invitation_accepted",
        "is_membership_locked": false
      }
    ]
    """

  Scenario: Request the second row of memberships requiring access to personal info
    Given I am the user with id "21"
    When I send a GET request to "/current-user/group-memberships?only_requiring_personal_info_access_approval=1&limit=1&from.id=6"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {
          "id": "9",
          "name": "Some other friends",
          "description": "Another friends group",
          "type": "Friends",
          "require_personal_info_access_approval": "view",
          "require_watch_approval": false
        },
        "action": "added_directly",
        "member_since": null,
        "is_membership_locked": false
      }
    ]
    """
