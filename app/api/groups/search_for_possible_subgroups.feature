Feature: Search for possible subgroups
  Background:
    Given the database has the following table 'groups':
      | id | type    | name                                      | description            |
      | 1  | Class   | (the) Our Class                           | Our class group        |
      | 2  | Team    | (the) Our Team ___                        | null                   |
      | 3  | Club    | (the) Our Club                            | Our club group         |
      | 4  | Friends | (the) \|\|\|Our Friends \\\\\\%\\\\%\\ :) | Group for our friends  |
      | 5  | Other   | Other people                              | Group for other people |
      | 6  | Class   | Another Class                             | Another class group    |
      | 7  | Team    | Another %%%Team                           | Another team group     |
      | 8  | Club    | Another %%%Club                           | Another club group     |
      | 9  | Friends | Some other friends                        | Another friends group  |
      | 10 | Class   | The third class                           | The third class        |
      | 11 | User    | Another %%%User                           | Another user group     |
      | 12 | Club    | Club                                      | Parent group           |
      | 21 | User    | (the) user self                           |                        |
    And the database has the following table 'users':
      | login | temp_user | group_id | first_name  | last_name | grade |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | 3     |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 5               | 21             |
      | 6               | 21             |
      | 9               | 21             |
      | 10              | 21             |
      | 12              | 7              |
      | 12              | 8              |
      | 1               | 7              |
      | 4               | 21             |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage            |
      | 1        | 21         | memberships           |
      | 4        | 21         | memberships_and_group |
      | 2        | 5          | memberships_and_group |
      | 12       | 5          | memberships_and_group |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         |
      | 1        | 21        | invitation   |
      | 3        | 21        | join_request |

  Scenario: Search for groups with "the"
    Given I am the user with id "21"
    When I send a GET request to "/groups/possible-subgroups?search=the"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "name": "(the) Our Team ___",
        "description": null,
        "type": "Team"
      },
      {
        "id": "4",
        "name": "(the) |||Our Friends \\\\\\%\\\\%\\ :)",
        "description": "Group for our friends",
        "type": "Friends"
      },
      {
        "id": "7",
        "name": "Another %%%Team",
        "description": "Another team group",
        "type": "Team"
      },
      {
        "id": "8",
        "name": "Another %%%Club",
        "description": "Another club group",
        "type": "Club"
      }
    ]
    """

  Scenario: Search for groups with "the" (limit=2)
    Given I am the user with id "21"
    When I send a GET request to "/groups/possible-subgroups?search=the&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "name": "(the) Our Team ___",
        "description": null,
        "type": "Team"
      },
      {
        "id": "4",
        "name": "(the) |||Our Friends \\\\\\%\\\\%\\ :)",
        "description": "Group for our friends",
        "type": "Friends"
      }
    ]
    """

  Scenario: Search for groups with percent signs ("%%%")
    Given I am the user with id "21"
    When I send a GET request to "/groups/possible-subgroups?search=%25%25%25"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "7",
        "name": "Another %%%Team",
        "description": "Another team group",
        "type": "Team"
      },
      {
        "id": "8",
        "name": "Another %%%Club",
        "description": "Another club group",
        "type": "Club"
      }
    ]
    """

  Scenario: Search for groups with underscore signs
    Given I am the user with id "21"
    When I send a GET request to "/groups/possible-subgroups?search=___"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "name": "(the) Our Team ___",
        "description": null,
        "type": "Team"
      }
    ]
    """

  Scenario: Search for groups with pipe signs ("|||")
    Given I am the user with id "21"
    When I send a GET request to "/groups/possible-subgroups?search=%7C%7C%7C"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "4",
        "name": "(the) |||Our Friends \\\\\\%\\\\%\\ :)",
        "description": "Group for our friends",
        "type": "Friends"
      }
    ]
    """

  Scenario: Search with percent sign and slashes ("\\\%\\%\")
    Given I am the user with id "21"
    When I send a GET request to "/groups/possible-subgroups?search=%5C%5C%5C%25%5C%5C%25%5C"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "4",
        "name": "(the) |||Our Friends \\\\\\%\\\\%\\ :)",
        "description": "Group for our friends",
        "type": "Friends"
      }
    ]
    """
