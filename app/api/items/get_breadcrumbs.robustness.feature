Feature: Get item breadcrumbs - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name    | grade | type  |
      | 11 | jdoe    | -2    | User  |
      | 13 | Group B | -2    | Class |
    And the database has the following table 'users':
      | login | temp_user | group_id |
      | jdoe  | 0         | 11       |
    And the database has the following table 'items':
      | id | no_score | type    | default_language_tag |
      | 21 | false    | Task    | fr                   |
      | 22 | false    | Task    | fr                   |
      | 23 | false    | Chapter | fr                   |
      | 24 | false    | Task    | fr                   |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title            |
      | 21      | en           | Graph: Methods   |
      | 22      | en           | DFS              |
      | 23      | en           | Reduce Graph     |
      | 21      | fr           | Graphe: Methodes |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 11             |
    And the groups ancestors are computed

  Scenario: Should fail when breadcrumb hierarchy is corrupt (one parent-child link missing), but user has full access to all
    Given the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 21      | content_with_descendants |
      | 13       | 22      | content_with_descendants |
      | 13       | 23      | content_with_descendants |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 22             | 23            | 1           |
    And the database has the following table 'attempts':
      | participant_id | id | parent_attempt_id | root_item_id |
      | 11             | 0  | null              | null         |
      | 11             | 1  | 0                 | 22           |
      | 11             | 2  | 0                 | 22           |
    And the database has the following table 'results':
      | participant_id | attempt_id | item_id | started_at          |
      | 11             | 0          | 21      | 2019-05-30 11:00:00 |
      | 11             | 1          | 22      | 2019-05-30 11:00:00 |
      | 11             | 2          | 22      | 2019-05-29 11:00:00 |
      | 11             | 1          | 23      | 2019-05-29 11:00:00 |
      | 11             | 2          | 23      | 2019-05-30 11:00:00 |
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/breadcrumbs?attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Item ids hierarchy is invalid or insufficient access rights"

  Scenario: Should fail when breadcrumb hierarchy is corrupt (one parent-child link missing at the end), but user has full access to all
    Given the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 21      | content_with_descendants |
      | 13       | 22      | content_with_descendants |
      | 13       | 23      | content_with_descendants |
      | 13       | 24      | content_with_descendants |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 21             | 22            | 1           |
      | 22             | 23            | 1           |
    And the database has the following table 'attempts':
      | participant_id | id | parent_attempt_id | root_item_id |
      | 11             | 0  | null              | null         |
      | 11             | 1  | 0                 | 22           |
      | 11             | 2  | 0                 | 22           |
    And the database has the following table 'results':
      | participant_id | attempt_id | item_id | started_at          |
      | 11             | 0          | 21      | 2019-05-30 11:00:00 |
      | 11             | 1          | 22      | 2019-05-30 11:00:00 |
      | 11             | 2          | 22      | 2019-05-29 11:00:00 |
      | 11             | 1          | 23      | 2019-05-29 11:00:00 |
      | 11             | 2          | 23      | 2019-05-30 11:00:00 |
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/24/breadcrumbs?parent_attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Item ids hierarchy is invalid or insufficient access rights"

  Scenario: Should fail when breadcrumb hierarchy is corrupt (one item missing), and user has full access to all
    Given the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 21      | content_with_descendants |
      | 13       | 22      | content_with_descendants |
      | 13       | 24      | content_with_descendants |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 21             | 22            | 1           |
      | 22             | 23            | 1           |
      | 23             | 24            | 1           |
    And the database has the following table 'attempts':
      | participant_id | id | parent_attempt_id | root_item_id |
      | 11             | 0  | null              | null         |
      | 11             | 1  | 0                 | 22           |
    And the database has the following table 'results':
      | participant_id | attempt_id | item_id | started_at          |
      | 11             | 0          | 21      | 2019-05-30 11:00:00 |
      | 11             | 1          | 22      | 2019-05-30 11:00:00 |
      | 11             | 1          | 23      | 2019-05-29 11:00:00 |
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/24/23/breadcrumbs?parent_attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Item ids hierarchy is invalid or insufficient access rights"

  Scenario: Should fail when the first item of hierarchy is not a root item, and user has full access to all
    Given the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 22      | content_with_descendants |
      | 13       | 23      | content_with_descendants |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 22             | 23            | 1           |
      | 23             | 24            | 1           |
    And the database has the following table 'attempts':
      | participant_id | id | parent_attempt_id | root_item_id |
      | 11             | 0  | null              | null         |
      | 11             | 1  | 0                 | 22           |
    And the database has the following table 'results':
      | participant_id | attempt_id | item_id | started_at          |
      | 11             | 0          | 21      | 2019-05-30 11:00:00 |
      | 11             | 1          | 22      | 2019-05-30 11:00:00 |
      | 11             | 1          | 23      | 2019-05-29 11:00:00 |
    And I am the user with id "11"
    When I send a GET request to "/items/22/23/breadcrumbs?attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Item ids hierarchy is invalid or insufficient access rights"

  Scenario: Should fail when the user has 'info' access to middle element, 'content' access to the rest
    Given the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 13       | 21      | content            |
      | 13       | 22      | info               |
      | 13       | 23      | content            |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 22             | 23            | 1           |
    And the database has the following table 'attempts':
      | participant_id | id | parent_attempt_id | root_item_id |
      | 11             | 0  | null              | null         |
      | 11             | 1  | 0                 | 22           |
    And the database has the following table 'results':
      | participant_id | attempt_id | item_id | started_at          |
      | 11             | 0          | 21      | 2019-05-30 11:00:00 |
      | 11             | 1          | 22      | 2019-05-30 11:00:00 |
      | 11             | 1          | 23      | 2019-05-29 11:00:00 |
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/breadcrumbs?attempt_id=1"
    Then the response code should be 403
    And the response error message should contain "Item ids hierarchy is invalid or insufficient access rights"

  Scenario: Should fail when the user doesn't exist
    And I am the user with id "404"
    When I send a GET request to "/items/21/22/23/breadcrumbs"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Invalid ids
    And I am the user with id "11"
    When I send a GET request to "/items/11111111111111111111111111111/2222222222222222222222222222/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: '11111111111111111111111111111', param: 'ids')"

  Scenario: More than 10 ids
    And I am the user with id "11"
    When I send a GET request to "/items/1/2/3/4/5/6/7/8/9/10/11/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "No more than 10 ids expected"

  Scenario: Invalid attempt_id
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/breadcrumbs?attempt_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for attempt_id (should be int64)"

  Scenario: Invalid parent_attempt_id
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/breadcrumbs?parent_attempt_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for parent_attempt_id (should be int64)"

  Scenario: Both attempt_id and parent_attempt_id are given
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/breadcrumbs?attempt_id=2&parent_attempt_id=1"
    Then the response code should be 400
    And the response error message should contain "Only one of attempt_id and parent_attempt_id can be given"

  Scenario: No attempt given
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "One of attempt_id and parent_attempt_id should be given"

  Scenario: Invalid as_team_id
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/breadcrumbs?attempt_id=1&as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: as_team_id is not the user's team
    And I am the user with id "11"
    When I send a GET request to "/items/21/22/23/breadcrumbs?attempt_id=1&as_team_id=13"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"
