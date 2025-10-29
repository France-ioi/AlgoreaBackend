Feature: List user descendants of the group (groupUserDescendantView)
  Background:
    Given the database has the following table "groups":
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
      | 22 | Club    | Club           | -2    |
      | 23 | Club    | School         | -2    |
    And the database has the following user:
      | group_id | login | grade | profile                                                |
      | 21       | owner | 10    | {"first_name": "Jean-Michel", "last_name": "Blanquer"} |
    And the database has the following table "group_managers":
      | group_id | manager_id |
      | 1        | 21         |
      | 23       | 22         |
    And the database has the following table "groups_groups":
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
      | 22              | 21             |
    And the groups ancestors are computed

  Scenario: One group with 5 grand children (different parents)
    Given the database table "groups" also has the following rows:
      | id | type | name  | grade |
      | 51 | User | johna | -2    |
      | 53 | User | johnb | -2    |
      | 55 | User | johnc | -2    |
      | 57 | User | jackd | -2    |
    And the database table "users" also has the following rows:
      | group_id | login | grade | profile                                      |
      | 51       | johna | 1     | {"last_name": "Adams"}                       |
      | 53       | johnb | null  | {"first_name": "John", "last_name": "Baker"} |
      | 55       | johnc | 3     | {"first_name": "John", "last_name": null}    |
      | 57       | jackd | 3     | {"first_name": "Jack", "last_name": "Doe"}   |
    And the database table "groups_groups" also has the following rows:
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 11              | 51             | null                           |
      | 11              | 21             | null                           |
      | 17              | 53             | 2019-05-30 11:00:00            |
      | 16              | 55             | null                           |
      | 18              | 57             | null                           |
      | 23              | 51             | 2019-05-30 11:00:00            |
    And the groups ancestors are computed
    And I am the user with id "21"
    When I send a GET request to "/groups/1/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "57",
        "name": "jackd",
        "parents": [{"id": "18", "name": "Our Club"}],
        "user": {"grade": 3, "login": "jackd"}
      },
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
        "user": {"grade": 3, "login": "johnc"}
      },
      {
        "id": "21",
        "name": "owner",
        "parents": [{"id": "11", "name": "Our Class"}],
        "user": {"first_name": "Jean-Michel", "grade": 10, "last_name": "Blanquer", "login": "owner"}
      }
    ]
    """
    When I send a GET request to "/groups/1/user-descendants?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "57",
        "name": "jackd",
        "parents": [{"id": "18", "name": "Our Club"}],
        "user": {"grade": 3, "login": "jackd"}
      }
    ]
    """
    When I send a GET request to "/groups/1/user-descendants?from.id=51"
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
        "user": {"grade": 3, "login": "johnc"}
      },
      {
        "id": "21",
        "name": "owner",
        "parents": [{"id": "11", "name": "Our Class"}],
        "user": {"first_name": "Jean-Michel", "grade": 10, "last_name": "Blanquer", "login": "owner"}
      }
    ]
    """

  Scenario: Non-descendant parents should not appear (one group with 1 grand child, having also a parent which is not descendant)
    Given the database also has the following users:
      | group_id | login | grade | profile                |
      | 51       | johna | 1     | {"last_name": "Adams"} |
    And the database table "groups_groups" also has the following rows:
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 11              | 51             | 2019-05-30 11:00:00            |
      | 13              | 51             | null                           |
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
    Given the database table "groups" also has the following rows:
      | id | type | name  | grade |
      | 51 | User | johna | -2    |
    And the database table "users" also has the following rows:
      | group_id | login | grade | profile                                      |
      | 51       | johna | 1     | {"first_name": "John", "last_name": "Adams"} |
    And the database table "groups_groups" also has the following rows:
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
    Given the database table "groups" also has the following rows:
      | id | type | name  | grade |
      | 51 | User | johna | -2    |
    And the database table "users" also has the following rows:
      | group_id | login | grade | profile                |
      | 51       | johna | 1     | {"last_name": "Adams"} |
    And the database table "groups_groups" also has the following rows:
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
        "user": {"grade": 1, "login": "johna"}
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
