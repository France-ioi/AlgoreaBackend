Feature: Get user info for the current user
  Background:
    Given the database has the following users:
      | group_id | temp_user | login | registered_at       | latest_profile_sync_at | profile                                    | default_language | is_admin |
      | 2        | 0         | user  | 2017-02-26 06:38:38 | 2019-05-30 12:00:00    | {"first_name": "John", "last_name": "Doe"} | en               | false    |
      | 3        | 1         | jane  | null                | null                   | null                                       | fr               | true     |

  Scenario: All field values are not nulls
    Given I am the user with id "2"
    When I send a GET request to "/current-user"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "2",
      "login": "user",
      "profile": {
        "first_name": "John",
        "last_name": "Doe"
      },
      "default_language": "en",
      "temp_user": false
    }
    """

  Scenario: All nullable field values are nulls
    Given I am the user with id "3"
    When I send a GET request to "/current-user"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "3",
      "login": "jane",
      "profile": null,
      "default_language": "fr",
      "temp_user": true
    }
    """
