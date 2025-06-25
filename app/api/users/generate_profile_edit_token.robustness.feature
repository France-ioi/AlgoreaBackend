Feature: Generate Profile Edit Token - robustness
  Scenario: Should be logged in
    Given there is a user @TargetUser
    When I send a POST request to "/users/@TargetUser/generate-profile-edit-token"
    Then the response code should be 401
    And the response error message should contain "No access token provided"

  Scenario: Should return an error when target_user_id is not an int64
    Given I am @CurrentUser
    When I send a POST request to "/users/xyz/generate-profile-edit-token"
    Then the response code should be 400
    And the response error message should contain "Wrong value for target_user_id (should be int64)"

  Scenario: Should be forbidden when the current user doesn't have login_id
    Given there are the following groups:
      | group     | parent        | members        | require_personal_info_access_approval |
      | @AllUsers |               | @Manager,@User |                                       |
      | @School   | @SchoolParent |                |                                       |
      | @Class    | @School       | @User          | edit                                  |
    And there are the following users:
      | user     | login_id |
      | @Manager | null     |
      | @User    | 2        |
    And I am @Manager
    And I am a manager of the group @School
    When I send a POST request to "/users/@User/generate-profile-edit-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should be forbidden when the target user doesn't have login_id
    Given there are the following groups:
      | group     | parent        | members        | require_personal_info_access_approval |
      | @AllUsers |               | @Manager,@User |                                       |
      | @School   | @SchoolParent |                |                                       |
      | @Class    | @School       | @User          | edit                                  |
    And there are the following users:
      | user     | login_id |
      | @Manager | 1        |
      | @User    | null     |
    And I am @Manager
    And I am a manager of the group @School
    When I send a POST request to "/users/@User/generate-profile-edit-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should be forbidden when the target user doesn't exist
    Given there are the following users:
      | user     | login_id |
      | @Manager | 1        |
    Given I am @Manager
    When I send a POST request to "/users/42/generate-profile-edit-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should be forbidden when the current-user is not a manager of a group of which the target user is a member
    Given there are the following groups:
      | group     | parent | members           | require_personal_info_access_approval |
      | @AllUsers |        | @NotManager,@User |                                       |
      | @Class    |        | @User             | edit                                  |
    And there are the following users:
      | user        | login_id |
      | @NotManager | 1        |
      | @User       | 2        |
    And I am @NotManager
    When I send a POST request to "/users/@User/generate-profile-edit-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario Outline: Should be forbidden when the group managed by the current-user doesn't have `require_personal_info_access_approval` === 'edit'
    Given there are the following groups:
      | group     | parent       | members        | require_personal_info_access_approval   |
      | @AllUsers |              | @Manager,@User |                                         |
      | @Class    | @ClassParent | @User          | <require_personal_info_access_approval> |
    And there are the following users:
      | user        | login_id |
      | @Manager    | 1        |
      | @User       | 2        |
    And I am @Manager
    And I am a manager of the group @ClassParent
    When I send a POST request to "/users/@User/generate-profile-edit-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
  Examples:
    | require_personal_info_access_approval |
    | none                                  |
    | view                                  |
