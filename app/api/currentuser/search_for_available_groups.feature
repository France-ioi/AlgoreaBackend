Feature: Search for groups available to the current user
  Background:
    Given the database has the following table 'groups':
      | id | type                | name                                  | description            | is_public |
      | 1  | Class               | amazing Class                         | Our class group        | 1         |
      | 2  | Team                | amazing Team                          | null                   | 1         |
      | 3  | Club                | amazing Club                          | Our club group         | 1         |
      | 4  | Friends             | the amazing Friends \\\\\\%\\\\%\\ :) | Group for our friends  | 1         |
      | 5  | Other               | Other people                          | Group for other people | 1         |
      | 6  | Class               | Another amazing Class                 | Another class group    | 1         |
      | 7  | Team                | Another amazing Team                  | Another team group     | 1         |
      | 8  | Club                | Another amazing Club                  | Another club group     | 1         |
      | 9  | Friends             | Some other friends                    | Another friends group  | 1         |
      | 10 | Class               | amazing third class                   | The third class        | 1         |
      | 11 | User                | Another amazing User                  | Another user group     | 1         |
      | 21 | User                | amazing user self                     |                        | 0         |
      | 22 | ContestParticipants | amazing                               |                        | 1         |
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

  Scenario: Search for groups with "amazing"
    Given I am the user with id "21"
    When I send a GET request to "/current-user/available-groups?search=amazing"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "name": "amazing Team",
        "description": null,
        "type": "Team"
      },
      {
        "id": "4",
        "name": "the amazing Friends \\\\\\%\\\\%\\ :)",
        "description": "Group for our friends",
        "type": "Friends"
      },
      {
        "description": "Another team group",
        "id": "7",
        "name": "Another amazing Team",
        "type": "Team"
      },
      {
        "description": "Another club group",
        "id": "8",
        "name": "Another amazing Club",
        "type": "Club"
      }
    ]
    """

  Scenario: Should treat the words in the search string as "AND"
    Given I am the user with id "21"
    When I send a GET request to "/current-user/available-groups?search=amazing%20team"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "name": "amazing Team",
        "description": null,
        "type": "Team"
      },
      {
        "description": "Another team group",
        "id": "7",
        "name": "Another amazing Team",
 		    "type": "Team"
 		  }
    ]
    """

  Scenario: Search for groups with "amazing", start from the second row
    Given I am the user with id "21"
    When I send a GET request to "/current-user/available-groups?search=amazing&from.id=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "4",
        "name": "the amazing Friends \\\\\\%\\\\%\\ :)",
        "description": "Group for our friends",
        "type": "Friends"
      },
      {
        "description": "Another team group",
        "id": "7",
        "name": "Another amazing Team",
        "type": "Team"
      },
      {
        "description": "Another club group",
        "id": "8",
        "name": "Another amazing Club",
        "type": "Club"
      }
    ]
    """

  Scenario: Search for groups with "amazing" (limit=2)
    Given I am the user with id "21"
    When I send a GET request to "/current-user/available-groups?search=amazing&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "name": "amazing Team",
        "description": null,
        "type": "Team"
      },
      {
        "description": "Group for our friends",
        "id": "4",
        "name": "the amazing Friends \\\\\\%\\\\%\\ :)",
        "type": "Friends"
      }
    ]
    """

  Scenario: Search for groups which begins with the search word
    Given I am the user with id "21"
    When I send a GET request to "/current-user/available-groups?search=friend"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "4",
        "name": "the amazing Friends \\\\\\%\\\\%\\ :)",
        "description": "Group for our friends",
        "type": "Friends"
      }
    ]
    """
