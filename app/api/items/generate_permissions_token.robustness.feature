Feature: Generate a permissions token for an item - robustness
  Background:
    Given the database has the following user:
      | group_id | login |
      | 101      | john  |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | default_language_tag |
      | 50 | fr                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 101      | 50      | none               |

  Scenario: Invalid item_id
    Given I am the user with id "101"
    When I send a POST request to "/items/abc/permissions-token"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: User not found
    Given I am the user with id "404"
    When I send a POST request to "/items/50/permissions-token"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: No access to the item (can_view is none)
    Given I am the user with id "101"
    When I send a POST request to "/items/50/permissions-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No permissions row for the item
    Given the database has the following table "items":
      | id  | default_language_tag |
      | 100 | fr                   |
    And I am the user with id "101"
    When I send a POST request to "/items/100/permissions-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Item does not exist
    Given I am the user with id "101"
    When I send a POST request to "/items/999/permissions-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
