Feature: Get threads - robustness
  Scenario: Should be logged
    Given There is a group Classroom
    When I send a GET request to "/threads?watched_group_id=@Classroom"
    Then the response code should be 401
    And the response error message should contain "No access token provided"

  Scenario: watched_group_id should be an integer
    Given I am John
    When I send a GET request to "/threads?watched_group_id=aaa"
    Then the response code should be 400
    And the response error message should contain "Wrong value for watched_group_id (should be int64)"

  Scenario: watched_group_id should be given because the alternative is not implemented yet
    Given I am John
    When I send a GET request to "/threads"
    Then the response code should be 400
    And the response error message should contain "Not implemented yet: watchedGroupID must be given"

  Scenario: The user should be a manager of watched_group_id group
    Given I am John
    And There is a group Classroom
    When I send a GET request to "/threads?watched_group_id=@Classroom"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"

  Scenario: The user should be able to watch the watched_group_id group
    Given I am John
    And I am a manager of the group Classroom
    When I send a GET request to "/threads?watched_group_id=@Classroom"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"
