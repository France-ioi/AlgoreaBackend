Feature: Get user info
  Background:
    Given the database has the following users:
      | group_id | temp_user | login | first_name | last_name | default_language | free_text | web_site   |
      | 2        | 0         | user  | John       | Doe       | en               | Some text | mysite.com |
      | 3        | 1         | jane  | null       | null      | fr               | null      | null       |
      | 4        | 0         | john  | null       | null      | fr               | null      | null       |
    And the database has the following table 'groups':
      | id |
      | 10 |
      | 11 |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 11              | 4              | null                           |
      | 10              | 2              | 2019-05-30 11:00:00            |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 10       | 11         |

  Scenario: All field values are not nulls
    Given I am the user with id "2"
    When I send a GET request to "/users/2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "2",
      "temp_user": false,
      "login": "user",
      "first_name": "John",
      "last_name": "Doe",
      "free_text": "Some text",
      "web_site": "mysite.com"
    }
    """

  Scenario: All nullable field values are nulls
    Given I am the user with id "3"
    When I send a GET request to "/users/3"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "first_name": null,
      "free_text": null,
      "group_id": "3",
      "last_name": null,
      "temp_user": true,
      "web_site": null,
      "login": "jane"
    }
    """

  Scenario: No access to personal info
    Given I am the user with id "3"
    When I send a GET request to "/users/2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "2",
      "temp_user": false,
      "login": "user",
      "free_text": "Some text",
      "web_site": "mysite.com"
    }
    """

  Scenario: A user has approved access to his personal info for a group managed by the current user
    Given I am the user with id "4"
    When I send a GET request to "/users/2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "2",
      "temp_user": false,
      "login": "user",
      "free_text": "Some text",
      "web_site": "mysite.com",
      "first_name": "John",
      "last_name": "Doe"
    }
    """
