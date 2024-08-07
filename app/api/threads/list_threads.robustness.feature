Feature: List threads - robustness
  Scenario: Should be logged in
    Given there is a group @Classroom
    When I send a GET request to "/threads?watched_group_id=@Classroom"
    Then the response code should be 401
    And the response error message should contain "No access token provided"

  Scenario: watched_group_id should be an integer
    Given I am @John
    When I send a GET request to "/threads?watched_group_id=aaa"
    Then the response code should be 400
    And the response error message should contain "Wrong value for watched_group_id (should be int64)"

  Scenario: item_id should be an integer
    Given I am @John
    When I send a GET request to "/threads?is_mine=1&item_id=aaa"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: The user should be a manager of watched_group_id group
    Given I am @John
    And there is a group @Classroom
    When I send a GET request to "/threads?watched_group_id=@Classroom"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"

  Scenario: The user should be able to watch the watched_group_id group
    Given I am @John
    And there is a group @Classroom
    And I am a manager of the group @ClassroomParent
    And @Classroom is a child of the group @ClassroomParent
    When I send a GET request to "/threads?watched_group_id=@Classroom"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"

  Scenario: Should have one of watched_group_id and is_mine given
    Given I am @John
    When I send a GET request to "/threads"
    Then the response code should be 400
    And the response error message should contain "One of watched_group_id or is_mine must be given"

  Scenario Outline: Should not have watched_group_id and is_mine set a the same time
    Given I am @John
    And there is a group @Classroom
    And I am a manager of the group @ClassroomParent and can watch its members
    And @Classroom is a child of the group @ClassroomParent
    When I send a GET request to "/threads?watched_group_id=@Classroom&is_mine=<is_mine>"
    Then the response code should be 400
    And the response error message should contain "Must not provide watched_group_id and is_mine at the same time"
    Examples:
      | is_mine |
      | 0       |
      | 1       |

  Scenario: Should return an error if sort parameter is invalid
    Given I am @John
    When I send a GET request to "/threads?is_mine=1&sort=invalid"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters"

  Scenario: Should return an error if latest_update_gt isn't in the right format
    Given I am @John
    When I send a GET request to "/threads?is_mine=1&latest_update_gt=2023-01-01T00:00:99"
    Then the response code should be 400
    And the response error message should contain "Wrong value for latest_update_gt (should be time (rfc3339Nano))"
