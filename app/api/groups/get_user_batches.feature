Feature: List user batches (userBatchesView)
  Background:
    Given the database has the following table 'groups':
      | id | name   | grade | type  |
      | 13 | class  | -2    | Class |
      | 14 | class2 | -2    | Class |
      | 21 | user   | -2    | User  |
    And the database has the following table 'users':
      | login | group_id |
      | owner | 21       |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 13       | 21         | memberships |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 21             |
    And the groups ancestors are computed
    And the database has the following table 'user_batch_prefixes':
      | group_prefix | group_id | allow_new |
      | test         | 13       | 1         |
      | test1        | 13       | 1         |
      | test2        | 13       | 0         |
      | test3        | 21       | 1         |
      | test4        | 14       | 1         |
    And the database has the following table 'user_batches':
      | group_prefix | custom_prefix | size | creator_id |
      | test         | custom        | 100  | null       |
      | test         | custom1       | 200  | 13         |
      | test1        | cust          | 300  | 21         |
      | test1        | cust1         | 400  | null       |
      | test2        | cus           | 300  | 21         |
      | test2        | cus1          | 400  | null       |
      | test3        | pref          | 500  | null       |
      | test4        | prf           | 600  | null       |

  Scenario: List user batches (default sort)
    Given I am the user with id "21"
    When I send a GET request to "/user-batches/by-group/21"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"creator_id": null, "custom_prefix": "custom", "group_prefix": "test", "size": 100},
      {"creator_id": "13", "custom_prefix": "custom1", "group_prefix": "test", "size": 200},
      {"creator_id": "21", "custom_prefix": "cust", "group_prefix": "test1", "size": 300},
      {"creator_id": null, "custom_prefix": "cust1", "group_prefix": "test1", "size": 400},
      {"creator_id": "21", "custom_prefix": "cus", "group_prefix": "test2", "size": 300},
      {"creator_id": null, "custom_prefix": "cus1", "group_prefix": "test2", "size": 400},
      {"creator_id": null, "custom_prefix": "pref", "group_prefix": "test3", "size": 500}
    ]
    """

  Scenario: List user batches (sorted by group_prefix desc, size desc)
    Given I am the user with id "21"
    When I send a GET request to "/user-batches/by-group/21?sort=-group_prefix,-size"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"creator_id": null, "custom_prefix": "pref", "group_prefix": "test3", "size": 500},
      {"creator_id": null, "custom_prefix": "cus1", "group_prefix": "test2", "size": 400},
      {"creator_id": "21", "custom_prefix": "cus", "group_prefix": "test2", "size": 300},
      {"creator_id": null, "custom_prefix": "cust1", "group_prefix": "test1", "size": 400},
      {"creator_id": "21", "custom_prefix": "cust", "group_prefix": "test1", "size": 300},
      {"creator_id": "13", "custom_prefix": "custom1", "group_prefix": "test", "size": 200},
      {"creator_id": null, "custom_prefix": "custom", "group_prefix": "test", "size": 100}
    ]
    """

  Scenario: List user batches (sorted by custom_prefix desc)
    Given I am the user with id "21"
    When I send a GET request to "/user-batches/by-group/21?sort=-custom_prefix"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"creator_id": null, "custom_prefix": "pref", "group_prefix": "test3", "size": 500},
      {"creator_id": "13", "custom_prefix": "custom1", "group_prefix": "test", "size": 200},
      {"creator_id": null, "custom_prefix": "custom", "group_prefix": "test", "size": 100},
      {"creator_id": null, "custom_prefix": "cust1", "group_prefix": "test1", "size": 400},
      {"creator_id": "21", "custom_prefix": "cust", "group_prefix": "test1", "size": 300},
      {"creator_id": null, "custom_prefix": "cus1", "group_prefix": "test2", "size": 400},
      {"creator_id": "21", "custom_prefix": "cus", "group_prefix": "test2", "size": 300}
    ]
    """

  Scenario: List user batches (sorted by size)
    Given I am the user with id "21"
    When I send a GET request to "/user-batches/by-group/21?sort=size"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"creator_id": null, "custom_prefix": "custom", "group_prefix": "test", "size": 100},
      {"creator_id": "13", "custom_prefix": "custom1", "group_prefix": "test", "size": 200},
      {"creator_id": "21", "custom_prefix": "cust", "group_prefix": "test1", "size": 300},
      {"creator_id": "21", "custom_prefix": "cus", "group_prefix": "test2", "size": 300},
      {"creator_id": null, "custom_prefix": "cust1", "group_prefix": "test1", "size": 400},
      {"creator_id": null, "custom_prefix": "cus1", "group_prefix": "test2", "size": 400},
      {"creator_id": null, "custom_prefix": "pref", "group_prefix": "test3", "size": 500}
    ]
    """

  Scenario: List user batches (sorted by size, start from the second row)
    Given I am the user with id "21"
    When I send a GET request to "/user-batches/by-group/21?sort=size&from.size=200&from.group_prefix=test&from.custom_prefix=custom"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"creator_id": "13", "custom_prefix": "custom1", "group_prefix": "test", "size": 200},
      {"creator_id": "21", "custom_prefix": "cust", "group_prefix": "test1", "size": 300},
      {"creator_id": "21", "custom_prefix": "cus", "group_prefix": "test2", "size": 300},
      {"creator_id": null, "custom_prefix": "cust1", "group_prefix": "test1", "size": 400},
      {"creator_id": null, "custom_prefix": "cus1", "group_prefix": "test2", "size": 400},
      {"creator_id": null, "custom_prefix": "pref", "group_prefix": "test3", "size": 500}
    ]
    """

  Scenario: List user batches (default sort, another group_id)
    Given I am the user with id "21"
    When I send a GET request to "/user-batches/by-group/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"creator_id": null, "custom_prefix": "custom", "group_prefix": "test", "size": 100},
      {"creator_id": "13", "custom_prefix": "custom1", "group_prefix": "test", "size": 200},
      {"creator_id": "21", "custom_prefix": "cust", "group_prefix": "test1", "size": 300},
      {"creator_id": null, "custom_prefix": "cust1", "group_prefix": "test1", "size": 400},
      {"creator_id": "21", "custom_prefix": "cus", "group_prefix": "test2", "size": 300},
      {"creator_id": null, "custom_prefix": "cus1", "group_prefix": "test2", "size": 400}
    ]
    """

  Scenario: User is not a manager of the group_id
    Given I am the user with id "21"
    When I send a GET request to "/user-batches/by-group/14"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """
