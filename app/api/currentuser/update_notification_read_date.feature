Feature: Set users.notification_read_date to NOW() for the current user
  Background:
    Given the database has the following table 'users':
      | id | login | notification_read_date |
      | 1  | user  | null                   |
      | 2  | admin | 2017-02-21 06:38:38    |

  Scenario: Successfully send a request
    Given I am the user with id "1"
    When I send a PUT request to "/current-user/notification-read-date"
    Then the response should be "updated"
    And the table "users" should stay unchanged but the row with id "1"
    And the table "users" at id "1" should be:
      | id | login | ABS(TIMESTAMPDIFF(SECOND, notification_read_date, NOW())) < 3 |
      | 1  | user  | 1                                                             |
