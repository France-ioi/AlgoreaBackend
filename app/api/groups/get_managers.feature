Feature: Get managers of group_id
  Background:
    Given the database has the following users:
      | login | group_id | first_name  | last_name  | grade |
      | owner | 21       | Jean-Michel | Blanquer   | 3     |
      | jeff  | 71       | Jeff        | Joe        | 3     |
      | larry | 81       | null        | null       | 3     |
    And the database has the following table 'groups':
      | id | name        | type    |
      | 11 | user        | User    |
      | 12 | Our Club    | Club    |
      | 13 | Our Class   | Class   |
      | 14 | Our Friends | Friends |
      | 15 | Other       | Other   |
      | 16 | Team        | Team    |
      | 31 | jane        | User    |
      | 41 | john        | User    |
      | 51 | billg       | User    |
      | 61 | zuck        | User    |
      | 91 | lp          | User    |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 12              | 13             |
      | 13              | 71             |
      | 15              | 13             |
      | 16              | 81             |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage            | can_grant_group_access | can_watch_members |
      | 12       | 81         | memberships           | 1                      | 0                 |
      | 13       | 14         | none                  | 0                      | 0                 |
      | 13       | 21         | none                  | 0                      | 0                 |
      | 13       | 51         | memberships           | 0                      | 1                 |
      | 13       | 61         | memberships_and_group | 1                      | 0                 |
      | 13       | 91         | none                  | 1                      | 0                 |
      | 14       | 21         | memberships           | 0                      | 1                 |
      | 14       | 31         | memberships           | 0                      | 1                 |
      | 14       | 41         | memberships           | 0                      | 1                 |
      | 15       | 81         | memberships_and_group | 0                      | 1                 |
    And the groups ancestors are computed

  Scenario: The user is a manager of the group, default sort (by name)
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/managers"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51", "name": "billg", "first_name": null, "last_name": null,
        "can_manage": "memberships", "can_grant_group_access": false, "can_watch_members": true
      },
      {
        "id": "91", "name": "lp", "first_name": null, "last_name": null,
        "can_manage": "none", "can_grant_group_access": true, "can_watch_members": false
      },
      {
        "id": "14", "name": "Our Friends",
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false
      },
      {
        "id": "21", "name": "owner", "first_name": "Jean-Michel", "last_name": "Blanquer",
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false
      },
      {
        "id": "61", "name": "zuck", "first_name": null, "last_name": null,
        "can_manage": "memberships_and_group", "can_grant_group_access": true, "can_watch_members": false
      }
    ]
    """

  Scenario: Sort by name in descending order (the current user is a manager of an ancestor group)
    Given I am the user with id "81"
    When I send a GET request to "/groups/13/managers?sort=-name"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "61", "name": "zuck", "first_name": null, "last_name": null,
        "can_manage": "memberships_and_group", "can_grant_group_access": true, "can_watch_members": false
      },
      {
        "id": "21", "name": "owner", "first_name": "Jean-Michel", "last_name": "Blanquer",
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false
      },
      {
        "id": "14", "name": "Our Friends",
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false
      },
      {
        "id": "91", "name": "lp", "first_name": null, "last_name": null,
        "can_manage": "none", "can_grant_group_access": true, "can_watch_members": false
      },
      {
        "id": "51", "name": "billg", "first_name": null, "last_name": null,
        "can_manage": "memberships", "can_grant_group_access": false, "can_watch_members": true
      }
    ]
    """

  Scenario: Sort by id
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/managers?sort=id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "14", "name": "Our Friends",
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false
      },
      {
        "id": "21", "name": "owner", "first_name": "Jean-Michel", "last_name": "Blanquer",
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false
      },
      {
        "id": "51", "name": "billg", "first_name": null, "last_name": null,
        "can_manage": "memberships", "can_grant_group_access": false, "can_watch_members": true
      },
      {
        "id": "61", "name": "zuck", "first_name": null, "last_name": null,
        "can_manage": "memberships_and_group", "can_grant_group_access": true, "can_watch_members": false
      },
      {
        "id": "91", "name": "lp", "first_name": null, "last_name": null,
        "can_manage": "none", "can_grant_group_access": true, "can_watch_members": false
      }
    ]
    """

  Scenario: Request the first row
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/managers?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51", "name": "billg", "first_name": null, "last_name": null,
        "can_manage": "memberships", "can_grant_group_access": false, "can_watch_members": true
      }
    ]
    """

  Scenario: Default sort (by name) including managers of ancestor groups
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/managers?include_managers_of_ancestor_groups=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51", "name": "billg", "first_name": null, "last_name": null,
        "can_manage": "memberships", "can_grant_group_access": false, "can_watch_members": true,
        "can_manage_through_ancestor_groups": "memberships", "can_grant_group_access_through_ancestor_groups": false,
        "can_watch_members_through_ancestor_groups": true
      },
      {
        "id": "81", "name": "larry", "first_name": null, "last_name": null,
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false,
        "can_manage_through_ancestor_groups": "memberships_and_group", "can_grant_group_access_through_ancestor_groups": true,
        "can_watch_members_through_ancestor_groups": true
      },
      {
        "id": "91", "name": "lp", "first_name": null, "last_name": null,
        "can_manage": "none", "can_grant_group_access": true, "can_watch_members": false,
        "can_manage_through_ancestor_groups": "none", "can_grant_group_access_through_ancestor_groups": true,
        "can_watch_members_through_ancestor_groups": false
      },
      {
        "id": "14", "name": "Our Friends",
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false,
        "can_manage_through_ancestor_groups": "none", "can_grant_group_access_through_ancestor_groups": false,
        "can_watch_members_through_ancestor_groups": false
      },
      {
        "id": "21", "name": "owner", "first_name": "Jean-Michel", "last_name": "Blanquer",
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false,
        "can_manage_through_ancestor_groups": "none", "can_grant_group_access_through_ancestor_groups": false,
        "can_watch_members_through_ancestor_groups": false
      },
      {
        "id": "61", "name": "zuck", "first_name": null, "last_name": null,
        "can_manage": "memberships_and_group", "can_grant_group_access": true, "can_watch_members": false,
        "can_manage_through_ancestor_groups": "memberships_and_group", "can_grant_group_access_through_ancestor_groups": true,
        "can_watch_members_through_ancestor_groups": false
      }
    ]
    """

  Scenario: Sort by id in descending order including managers of ancestor groups, get first two rows
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/managers?sort=-id&include_managers_of_ancestor_groups=1&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "91", "name": "lp", "first_name": null, "last_name": null,
        "can_manage": "none", "can_grant_group_access": true, "can_watch_members": false,
        "can_manage_through_ancestor_groups": "none", "can_grant_group_access_through_ancestor_groups": true,
        "can_watch_members_through_ancestor_groups": false
      },
      {
        "id": "81", "name": "larry", "first_name": null, "last_name": null,
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false,
        "can_manage_through_ancestor_groups": "memberships_and_group", "can_grant_group_access_through_ancestor_groups": true,
        "can_watch_members_through_ancestor_groups": true
      }
    ]
    """

  Scenario: The user is a member of the group
    Given I am the user with id "71"
    When I send a GET request to "/groups/13/managers"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51", "name": "billg", "first_name": null, "last_name": null,
        "can_manage": "memberships", "can_grant_group_access": false, "can_watch_members": true
      },
      {
        "id": "91", "name": "lp", "first_name": null, "last_name": null,
        "can_manage": "none", "can_grant_group_access": true, "can_watch_members": false
      },
      {
        "id": "14", "name": "Our Friends",
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false
      },
      {
        "id": "21", "name": "owner", "first_name": "Jean-Michel", "last_name": "Blanquer",
        "can_manage": "none", "can_grant_group_access": false, "can_watch_members": false
      },
      {
        "id": "61", "name": "zuck", "first_name": null, "last_name": null,
        "can_manage": "memberships_and_group", "can_grant_group_access": true, "can_watch_members": false
      }
    ]
    """

  Scenario: The user is a member of a descendant group
    Given I am the user with id "71"
    When I send a GET request to "/groups/12/managers"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "81", "name": "larry", "first_name": null, "last_name": null,
        "can_manage": "memberships", "can_grant_group_access": true, "can_watch_members": false
      }
    ]
    """

  Scenario: The user is a member of a team
    Given I am the user with id "81"
    When I send a GET request to "/groups/16/managers"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
