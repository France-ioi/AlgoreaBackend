Feature: Get requests for group_id
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | 3      |
      | 2  | user   | 0        | 11          | 12           | John        | Doe       | 1      |
      | 3  | jane   | 0        | 31          | 32           | Jane        | Doe       | 2      |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 75 | 22              | 13           | 0       | 0        |
      | 76 | 13              | 11           | 0       | 0        |
      | 77 | 22              | 11           | 0       | 0        |
      | 78 | 21              | 21           | 1       | 0        |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate          | idUserInviting |
      | 1  | 13            | 21           | invitationSent     | relativeTime(-170h)  | 2              |
      | 2  | 13            | 11           | invitationRefused  | relativeTime(-169h)  | 3              |
      | 3  | 13            | 31           | requestSent        | relativeTime(-168h)  | 1              |
      | 4  | 13            | 22           | requestRefused     | relativeTime(-167h)  | 2              |
      | 5  | 14            | 11           | invitationSent     | null                 | 2              |
      | 6  | 14            | 31           | invitationRefused  | null                 | 3              |
      | 7  | 14            | 21           | requestSent        | null                 | 1              |
      | 8  | 14            | 22           | requestRefused     | null                 | 2              |
      | 9  | 13            | 121          | invitationAccepted | 2017-05-29T06:38:38Z | 2              |
      | 10 | 13            | 111          | requestAccepted    | null                 | 3              |
      | 11 | 13            | 131          | removed            | null                 | 1              |
      | 12 | 13            | 122          | left               | null                 | 2              |
      | 13 | 13            | 123          | direct             | null                 | 2              |

  Scenario: User is an admin of the group (default sort)
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/requests"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": 4,
        "inviting_user": {
          "first_name": "John",
          "id": 2,
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {},
        "status_date": "{groups_groups[4][sStatusDate]}",
        "type": "requestRefused"
      },
      {
        "id": 3,
        "inviting_user": {
          "first_name": "Jean-Michel",
          "id": 1,
          "last_name": "Blanquer",
          "login": "owner"
        },
        "joining_user": {
          "first_name": "Jane",
          "grade": 2,
          "id": 3,
          "last_name": "Doe",
          "login": "jane"
        },
        "status_date": "{groups_groups[3][sStatusDate]}",
        "type": "requestSent"
      },
      {
        "id": 2,
        "inviting_user": {
          "first_name": "Jane",
          "id": 3,
          "last_name": "Doe",
          "login": "jane"
        },
        "joining_user": {
          "grade": 1,
          "id": 2,
          "login": "user"
        },
        "status_date": "{groups_groups[2][sStatusDate]}",
        "type": "invitationRefused"
      },
      {
        "id": 1,
        "inviting_user": {
          "first_name": "John",
          "id": 2,
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {
          "grade": 3,
          "id": 1,
          "login": "owner"
        },
        "status_date": "{groups_groups[1][sStatusDate]}",
        "type": "invitationSent"
      }
    ]
    """

  Scenario: User is an admin of the group (sort by type)
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/requests?sort=type"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": 1,
        "inviting_user": {
          "first_name": "John",
          "id": 2,
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {
          "grade": 3,
          "id": 1,
          "login": "owner"
        },
        "status_date": "{groups_groups[1][sStatusDate]}",
        "type": "invitationSent"
      },
      {
        "id": 3,
        "inviting_user": {
          "first_name": "Jean-Michel",
          "id": 1,
          "last_name": "Blanquer",
          "login": "owner"
        },
        "joining_user": {
          "first_name": "Jane",
          "grade": 2,
          "id": 3,
          "last_name": "Doe",
          "login": "jane"
        },
        "status_date": "{groups_groups[3][sStatusDate]}",
        "type": "requestSent"
      },
      {
        "id": 2,
        "inviting_user": {
          "first_name": "Jane",
          "id": 3,
          "last_name": "Doe",
          "login": "jane"
        },
        "joining_user": {
          "grade": 1,
          "id": 2,
          "login": "user"
        },
        "status_date": "{groups_groups[2][sStatusDate]}",
        "type": "invitationRefused"
      },
      {
        "id": 4,
        "inviting_user": {
          "first_name": "John",
          "id": 2,
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {},
        "status_date": "{groups_groups[4][sStatusDate]}",
        "type": "requestRefused"
      }
    ]
    """

  Scenario: User is an admin of the group (sort by joining user's login)
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/requests?sort=joining_user.login"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": 4,
        "inviting_user": {
          "first_name": "John",
          "id": 2,
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {},
        "status_date": "{groups_groups[4][sStatusDate]}",
        "type": "requestRefused"
      },
      {
        "id": 3,
        "inviting_user": {
          "first_name": "Jean-Michel",
          "id": 1,
          "last_name": "Blanquer",
          "login": "owner"
        },
        "joining_user": {
          "first_name": "Jane",
          "grade": 2,
          "id": 3,
          "last_name": "Doe",
          "login": "jane"
        },
        "status_date": "{groups_groups[3][sStatusDate]}",
        "type": "requestSent"
      },
      {
        "id": 1,
        "inviting_user": {
          "first_name": "John",
          "id": 2,
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {
          "grade": 3,
          "id": 1,
          "login": "owner"
        },
        "status_date": "{groups_groups[1][sStatusDate]}",
        "type": "invitationSent"
      },
      {
        "id": 2,
        "inviting_user": {
          "first_name": "Jane",
          "id": 3,
          "last_name": "Doe",
          "login": "jane"
        },
        "joining_user": {
          "grade": 1,
          "id": 2,
          "login": "user"
        },
        "status_date": "{groups_groups[2][sStatusDate]}",
        "type": "invitationRefused"
      }
    ]
    """

  Scenario: User is an admin of the group; request the first row
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/requests?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": 4,
        "inviting_user": {
          "first_name": "John",
          "id": 2,
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {},
        "status_date": "{groups_groups[4][sStatusDate]}",
        "type": "requestRefused"
      }
    ]
    """

  Scenario: User is an admin of the group; filter out old rejections
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/requests?rejections_within_weeks=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": 4,
        "inviting_user": {
          "first_name": "John",
          "id": 2,
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {},
        "status_date": "{groups_groups[4][sStatusDate]}",
        "type": "requestRefused"
      },
      {
        "id": 3,
        "inviting_user": {
          "first_name": "Jean-Michel",
          "id": 1,
          "last_name": "Blanquer",
          "login": "owner"
        },
        "joining_user": {
          "first_name": "Jane",
          "grade": 2,
          "id": 3,
          "last_name": "Doe",
          "login": "jane"
        },
        "status_date": "{groups_groups[3][sStatusDate]}",
        "type": "requestSent"
      },
      {
        "id": 1,
        "inviting_user": {
          "first_name": "John",
          "id": 2,
          "last_name": "Doe",
          "login": "user"
        },
        "joining_user": {
          "grade": 3,
          "id": 1,
          "login": "owner"
        },
        "status_date": "{groups_groups[1][sStatusDate]}",
        "type": "invitationSent"
      }
    ]
    """
