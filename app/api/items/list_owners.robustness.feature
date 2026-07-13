Feature: List owner groups for an item - robustness
  Background:
    Given the database has the following user:
      | group_id | login |
      | 11       | jdoe  |
    And the database has the following table "items":
      | id | default_language_tag |
      | 21 | en                   |
      | 22 | en                   |
      | 50 | en                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_edit_generated |
      | 11       | 21      | solution           | children           |
      | 11       | 22      | info               | none               |
      | 11       | 50      | solution           | all                |

  Scenario: User not found
    Given I am the user with id "404"
    When I send a GET request to "/items/50/owners"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Wrong item_id
    Given I am the user with id "11"
    When I send a GET request to "/items/abc/owners"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Insufficient can_edit (children)
    Given I am the user with id "11"
    When I send a GET request to "/items/21/owners"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Insufficient can_edit (view-only)
    Given I am the user with id "11"
    When I send a GET request to "/items/22/owners"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
