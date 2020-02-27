Feature: List user-batch prefixes (userBatchPrefixesView)
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
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 13                | 13             |
      | 13                | 21             |
      | 14                | 14             |
      | 21                | 21             |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 21             |
    And the database has the following table 'user_batch_prefixes':
      | group_prefix | group_id | allow_new | max_users |
      | test         | 13       | 1         | 90        |
      | test1        | 13       | 1         | null      |
      | test2        | 13       | 0         | 80        |
      | test3        | 21       | 1         | 70        |
      | test4        | 14       | 1         | 60        |
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

  Scenario: List user-batch prefixes (default sort)
    Given I am the user with id "21"
    When I send a GET request to "/groups/21/user-batch-prefixes"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"group_id": "13", "group_prefix": "test", "max_users": 90, "total_size": 2},
      {"group_id": "13", "group_prefix": "test1", "max_users": null, "total_size": 2},
      {"group_id": "21", "group_prefix": "test3", "max_users": 70, "total_size": 1}
    ]
    """

  Scenario: List user-batch prefixes (sorted by group_prefix desc, start from the second row)
    Given I am the user with id "21"
    When I send a GET request to "/groups/21/user-batch-prefixes?sort=-group_prefix&from.group_prefix=test3"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"group_id": "13", "group_prefix": "test1", "max_users": null, "total_size": 2},
      {"group_id": "13", "group_prefix": "test", "max_users": 90, "total_size": 2}
    ]
    """

  Scenario: List user-batch prefixes (default sort, another group_id)
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/user-batch-prefixes"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"group_id": "13", "group_prefix": "test", "max_users": 90, "total_size": 2},
      {"group_id": "13", "group_prefix": "test1", "max_users": null, "total_size": 2}
    ]
    """
