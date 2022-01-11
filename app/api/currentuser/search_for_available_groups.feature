Feature: Search for groups available to the current user
  Background:
    Given the database has the following table 'groups':
      | id | type                | name                                      | description            | is_public |
      | 1  | Class               | (the) Our Class                           | Our class group        | 1         |
      | 2  | Team                | (the) Our Team ___                        | null                   | 1         |
      | 3  | Club                | (the) Our Club                            | Our club group         | 1         |
      | 4  | Friends             | (the) \|\|\|Our Friends \\\\\\%\\\\%\\ :) | Group for our friends  | 1         |
      | 5  | Other               | Other people                              | Group for other people | 1         |
      | 6  | Class               | Another Class                             | Another class group    | 1         |
      | 7  | Team                | Another %%%Team                           | Another team group     | 1         |
      | 8  | Club                | Another %%%Club                           | Another club group     | 1         |
      | 9  | Friends             | Some other friends                        | Another friends group  | 1         |
      | 10 | Class               | The third class                           | The third class        | 1         |
      | 11 | User                | Another %%%User                           | Another user group     | 1         |
      | 21 | User                | (the) user self                           |                        | 0         |
      | 22 | ContestParticipants | the                                       |                        | 1         |
    And the database has the following table 'users':
      | login | temp_user | group_id | first_name  | last_name | grade |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | 3     |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | expires_at          |
      | 2               | 21             | 2019-05-30 11:00:00 |
      | 5               | 21             | 9999-12-31 23:59:59 |
      | 6               | 21             | 9999-12-31 23:59:59 |
      | 9               | 21             | 9999-12-31 23:59:59 |
      | 10              | 21             | 9999-12-31 23:59:59 |
      | 1               | 7              | 9999-12-31 23:59:59 |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         |
      | 1        | 21        | invitation   |
      | 3        | 21        | join_request |

  Scenario: Search for groups with "the"
    Given I am the user with id "21"
    When I send a GET request to "/current-user/available-groups?search=the"
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

  Scenario: Search for groups with "the", start from the second row
    Given I am the user with id "21"
    When I send a GET request to "/current-user/available-groups?search=the&from.id=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
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
    When I send a GET request to "/current-user/available-groups?search=the&limit=2"
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
    When I send a GET request to "/current-user/available-groups?search=%25%25%25"
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
    When I send a GET request to "/current-user/available-groups?search=___"
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
    When I send a GET request to "/current-user/available-groups?search=%7C%7C%7C"
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
    When I send a GET request to "/current-user/available-groups?search=%5C%5C%5C%25%5C%5C%25%5C"
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
