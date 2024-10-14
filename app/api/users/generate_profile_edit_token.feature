Feature: Generate Profile Edit Token
  Scenario: Should return a token when the requester is a manager of the target group
    Given there are the following groups:
      | group     | parent       | members        | require_personal_info_access_approval |
      | @AllUsers |              | @Manager,@User |                                       |
      | @Class    | @ClassParent | @User          | edit                                  |
    And there are the following users:
      | user     | login_id |
      | @Manager | 1        |
      | @User    | 2        |
    And the time now is "2020-01-01T00:00:00Z"
    And I am @Manager
    And I am a manager of the group @ClassParent
    When I send a POST request to "/users/@User/generate-profile-edit-token"
    Then the response code should be 200
    And the response at $.token should be the base64 of an AES-256-GCM encrypted JSON object containing:
      """
        {
          "requester_id": "1",
          "target_id": "2",
          "exp": "1577838600"
        }
      """
    And the response at $.alg should be "AES-256-GCM"

  Scenario: Should return a token when the requester is a manager of the target group's parent group
    Given there are the following groups:
      | group     | parent        | members        | require_personal_info_access_approval |
      | @AllUsers |               | @Manager,@User |                                       |
      | @School   | @SchoolParent |                |                                       |
      | @Class    | @School       | @User          | edit                                  |
    And there are the following users:
      | user     | login_id |
      | @Manager | 1        |
      | @User    | 2        |
    And the time now is "2020-01-01T00:00:00Z"
    And I am @Manager
    And I am a manager of the group @SchoolParent
    When I send a POST request to "/users/@User/generate-profile-edit-token"
    Then the response code should be 200
    And the response at $.token should be the base64 of an AES-256-GCM encrypted JSON object containing:
      """
        {
          "requester_id": "1",
          "target_id": "2",
          "exp": "1577838600"
        }
      """
    And the response at $.alg should be "AES-256-GCM"

  Scenario: >
      Should return a token when the requester is a manager of the target group's parent group,
      and `require_personal_info_access_approval`=`edit` is set on the parent of the group the target user belongs to
    Given there are the following groups:
      | group     | parent  | members        | require_personal_info_access_approval |
      | @AllUsers |         | @Manager,@User |                                       |
      | @School   | @City   |                | edit                                  |
      | @Class    | @School | @User          |                                       |
    And there are the following users:
      | user     | login_id |
      | @Manager | 1        |
      | @User    | 2        |
    And the time now is "2020-01-01T00:00:00Z"
    And I am @Manager
    And I am a manager of the group @CityParent
    And the group @City is a child of the group @CityParent
    When I send a POST request to "/users/@User/generate-profile-edit-token"
    Then the response code should be 200
    And the response at $.token should be the base64 of an AES-256-GCM encrypted JSON object containing:
      """
        {
          "requester_id": "1",
          "target_id": "2",
          "exp": "1577838600"
        }
      """
    And the response at $.alg should be "AES-256-GCM"
