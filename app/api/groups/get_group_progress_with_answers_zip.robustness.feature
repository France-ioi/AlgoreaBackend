Feature: Export the current progress of a group with answers on a subset of items as a ZIP file (groupGroupProgressWithAnswersZIP) - robustness
  Scenario: User is not able to watch group members
    Given I am @John
    And there is a group @Classroom
    And there are the following items:
      | item  | type |
      | @Item | Task |
    And I am a manager of the group @Classroom
    When I send a GET request to "/groups/@Classroom/group-progress-with-answers-zip?parent_item_ids=@Item"
    Then the response code should be 403
    And the response error message should contain "No rights to watch group members"

  Scenario: Group id is incorrect
    Given I am @John
    And there are the following items:
      | item  | type |
      | @Item | Task |
    When I send a GET request to "/groups/abc/group-progress-with-answers-zip?parent_item_ids=@Item"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: parent_item_ids is incorrect
    Given I am @John
    And there is a group @Classroom
    And I am a manager of the group @Classroom and can watch its members
    And there are the following items:
      | item  | type |
      | @Item | Task |
    When I send a GET request to "/groups/@Classroom/group-progress-with-answers-zip?parent_item_ids=abc,@Item"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: 'abc', param: 'parent_item_ids')"

  Scenario: Not enough permissions to watch answers on parent_item_ids
    Given I am @John
    And there is a group @Classroom
    And I am a manager of the group @Classroom and can watch its members
    And there are the following items:
      | item                | type |
      | @ItemCanWatchAnswer | Task |
      | @ItemCanWatchResult | Task |
    And there are the following item permissions:
      | item                | group | can_watch |
      | @ItemCanWatchAnswer | @John | answer    |
      | @ItemCanWatchResult | @John | result    |
    When I send a GET request to "/groups/@Classroom/group-progress-with-answers-zip?parent_item_ids=@ItemCanWatchAnswer,@ItemCanWatchResult"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User not found
    Given I am the user with id "404"
    And there is a group @Classroom
    And there are the following items:
      | item  | type |
      | @Item | Task |
    When I send a GET request to "/groups/@Classroom/group-progress-with-answers-zip?parent_item_ids=@Item"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
