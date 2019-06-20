Feature: Set users.sNotificationReadDate to NOW() for the current user
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | sNotificationReadDate |
      | 1  | user   | null                  |
      | 2  | admin  | 2017-02-21T06:38:38Z  |

  Scenario: Successfully send a request
    Given I am the user with ID "1"
    When I send a PUT request to "/current-user/notification-read-date"
    Then the response should be "updated"
    And the table "users" should stay unchanged but the row with ID "1"
    And the table "users" at ID "1" should be:
      | ID | sLogin | ABS(NOW() - sNotificationReadDate) < 3 |
      | 1  | user   | 1                                      |
