Feature: Generate a permissions token for an item
  Background:
    Given the database has the following users:
      | group_id | login |
      | 101      | john  |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | default_language_tag |
      | 50 | fr                   |
      | 60 | fr                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 101      | 50      | content            | content                  | result              | children           | true               |
      | 101      | 60      | info               | none                     | none                | none               | false              |
    And the time now is "2019-07-16T22:02:28Z"

  Scenario: Successfully generate a permissions token with full permissions
    Given I am the user with id "101"
    And "expectedPermissionsToken" is a token signed by the app with the following payload:
      """
      {
        "user_id": "101",
        "item_id": "50",
        "can_view": "content",
        "can_grant_view": "content",
        "can_watch": "result",
        "can_edit": "children",
        "is_owner": true,
        "exp": 1563321748
      }
      """
    When I send a POST request to "/items/50/permissions-token"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "permissions_token": "{{expectedPermissionsToken}}",
          "expires_in": 7200
        }
      }
      """

  Scenario: Successfully generate a permissions token with minimal permissions
    Given I am the user with id "101"
    And "expectedPermissionsToken" is a token signed by the app with the following payload:
      """
      {
        "user_id": "101",
        "item_id": "60",
        "can_view": "info",
        "can_grant_view": "none",
        "can_watch": "none",
        "can_edit": "none",
        "is_owner": false,
        "exp": 1563321748
      }
      """
    When I send a POST request to "/items/60/permissions-token"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "permissions_token": "{{expectedPermissionsToken}}",
          "expires_in": 7200
        }
      }
      """

  Scenario: Permissions are aggregated across group ancestors
    Given the database has the following table "groups":
      | id  | type  |
      | 102 | Class |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 102             | 101            |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | default_language_tag |
      | 70 | fr                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 101      | 70      | content            | none                     | none                | none               | false              |
      | 102      | 70      | info               | content                  | answer              | all                | true               |
    And I am the user with id "101"
    And "expectedPermissionsToken" is a token signed by the app with the following payload:
      """
      {
        "user_id": "101",
        "item_id": "70",
        "can_view": "content",
        "can_grant_view": "content",
        "can_watch": "answer",
        "can_edit": "all",
        "is_owner": true,
        "exp": 1563321748
      }
      """
    When I send a POST request to "/items/70/permissions-token"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {
          "permissions_token": "{{expectedPermissionsToken}}",
          "expires_in": 7200
        }
      }
      """
