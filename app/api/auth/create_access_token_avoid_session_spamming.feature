Feature: To avoid session creation spamming, we allow a maximum of 10 sessions per user
  Background:
    Given the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
      domains:
        -
          domains: [127.0.0.1]
          allUsersGroup: @AllUsers
          TempUsersGroup: @TempUsers
      """
    And the time now is "2020-01-01T01:00:00Z"
    And the DB time now is "2020-01-01 01:00:00"
    # login_id is used to match with the "id" returned by the login module
    And there are the following users:
      | user                | login_id |
      | @UserWith11Sessions | 11       |
      | @UserWith10Sessions | 10       |
      | @UserWith9Sessions  | 9        |
    And there are the following groups:
      | group      | members                                                    |
      | @AllUsers  | @UserWith9Sessions,@UserWith10Sessions,@UserWith11Sessions |
      | @TempUsers |                                                            |
    And there are the following sessions:
      | session                        | user                | refresh_token         |
      | @Session_UserWith11Sessions_1  | @UserWith11Sessions | rt_user_11_session_1  | # shouldn't be deleted because it's the newest one
      | @Session_UserWith11Sessions_2  | @UserWith11Sessions | rt_user_11_session_2  | # should be deleted
      | @Session_UserWith11Sessions_3  | @UserWith11Sessions | rt_user_11_session_3  | # should be deleted
      | @Session_UserWith11Sessions_4  | @UserWith11Sessions | rt_user_11_session_4  |
      | @Session_UserWith11Sessions_5  | @UserWith11Sessions | rt_user_11_session_5  |
      | @Session_UserWith11Sessions_6  | @UserWith11Sessions | rt_user_11_session_6  |
      | @Session_UserWith11Sessions_7  | @UserWith11Sessions | rt_user_11_session_7  |
      | @Session_UserWith11Sessions_8  | @UserWith11Sessions | rt_user_11_session_8  |
      | @Session_UserWith11Sessions_9  | @UserWith11Sessions | rt_user_11_session_9  |
      | @Session_UserWith11Sessions_10 | @UserWith11Sessions | rt_user_11_session_10 |
      | @Session_UserWith11Sessions_11 | @UserWith11Sessions | rt_user_11_session_11 |
      | @Session_UserWith10Sessions_1  | @UserWith10Sessions | rt_user_10_session_1  | # should be deleted
      | @Session_UserWith10Sessions_2  | @UserWith10Sessions | rt_user_10_session_2  |
      | @Session_UserWith10Sessions_3  | @UserWith10Sessions | rt_user_10_session_3  |
      | @Session_UserWith10Sessions_4  | @UserWith10Sessions | rt_user_10_session_4  |
      | @Session_UserWith10Sessions_5  | @UserWith10Sessions | rt_user_10_session_5  |
      | @Session_UserWith10Sessions_6  | @UserWith10Sessions | rt_user_10_session_6  |
      | @Session_UserWith10Sessions_7  | @UserWith10Sessions | rt_user_10_session_7  |
      | @Session_UserWith10Sessions_8  | @UserWith10Sessions | rt_user_10_session_8  |
      | @Session_UserWith10Sessions_9  | @UserWith10Sessions | rt_user_10_session_9  |
      | @Session_UserWith10Sessions_10 | @UserWith10Sessions | rt_user_10_session_10 |
      | @Session_UserWith9Sessions_1   | @UserWith9Sessions  | rt_user_9_session_1   |
      | @Session_UserWith9Sessions_2   | @UserWith9Sessions  | rt_user_9_session_2   |
      | @Session_UserWith9Sessions_3   | @UserWith9Sessions  | rt_user_9_session_3   |
      | @Session_UserWith9Sessions_4   | @UserWith9Sessions  | rt_user_9_session_4   |
      | @Session_UserWith9Sessions_5   | @UserWith9Sessions  | rt_user_9_session_5   |
      | @Session_UserWith9Sessions_6   | @UserWith9Sessions  | rt_user_9_session_6   |
      | @Session_UserWith9Sessions_7   | @UserWith9Sessions  | rt_user_9_session_7   |
      | @Session_UserWith9Sessions_8   | @UserWith9Sessions  | rt_user_9_session_8   |
      | @Session_UserWith9Sessions_9   | @UserWith9Sessions  | rt_user_9_session_9   |
    And there are the following access tokens:
      | session                        | token                  | issued_at           | expires_at          |
      | @Session_UserWith11Sessions_1  | rt_user_11_session_1a  | 2020-01-01 00:00:01 | 2020-01-01 02:00:01 | # has a more recent issued_at, shouldn't be deleted
      | @Session_UserWith11Sessions_1  | rt_user_11_session_1b  | 2020-01-01 00:01:01 | 2020-01-01 02:01:01 | # newest one
      | @Session_UserWith11Sessions_2  | rt_user_11_session_2   | 2020-01-01 00:00:02 | 2020-01-01 02:00:02 | # oldest
      | @Session_UserWith11Sessions_3  | rt_user_11_session_3   | 2020-01-01 00:00:03 | 2020-01-01 02:00:03 | #second oldest, should be deleted
      | @Session_UserWith11Sessions_4  | rt_user_11_session_4   | 2020-01-01 00:00:04 | 2020-01-01 02:00:04 |
      | @Session_UserWith11Sessions_5  | rt_user_11_session_5   | 2020-01-01 00:00:05 | 2020-01-01 02:00:05 |
      | @Session_UserWith11Sessions_6  | rt_user_11_session_6   | 2020-01-01 00:00:06 | 2020-01-01 02:00:06 |
      | @Session_UserWith11Sessions_7  | rt_user_11_session_7   | 2020-01-01 00:00:07 | 2020-01-01 02:00:07 |
      | @Session_UserWith11Sessions_8  | rt_user_11_session_8   | 2020-01-01 00:00:08 | 2020-01-01 02:00:08 |
      | @Session_UserWith11Sessions_9  | rt_user_11_session_9   | 2020-01-01 00:00:09 | 2020-01-01 02:00:09 |
      | @Session_UserWith11Sessions_10 | rt_user_11_session_10  | 2020-01-01 00:00:10 | 2020-01-01 02:00:10 |
      | @Session_UserWith11Sessions_11 | rt_user_11_session_11  | 2020-01-01 00:00:11 | 2020-01-01 02:00:11 |
      | @Session_UserWith10Sessions_1  | rt_user_10_session_1   | 2020-01-01 00:00:01 | 2020-01-01 02:00:01 | # oldest one
      | @Session_UserWith10Sessions_2  | rt_user_10_session_2   | 2020-01-01 00:00:02 | 2020-01-01 02:00:02 |
      | @Session_UserWith10Sessions_3  | rt_user_10_session_3   | 2020-01-01 00:00:03 | 2020-01-01 02:00:03 |
      | @Session_UserWith10Sessions_4  | rt_user_10_session_4   | 2020-01-01 00:00:04 | 2020-01-01 02:00:04 |
      | @Session_UserWith10Sessions_5  | rt_user_10_session_5   | 2020-01-01 00:00:05 | 2020-01-01 02:00:05 |
      | @Session_UserWith10Sessions_6  | rt_user_10_session_6   | 2020-01-01 00:00:06 | 2020-01-01 02:00:06 |
      | @Session_UserWith10Sessions_7  | rt_user_10_session_7   | 2020-01-01 00:00:07 | 2020-01-01 02:00:07 |
      | @Session_UserWith10Sessions_8  | rt_user_10_session_8   | 2020-01-01 00:00:08 | 2020-01-01 02:00:08 |
      | @Session_UserWith10Sessions_9  | rt_user_10_session_9   | 2020-01-01 00:00:09 | 2020-01-01 02:00:09 |
      | @Session_UserWith10Sessions_10 | rt_user_10_session_10a | 2020-01-01 00:00:10 | 2020-01-01 02:00:10 |
      | @Session_UserWith10Sessions_10 | rt_user_10_session_10b | 2020-01-01 00:01:10 | 2020-01-01 02:01:10 |
      | @Session_UserWith9Sessions_1   | rt_user_9_session_1a   | 2020-01-01 00:00:01 | 2020-01-01 02:00:01 |
      | @Session_UserWith9Sessions_3   | rt_user_9_session_3    | 2020-01-01 00:00:03 | 2020-01-01 02:00:03 |
      | @Session_UserWith9Sessions_4   | rt_user_9_session_4    | 2020-01-01 00:00:04 | 2020-01-01 02:00:04 |
      | @Session_UserWith9Sessions_5   | rt_user_9_session_5    | 2020-01-01 00:00:05 | 2020-01-01 02:00:05 |
      | @Session_UserWith9Sessions_6   | rt_user_9_session_6    | 2020-01-01 00:00:06 | 2020-01-01 02:00:06 |
      | @Session_UserWith9Sessions_7   | rt_user_9_session_7    | 2020-01-01 00:00:07 | 2020-01-01 02:00:07 |
      | @Session_UserWith9Sessions_8   | rt_user_9_session_8    | 2020-01-01 00:00:08 | 2020-01-01 02:00:08 |
      | @Session_UserWith9Sessions_9   | rt_user_9_session_9a   | 2020-01-01 00:00:09 | 2020-01-01 02:00:09 |
      | @Session_UserWith9Sessions_9   | rt_user_9_session_9b   | 2020-01-01 00:01:09 | 2020-01-01 02:01:09 | # verify we count the number of sessions and not access tokens.

  Scenario: Should just add the new session when the user have 9 sessions
    Given  the login module "token" endpoint for code "codefromauth" returns 200 with body:
      """
      {
        "token_type": "Bearer",
        "expires_in": 31622420,
        "access_token": "access_token_from_oauth",
        "refresh_token": "refresh_token_from_oauth"
      }
      """
    And the login module "account" endpoint for token "access_token_from_oauth" returns 200 with body:
      """
      {
        "id":9, "login":"login","login_updated_at":null,"login_fixed":0,
        "login_revalidate_required":0,"login_change_required":0,"language":null,"first_name":null,
        "last_name":null,"real_name_visible":false,"timezone":null,"country_code":null,
        "address":null,"city":null,"zipcode":null,"primary_phone":null,"secondary_phone":null,
        "role":null,"school_grade":null,"student_id":null,"ministry_of_education":null,
        "ministry_of_education_fr":false,"birthday":null,"presentation":null,
        "website":null,"ip":null,"picture":null,
        "gender":null,"graduation_year":null,"graduation_grade_expire_at":null,
        "graduation_grade":null,"created_at":null,"last_login":null,
        "logout_config":null,"last_password_recovery_at":null,"merge_group_id":null,
        "origin_instance_id":null,"creator_client_id":null,"nationality":null,
        "primary_email":null,"secondary_email":null,
        "primary_email_verified":null,"secondary_email_verified":null,"has_picture":false,
        "badges":null,"client_id":null,"verification":null,"subscription_news":false
      }
      """
    When I send a POST request to "/auth/token?code=codefromauth"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "access_token": "access_token_from_oauth",
          "expires_in": 31622420
        }
      }
      """
    And there are 10 sessions for user @UserWith9Sessions

  Scenario: Should add the new session and delete the oldest one when the user have 10 sessions
    Given  the login module "token" endpoint for code "codefromauth" returns 200 with body:
      """
      {
        "token_type": "Bearer",
        "expires_in": 31622420,
        "access_token": "access_token_from_oauth",
        "refresh_token": "refresh_token_from_oauth"
      }
      """
    And the login module "account" endpoint for token "access_token_from_oauth" returns 200 with body:
      """
      {
        "id":10, "login":"login","login_updated_at":null,"login_fixed":0,
        "login_revalidate_required":0,"login_change_required":0,"language":null,"first_name":null,
        "last_name":null,"real_name_visible":false,"timezone":null,"country_code":null,
        "address":null,"city":null,"zipcode":null,"primary_phone":null,"secondary_phone":null,
        "role":null,"school_grade":null,"student_id":null,"ministry_of_education":null,
        "ministry_of_education_fr":false,"birthday":null,"presentation":null,
        "website":null,"ip":null,"picture":null,
        "gender":null,"graduation_year":null,"graduation_grade_expire_at":null,
        "graduation_grade":null,"created_at":null,"last_login":null,
        "logout_config":null,"last_password_recovery_at":null,"merge_group_id":null,
        "origin_instance_id":null,"creator_client_id":null,"nationality":null,
        "primary_email":null,"secondary_email":null,
        "primary_email_verified":null,"secondary_email_verified":null,"has_picture":false,
        "badges":null,"client_id":null,"verification":null,"subscription_news":false
      }
      """
    When I send a POST request to "/auth/token?code=codefromauth"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "access_token": "access_token_from_oauth",
          "expires_in": 31622420
        }
      }
      """
    And there are 10 sessions for user @UserWith10Sessions
    And there is no session @Session_UserWith10Sessions_1

  Scenario: Should add the new session and delete the oldest one when the user have 10 sessions
    Given  the login module "token" endpoint for code "codefromauth" returns 200 with body:
      """
      {
        "token_type": "Bearer",
        "expires_in": 31622420,
        "access_token": "access_token_from_oauth",
        "refresh_token": "refresh_token_from_oauth"
      }
      """
    And the login module "account" endpoint for token "access_token_from_oauth" returns 200 with body:
      """
      {
        "id":11, "login":"login","login_updated_at":null,"login_fixed":0,
        "login_revalidate_required":0,"login_change_required":0,"language":null,"first_name":null,
        "last_name":null,"real_name_visible":false,"timezone":null,"country_code":null,
        "address":null,"city":null,"zipcode":null,"primary_phone":null,"secondary_phone":null,
        "role":null,"school_grade":null,"student_id":null,"ministry_of_education":null,
        "ministry_of_education_fr":false,"birthday":null,"presentation":null,
        "website":null,"ip":null,"picture":null,
        "gender":null,"graduation_year":null,"graduation_grade_expire_at":null,
        "graduation_grade":null,"created_at":null,"last_login":null,
        "logout_config":null,"last_password_recovery_at":null,"merge_group_id":null,
        "origin_instance_id":null,"creator_client_id":null,"nationality":null,
        "primary_email":null,"secondary_email":null,
        "primary_email_verified":null,"secondary_email_verified":null,"has_picture":false,
        "badges":null,"client_id":null,"verification":null,"subscription_news":false
      }
      """
    When I send a POST request to "/auth/token?code=codefromauth"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "access_token": "access_token_from_oauth",
          "expires_in": 31622420
        }
      }
      """
    And there are 10 sessions for user @UserWith10Sessions
    And there is no session @Session_UserWith11Sessions_2
    And there is no session @Session_UserWith11Sessions_3
