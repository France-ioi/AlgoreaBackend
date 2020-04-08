Feature: Get additional times for a group of users/teams on a contest (contestListMembersAdditionalTime) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name    |
      | 12 | Group A |
      | 13 | Group B |
      | 21 | owner   |
    And the database has the following table 'users':
      | login | group_id |
      | owner | 21       |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id | duration | default_language_tag |
      | 50 | 00:00:00 | fr                   |
      | 60 | null     | fr                   |
      | 10 | 00:00:02 | fr                   |
      | 70 | 00:00:03 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 50      | content                  |
      | 13       | 60      | info                     |
      | 13       | 70      | content_with_descendants |
      | 21       | 50      | none                     |
      | 21       | 60      | content_with_descendants |
      | 21       | 70      | content_with_descendants |

  Scenario: Wrong item_id
    Given I am the user with id "21"
    When I send a GET request to "/contests/abc/groups/13/members/additional-times"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: No such item
    Given I am the user with id "21"
    When I send a GET request to "/contests/90/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the item
    Given I am the user with id "21"
    When I send a GET request to "/contests/10/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item is not a timed contest
    Given I am the user with id "21"
    When I send a GET request to "/contests/60/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is not a contest admin
    Given I am the user with id "21"
    When I send a GET request to "/contests/50/groups/13/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Wrong group_id
    Given I am the user with id "21"
    When I send a GET request to "/contests/70/groups/abc/members/additional-times"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: The user is not a manager of the group
    Given I am the user with id "21"
    When I send a GET request to "/contests/70/groups/12/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No such group
    Given I am the user with id "21"
    When I send a GET request to "/contests/70/groups/404/members/additional-times"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Wrong sort
    Given I am the user with id "21"
    When I send a GET request to "/contests/70/groups/13/members/additional-times?sort=title"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "title""

