Feature: Get additional times for a group of users/teams on an item with duration (itemListMembersAdditionalTime) - robustness
  Background:
    Given the database has the following table "groups":
      | id | name    |
      | 12 | Group A |
      | 13 | Group B |
      | 14 | Group C |
      | 15 | Group D |
    And the database has the following user:
      | group_id | login |
      | 21       | owner |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access | can_watch_members |
      | 13       | 21         | true                   | true              |
      | 14       | 21         | true                   | false             |
      | 15       | 21         | false                  | true              |
    And the database has the following table "items":
      | id | duration | default_language_tag |
      | 50 | 00:00:00 | fr                   |
      | 60 | null     | fr                   |
      | 10 | 00:00:02 | fr                   |
      | 70 | 00:00:03 | fr                   |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_watch_generated |
      | 13       | 50      | content                  | none                     | none                |
      | 13       | 60      | info                     | none                     | none                |
      | 13       | 70      | content_with_descendants | none                     | none                |
      | 21       | 50      | none                     | none                     | none                |
      | 21       | 60      | content_with_descendants | none                     | none                |
      | 21       | 70      | content_with_descendants | enter                    | result              |

  Scenario: Wrong item_id
    Given I am the user with id "21"
    When I send a GET request to "/items/abc/groups/13/members/additional-times"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: No such item
    Given I am the user with id "21"
    When I send a GET request to "/items/90/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the item
    Given I am the user with id "21"
    When I send a GET request to "/items/10/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item is not a timed contest
    Given I am the user with id "21"
    When I send a GET request to "/items/60/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is not a contest admin
    Given I am the user with id "21"
    When I send a GET request to "/items/50/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Wrong group_id
    Given I am the user with id "21"
    When I send a GET request to "/items/70/groups/abc/members/additional-times"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: The user is not a manager of the group
    Given I am the user with id "21"
    When I send a GET request to "/items/70/groups/12/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user cannot watch for members of the group
    Given I am the user with id "21"
    When I send a GET request to "/items/70/groups/14/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user cannot grant access to the group
    Given I am the user with id "21"
    When I send a GET request to "/items/70/groups/15/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No such group
    Given I am the user with id "21"
    When I send a GET request to "/items/70/groups/404/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Wrong sort
    Given I am the user with id "21"
    When I send a GET request to "/items/70/groups/13/members/additional-times?sort=title"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "title""
