Feature: Set additional time in the contest for the group (contestSetAdditionalTime) - robustness
  Background:
    Given the database has the following table 'users':
      | id | login | group_self_id | group_owned_id |
      | 1  | owner | 21            | 22             |
      | 2  | john  | 31            | 32             |
    And the database has the following table 'groups':
      | id | name        | type      |
      | 12 | Group A     | Class     |
      | 13 | Group B     | Other     |
      | 21 | owner       | UserSelf  |
      | 22 | owner-admin | UserAdmin |
      | 31 | john        | UserSelf  |
      | 32 | john-admin  | UserAdmin |
    And the database has the following table 'groups_ancestors':
      | group_ancestor_id | group_child_id | is_self |
      | 12                | 12             | 1       |
      | 13                | 13             | 1       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |
      | 22                | 31             | 0       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
    And the database has the following table 'items':
      | id | duration | has_attempts |
      | 50 | 00:00:00 | 0            |
      | 60 | null     | 0            |
      | 10 | 00:00:02 | 0            |
      | 70 | 00:00:03 | 0            |
      | 80 | 00:00:04 | 1            |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_partial_access_date | cached_grayed_access_date | cached_full_access_date | cached_access_solutions_date | additional_time | user_created_id |
      | 13       | 50      | 2017-05-29 06:38:38        | null                      | null                    | null                         | 01:00:00        | 1               |
      | 13       | 60      | null                       | 2017-05-29 06:38:38       | null                    | null                         | 01:01:00        | 1               |
      | 13       | 70      | null                       | null                      | 2017-05-29 06:38:38     | null                         | null            | 1               |
      | 21       | 50      | null                       | null                      | null                    | null                         | null            | 1               |
      | 21       | 60      | null                       | null                      | 2018-05-29 06:38:38     | null                         | null            | 1               |
      | 21       | 70      | null                       | null                      | 2018-05-29 06:38:38     | null                         | null            | 1               |
      | 21       | 80      | null                       | null                      | 2018-05-29 06:38:38     | null                         | null            | 1               |

  Scenario: Wrong item_id
    Given I am the user with id "1"
    When I send a PUT request to "/contests/abc/groups/13/additional-times?seconds=0"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Wrong group_id
    Given I am the user with id "1"
    When I send a PUT request to "/contests/50/groups/abc/additional-times?seconds=0"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Wrong 'seconds'
    Given I am the user with id "1"
    When I send a PUT request to "/contests/50/groups/13/additional-times?seconds=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for seconds (should be int64)"

  Scenario: 'seconds' is too big
    Given I am the user with id "1"
    When I send a PUT request to "/contests/50/groups/13/additional-times?seconds=3020400"
    Then the response code should be 400
    And the response error message should contain "'seconds' should be between -3020399 and 3020399"

  Scenario: 'seconds' is too small
    Given I am the user with id "1"
    When I send a PUT request to "/contests/50/groups/13/additional-times?seconds=-3020400"
    Then the response code should be 400
    And the response error message should contain "'seconds' should be between -3020399 and 3020399"

  Scenario: No such item
    Given I am the user with id "1"
    When I send a PUT request to "/contests/90/groups/13/additional-times?seconds=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the item
    Given I am the user with id "1"
    When I send a PUT request to "/contests/10/groups/13/additional-times?seconds=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item is not a timed contest
    Given I am the user with id "1"
    When I send a PUT request to "/contests/60/groups/13/additional-times?seconds=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is not a contest admin
    Given I am the user with id "1"
    When I send a PUT request to "/contests/50/groups/13/additional-times?seconds=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The group is not owned by the user
    Given I am the user with id "1"
    When I send a PUT request to "/contests/70/groups/12/additional-times?seconds=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No such group
    Given I am the user with id "1"
    When I send a PUT request to "/contests/70/groups/404/additional-times?seconds=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Team contest and the UserSelf group
    Given I am the user with id "1"
    When I send a PUT request to "/contests/80/groups/31/additional-times?seconds=0"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
