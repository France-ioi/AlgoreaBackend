Feature: List groups managed by the current user
  Background:
    Given the database has the following table "groups":
      | id | name          | type  | description |
      | 5  | Group         | Class | null        |
      | 13 | Our Class     | Class | null        |
      | 14 | Our Friends   | Other | null        |
      | 15 | Another Group | Other | Super Group |
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
        "description": null,
        "id": "13",
        "name": "Our Class",
        "type": "Class"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "none",
        "can_watch_members": false,
        "description": "Super Group",
        "id": "15",
        "name": "Another Group",
        "type": "Other"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "memberships",
        "can_watch_members": true,
        "description": null,
        "id": "14",
        "name": "Our Friends",
        "type": "Other"
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
        "can_manage": "memberships",
        "can_watch_members": true,
        "description": null,
        "id": "14",
        "name": "Our Friends",
        "type": "Other"
      },
      {
        "can_grant_group_access": true,
        "can_manage": "memberships_and_group",
        "can_watch_members": false,
        "description": null,
        "id": "13",
        "name": "Our Class",
        "type": "Class"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "none",
        "can_watch_members": false,
        "description": "Super Group",
        "id": "15",
        "name": "Another Group",
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
        "description": null,
        "id": "13",
        "name": "Our Class",
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
        "description": "Super Group",
        "id": "15",
        "name": "Another Group",
        "type": "Other"
      },
      {
        "can_grant_group_access": false,
        "can_manage": "memberships",
        "can_watch_members": true,
        "description": null,
        "id": "14",
        "name": "Our Friends",
        "type": "Other"
      }
    ]
    """
