Feature: List user descendants of the group (groupUserDescendantView)
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
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name | grade |
      | owner | 21       | Jean-Michel | Blanquer  | 10    |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 1        | 21         |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 1               | 11             |
      | 3               | 13             |
      | 3               | 15             |
      | 11              | 14             |
      | 11              | 16             |
      | 11              | 17             |
      | 11              | 18             |
      | 13              | 14             |
      | 13              | 15             |
    And the groups ancestors are computed

  Scenario: One group with 4 grand children (different parents)
    Given the database table 'groups' has also the following rows:
      | id | type | name  | grade |
      | 51 | User | johna | -2    |
      | 53 | User | johnb | -2    |
      | 55 | User | johnc | -2    |
      | 57 | User | johnd | -2    |
    And the database table 'users' has also the following rows:
      | login | group_id | first_name | last_name | grade |
      | johna | 51       | null       | Adams     | 1     |
      | johnb | 53       | John       | Baker     | null  |
      | johnc | 55       | John       | null      | 3     |
      | johnd | 57       | John       | Doe       | 3     |
    And the database table 'groups_groups' has also the following rows:
      | parent_group_id | child_group_id |
      | 11              | 51             |
      | 17              | 53             |
      | 16              | 55             |
      | 18              | 57             |
    And the groups ancestors are computed
    And I am the user with id "21"
    When I send a GET request to "/groups/1/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "name": "johna",
        "parents": [{"id": "11", "name": "Our Class"}],
        "user": {"first_name": null, "grade": 1, "last_name": "Adams", "login": "johna"}
      },
      {
        "id": "53",
        "name": "johnb",
        "parents": [{"id": "17", "name": "A custom group"}],
        "user": {"first_name": "John", "grade": null, "last_name": "Baker", "login": "johnb"}
      },
      {
        "id": "55",
        "name": "johnc",
        "parents": [{"id": "16", "name": "First Team"}],
        "user": {"first_name": "John", "grade": 3, "last_name": null, "login": "johnc"}
      },
      {
        "id": "57",
        "name": "johnd",
        "parents": [{"id": "18", "name": "Our Club"}],
        "user": {"first_name": "John", "grade": 3, "last_name": "Doe", "login": "johnd"}
      }
    ]
    """
    When I send a GET request to "/groups/1/user-descendants?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "name": "johna",
        "parents": [{"id": "11", "name": "Our Class"}],
        "user": {"first_name": null, "grade": 1, "last_name": "Adams", "login": "johna"}
      }
    ]
    """
    When I send a GET request to "/groups/1/user-descendants?from.name=johna&from.id=51"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "53",
        "name": "johnb",
        "parents": [{"id": "17", "name": "A custom group"}],
        "user": {"first_name": "John", "grade": null, "last_name": "Baker", "login": "johnb"}
      },
      {
        "id": "55",
        "name": "johnc",
        "parents": [{"id": "16", "name": "First Team"}],
        "user": {"first_name": "John", "grade": 3, "last_name": null, "login": "johnc"}
      },
      {
        "id": "57",
        "name": "johnd",
        "parents": [{"id": "18", "name": "Our Club"}],
        "user": {"first_name": "John", "grade": 3, "last_name": "Doe", "login": "johnd"}
      }
    ]
    """

  Scenario: Non-descendant parents should not appear (one group with 1 grand child, having also a parent which is not descendant)
    Given the database table 'groups' has also the following rows:
      | id | type | name  | grade |
      | 51 | User | johna | -2    |
    And the database table 'users' has also the following rows:
      | login | group_id | first_name | last_name | grade |
      | johna | 51       | null       | Adams     | 1     |
    And the database table 'groups_groups' has also the following rows:
      | parent_group_id | child_group_id |
      | 11              | 51             |
      | 13              | 51             |
    And the groups ancestors are computed
    And I am the user with id "21"
    When I send a GET request to "/groups/1/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "name": "johna",
        "parents": [{"id": "11", "name": "Our Class"}],
        "user": {"first_name": null, "grade": 1, "last_name": "Adams", "login": "johna"}
      }
    ]
    """

  Scenario: Only actual memberships count
    Given the database table 'groups' has also the following rows:
      | id | type | name  | grade |
      | 51 | User | johna | -2    |
    And the database table 'users' has also the following rows:
      | login | group_id | first_name | last_name | grade |
      | johna | 51       | John       | Adams     | 1     |
    And the database table 'groups_groups' has also the following rows:
      | parent_group_id | child_group_id | expires_at          |
      | 11              | 51             | 2019-05-30 11:00:00 |
    And the groups ancestors are computed
    And I am the user with id "21"
    When I send a GET request to "/groups/1/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: No duplication (one group with 1 grand children connected through 2 different parents)
    Given the database table 'groups' has also the following rows:
      | id | type | name  | grade |
      | 51 | User | johna | -2    |
    And the database table 'users' has also the following rows:
      | login | group_id | first_name | last_name | grade |
      | johna | 51       | null       | Adams     | 1     |
    And the database table 'groups_groups' has also the following rows:
      | parent_group_id | child_group_id |
      | 11              | 51             |
      | 14              | 51             |
    And the groups ancestors are computed
    And I am the user with id "21"
    When I send a GET request to "/groups/1/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "name": "johna",
        "parents": [{"id": "11", "name": "Our Class"}, {"id": "14", "name": "Super Team"}],
        "user": {"first_name": null, "grade": 1, "last_name": "Adams", "login": "johna"}
      }
    ]
    """

  Scenario: No users
    Given I am the user with id "21"
    When I send a GET request to "/groups/18/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
