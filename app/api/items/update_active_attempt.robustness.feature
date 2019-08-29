Feature: Update active attempt for an item - robustness
  Background:
    Given the database has the following table 'users':
      | ID  | sLogin | idGroupSelf |
      | 10  | john   | 101         |
      | 11  | jane   | 111         |
      | 12  | guest  | 121         |
    And the database has the following table 'groups':
      | ID  | idTeamItem | sType    |
      | 101 | null       | UserSelf |
      | 102 | 10         | Team     |
      | 103 | 10         | Team     |
      | 104 | 10         | Team     |
      | 105 | 10         | Team     |
      | 108 | 10         | Team     |
      | 109 | 10         | Team     |
      | 111 | null       | UserSelf |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType              |
      | 102           | 101          | requestAccepted    |
      | 103           | 101          | invitationSent     |
      | 104           | 101          | requestSent        |
      | 105           | 101          | invitationRefused  |
      | 106           | 101          | requestRefused     |
      | 107           | 101          | removed            |
      | 108           | 101          | left               |
      | 109           | 101          | direct            |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 101             | 101          | 1       |
      | 102             | 101          | 0       |
      | 102             | 102          | 1       |
      | 111             | 111          | 1       |
      | 121             | 121          | 1       |
    And the database has the following table 'items':
      | ID | sUrl                                                                    | sType   | bHasAttempts |
      | 10 | null                                                                    | Chapter | 0            |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | 0            |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course  | 1            |
    And the database has the following table 'items_ancestors':
      | idItemAncestor | idItemChild |
      | 10             | 60          |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate | sCachedFullAccessDate | sCachedGrayedAccessDate |
      | 101     | 50     | 2017-05-29T06:38:38Z     | null                  | null                    |
      | 101     | 60     | 2017-05-29T06:38:38Z     | null                  | null                    |
      | 111     | 50     | null                     | 2017-05-29T06:38:38Z  | null                    |
      | 121     | 50     | null                     | null                  | 2017-05-29T06:38:38Z    |

  Scenario: Invalid groups_attempt_id
    Given I am the user with ID "10"
    When I send a PUT request to "/attempts/abc/active"
    Then the response code should be 400
    And the response error message should contain "Wrong value for groups_attempt_id (should be int64)"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User not found
    Given I am the user with ID "404"
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User doesn't have access to the item
    Given I am the user with ID "12"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 12     | 50     | null            |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem |
      | 100 | 121     | 50     |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No groups_attempts
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 50     | null            |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: Wrong item in groups_attempts
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 50     | null            |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem |
      | 100 | 101     | 51     |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No users_items
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 51     | null            |
      | 11     | 50     | null            |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem |
      | 100 | 101     | 50     |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (invitationSent)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 60     | null            |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem |
      | 100 | 103     | 60     |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (requestSent)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 60     | null            |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem |
      | 100 | 104     | 60     |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (invitationRefused)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 60     | null            |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem |
      | 100 | 105     | 60     |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (requestRefused)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 60     | null            |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem |
      | 100 | 106     | 60     |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (removed)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 60     | null            |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem |
      | 100 | 107     | 60     |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (left)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 60     | null            |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem |
      | 100 | 108     | 60     |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: groups_attempts.idGroup is not user's self group
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive |
      | 10     | 50     | null            |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem |
      | 100 | 102     | 50     |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged
