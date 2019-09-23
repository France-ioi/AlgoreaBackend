Feature: Get user info the current user
  Background:
    Given the database has the following table 'users':
      | id | temp_user | login | registration_date   | email          | first_name  | last_name | student_id | country_code | time_zone | birth_date | graduation_year | grade | sex  | address          | zipcode  | city          | land_line_number | cell_phone_number | default_language | public_first_name | public_last_name | notify_news | notify | free_text | web_site   | photo_autoload | lang_prog | basic_editor_mode | spaces_for_tab | step_level_in_site | is_admin | no_ranking | login_module_prefix | allow_subgroups |
      | 2  | 0         | user  | 2017-02-26 06:38:38 | user@gmail.com | John        | Doe       | Some id    | us           | PT        | 1975-12-13 | 1997            | 10    | Male | 314 N Beverly Dr | 90210    | Beverly Hills | +1 310-435-9669  | +1 310-860-9581   | en               | true              | true            | true        | Answers | Some text | mysite.com | true           | Python    | true              | 3              | 11                 | false    | false      | my_prefix           | false           |
      | 3  | 1         | jane  | null                | null           | null        | null      | null       |              | null      | null       | 0               | null  | null | null             | null     | null          | null             | null              | fr               | false             | false           | false       | Never   | null      | null       | false          | null      | false             | 0              | 0                  | true     | true       | null                | null            |

  Scenario: All field values are not nulls
    Given I am the user with id "2"
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
      "email_verified": false,
      "first_name": "John",
      "last_name": "Doe",
      "student_id": "Some id",
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
    Given I am the user with id "3"
    When I send a GET request to "/current-user"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "address": null,
      "allow_subgroups": null,
      "birth_date": null,
      "cell_phone_number": null,
      "city": null,
      "email": null,
      "email_verified": false,
      "first_name": null,
      "free_text": null,
      "grade": null,
      "id": "3",
      "land_line_number": null,
      "lang_prog": null,
      "last_name": null,
      "login_module_prefix": null,
      "registration_date": null,
      "sex": null,
      "student_id": null,
      "time_zone": null,
      "temp_user": true,
      "web_site": null,
      "zip_code": null,
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
