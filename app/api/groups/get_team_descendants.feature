Feature: List team descendants of the group (groupTeamDescendantView)
  Background:
    Given the database has the following table 'groups':
      | id | type    | name           | grade |
      | 1  | Base    | Root 1         | -2    |
      | 3  | Base    | Root 2         | -2    |
      | 11 | Class   | Our Class      | -2    |
      | 12 | Class   | Other Class    | -2    |
      | 13 | Class   | Special Class  | -2    |
      | 14 | Team    | Super Team     | -2    |
      | 15 | Team    | Our Team       | -1    |
      | 16 | Team    | First Team     | 0     |
      | 17 | Other   | A custom group | -2    |
      | 18 | Club    | Our Club       | -2    |
      | 20 | Friends | My Friends     | -2    |
      | 21 | User    | owner          | -2    |
      | 22 | Class   | Managed Class  | -2    |
      | 51 | User    | johna          | -2    |
      | 53 | User    | johnb          | -2    |
      | 55 | User    | johnc          | -2    |
      | 57 | User    | johnd          | -2    |
      | 59 | User    | johne          | -2    |
      | 61 | User    | janea          | -2    |
      | 63 | User    | janeb          | -2    |
      | 65 | User    | janec          | -2    |
      | 67 | User    | janed          | -2    |
      | 69 | User    | janee          | -2    |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name | grade |
      | owner | 21       | Jean-Michel | Blanquer  | 10    |
      | johna | 51       | John        | Adams     | 1     |
      | johnb | 53       | John        | Baker     | 2     |
      | johnc | 55       | John        | null      | 3     |
      | johnd | 57       | null        | Davis     | -1    |
      | johne | 59       | John        | Edwards   | null  |
      | janea | 61       | Jane        | Adams     | 3     |
      | janeb | 63       | Jane        | Baker     | null  |
      | janec | 65       | Jane        | null      | 4     |
      | janed | 67       | Jane        | Doe       | -2    |
      | janee | 69       | Jane        | Edwards   | null  |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 1        | 21         |
      | 22       | 20         |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 1               | 11             | null                           |
      | 3               | 13             | null                           |
      | 11              | 14             | null                           |
      | 11              | 16             | null                           |
      | 11              | 17             | null                           |
      | 11              | 18             | null                           |
      | 11              | 59             | null                           |
      | 13              | 14             | null                           |
      | 13              | 15             | null                           |
      | 13              | 69             | null                           |
      | 14              | 51             | null                           |
      | 14              | 53             | null                           |
      | 14              | 55             | null                           |
      | 15              | 57             | null                           |
      | 15              | 59             | null                           |
      | 15              | 61             | null                           |
      | 16              | 63             | null                           |
      | 16              | 65             | null                           |
      | 16              | 67             | null                           |
      | 20              | 21             | null                           |
      | 20              | 67             | 2019-05-30 11:00:00            |
      | 22              | 63             | 2019-05-30 11:00:00            |
      | 22              | 65             | 2019-05-30 11:00:00            |
    And the groups ancestors are computed

  Scenario: Get descendant teams
    Given I am the user with id "21"
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
            "last_name": "Baker",
            "login": "janeb",
            "group_id": 63
          },
          {
            "first_name": "Jane",
            "grade": 4,
            "last_name": null,
            "login": "janec",
            "group_id": 65
          },
          {
            "grade": -2,
            "login": "janed",
            "group_id": 67
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
            "grade": 1,
            "login": "johna",
            "group_id": 51
          },
          {
            "grade": 2,
            "login": "johnb",
            "group_id": 53
          },
          {
            "grade": 3,
            "login": "johnc",
            "group_id": 55
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
    Given I am the user with id "21"
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
            "last_name": "Baker",
            "login": "janeb",
            "group_id": 63
          },
          {
            "first_name": "Jane",
            "grade": 4,
            "last_name": null,
            "login": "janec",
            "group_id": 65
          },
          {
            "grade": -2,
            "login": "janed",
            "group_id": 67
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
    Given I am the user with id "21"
    When I send a GET request to "/groups/1/team-descendants?from.id=16"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "grade": -2,
        "id": "14",
        "members": [
          {
            "grade": 1,
            "login": "johna",
            "group_id": 51
          },
          {
            "grade": 2,
            "login": "johnb",
            "group_id": 53
          },
          {
            "grade": 3,
            "login": "johnc",
            "group_id": 55
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
    Given I am the user with id "21"
    When I send a GET request to "/groups/16/team-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
