Feature: Get user info the current user
  Background:
    Given the database has the following table 'users':
      | ID | tempUser | sLogin | sRegistrationDate    | sEmail         | sFirstName  | sLastName | sStudentId | sCountryCode | sTimeZone | sBirthDate | iGraduationYear | iGrade | sSex | sAddress         | sZipcode | sCity         | sLandLineNumber | sCellPhoneNumber | sDefaultLanguage | bPublicFirstName | bPublicLastName | bNotifyNews | sNotify | sFreeText | sWebSite   | bPhotoAutoload | sLangProg | bBasicEditorMode | nbSpacesForTab | iStepLevelInSite | bIsAdmin | bNoRanking | loginModulePrefix | allowSubgroups |
      | 2  | 0        | user   | 2017-02-26T06:38:38Z | user@gmail.com | John        | Doe       | Some ID    | us           | PT        | 1975-12-13 | 1997            | 10     | Male | 314 N Beverly Dr | 90210    | Beverly Hills | +1 310-435-9669 | +1 310-860-9581  | en               | true             | true            | true        | Answers | Some text | mysite.com | true           | Python    | true             | 3              | 11               | false    | false      | my_prefix         | false          |
      | 3  | 1        | jane   | null                 | null           | null        | null      | null       |              | null      | null       | 0               | null   | null | null             | null     | null          | null            | null             | fr               | false            | false           | false       | Never   | null      | null       | false          | null      | false            | 0              | 0                | true     | true       | null              | null           |

  Scenario: All field values are not nulls
    Given I am the user with ID "2"
    When I send a GET request to "/current-user"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "2",
      "temp_user": false,
      "login": "user",
      "registration_date": "2017-02-26T06:38:38Z",
      "email": "user@gmail.com",
      "first_name": "John",
      "last_name": "Doe",
      "student_id": "Some ID",
      "country_code": "us",
      "time_zone": "PT",
      "birth_date": "1975-12-13",
      "graduation_year": 1997,
      "grade": 10,
      "sex": "Male",
      "address": "314 N Beverly Dr",
      "zip_code": "90210",
      "city": "Beverly Hills",
      "land_line_number": "+1 310-435-9669",
      "cell_phone_number": "+1 310-860-9581",
      "default_language": "en",
      "public_first_name": true,
      "public_last_name": true,
      "notify_news": true,
      "notify": "Answers",
      "free_text": "Some text",
      "web_site": "mysite.com",
      "photo_autoload": true,
      "lang_prog": "Python",
      "basic_editor_mode": true,
      "spaces_for_tab": 3,
      "step_level_in_site": 11,
      "is_admin": false,
      "no_ranking": false,
      "login_module_prefix": "my_prefix",
      "allow_subgroups": false
    }
    """

  Scenario: All nullable field values are nulls
    Given I am the user with ID "3"
    When I send a GET request to "/current-user"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "3",
      "temp_user": true,
      "login": "jane",
      "country_code": "",
      "graduation_year": 0,
      "default_language": "fr",
      "public_first_name": false,
      "public_last_name": false,
      "notify_news": false,
      "notify": "Never",
      "photo_autoload": false,
      "basic_editor_mode": false,
      "spaces_for_tab": 0,
      "step_level_in_site": 0,
      "is_admin": true,
      "no_ranking": true
    }
    """
