Feature: Get group invitations for the current user
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
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate          | idUserInviting |
      | 2  | 1             | 21           | invitationSent     | relativeTime(-169h)  | null           |
      | 3  | 2             | 21           | invitationRefused  | relativeTime(-168h)  | 1              |
      | 4  | 3             | 21           | requestSent        | relativeTime(-167h)  | 1              |
      | 5  | 4             | 21           | requestRefused     | relativeTime(-166h)  | 2              |
      | 6  | 5             | 21           | invitationAccepted | relativeTime(-165h)  | 2              |
      | 7  | 6             | 21           | requestAccepted    | relativeTime(-164h)  | 2              |
      | 8  | 7             | 21           | removed            | relativeTime(-163h)  | 1              |
      | 9  | 8             | 21           | left               | relativeTime(-162h)  | 1              |
      | 10 | 9             | 21           | direct             | relativeTime(-161h)  | 2              |
      | 11 | 1             | 22           | invitationSent     | relativeTime(-170h)  | 2              |

  Scenario: Show all invitations
    Given I am the user with ID "1"
    When I send a GET request to "/user/invitations"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "5",
        "inviting_user": {
          "id": "2",
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "group": {
          "id": "4",
          "name": "Our Friends",
          "description": "Group for our friends",
          "type": "Friends"
        },
        "status_date": "{groups_groups[4][sStatusDate]}",
        "type": "requestRefused"
      },
      {
        "id": "4",
        "inviting_user": {
          "id": "1",
          "first_name": "Jean-Michel",
          "last_name": "Blanquer",
          "login": "owner"
        },
        "group": {
          "id": "3",
          "name": "Our Club",
          "description": "Our club group",
          "type": "Club"
        },
        "status_date": "{groups_groups[3][sStatusDate]}",
        "type": "requestSent"
      },
      {
        "id": "2",
        "inviting_user": {},
        "group": {
          "id": "1",
          "name": "Our Class",
          "description": "Our class group",
          "type": "Class"
        },
        "status_date": "{groups_groups[1][sStatusDate]}",
        "type": "invitationSent"
      }
    ]
    """

  Scenario: Request the first row
    Given I am the user with ID "1"
    When I send a GET request to "/user/invitations?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "5",
        "inviting_user": {
          "id": "2",
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "group": {
          "id": "4",
          "name": "Our Friends",
          "description": "Group for our friends",
          "type": "Friends"
        },
        "status_date": "{groups_groups[4][sStatusDate]}",
        "type": "requestRefused"
      }
    ]
    """

  Scenario: Filter out old invitations
    Given I am the user with ID "1"
    When I send a GET request to "/user/invitations?within_weeks=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "5",
        "inviting_user": {
          "id": "2",
          "first_name": "John",
          "last_name": "Doe",
          "login": "user"
        },
        "group": {
          "id": "4",
          "name": "Our Friends",
          "description": "Group for our friends",
          "type": "Friends"
        },
        "status_date": "{groups_groups[4][sStatusDate]}",
        "type": "requestRefused"
      },
      {
        "id": "4",
        "inviting_user": {
          "id": "1",
          "first_name": "Jean-Michel",
          "last_name": "Blanquer",
          "login": "owner"
        },
        "group": {
          "id": "3",
          "name": "Our Club",
          "description": "Our club group",
          "type": "Club"
        },
        "status_date": "{groups_groups[3][sStatusDate]}",
        "type": "requestSent"
      }
    ]
    """
