Feature: List groups managed by the current user - robustness
  Background:
    Given the database has the following user:
      | group_id | login |
      | 21       | owner |

  Scenario: Wrong sort
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?sort=description"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "description""

  Scenario: Wrong from
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?from.type=Class"
    Then the response code should be 400
    And the response error message should contain "Unallowed paging parameters (from.type)"

  Scenario: Unknown type
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?type=User"
    Then the response code should be 400
    And the response error message should contain "Wrong value for type: "User""

  Scenario: Unparsable root_activity_id_among
    Given I am the user with id "21"
    When I send a GET request to "/current-user/managed-groups?root_activity_id_among=200,abc"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: 'abc', param: 'root_activity_id_among')"
