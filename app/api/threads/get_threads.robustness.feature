Feature: Get threads - robustness
  Scenario: Should be logged
    When I send a GET request to "/threads"
    Then the response code should be 401
    And the response error message should contain "No access token provided"

  Scenario: Watched group should be an integer
    Given I am John
    When I send a GET request to "/threads?watched_group_id=aaa"
    Then the response code should be 400
    And the response error message should contain "Wrong value for watched_group_id (should be int64)"

    #TODO
  Scenario: Check that watched group has the rights (must be a manager and have can_watch_members on watched_group_id)

  Scenario: The user should be a manager of watched_group_id group
    Given I am John
    And There is a group with "id=10"
    When I send a GET request to "/threads?watched_group_id=10"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"

  Scenario: The user should be able to watch the watched group id
    Given I am John
    And I am a manager of the group with id "10"
    When I send a GET request to "/threads?watched_group_id=10"
    Then the response code should be 400
    And the response error message should contain "eee"
