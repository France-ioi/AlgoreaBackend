Feature: Update the current user's profile
  Background:
    Given the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
        profileUpdateEnabled: true
      """
    And the server time now is "2019-07-16T22:02:29Z"
    And the DB time now is "2019-07-16 22:02:28"
    And the database has the following users:
      | group_id | latest_login_at     | latest_activity_at  | registered_at       | latest_profile_sync_at | login_id  | login    | profile                    | default_language | last_ip     |
      | 11       | 2019-06-16 21:01:25 | 2019-06-16 22:05:44 | 2019-05-10 10:42:11 | 2019-06-15 22:05:44    | 100000001 | mohammed | {"first_name": "Mohammed"} | en               | 192.168.0.1 |
      | 13       | 2018-06-16 21:01:25 | 2018-06-16 22:05:44 | 2018-05-10 10:42:11 | null                   | 100000002 | john     | {"first_name": "John"}     | en               | 110.55.10.2 |
    And the database has the following table "sessions":
      | session_id | user_id |
      | 1          | 11      |
    And the database has the following table "access_tokens":
      | session_id | token       | expires_at          | issued_at           |
      | 1          | accesstoken | 2020-06-16 22:02:49 | 2019-06-16 22:02:28 |
    And the "Authorization" request header is "Bearer accesstoken"

  Scenario: Forward the profile with the caller's token and store the returned profile
    Given the login module "account" PUT endpoint for token "accesstoken" and body '{"profile":{"birth_date":"2010-03-14","first_name":"Jane 🐱","gender":"female","nickname":""}}' returns 200 with body:
      """
      {
        "id":100000001, "login":"jane", "language":"fr", "badges":[],
        "profile": {"first_name":"Jane 🐱","last_name":"Doe","birth_date":"2010-03-14","gender":"female"}
      }
      """
    When I send a PUT request to "/current-user/profile" with the following body:
      """
      {"profile": {"gender": "female", "first_name": "Jane 🐱", "birth_date": "2010-03-14", "nickname": ""}}
      """
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "updated"
      }
      """
    And the table "users" should remain unchanged, regardless of the row with group_id "11"
    And the table "users" at group_id "11" should be:
      | group_id | latest_login_at     | latest_activity_at  | latest_profile_sync_at | temp_user | registered_at       | login_id  | login | default_language | last_ip     |
      | 11       | 2019-06-16 21:01:25 | 2019-07-16 22:02:28 | 2019-07-16 22:02:28    | 0         | 2019-05-10 10:42:11 | 100000001 | jane  | en               | 192.168.0.1 |
    And the column "users.profile" at group_id "11" should be, in JSON:
      """
      {"first_name":"Jane 🐱","last_name":"Doe","birth_date":"2010-03-14","gender":"female"}
      """

  Scenario: A refusal from the login module becomes a 400 carrying its message
    Given the login module "account" PUT endpoint for token "accesstoken" and body '{"profile":{"gender":"unknown"}}' returns 400 with body:
      """
      invalid value for gender
      """
    When I send a PUT request to "/current-user/profile" with the following body:
      """
      {"profile": {"gender": "unknown"}}
      """
    Then the response code should be 400
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Request",
        "error_text": "Invalid value for gender"
      }
      """
    And the table "users" should remain unchanged

  Scenario: A failure of the login module becomes a 502
    Given the login module "account" PUT endpoint for token "accesstoken" and body '{"profile":{"gender":"female"}}' returns 500 with body:
      """
      Internal error
      """
    When I send a PUT request to "/current-user/profile" with the following body:
      """
      {"profile": {"gender": "female"}}
      """
    Then the response code should be 502
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "Bad Gateway",
        "error_text": "The login module failed to update the profile"
      }
      """
    And the table "users" should remain unchanged
