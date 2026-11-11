Feature: List groups managed by the current user
  Background:
    Given the database has the following table "groups":
      | id | name          | type    | description | root_activity_id | expected_start      | organizer | code       | code_expires_at     |
      | 1  | Friends       | Friends | null        | null             | null                | null      | null       | null                |
      | 5  | Group         | Class   | null        | null             | null                | null      | null       | null                |
      | 6  | Club          | Club    | null        | null             | null                | null      | null       | null                |
      | 9  | Other         | Other   | null        | null             | null                | null      | null       | null                |
      | 13 | Our Class     | Class   | null        | null             | null                | null      | null       | null                |
      | 14 | Our Friends   | Other   | null        | null             | null                | null      | null       | null                |
      | 15 | Another Group | Other   | Super Group | null             | null                | null      | null       | null                |
      | 16 | Stage A      | Session | null        | 200              | 2019-05-30 12:00:00 | School A  | ybabbxnlyo | 2019-12-31 23:59:59 |
      | 17 | Stage B      | Session | null        | 210              | null                | null      | abcdefghij | null                |
      | 18 | Stage C      | Session | null        | 200              | 2019-06-01 08:00:00 | School C  | null       | null                |
    And the database has the following users:
      | group_id | login |
      | 21       | owner |
      | 11       | user  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 5               | 21             |
      | 6               | 21             |
      | 9               | 21             |
      | 1               | 11             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage            | can_grant_group_access | can_watch_members |
      | 13       | 5          | memberships_and_group | 1                      | 0                 |
      | 13       | 21         | none                  | 0                      | 0                 |
      | 14       | 21         | memberships           | 0                      | 1                 |
      | 15       | 5          | none                  | 0                      | 0                 |
      | 16       | 21         | memberships_and_group | 1                      | 1                 |
      | 17       | 21         | memberships           | 0                      | 0                 |
      | 18       | 21         | none                  | 0                      | 0                 |

  Scenario: Show all managed groups
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "can_grant_group_access": true,
        "can_manage": "memberships_and_group",
        "can_watch_members": false,
        "code": null,
        "code_expires_at": null,
        "description": null,
        "expected_start": null,
        "id": "13",
        "name": "Our Class",
        "organizer": null,
        "root_activity_id": null,
        "type": "Class"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "none",
        "can_watch_members": false,
        "code": null,
        "code_expires_at": null,
        "description": "Super Group",
        "expected_start": null,
        "id": "15",
        "name": "Another Group",
        "organizer": null,
        "root_activity_id": null,
        "type": "Other"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "memberships",
        "can_watch_members": true,
        "code": null,
        "code_expires_at": null,
        "description": null,
        "expected_start": null,
        "id": "14",
        "name": "Our Friends",
        "organizer": null,
        "root_activity_id": null,
        "type": "Other"
      },
      {
        "can_grant_group_access": true,
        "can_manage": "memberships_and_group",
        "can_watch_members": true,
        "code": "ybabbxnlyo",
        "code_expires_at": "2019-12-31T23:59:59Z",
        "description": null,
        "expected_start": "2019-05-30T12:00:00Z",
        "id": "16",
        "name": "Stage A",
        "organizer": "School A",
        "root_activity_id": "200",
        "type": "Session"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "memberships",
        "can_watch_members": false,
        "code": "abcdefghij",
        "code_expires_at": null,
        "description": null,
        "expected_start": null,
        "id": "17",
        "name": "Stage B",
        "organizer": null,
        "root_activity_id": "210",
        "type": "Session"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "none",
        "can_watch_members": false,
        "code": null,
        "code_expires_at": null,
        "description": null,
        "expected_start": "2019-06-01T08:00:00Z",
        "id": "18",
        "name": "Stage C",
        "organizer": "School C",
        "root_activity_id": "200",
        "type": "Session"
      }
    ]
    """

  Scenario: Show all managed groups (different order)
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?sort=-name"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "can_grant_group_access": false,
        "can_manage": "none",
        "can_watch_members": false,
        "code": null,
        "code_expires_at": null,
        "description": null,
        "expected_start": "2019-06-01T08:00:00Z",
        "id": "18",
        "name": "Stage C",
        "organizer": "School C",
        "root_activity_id": "200",
        "type": "Session"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "memberships",
        "can_watch_members": false,
        "code": "abcdefghij",
        "code_expires_at": null,
        "description": null,
        "expected_start": null,
        "id": "17",
        "name": "Stage B",
        "organizer": null,
        "root_activity_id": "210",
        "type": "Session"
      },
      {
        "can_grant_group_access": true,
        "can_manage": "memberships_and_group",
        "can_watch_members": true,
        "code": "ybabbxnlyo",
        "code_expires_at": "2019-12-31T23:59:59Z",
        "description": null,
        "expected_start": "2019-05-30T12:00:00Z",
        "id": "16",
        "name": "Stage A",
        "organizer": "School A",
        "root_activity_id": "200",
        "type": "Session"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "memberships",
        "can_watch_members": true,
        "code": null,
        "code_expires_at": null,
        "description": null,
        "expected_start": null,
        "id": "14",
        "name": "Our Friends",
        "organizer": null,
        "root_activity_id": null,
        "type": "Other"
      },
      {
        "can_grant_group_access": true,
        "can_manage": "memberships_and_group",
        "can_watch_members": false,
        "code": null,
        "code_expires_at": null,
        "description": null,
        "expected_start": null,
        "id": "13",
        "name": "Our Class",
        "organizer": null,
        "root_activity_id": null,
        "type": "Class"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "none",
        "can_watch_members": false,
        "code": null,
        "code_expires_at": null,
        "description": "Super Group",
        "expected_start": null,
        "id": "15",
        "name": "Another Group",
        "organizer": null,
        "root_activity_id": null,
        "type": "Other"
      }
    ]
    """

  Scenario: Request the first row
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "can_grant_group_access": true,
        "can_manage": "memberships_and_group",
        "can_watch_members": false,
        "code": null,
        "code_expires_at": null,
        "description": null,
        "expected_start": null,
        "id": "13",
        "name": "Our Class",
        "organizer": null,
        "root_activity_id": null,
        "type": "Class"
      }
    ]
    """

  Scenario: Start from the second row
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?from.id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "can_grant_group_access": false,
        "can_manage": "none",
        "can_watch_members": false,
        "code": null,
        "code_expires_at": null,
        "description": "Super Group",
        "expected_start": null,
        "id": "15",
        "name": "Another Group",
        "organizer": null,
        "root_activity_id": null,
        "type": "Other"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "memberships",
        "can_watch_members": true,
        "code": null,
        "code_expires_at": null,
        "description": null,
        "expected_start": null,
        "id": "14",
        "name": "Our Friends",
        "organizer": null,
        "root_activity_id": null,
        "type": "Other"
      },
      {
        "can_grant_group_access": true,
        "can_manage": "memberships_and_group",
        "can_watch_members": true,
        "code": "ybabbxnlyo",
        "code_expires_at": "2019-12-31T23:59:59Z",
        "description": null,
        "expected_start": "2019-05-30T12:00:00Z",
        "id": "16",
        "name": "Stage A",
        "organizer": "School A",
        "root_activity_id": "200",
        "type": "Session"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "memberships",
        "can_watch_members": false,
        "code": "abcdefghij",
        "code_expires_at": null,
        "description": null,
        "expected_start": null,
        "id": "17",
        "name": "Stage B",
        "organizer": null,
        "root_activity_id": "210",
        "type": "Session"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "none",
        "can_watch_members": false,
        "code": null,
        "code_expires_at": null,
        "description": null,
        "expected_start": "2019-06-01T08:00:00Z",
        "id": "18",
        "name": "Stage C",
        "organizer": "School C",
        "root_activity_id": "200",
        "type": "Session"
      }
    ]
    """

  Scenario: Filter by type
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?type=Session"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "can_grant_group_access": true,
        "can_manage": "memberships_and_group",
        "can_watch_members": true,
        "code": "ybabbxnlyo",
        "code_expires_at": "2019-12-31T23:59:59Z",
        "description": null,
        "expected_start": "2019-05-30T12:00:00Z",
        "id": "16",
        "name": "Stage A",
        "organizer": "School A",
        "root_activity_id": "200",
        "type": "Session"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "memberships",
        "can_watch_members": false,
        "code": "abcdefghij",
        "code_expires_at": null,
        "description": null,
        "expected_start": null,
        "id": "17",
        "name": "Stage B",
        "organizer": null,
        "root_activity_id": "210",
        "type": "Session"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "none",
        "can_watch_members": false,
        "code": null,
        "code_expires_at": null,
        "description": null,
        "expected_start": "2019-06-01T08:00:00Z",
        "id": "18",
        "name": "Stage C",
        "organizer": "School C",
        "root_activity_id": "200",
        "type": "Session"
      }
    ]
    """

  Scenario: Filter by root_activity_id_among with ids
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?root_activity_id_among=200,999"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "can_grant_group_access": true,
        "can_manage": "memberships_and_group",
        "can_watch_members": true,
        "code": "ybabbxnlyo",
        "code_expires_at": "2019-12-31T23:59:59Z",
        "description": null,
        "expected_start": "2019-05-30T12:00:00Z",
        "id": "16",
        "name": "Stage A",
        "organizer": "School A",
        "root_activity_id": "200",
        "type": "Session"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "none",
        "can_watch_members": false,
        "code": null,
        "code_expires_at": null,
        "description": null,
        "expected_start": "2019-06-01T08:00:00Z",
        "id": "18",
        "name": "Stage C",
        "organizer": "School C",
        "root_activity_id": "200",
        "type": "Session"
      }
    ]
    """

  Scenario: Empty root_activity_id_among returns no groups
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?root_activity_id_among="
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: Combine type filter with root_activity_id_among and paging
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?type=Session&root_activity_id_among=200,210&limit=1&from.id=16"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "can_grant_group_access": false,
        "can_manage": "memberships",
        "can_watch_members": false,
        "code": "abcdefghij",
        "code_expires_at": null,
        "description": null,
        "expected_start": null,
        "id": "17",
        "name": "Stage B",
        "organizer": null,
        "root_activity_id": "210",
        "type": "Session"
      }
    ]
    """
