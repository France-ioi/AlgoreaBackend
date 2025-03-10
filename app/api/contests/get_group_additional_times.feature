Feature: Get additional times for a group on a contest (contestGetAdditionalTime)
  Background:
    Given the database has the following table "groups":
      | id | name        | type    |
      | 10 | Parent      | Club    |
      | 11 | Group A     | Friends |
      | 14 | Group B     | Other   |
    And the database has the following users:
      | group_id | login |
      | 21       | owner |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access | can_watch_members |
      | 11       | 21         | true                   | true              |
      | 14       | 21         | true                   | true              |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 10              | 11             |
      | 11              | 14             |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | duration | entry_participant_type | default_language_tag |
      | 50 | 00:00:00 | User                   | fr                   |
      | 60 | 00:00:01 | Team                   | fr                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_watch_generated |
      | 21       | 50      | content                  | enter                    | result              |
      | 21       | 60      | content_with_descendants | content                  | answer              |
    And the database has the following table "groups_contest_items":
      | group_id | item_id | additional_time |
      | 10       | 50      | 01:00:00        |
      | 11       | 50      | 00:01:00        |

  Scenario: With additional time
    Given I am the user with id "21"
    When I send a GET request to "/contests/50/groups/11/additional-times"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "additional_time": 60,
      "total_additional_time": 3660
    }
    """

  Scenario: Without additional time
    Given I am the user with id "21"
    When I send a GET request to "/contests/60/groups/14/additional-times"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "additional_time": 0,
      "total_additional_time": 0
    }
    """
