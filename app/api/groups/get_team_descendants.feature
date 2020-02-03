Feature: List team descendants of the group (groupTeamDescendantView)
  Background:
    Given the database has the following table 'groups':
      | id | type     | name           | grade |
      | 1  | Base     | Root 1         | -2    |
      | 3  | Base     | Root 2         | -2    |
      | 11 | Class    | Our Class      | -2    |
      | 12 | Class    | Other Class    | -2    |
      | 13 | Class    | Special Class  | -2    |
      | 14 | Team     | Super Team     | -2    |
      | 15 | Team     | Our Team       | -1    |
      | 16 | Team     | First Team     | 0     |
      | 17 | Other    | A custom group | -2    |
      | 18 | Club     | Our Club       | -2    |
      | 20 | Friends  | My Friends     | -2    |
      | 21 | UserSelf | owner          | -2    |
      | 51 | UserSelf | johna          | -2    |
      | 53 | UserSelf | johnb          | -2    |
      | 55 | UserSelf | johnc          | -2    |
      | 57 | UserSelf | johnd          | -2    |
      | 59 | UserSelf | johne          | -2    |
      | 61 | UserSelf | janea          | -2    |
      | 63 | UserSelf | janeb          | -2    |
      | 65 | UserSelf | janec          | -2    |
      | 67 | UserSelf | janed          | -2    |
      | 69 | UserSelf | janee          | -2    |
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
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 1               | 11             |
      | 3               | 13             |
      | 11              | 14             |
      | 11              | 16             |
      | 11              | 17             |
      | 11              | 18             |
      | 11              | 59             |
      | 13              | 14             |
      | 13              | 15             |
      | 13              | 69             |
      | 14              | 51             |
      | 14              | 53             |
      | 14              | 55             |
      | 15              | 57             |
      | 15              | 59             |
      | 15              | 61             |
      | 16              | 63             |
      | 16              | 65             |
      | 16              | 67             |
      | 20              | 21             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 1                 | 1              |
      | 1                 | 11             |
      | 1                 | 12             |
      | 1                 | 14             |
      | 1                 | 16             |
      | 1                 | 17             |
      | 1                 | 18             |
      | 1                 | 51             |
      | 1                 | 53             |
      | 1                 | 55             |
      | 1                 | 59             |
      | 1                 | 63             |
      | 1                 | 65             |
      | 1                 | 67             |
      | 3                 | 3              |
      | 3                 | 13             |
      | 3                 | 15             |
      | 3                 | 51             |
      | 3                 | 53             |
      | 3                 | 55             |
      | 3                 | 61             |
      | 3                 | 63             |
      | 3                 | 65             |
      | 3                 | 69             |
      | 11                | 11             |
      | 11                | 14             |
      | 11                | 16             |
      | 11                | 17             |
      | 11                | 18             |
      | 11                | 51             |
      | 11                | 53             |
      | 11                | 55             |
      | 11                | 59             |
      | 11                | 63             |
      | 11                | 65             |
      | 11                | 67             |
      | 12                | 12             |
      | 13                | 13             |
      | 13                | 14             |
      | 13                | 15             |
      | 13                | 51             |
      | 13                | 53             |
      | 13                | 55             |
      | 13                | 61             |
      | 13                | 63             |
      | 13                | 65             |
      | 13                | 69             |
      | 14                | 14             |
      | 14                | 51             |
      | 14                | 53             |
      | 14                | 55             |
      | 15                | 15             |
      | 15                | 61             |
      | 15                | 63             |
      | 15                | 65             |
      | 16                | 16             |
      | 16                | 63             |
      | 16                | 65             |
      | 16                | 67             |
      | 20                | 20             |
      | 20                | 21             |
      | 21                | 21             |

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
            "first_name": "Jane",
            "grade": -2,
            "last_name": "Doe",
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
            "first_name": "John",
            "grade": 1,
            "last_name": "Adams",
            "login": "johna",
            "group_id": 51
          },
          {
            "first_name": "John",
            "grade": 2,
            "last_name": "Baker",
            "login": "johnb",
            "group_id": 53
          },
          {
            "first_name": "John",
            "grade": 3,
            "last_name": null,
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
            "first_name": "Jane",
            "grade": -2,
            "last_name": "Doe",
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
            "last_name": "Adams",
            "login": "johna",
            "group_id": 51
          },
          {
            "first_name": "John",
            "grade": 2,
            "last_name": "Baker",
            "login": "johnb",
            "group_id": 53
          },
          {
            "first_name": "John",
            "grade": 3,
            "last_name": null,
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
