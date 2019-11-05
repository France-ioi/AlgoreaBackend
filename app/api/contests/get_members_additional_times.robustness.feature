Feature: Get additional times for a group of users/teams on a contest (contestListMembersAdditionalTime) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name        |
      | 12 | Group A     |
      | 13 | Group B     |
      | 21 | owner       |
      | 22 | owner-admin |
    And the database has the following table 'users':
      | login | group_id | owned_group_id |
      | owner | 21       | 22             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 12                | 12             | 1       |
      | 13                | 13             | 1       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |
    And the database has the following table 'items':
      | id | duration |
      | 50 | 00:00:00 |
      | 60 | null     |
      | 10 | 00:00:02 |
      | 70 | 00:00:03 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 50      | content                  |
      | 13       | 60      | info                     |
      | 13       | 70      | content_with_descendants |
      | 21       | 50      | none                     |
      | 21       | 60      | content_with_descendants |
      | 21       | 70      | content_with_descendants |

  Scenario: Wrong item_id
    Given I am the user with group_id "21"
    When I send a GET request to "/contests/abc/groups/13/members/additional-times"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: No such item
    Given I am the user with group_id "21"
    When I send a GET request to "/contests/90/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the item
    Given I am the user with group_id "21"
    When I send a GET request to "/contests/10/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item is not a timed contest
    Given I am the user with group_id "21"
    When I send a GET request to "/contests/60/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is not a contest admin
    Given I am the user with group_id "21"
    When I send a GET request to "/contests/50/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Wrong group_id
    Given I am the user with group_id "21"
    When I send a GET request to "/contests/70/groups/abc/members/additional-times"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: The group is not owned by the user
    Given I am the user with group_id "21"
    When I send a GET request to "/contests/70/groups/12/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No such group
    Given I am the user with group_id "21"
    When I send a GET request to "/contests/70/groups/404/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Wrong sort
    Given I am the user with group_id "21"
    When I send a GET request to "/contests/70/groups/13/members/additional-times?sort=title"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "title""

