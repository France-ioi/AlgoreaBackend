Feature: Generate a group results token (groupResultsTokenGenerate)
  Background:
    Given the database has the following users:
      | group_id | login | default_language |
      | 21       | owner | en               |
    And the groups ancestors are computed
    And the time now is "2019-07-16T22:02:28Z"

  Scenario: Successfully generate a group results token
    Given I am the user with id "21"
    And the database has the following table "groups":
      | id | type  | name      |
      | 1  | Base  | Root 1    |
      | 11 | Class | Our Class |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 1        | 21         | true              |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | is_team_membership |
      | 1               | 11             | 0                  |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id  | type    | default_language_tag |
      | 210 | Chapter | fr                   |
      | 220 | Chapter | fr                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 21       | 210     | info               | answer              |
      | 21       | 220     | info               | answer              |
    And "expectedGroupResultsToken" is a token signed by the app with the following payload:
      """
      {
        "user_id": "21",
        "group_id": "11",
        "item_ids": ["210", "220"],
        "exp": 1563318148
      }
      """
    When I send a POST request to "/groups/11/group-results-token?parent_item_ids=210,220"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "group_results_token": "{{expectedGroupResultsToken}}",
          "expires_in": 3600
        }
      }
      """

  Scenario: Successfully generate a token with empty parent_item_ids
    Given I am the user with id "21"
    And the database has the following table "groups":
      | id | type  | name      |
      | 1  | Base  | Root 1    |
      | 11 | Class | Our Class |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 1        | 21         | true              |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | is_team_membership |
      | 1               | 11             | 0                  |
    And the groups ancestors are computed
    And "expectedGroupResultsToken" is a token signed by the app with the following payload:
      """
      {
        "user_id": "21",
        "group_id": "11",
        "item_ids": [],
        "exp": 1563318148
      }
      """
    When I send a POST request to "/groups/11/group-results-token?parent_item_ids="
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "group_results_token": "{{expectedGroupResultsToken}}",
          "expires_in": 3600
        }
      }
      """
