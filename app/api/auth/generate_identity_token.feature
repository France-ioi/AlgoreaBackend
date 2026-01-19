Feature: Generate identity token
  Background:
    Given the database has the following table "users":
      | group_id | login |
      | 21       | john  |
    And the DB time now is "2019-07-16 22:02:28"
    And the database has the following table "sessions":
      | session_id | user_id |
      | 1          | 21      |
    And the database has the following table "access_tokens":
      | session_id | token           | expires_at          |
      | 1          | someaccesstoken | 2019-07-16 23:02:28 |

  Scenario: Successfully generate an identity token
    Given the time now is "2019-07-16T22:02:28Z"
    And the "Authorization" request header is "Bearer someaccesstoken"
    And "expectedIdentityToken" is a token signed by the app with the following payload:
      """
      {
        "user_id": "21",
        "is_temp_user": false,
        "exp": 1563321748
      }
      """
    When I send a POST request to "/auth/identity-token"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "identity_token": "{{expectedIdentityToken}}",
          "expires_in": 7200
        }
      }
      """

  Scenario: Successfully generate an identity token for a temp user
    Given the database has the following table "users":
      | group_id | login    | temp_user |
      | 22       | tmp-user | true      |
    And the database has the following table "sessions":
      | session_id | user_id |
      | 2          | 22      |
    And the database has the following table "access_tokens":
      | session_id | token          | expires_at          |
      | 2          | tempacesstoken | 2019-07-17 00:02:28 |
    And the time now is "2019-07-16T22:02:28Z"
    And the "Authorization" request header is "Bearer tempacesstoken"
    And "expectedIdentityToken" is a token signed by the app with the following payload:
      """
      {
        "user_id": "22",
        "is_temp_user": true,
        "exp": 1563321748
      }
      """
    When I send a POST request to "/auth/identity-token"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "identity_token": "{{expectedIdentityToken}}",
          "expires_in": 7200
        }
      }
      """
