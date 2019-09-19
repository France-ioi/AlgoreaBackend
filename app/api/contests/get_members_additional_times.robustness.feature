Feature: Get additional times for a group of users/teams on a contest (contestListMembersAdditionalTime) - robustness
  Background:
    Given the database has the following table 'users':
      | id | login | group_self_id | group_owned_id |
      | 1  | owner | 21            | 22             |
    And the database has the following table 'groups':
      | id | name    |
      | 12 | Group A |
      | 13 | Group B |
    And the database has the following table 'groups_ancestors':
      | group_ancestor_id | group_child_id | is_self |
      | 12                | 12             | 1       |
      | 13                | 13             | 1       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |
    And the database has the following table 'items':
      | id | duration |
      | 50 | 00:00:00 |
      | 60 | null     |
      | 10 | 00:00:02 |
      | 70 | 00:00:03 |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_partial_access_date | cached_grayed_access_date | cached_full_access_date | cached_access_solutions_date | user_created_id |
      | 13       | 50      | 2017-05-29 06:38:38        | null                      | null                    | null                         | 1               |
      | 13       | 60      | null                       | 2017-05-29 06:38:38       | null                    | null                         | 1               |
      | 13       | 70      | null                       | null                      | 2017-05-29 06:38:38     | null                         | 1               |
      | 21       | 50      | null                       | null                      | null                    | null                         | 1               |
      | 21       | 60      | null                       | null                      | 2018-05-29 06:38:38     | null                         | 1               |
      | 21       | 70      | null                       | null                      | 2018-05-29 06:38:38     | null                         | 1               |

  Scenario: Wrong item_id
    Given I am the user with id "1"
    When I send a GET request to "/contests/abc/groups/13/members/additional-times"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: No such item
    Given I am the user with id "1"
    When I send a GET request to "/contests/90/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the item
    Given I am the user with id "1"
    When I send a GET request to "/contests/10/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item is not a timed contest
    Given I am the user with id "1"
    When I send a GET request to "/contests/60/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is not a contest admin
    Given I am the user with id "1"
    When I send a GET request to "/contests/50/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Wrong group_id
    Given I am the user with id "1"
    When I send a GET request to "/contests/70/groups/abc/members/additional-times"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: The group is not owned by the user
    Given I am the user with id "1"
    When I send a GET request to "/contests/70/groups/12/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No such group
    Given I am the user with id "1"
    When I send a GET request to "/contests/70/groups/404/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Wrong sort
    Given I am the user with id "1"
    When I send a GET request to "/contests/70/groups/13/members/additional-times?sort=title"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "title""

