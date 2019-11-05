Feature: Update active attempt for an item - robustness
  Background:
    Given the database has the following table 'groups':
      | id  | team_item_id | type     |
      | 101 | null         | UserSelf |
      | 102 | 10           | Team     |
      | 103 | 10           | Team     |
      | 104 | 10           | Team     |
      | 105 | 10           | Team     |
      | 108 | 10           | Team     |
      | 109 | 10           | Team     |
      | 111 | null         | UserSelf |
      | 121 | null         | UserSelf |
    And the database has the following table 'users':
      | login | group_id |
      | john  | 101      |
      | jane  | 111      |
      | guest | 121      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type              |
      | 102             | 101            | requestAccepted   |
      | 103             | 101            | invitationSent    |
      | 104             | 101            | requestSent       |
      | 105             | 101            | invitationRefused |
      | 106             | 101            | requestRefused    |
      | 107             | 101            | removed           |
      | 108             | 101            | left              |
      | 109             | 101            | direct            |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 101               | 101            | 1       |
      | 102               | 101            | 0       |
      | 102               | 102            | 1       |
      | 111               | 111            | 1       |
      | 121               | 121            | 1       |
    And the database has the following table 'items':
      | id | url                                                                     | type    | has_attempts |
      | 10 | null                                                                    | Chapter | 0            |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | 0            |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course  | 1            |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 101      | 50      | content                  |
      | 101      | 60      | content                  |
      | 111      | 50      | content_with_descendants |
      | 121      | 50      | info                     |

  Scenario: Invalid groups_attempt_id
    Given I am the user with group_id "101"
    When I send a PUT request to "/attempts/abc/active"
    Then the response code should be 400
    And the response error message should contain "Wrong value for groups_attempt_id (should be int64)"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User not found
    Given I am the user with group_id "404"
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User doesn't have access to the item
    Given I am the user with group_id "121"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 100 | 121      | 50      | 1     |
      | 101 | 121      | 50      | 2     |
    And the database has the following table 'users_items':
      | user_group_id | item_id | active_attempt_id |
      | 121           | 50      | 101               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No groups_attempts
    Given I am the user with group_id "101"
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: Wrong item in groups_attempts
    Given I am the user with group_id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 100 | 101      | 51      | 1     |
      | 101 | 101      | 50      | 2     |
    And the database has the following table 'users_items':
      | user_group_id | item_id | active_attempt_id |
      | 101           | 50      | 101               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (invitationSent)
    Given I am the user with group_id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 100 | 103      | 60      | 1     |
      | 200 | 102      | 60      | 2     |
    And the database has the following table 'users_items':
      | user_group_id | item_id | active_attempt_id |
      | 101           | 60      | 200               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (requestSent)
    Given I am the user with group_id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 100 | 104      | 60      | 1     |
      | 200 | 102      | 60      | 2     |
    And the database has the following table 'users_items':
      | user_group_id | item_id | active_attempt_id |
      | 101           | 60      | 200               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (invitationRefused)
    Given I am the user with group_id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 100 | 105      | 60      | 1     |
      | 200 | 102      | 60      | 2     |
    And the database has the following table 'users_items':
      | user_group_id | item_id | active_attempt_id |
      | 101           | 60      | 200               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (requestRefused)
    Given I am the user with group_id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 100 | 106      | 60      | 1     |
      | 200 | 102      | 60      | 2     |
    And the database has the following table 'users_items':
      | user_group_id | item_id | active_attempt_id |
      | 101           | 60      | 200               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (removed)
    Given I am the user with group_id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 100 | 107      | 60      | 1     |
      | 200 | 102      | 60      | 2     |
    And the database has the following table 'users_items':
      | user_group_id | item_id | active_attempt_id |
      | 101           | 60      | 200               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User is not a member of the team (left)
    Given I am the user with group_id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 100 | 108      | 60      | 1     |
      | 200 | 102      | 60      | 2     |
    And the database has the following table 'users_items':
      | user_group_id | item_id | active_attempt_id |
      | 101           | 60      | 200               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: groups_attempts.group_id is not user's self group
    Given I am the user with group_id "101"
    And the database has the following table 'groups_attempts':
      | id  | group_id | item_id | order |
      | 100 | 102      | 50      | 1     |
      | 200 | 102      | 50      | 1     |
    And the database has the following table 'users_items':
      | user_group_id | item_id | active_attempt_id |
      | 101           | 50      | 200               |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged
