Feature: Get group by groupID (groupView)
  Background:
    Given the database has the following table 'groups':
      | id | name    | grade | description     | created_at          | type  | root_activity_id    | root_skill_id | is_open | is_public | code       | code_lifetime | code_expires_at     | open_activity_when_joining | require_lock_membership_approval_until | frozen_membership |
      | 8  | Other   | -4    | Parent of 9     | 2019-04-06 09:26:40 | Other | null                | null          | false   | false     | efghijklmn | null          | null                | false                      | null                                   | false             |
      | 9  | Club    | -4    | Club            | 2019-04-06 09:26:40 | Other | null                | null          | false   | false     | null       | null          | null                | false                      | null                                   | false             |
      | 10 | Parent  | -3    | Parent of 16    | 2019-02-06 09:26:40 | Class | 7297887146214536132 | 123456        | true    | false     | defghijklm | 02:00:00      | 2019-10-13 05:39:48 | true                       | null                                   | false             |
      | 11 | Group A | -3    | Group A is here | 2019-02-06 09:26:40 | Class | 1672978871462145361 | 567890        | true    | false     | ybqybxnlyo | 01:00:00      | 2017-10-13 05:39:48 | true                       | null                                   | false             |
      | 13 | Group B | -2    | Group B is here | 2019-03-06 09:26:40 | Class | 1672978871462145461 | 789012        | true    | false     | ybabbxnlyo | 01:00:00      | 2017-10-14 05:39:48 | true                       | null                                   | false             |
      | 15 | Group D | -4    | Other Group     | 2019-04-06 09:26:40 | Other | null                | null          | false   | true      | abcdefghij | null          | null                | false                      | null                                   | false             |
      | 16 | Group E | -4    | Other Group     | 2019-04-06 09:26:40 | Other | null                | null          | false   | true      | null       | null          | null                | false                      | null                                   | false             |
      | 17 | Team 1  | -4    | Team 1          | 2019-04-06 09:26:40 | Team  | null                | null          | false   | false     | null       | null          | null                | false                      | 3019-05-30 11:00:00                    | false             |
      | 18 | Team 2  | -4    | Team 2          | 2019-04-06 09:26:40 | Team  | null                | null          | false   | false     | null       | null          | null                | false                      | 2019-05-30 11:00:00                    | true              |
      | 19 | Team 3  | -4    | Team 3          | 2019-04-06 09:26:40 | Team  | null                | null          | false   | false     | null       | null          | null                | false                      | null                                   | false             |
      | 21 | owner   | 0     | null            | 2019-01-06 09:26:40 | User  | null                | null          | false   | false     | null       | null          | null                | false                      | null                                   | false             |
      | 31 | john    | 0     | null            | 2019-01-06 09:26:40 | User  | null                | null          | false   | false     | null       | null          | null                | false                      | null                                   | false             |
      | 41 | jane    | 0     | null            | 2019-01-06 09:26:40 | User  | null                | null          | false   | false     | null       | null          | null                | false                      | null                                   | false             |
      | 51 | rick    | 0     | null            | 2019-01-06 09:26:40 | User  | null                | null          | false   | false     | null       | null          | null                | false                      | null                                   | false             |
      | 61 | ian     | 0     | null            | 2019-01-06 09:26:40 | User  | null                | null          | false   | false     | null       | null          | null                | false                      | null                                   | false             |
      | 71 | dirk    | 0     | null            | 2019-01-06 09:26:40 | User  | null                | null          | false   | false     | null       | null          | null                | false                      | null                                   | false             |
      | 81 | chuck   | 0     | null            | 2019-01-06 09:26:40 | User  | null                | null          | false   | false     | null       | null          | null                | false                      | null                                   | false             |
    And the database has the following table 'users':
      | login | group_id |
      | owner | 21       |
      | john  | 31       |
      | jane  | 41       |
      | rick  | 51       |
      | ian   | 61       |
      | dirk  | 71       |
      | chuck | 81       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | lock_membership_approved_at |
      | 8               | 9              | null                        |
      | 9               | 21             | null                        |
      | 10              | 16             | null                        |
      | 11              | 31             | null                        |
      | 13              | 11             | null                        |
      | 13              | 51             | null                        |
      | 13              | 61             | null                        |
      | 13              | 71             | null                        |
      | 13              | 81             | null                        |
      | 17              | 51             | 3019-05-30 11:00:00         |
      | 17              | 61             | null                        |
      | 18              | 71             | null                        |
      | 19              | 81             | null                        |
      | 15              | 11             | null                        |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage            | can_grant_group_access | can_watch_members |
      | 13       | 21         | none                  | false                  | false             |
      | 15       | 21         | memberships_and_group | true                   | true              |
      | 16       | 9          | none                  | false                  | false             |
      | 19       | 21         | none                  | false                  | false             |
    And the database has the following table 'items':
      | id | default_language_tag | entry_min_admitted_members_ratio |
      | 2  | fr                   | All                              |
    And the database table 'attempts' has also the following row:
      | participant_id | id | root_item_id |
      | 17             | 1  | 2            |
    And the database has the following table 'results':
      | participant_id | attempt_id | item_id | started_at          |
      | 17             | 1          | 2       | 2019-05-30 11:00:00 |

  Scenario: The user is a manager of the group
    Given I am the user with id "21"
    When I send a GET request to "/groups/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "13",
      "name": "Group B",
      "grade": -2,
      "description": "Group B is here",
      "created_at": "2019-03-06T09:26:40Z",
      "type": "Class",
      "root_activity_id": "1672978871462145461",
      "root_skill_id": "789012",
      "is_open": true,
      "is_public": false,
      "code": "ybabbxnlyo",
      "code_lifetime": "01:00:00",
      "code_expires_at": "2017-10-14T05:39:48Z",
      "open_activity_when_joining": true,
      "current_user_managership": "direct",
      "current_user_can_manage": "none",
      "current_user_can_grant_group_access": false,
      "current_user_can_watch_members": false,
      "current_user_membership": "none",
      "descendants_current_user_is_member_of": [],
      "ancestors_current_user_is_manager_of": [],
      "descendants_current_user_is_manager_of": [{"id": "11", "name": "Group A"}],
      "is_membership_locked": false
    }
    """

  Scenario: The user is a manager of the group's ancestor
    Given I am the user with id "21"
    When I send a GET request to "/groups/11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "11",
      "name": "Group A",
      "grade": -3,
      "description": "Group A is here",
      "created_at": "2019-02-06T09:26:40Z",
      "type": "Class",
      "root_activity_id": "1672978871462145361",
      "root_skill_id": "567890",
      "is_open": true,
      "is_public": false,
      "code": "ybqybxnlyo",
      "code_lifetime": "01:00:00",
      "code_expires_at": "2017-10-13T05:39:48Z",
      "open_activity_when_joining": true,
      "current_user_managership": "ancestor",
      "current_user_can_manage": "memberships_and_group",
      "current_user_can_grant_group_access": true,
      "current_user_can_watch_members": true,
      "current_user_membership": "none",
      "descendants_current_user_is_member_of": [],
      "ancestors_current_user_is_manager_of": [{"id": "13", "name": "Group B"}, {"id": "15", "name": "Group D"}],
      "descendants_current_user_is_manager_of": [],
      "is_membership_locked": false
    }
    """

  Scenario: The user is a manager of the group's non-user descendant
    Given I am the user with id "21"
    When I send a GET request to "/groups/10"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "10",
      "name": "Parent",
      "grade": -3,
      "description": "Parent of 16",
      "created_at": "2019-02-06T09:26:40Z",
      "type": "Class",
      "root_activity_id": "7297887146214536132",
      "root_skill_id": "123456",
      "is_open": true,
      "is_public": false,
      "open_activity_when_joining": true,
      "current_user_managership": "descendant",
      "current_user_membership": "none",
      "descendants_current_user_is_member_of": [],
      "ancestors_current_user_is_manager_of": [],
      "descendants_current_user_is_manager_of": [{"id": "16", "name": "Group E"}],
      "is_membership_locked": false
    }
    """

  Scenario: The user is a member of the group's descendant
    Given I am the user with id "21"
    When I send a GET request to "/groups/8"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "8",
      "name": "Other",
      "grade": -4,
      "description": "Parent of 9",
      "created_at": "2019-04-06T09:26:40Z",
      "type": "Other",
      "root_activity_id": null,
      "root_skill_id": null,
      "is_open": false,
      "is_public": false,
      "open_activity_when_joining": false,
      "current_user_managership": "none",
      "current_user_membership": "descendant",
      "descendants_current_user_is_member_of": [{"id": "9", "name": "Club"}],
      "ancestors_current_user_is_manager_of": [],
      "descendants_current_user_is_manager_of": [],
      "is_membership_locked": false
    }
    """

  Scenario: The user is a descendant of the group
    Given I am the user with id "31"
    When I send a GET request to "/groups/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "13",
      "name": "Group B",
      "grade": -2,
      "description": "Group B is here",
      "created_at": "2019-03-06T09:26:40Z",
      "type": "Class",
      "root_activity_id": "1672978871462145461",
      "root_skill_id": "789012",
      "is_open": true,
      "is_public": false,
      "open_activity_when_joining": true,
      "current_user_managership": "none",
      "current_user_membership": "descendant",
      "descendants_current_user_is_member_of": [{"id": "11", "name": "Group A"}],
      "ancestors_current_user_is_manager_of": [],
      "descendants_current_user_is_manager_of": [],
      "is_membership_locked": false
    }
    """

  Scenario Outline: The user is a member of the group
    Given I am the user with id "<user_id>"
    When I send a GET request to "/groups/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "13",
      "name": "Group B",
      "grade": -2,
      "description": "Group B is here",
      "created_at": "2019-03-06T09:26:40Z",
      "type": "Class",
      "root_activity_id": "1672978871462145461",
      "root_skill_id": "789012",
      "is_open": true,
      "is_public": false,
      "open_activity_when_joining": true,
      "current_user_managership": "none",
      "current_user_membership": "direct",
      "descendants_current_user_is_member_of": [],
      "ancestors_current_user_is_manager_of": [],
      "descendants_current_user_is_manager_of": [],
      "is_membership_locked": false
    }
    """
  Examples:
    | user_id |
    | 51      |
    | 61      |
    | 71      |
    | 81      |

  Scenario: The group has is_public = 1
    Given I am the user with id "41"
    When I send a GET request to "/groups/15"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "15",
      "name": "Group D",
      "grade": -4,
      "description": "Other Group",
      "created_at": "2019-04-06T09:26:40Z",
      "type": "Other",
      "root_activity_id": null,
      "root_skill_id": null,
      "is_open": false,
      "is_public": true,
      "open_activity_when_joining": false,
      "current_user_managership": "none",
      "current_user_membership": "none",
      "descendants_current_user_is_member_of": [],
      "ancestors_current_user_is_manager_of": [],
      "descendants_current_user_is_manager_of": [],
      "is_membership_locked": false
    }
    """

  Scenario Outline: The user is a member of the team
    Given I am the user with id "<user_id>"
    When I send a GET request to "/groups/<group_id>"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "<group_id>",
      "name": "<team_name>",
      "grade": -4,
      "description": "<team_name>",
      "created_at": "2019-04-06T09:26:40Z",
      "type": "Team",
      "root_activity_id": null,
      "root_skill_id": null,
      "is_open": false,
      "is_public": false,
      "open_activity_when_joining": false,
      "current_user_managership": "none",
      "current_user_membership": "direct",
      "descendants_current_user_is_member_of": [],
      "ancestors_current_user_is_manager_of": [],
      "descendants_current_user_is_manager_of": [],
      "is_membership_locked": <is_membership_locked>,
      "can_leave_team": "<can_leave_team>"
    }
    """
    Examples:
      | group_id | user_id | team_name | is_membership_locked | can_leave_team               |
      | 17       | 51      | Team 1    | true                 | would_break_entry_conditions |
      | 17       | 61      | Team 1    | false                | would_break_entry_conditions |
      | 18       | 71      | Team 2    | false                | frozen_membership            |
      | 19       | 81      | Team 3    | false                | free_to_leave                |

  Scenario: The user is a manager of the team
    Given I am the user with id "21"
    When I send a GET request to "/groups/19"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "19",
      "name": "Team 3",
      "grade": -4,
      "description": "Team 3",
      "created_at": "2019-04-06T09:26:40Z",
      "type": "Team",
      "root_activity_id": null,
      "root_skill_id": null,
      "is_open": false,
      "is_public": false,
      "code": null,
      "code_lifetime": null,
      "code_expires_at": null,
      "open_activity_when_joining": false,
      "current_user_managership": "direct",
      "current_user_can_manage": "none",
      "current_user_can_grant_group_access": false,
      "current_user_can_watch_members": false,
      "current_user_membership": "none",
      "descendants_current_user_is_member_of": [],
      "ancestors_current_user_is_manager_of": [],
      "descendants_current_user_is_manager_of": [],
      "is_membership_locked": false
    }
    """
