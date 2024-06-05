Feature: Generate Profile Edit Token
  Scenario: Should return all fields when the thread exists
    Given there are the following groups:
      | group     | parent | members        | require_personal_info_access_approval |
      | @AllUsers |        | @Manager,@User |                                       |
      | @Class    |        | @User          | edit                                  |
    And the time now is "2020-01-01T00:00:00Z"
    And I am @Manager
    And I am a manager of the group @Class
    When I send a POST request to "/users/@User/generate-profile-edit-token"
    Then the response code should be 200
    And the response at $.token should be the base64 of an AES-256-GCM encrypted JSON object containing:
      """
        {
          "requester_id": "@Manager",
          "target_id": "@User",
          "exp": "1577838600"
        }
      """
