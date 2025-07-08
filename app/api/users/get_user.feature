Feature: Get user info
  Background:
    Given the database has the following users:
      | group_id | temp_user | login | first_name | last_name | default_language | free_text | web_site   |
      | 2        | 0         | user  | John       | Doe       | en               | Some text | mysite.com |
      | 3        | 1         | jane  | null       | null      | fr               | null      | null       |
      | 4        | 0         | john  | null       | null      | fr               | null      | null       |
      | 5        | 0         | paul  | null       | null      | fr               | null      | null       |
    And the database has the following table "groups":
      | id | name          |
      | 10 | Some group    |
      | 11 | Another group |
      | 12 | Friends       |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 11              | 4              | null                           |
      | 10              | 2              | 2019-05-30 11:00:00            |
      | 12              | 5              | null                           |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members | can_grant_group_access |
      | 10       | 11         | false             | false                  |
      | 12       | 4          | true              | false                  |
      | 12       | 10         | false             | true                   |

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
      "web_site": "mysite.com",
      "is_current_user": true,
      "ancestors_current_user_is_manager_of": []
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
      "login": "jane",
      "is_current_user": true,
      "ancestors_current_user_is_manager_of": []
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
      "web_site": "mysite.com",
      "is_current_user": false,
      "ancestors_current_user_is_manager_of": []
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
      "last_name": "Doe",
      "is_current_user": false,
      "ancestors_current_user_is_manager_of": [{"id": "10", "name": "Some group"}],
      "current_user_can_grant_user_access": false,
      "current_user_can_watch_user": false,
      "personal_info_access_approval_to_current_user": "none"
    }
    """

  Scenario: The current user can watch members of one of user's groups
    Given I am the user with id "4"
    When I send a GET request to "/users/5"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "5",
      "temp_user": false,
      "login": "paul",
      "free_text": null,
      "web_site": null,
      "is_current_user": false,
      "ancestors_current_user_is_manager_of": [{"id": "12", "name": "Friends"}],
      "current_user_can_grant_user_access": false,
      "current_user_can_watch_user": true,
      "personal_info_access_approval_to_current_user": "none"
    }
    """

  Scenario: The current user can grant group access to members of one of user's groups
    Given I am the user with id "2"
    When I send a GET request to "/users/5"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "group_id": "5",
      "temp_user": false,
      "login": "paul",
      "free_text": null,
      "web_site": null,
      "is_current_user": false,
      "ancestors_current_user_is_manager_of": [{"id": "12", "name": "Friends"}],
      "current_user_can_grant_user_access": true,
      "current_user_can_watch_user": false,
      "personal_info_access_approval_to_current_user": "none"
    }
    """
