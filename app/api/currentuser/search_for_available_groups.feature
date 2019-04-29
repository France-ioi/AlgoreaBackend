Feature: Search for groups available to the current user
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | 3      |
    And the database has the following table 'groups':
      | ID | sType     | sName                                     | sDescription           | bFreeAccess |
      | 1  | Class     | (the) Our Class                           | Our class group        | 1           |
      | 2  | Team      | (the) Our Team ___                        | Our team group         | 1           |
      | 3  | Club      | (the) Our Club                            | Our club group         | 1           |
      | 4  | Friends   | (the) \|\|\|Our Friends \\\\\\%\\\\%\\ :) | Group for our friends  | 1           |
      | 5  | Other     | Other people                              | Group for other people | 1           |
      | 6  | Class     | Another Class                             | Another class group    | 1           |
      | 7  | Team      | Another %%%Team                           | Another team group     | 1           |
      | 8  | Club      | Another %%%Club                           | Another club group     | 1           |
      | 9  | Friends   | Some other friends                        | Another friends group  | 1           |
      | 21 | UserSelf  | (the) user self                           |                        | 0           |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate          |
      | 2  | 1             | 21           | invitationSent     | 2017-02-29T06:38:38Z |
      | 3  | 2             | 21           | invitationRefused  | 2017-03-29T06:38:38Z |
      | 4  | 3             | 21           | requestSent        | 2017-04-29T06:38:38Z |
      | 5  | 4             | 21           | requestRefused     | 2017-05-29T06:38:38Z |
      | 6  | 5             | 21           | invitationAccepted | 2017-06-29T06:38:38Z |
      | 7  | 6             | 21           | requestAccepted    | 2017-07-29T06:38:38Z |
      | 8  | 7             | 21           | removed            | 2017-08-29T06:38:38Z |
      | 9  | 8             | 21           | left               | 2017-09-29T06:38:38Z |
      | 10 | 9             | 21           | direct             | 2017-10-29T06:38:38Z |
      | 11 | 1             | 22           | direct             | 2017-11-29T06:38:38Z |

  Scenario: Search for groups with "the"
    Given I am the user with ID "1"
    When I send a GET request to "/current-user/available-groups?search=the"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "name": "(the) Our Team ___",
        "description": "Our team group",
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
    Given I am the user with ID "1"
    When I send a GET request to "/current-user/available-groups?search=the&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "name": "(the) Our Team ___",
        "description": "Our team group",
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

  Scenario: Search for groups with percent signs
    Given I am the user with ID "1"
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
    Given I am the user with ID "1"
    When I send a GET request to "/current-user/available-groups?search=___"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "2",
        "name": "(the) Our Team ___",
        "description": "Our team group",
        "type": "Team"
      }
    ]
    """

  Scenario: Search for groups with pipe signs
    Given I am the user with ID "1"
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

  Scenario: Search with percent sign and slashes
    Given I am the user with ID "1"
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
