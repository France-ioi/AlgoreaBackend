Feature: Fetch active attempt for an item - robustness
  Background:
    Given the database has the following table 'users':
      | ID  | sLogin | idGroupSelf |
      | 10  | john   | 101         |
    And the database has the following table 'groups':
      | ID  | idTeamItem | sType    |
      | 101 | null       | UserSelf |
      | 102 | 60         | Team     |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 101             | 101          | 1       |
      | 102             | 102          | 1       |
    And the database has the following table 'items':
      | ID | sUrl                                                                    | sType    | bHasAttempts |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task     | 0            |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course   | 1            |
      | 70 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Root     | 1            |
      | 80 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Category | 1            |
      | 90 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Chapter  | 1            |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate |
      | 101     | 50     | 2017-05-29T06:38:38Z     |
      | 101     | 60     | 2017-05-29T06:38:38Z     |
      | 101     | 70     | 2017-05-29T06:38:38Z     |
      | 101     | 80     | 2017-05-29T06:38:38Z     |
      | 101     | 90     | 2017-05-29T06:38:38Z     |
    And time is frozen

  Scenario: Invalid item_id
    Given I am the user with ID "10"
    When I send a PUT request to "/items/abc/active-attempt"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User not found
    Given I am the user with ID "404"
    When I send a PUT request to "/items/50/active-attempt"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No access to the item (no item)
    Given I am the user with ID "10"
    When I send a PUT request to "/items/404/active-attempt"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No access to the item (sType='Root')
    Given I am the user with ID "10"
    When I send a PUT request to "/items/70/active-attempt"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No access to the item (sType='Category')
    Given I am the user with ID "10"
    When I send a PUT request to "/items/80/active-attempt"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No access to the item (sType='Chapter')
    Given I am the user with ID "10"
    When I send a PUT request to "/items/90/active-attempt"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a team member
    Given I am the user with ID "10"
    When I send a PUT request to "/items/60/active-attempt"
    Then the response code should be 403
    And the response error message should contain "No team found for the user"
    And the table "users_answers" should stay unchanged
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged
