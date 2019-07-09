Feature: List team descendants of the group (groupTeamDescendants)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 1  | owner  | 21          | 22           | Jean-Michel | Blanquer  | 10     |
      | 11 | johna  | 51          | 52           | John        | Adams     | 1      |
      | 12 | johnb  | 53          | 54           | John        | Baker     | 2      |
      | 13 | johnc  | 55          | 56           | John        | null      | 3      |
      | 14 | johnd  | 57          | 58           | null        | Davis     | -1     |
      | 15 | johne  | 59          | 60           | John        | Edwards   | null   |
      | 16 | janea  | 61          | 62           | Jane        | Adams     | 3      |
      | 17 | janeb  | 63          | 64           | Jane        | Baker     | null   |
      | 18 | janec  | 65          | 66           | Jane        | null      | 4      |
      | 19 | janed  | 67          | 68           | Jane        | Doe       | -2     |
      | 20 | janee  | 69          | 70           | Jane        | Edwards   | null   |
    And the database has the following table 'groups':
      | ID | sType     | sName          | iGrade |
      | 1  | Root      | Root 1         | -2     |
      | 3  | Root      | Root 2         | -2     |
      | 11 | Class     | Our Class      | -2     |
      | 12 | Class     | Other Class    | -2     |
      | 13 | Class     | Special Class  | -2     |
      | 14 | Team      | Super Team     | -2     |
      | 15 | Team      | Our Team       | -1     |
      | 16 | Team      | First Team     | 0      |
      | 17 | Other     | A custom group | -2     |
      | 18 | Club      | Our Club       | -2     |
      | 20 | Friends   | My Friends     | -2     |
      | 21 | UserSelf  | owner          | -2     |
      | 51 | UserSelf  | johna          | -2     |
      | 53 | UserSelf  | johnb          | -2     |
      | 55 | UserSelf  | johnc          | -2     |
      | 57 | UserSelf  | johnd          | -2     |
      | 59 | UserSelf  | johne          | -2     |
      | 61 | UserSelf  | janea          | -2     |
      | 63 | UserSelf  | janeb          | -2     |
      | 65 | UserSelf  | janec          | -2     |
      | 67 | UserSelf  | janed          | -2     |
      | 69 | UserSelf  | janee          | -2     |
      | 22 | UserAdmin | owner-admin    | -2     |
      | 52 | UserAdmin | johna-admin    | -2     |
      | 54 | UserAdmin | johnb-admin    | -2     |
      | 56 | UserAdmin | johnc-admin    | -2     |
      | 58 | UserAdmin | johnd-admin    | -2     |
      | 60 | UserAdmin | johne-admin    | -2     |
      | 62 | UserAdmin | janea-admin    | -2     |
      | 64 | UserAdmin | janeb-admin    | -2     |
      | 66 | UserAdmin | janec-admin    | -2     |
      | 68 | UserAdmin | janed-admin    | -2     |
      | 70 | UserAdmin | janee-admin    | -2     |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType              |
      | 1             | 11           | direct             |
      | 3             | 13           | direct             |
      | 11            | 14           | direct             |
      | 11            | 16           | direct             |
      | 11            | 17           | direct             |
      | 11            | 18           | direct             |
      | 11            | 59           | requestAccepted    |
      | 13            | 14           | direct             |
      | 13            | 15           | direct             |
      | 13            | 69           | invitationAccepted |
      | 14            | 51           | requestAccepted    |
      | 14            | 53           | requestAccepted    |
      | 14            | 55           | invitationAccepted |
      | 15            | 57           | direct             |
      | 15            | 59           | requestAccepted    |
      | 15            | 61           | invitationAccepted |
      | 15            | 63           | invitationRejected |
      | 15            | 65           | left               |
      | 15            | 67           | invitationSent     |
      | 15            | 69           | requestSent        |
      | 16            | 51           | invitationRefused  |
      | 16            | 53           | requestRefused     |
      | 16            | 55           | removed            |
      | 16            | 63           | direct             |
      | 16            | 65           | requestAccepted    |
      | 16            | 67           | invitationAccepted |
      | 20            | 21           | direct             |
      | 22            | 1            | direct             |
      | 22            | 3            | direct             |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 1               | 1            | 1       |
      | 1               | 11           | 0       |
      | 1               | 12           | 0       |
      | 1               | 14           | 0       |
      | 1               | 16           | 0       |
      | 1               | 17           | 0       |
      | 1               | 18           | 0       |
      | 1               | 51           | 0       |
      | 1               | 53           | 0       |
      | 1               | 55           | 0       |
      | 1               | 59           | 0       |
      | 1               | 63           | 0       |
      | 1               | 65           | 0       |
      | 1               | 67           | 0       |
      | 3               | 3            | 1       |
      | 3               | 13           | 0       |
      | 3               | 15           | 0       |
      | 3               | 51           | 0       |
      | 3               | 53           | 0       |
      | 3               | 55           | 0       |
      | 3               | 61           | 0       |
      | 3               | 63           | 0       |
      | 3               | 65           | 0       |
      | 3               | 69           | 0       |
      | 11              | 11           | 1       |
      | 11              | 14           | 0       |
      | 11              | 16           | 0       |
      | 11              | 17           | 0       |
      | 11              | 18           | 0       |
      | 11              | 51           | 0       |
      | 11              | 53           | 0       |
      | 11              | 55           | 0       |
      | 11              | 59           | 0       |
      | 11              | 63           | 0       |
      | 11              | 65           | 0       |
      | 11              | 67           | 0       |
      | 12              | 12           | 1       |
      | 13              | 13           | 1       |
      | 13              | 14           | 0       |
      | 13              | 15           | 0       |
      | 13              | 51           | 0       |
      | 13              | 53           | 0       |
      | 13              | 55           | 0       |
      | 13              | 61           | 0       |
      | 13              | 63           | 0       |
      | 13              | 65           | 0       |
      | 13              | 69           | 0       |
      | 14              | 14           | 1       |
      | 14              | 51           | 0       |
      | 14              | 53           | 0       |
      | 14              | 55           | 0       |
      | 15              | 15           | 1       |
      | 15              | 61           | 0       |
      | 15              | 63           | 0       |
      | 15              | 65           | 0       |
      | 16              | 16           | 1       |
      | 16              | 63           | 0       |
      | 16              | 65           | 0       |
      | 16              | 67           | 0       |
      | 20              | 20           | 1       |
      | 20              | 21           | 0       |
      | 21              | 21           | 1       |
      | 22              | 1            | 0       |
      | 22              | 11           | 0       |
      | 22              | 12           | 0       |
      | 22              | 13           | 0       |
      | 22              | 14           | 0       |
      | 22              | 15           | 0       |
      | 22              | 16           | 0       |
      | 22              | 17           | 0       |
      | 22              | 18           | 0       |
      | 22              | 22           | 1       |
      | 22              | 51           | 0       |
      | 22              | 53           | 0       |
      | 22              | 55           | 0       |
      | 22              | 59           | 0       |
      | 22              | 61           | 0       |
      | 22              | 63           | 0       |
      | 22              | 65           | 0       |
      | 22              | 67           | 0       |
      | 22              | 69           | 0       |

  Scenario: Get descendant teams
    Given I am the user with ID "1"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30T20:19:05Z"
    When I send a GET request to "/groups/1/team-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "grade": 0,
        "id": "16",
        "members": [
          {
            "first_name": "Jane",
            "grade": null,
            "id": "17",
            "last_name": "Baker",
            "login": "janeb",
            "self_group_id": 63
          },
          {
            "first_name": "Jane",
            "grade": 4,
            "id": "18",
            "last_name": null,
            "login": "janec",
            "self_group_id": 65
          },
          {
            "first_name": "Jane",
            "grade": -2,
            "id": "19",
            "last_name": "Doe",
            "login": "janed",
            "self_group_id": 67
          }
        ],
        "name": "First Team",
        "parents": [
          {
            "id": "11",
            "name": "Our Class"
          }
        ]
      },
      {
        "grade": -2,
        "id": "14",
        "members": [
          {
            "first_name": "John",
            "grade": 1,
            "id": "11",
            "last_name": "Adams",
            "login": "johna",
            "self_group_id": 51
          },
          {
            "first_name": "John",
            "grade": 2,
            "id": "12",
            "last_name": "Baker",
            "login": "johnb",
            "self_group_id": 53
          },
          {
            "first_name": "John",
            "grade": 3,
            "id": "13",
            "last_name": null,
            "login": "johnc",
            "self_group_id": 55
          }
        ],
        "name": "Super Team",
        "parents": [
          {
            "id": "11",
            "name": "Our Class"
          }
        ]
      }
    ]
    """

  Scenario: Get the first team from the list
    Given I am the user with ID "1"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30T20:19:05Z"
    When I send a GET request to "/groups/1/team-descendants?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "grade": 0,
        "id": "16",
        "members": [
          {
            "first_name": "Jane",
            "grade": null,
            "id": "17",
            "last_name": "Baker",
            "login": "janeb",
            "self_group_id": 63
          },
          {
            "first_name": "Jane",
            "grade": 4,
            "id": "18",
            "last_name": null,
            "login": "janec",
            "self_group_id": 65
          },
          {
            "first_name": "Jane",
            "grade": -2,
            "id": "19",
            "last_name": "Doe",
            "login": "janed",
            "self_group_id": 67
          }
        ],
        "name": "First Team",
        "parents": [
          {
            "id": "11",
            "name": "Our Class"
          }
        ]
      }
    ]
    """

  Scenario: Get teams skipping the first one
    Given I am the user with ID "1"
    # here we fixate time_spent even if it depends on NOW()
    And the DB time now is "2019-06-30T20:19:05Z"
    When I send a GET request to "/groups/1/team-descendants?from.name=First%20Team&from.id=16"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "grade": -2,
        "id": "14",
        "members": [
          {
            "first_name": "John",
            "grade": 1,
            "id": "11",
            "last_name": "Adams",
            "login": "johna",
            "self_group_id": 51
          },
          {
            "first_name": "John",
            "grade": 2,
            "id": "12",
            "last_name": "Baker",
            "login": "johnb",
            "self_group_id": 53
          },
          {
            "first_name": "John",
            "grade": 3,
            "id": "13",
            "last_name": null,
            "login": "johnc",
            "self_group_id": 55
          }
        ],
        "name": "Super Team",
        "parents": [
          {
            "id": "11",
            "name": "Our Class"
          }
        ]
      }
    ]
    """

  Scenario: No teams
    Given I am the user with ID "1"
    When I send a GET request to "/groups/16/team-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
