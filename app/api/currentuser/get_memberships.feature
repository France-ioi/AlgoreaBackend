Feature: Get group memberships for the current user
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | 3      |
      | 2  | user   | 0        | 11          | 12           | John        | Doe       | 1      |
    And the database has the following table 'groups':
      | ID | sType     | sName              | sDescription           |
      | 1  | Class     | Our Class          | Our class group        |
      | 2  | Team      | Our Team           | Our team group         |
      | 3  | Club      | Our Club           | Our club group         |
      | 4  | Friends   | Our Friends        | Group for our friends  |
      | 5  | Other     | Other people       | Group for other people |
      | 6  | Class     | Another Class      | Another class group    |
      | 7  | Team      | Another Team       | Another team group     |
      | 8  | Club      | Another Club       | Another club group     |
      | 9  | Friends   | Some other friends | Another friends group  |
      | 11 | UserSelf  | user self          |                        |
      | 12 | UserAdmin | user admin         |                        |
      | 21 | UserSelf  | owner self         |                        |
      | 22 | UserAdmin | owner admin        |                        |
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

  Scenario: Show all invitations
    Given I am the user with ID "1"
    When I send a GET request to "/current-user/memberships"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "10",
        "group": {
          "id": "9",
          "name": "Some other friends",
          "description": "Another friends group",
          "type": "Friends"
        },
        "status_date": "2017-10-29T06:38:38Z",
        "type": "direct"
      },
      {
        "id": "7",
        "group": {
          "id": "6",
          "name": "Another Class",
          "description": "Another class group",
          "type": "Class"
        },
        "status_date": "2017-07-29T06:38:38Z",
        "type": "requestAccepted"
      },
      {
        "id": "6",
        "group": {
          "id": "5",
          "name": "Other people",
          "description": "Group for other people",
          "type": "Other"
        },
        "status_date": "2017-06-29T06:38:38Z",
        "type": "invitationAccepted"
      }
    ]
    """

  Scenario: Request the first row
    Given I am the user with ID "1"
    When I send a GET request to "/current-user/memberships?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "10",
        "group": {
          "id": "9",
          "name": "Some other friends",
          "description": "Another friends group",
          "type": "Friends"
        },
        "status_date": "2017-10-29T06:38:38Z",
        "type": "direct"
      }
    ]
    """

