Feature: Display the current progress of users on a subset of items (groupUserProgress) - robustness
  Background:
    Given the database has the following users:
      | login | group_id | owned_group_id |
      | owner | 21       | 22             |
      | user  | 11       | 12             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 13                | 13             | 1       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |
    And the database has the following table 'items':
      | id  | type     |
      | 200 | Category |
      | 210 | Chapter  |
      | 211 | Task     |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_full_access_since | cached_partial_access_since | cached_grayed_access_since |
      | 21       | 211     | null                     | null                        | 2017-05-29 06:38:38        |
      | 20       | 212     | null                     | 2017-05-29 06:38:38         | null                       |
      | 21       | 213     | 2017-05-29 06:38:38      | null                        | null                       |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 200            | 210           | 0           |
      | 200            | 220           | 1           |
      | 210            | 211           | 0           |

  Scenario: User is not an admin of the group
    Given I am the user with id "11"
    When I send a GET request to "/groups/13/user-progress?parent_item_ids=210,220,310"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Group id is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/abc/user-progress?parent_item_ids=210,220,310"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: parent_item_ids is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/user-progress?parent_item_ids=abc,123"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: 'abc', param: 'parent_item_ids')"

  Scenario: User not found
    Given I am the user with id "404"
    When I send a GET request to "/groups/13/user-progress?parent_item_ids=210,220"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: sort is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/user-progress?parent_item_ids=210,220,310&sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""
