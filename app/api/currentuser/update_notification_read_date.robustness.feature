Feature: Set users.sNotificationReadDate to NOW() for the current user - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | sNotificationReadDate |
      | 1  | user   | null                  |
      | 2  | admin  | 2017-02-21T06:38:38Z  |

  Scenario: No such user
    Given I am the user with ID "404"
    When I send a PUT request to "/current-user/notification-read-date"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users" should stay unchanged
