Feature: Get group memberships history for the current user
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade | sNotificationReadDate |
      | 1  | owner  | 21          | 22           | Jean-Michel | Blanquer  | 3      | 2017-06-29T06:38:38Z  |
      | 2  | user   | 11          | 12           | John        | Doe       | 1      | null                  |
    And the database has the following table 'groups':
      | ID | sType     | sName              |
      | 1  | Class     | Our Class          |
      | 2  | Team      | Our Team           |
      | 3  | Club      | Our Club           |
      | 4  | Friends   | Our Friends        |
      | 5  | Other     | Other people       |
      | 6  | Class     | Another Class      |
      | 7  | Team      | Another Team       |
      | 8  | Club      | Another Club       |
      | 9  | Friends   | Some other friends |
      | 11 | UserSelf  | user self          |
      | 12 | UserAdmin | user admin         |
      | 21 | UserSelf  | owner self         |
      | 22 | UserAdmin | owner admin        |
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
      | 12 | 1             | 11           | invitationSent     | 2016-02-29T06:38:38Z |
      | 13 | 2             | 11           | invitationRefused  | 2016-03-29T06:38:38Z |
      | 14 | 3             | 11           | requestSent        | 2016-04-29T06:38:38Z |
      | 15 | 4             | 11           | requestRefused     | 2016-05-29T06:38:38Z |
      | 16 | 5             | 11           | invitationAccepted | 2016-06-29T06:38:38Z |
      | 17 | 6             | 11           | requestAccepted    | 2016-07-29T06:38:38Z |
      | 18 | 7             | 11           | removed            | 2016-08-29T06:38:38Z |
      | 19 | 8             | 11           | left               | 2016-09-29T06:38:38Z |
      | 20 | 9             | 11           | direct             | 2016-10-29T06:38:38Z |

  Scenario: Show all the history (with sNotificationReadDate set)
    Given I am the user with ID "1"
    When I send a GET request to "/current-user/group-memberships-history"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "9",
        "group": {
          "name": "Another Club",
          "type": "Club"
        },
        "status_date": "2017-09-29T06:38:38Z",
        "type": "left"
      },
      {
        "id": "8",
        "group": {
          "name": "Another Team",
          "type": "Team"
        },
        "status_date": "2017-08-29T06:38:38Z",
        "type": "removed"
      },
      {
        "id": "7",
        "group": {
          "name": "Another Class",
          "type": "Class"
        },
        "status_date": "2017-07-29T06:38:38Z",
        "type": "requestAccepted"
      },
      {
        "id": "6",
        "group": {
          "name": "Other people",
          "type": "Other"
        },
        "status_date": "2017-06-29T06:38:38Z",
        "type": "invitationAccepted"
      }
    ]
    """

  Scenario: Show all the history (without sNotificationReadDate set)
    Given I am the user with ID "2"
    When I send a GET request to "/current-user/group-memberships-history"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "19",
        "group": {
          "name": "Another Club",
          "type": "Club"
        },
        "status_date": "2016-09-29T06:38:38Z",
        "type": "left"
      },
      {
        "id": "18",
        "group": {
          "name": "Another Team",
          "type": "Team"
        },
        "status_date": "2016-08-29T06:38:38Z",
        "type": "removed"
      },
      {
        "id": "17",
        "group": {
          "name": "Another Class",
          "type": "Class"
        },
        "status_date": "2016-07-29T06:38:38Z",
        "type": "requestAccepted"
      },
      {
        "id": "16",
        "group": {
          "name": "Other people",
          "type": "Other"
        },
        "status_date": "2016-06-29T06:38:38Z",
        "type": "invitationAccepted"
      },
      {
        "id": "15",
        "group": {
          "name": "Our Friends",
          "type": "Friends"
        },
        "status_date": "2016-05-29T06:38:38Z",
        "type": "requestRefused"
      },
      {
        "id": "14",
        "group": {
          "name": "Our Club",
          "type": "Club"
        },
        "status_date": "2016-04-29T06:38:38Z",
        "type": "requestSent"
      },
      {
        "id": "13",
        "group": {
          "name": "Our Team",
          "type": "Team"
        },
        "status_date": "2016-03-29T06:38:38Z",
        "type": "invitationRefused"
      },
      {
        "id": "12",
        "group": {
          "name": "Our Class",
          "type": "Class"
        },
        "status_date": "2016-02-29T06:38:38Z",
        "type": "invitationSent"
      }
    ]
    """

  Scenario: Request the first row
    Given I am the user with ID "1"
    When I send a GET request to "/current-user/group-memberships-history?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "9",
        "group": {
          "name": "Another Club",
          "type": "Club"
        },
        "status_date": "2017-09-29T06:38:38Z",
        "type": "left"
      }
    ]
    """

  Scenario: Request the first row starting from some date
    Given I am the user with ID "1"
    When I send a GET request to "/current-user/group-memberships-history?limit=1&from.status_date=2017-07-29T06:38:38Z&from.id=7"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "6",
        "group": {
          "name": "Other people",
          "type": "Other"
        },
        "status_date": "2017-06-29T06:38:38Z",
        "type": "invitationAccepted"
      }
    ]
    """

