Feature: Update the local user info cache
  Background:
    Given the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
      """

  Scenario Outline: Update an existing user
    Given the time now is "2019-07-16T22:02:29Z"
    And the DB time now is "2019-07-16 22:02:28"
    And the template constant "profile_with_all_fields_set" is:
      """
      {
        "id":100000001, "login":"jane","login_updated_at":"2019-07-16 01:56:25","login_fixed":0,
        "login_revalidate_required":0,"login_change_required":0,"language":"en","first_name":"Jane",
        "last_name":"Doe","real_name_visible":false,"timezone":"Europe\/London","country_code":"GB",
        "address":null,"city":null,"zipcode":null,"primary_phone":null,"secondary_phone":null,
        "role":"student","school_grade":null,"student_id":"456789012","ministry_of_education":null,
        "ministry_of_education_fr":false,"birthday":"2001-08-03","presentation":"I'm Jane Doe",
        "website":"http://jane.freepages.com","ip":"192.168.11.1","picture":"http:\/\/127.0.0.1:8000\/images\/user.png",
        "gender":"f","graduation_year":2021,"graduation_grade_expire_at":"2020-07-01 00:00:00",
        "graduation_grade":0,"created_at":"2019-07-16 01:56:25","last_login":"2019-07-22 14:47:18",
        "logout_config":null,"last_password_recovery_at":null,"merge_group_id":null,
        "origin_instance_id":null,"creator_client_id":null,"nationality":"GB",
        "primary_email":"janedoe@gmail.com","secondary_email":"jane.doe@gmail.com",
        "primary_email_verified":1,"secondary_email_verified":null,"has_picture":false,
        "badges":[],"client_id":1,"verification":[]
      }
      """
    And the template constant "profile_with_null_fields" is:
      """
      {
        "id":100000001, "login":"jane","login_updated_at":null,"login_fixed":0,
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
        "badges":null,"client_id":null,"verification":null
      }
      """
    And the database has the following table 'users':
      | id | self_group_id | owned_group_id | last_login_date     | last_activity_date  | registration_date   | login_id  | login    | email                | first_name | last_name | student_id | country_code | birth_date | graduation_year | grade | address           | zipcode  | city                | land_line_number  | cell_phone_number | default_language | free_text           | web_site                      | sex  | email_verified | last_ip     |
      | 1  | 11            | 12             | 2019-06-16 21:01:25 | 2019-06-16 22:05:44 | 2019-05-10 10:42:11 | 100000001 | mohammed | mohammedam@gmail.com | Mohammed   | Amrani    | 123456789  | dz           | 2000-07-02 | 2020            | 0     | Rue Tebessi Larbi | 16000    | Algiers             | +213 778 02 85 31 | null              | en               | I'm Mohammed Amrani | http://mohammed.freepages.com | Male | 0              | 192.168.0.1 |
      | 2  | 13            | 14             | 2018-06-16 21:01:25 | 2018-06-16 22:05:44 | 2018-05-10 10:42:11 | 100000002 | john     | johndoe@gmail.com    | John       | Doe       | 987654321  | gb           | 1999-03-20 | 2021            | 1     | 1, Trafalgar sq.  | WC2N 5DN | City of Westminster | +44 20 7747 2885  | +44 333 300 7774  | en               | I'm John Doe        | http://johndoe.freepages.com  | Male | 1              | 110.55.10.2 |
    And the database has the following table 'sessions':
      | expiration_date     | user_id | issuer       | issued_at_date      | access_token |
      | 2020-06-16 22:02:49 | 1       | login-module | 2019-06-16 22:02:28 | accesstoken  |
    And the login module "account" endpoint for token "accesstoken" returns 200 with body:
      """
      {{<profile_response_name>}}
      """
    And the "Authorization" request header is "Bearer accesstoken"
    When I send a PUT request to "/current-user/refresh"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "updated"
      }
      """
    And the table "users" should stay unchanged but the row with id "1"
    And the table "users" at id "1" should be:
      | id | self_group_id | owned_group_id | last_login_date     | last_activity_date  | temp_user | registration_date   | login_id  | login | email   | first_name   | last_name   | student_id   | country_code   | birth_date   | graduation_year   | grade   | address | zipcode | city | land_line_number | cell_phone_number | default_language   | free_text   | web_site   | sex   | email_verified   | last_ip     |
      | 1  | 11            | 12             | 2019-06-16 21:01:25 | 2019-07-16 22:02:28 | 0         | 2019-05-10 10:42:11 | 100000001 | jane  | <email> | <first_name> | <last_name> | <student_id> | <country_code> | <birth_date> | <graduation_year> | <grade> | null    | null    | null | null             | null              | <default_language> | <free_text> | <web_site> | <sex> | <email_verified> | 192.168.0.1 |
  Examples:
    | profile_response_name       | email             | first_name | last_name | student_id | country_code | birth_date | graduation_year | grade | default_language | free_text    | web_site                  | sex    | email_verified |
    | profile_with_all_fields_set | janedoe@gmail.com | Jane       | Doe       | 456789012  | gb           | 2001-08-03 | 2021            | 0     | en               | I'm Jane Doe | http://jane.freepages.com | Female | true           |
    | profile_with_null_fields    | null              | null       | null      | null       |              | null       | 0               | null  | fr               | null         | null                      | null   | false          |
