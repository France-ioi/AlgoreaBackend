Feature: Display the current progress of teams on a subset of items (groupTeamProgress) - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned |
      | 1  | owner  | 21          | 22           |
      | 2  | user   | 11          | 12           |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 12              | 12           | 1       |
      | 13              | 13           | 1       |
      | 21              | 21           | 1       |
      | 22              | 13           | 0       |
      | 22              | 22           | 1       |
    And the database has the following table 'items':
      | ID  | sType    |
      | 200 | Category |
      | 210 | Chapter  |
      | 211 | Task     |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedFullAccessDate | sCachedPartialAccessDate | sCachedGrayedAccessDate |
      | 21      | 211    | null                  | null                     | 2017-05-29T06:38:38Z    |
      | 20      | 212    | null                  | 2017-05-29T06:38:38Z     | null                    |
      | 21      | 213    | 2017-05-29T06:38:38Z  | null                     | null                    |
    And the database has the following table 'items_items':
      | idItemParent | idItemChild |
      | 200          | 210         |
      | 200          | 220         |
      | 210          | 211         |

  Scenario: User is not an admin of the group
    Given I am the user with ID "2"
    When I send a GET request to "/groups/13/team-progress?parent_item_ids=210,220,310"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Group ID is incorrect
    Given I am the user with ID "1"
    When I send a GET request to "/groups/abc/team-progress?parent_item_ids=210,220,310"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: parent_item_ids is incorrect
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/team-progress?parent_item_ids=abc,123"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: 'abc', param: 'parent_item_ids')"

  Scenario: User not found
    Given I am the user with ID "404"
    When I send a GET request to "/groups/13/team-progress?parent_item_ids=210,220"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: sort is incorrect
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/team-progress?parent_item_ids=210,220,310&sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""

