Feature: Get breadcrumbs (groupBreadcrumbsView) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name                | type                | is_public |
      | 1  | Team                | Team                | true      |
      | 2  | Managed By Team     | Class               | false     |
      | 41 | user                | User                | true      |
      | 42 | ContestParticipants | ContestParticipants | false     |
      | 43 | Another Team        | Team                | false     |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 41       | Jean-Michel | Blanquer  |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 2        | 1          |
      | 42       | 41         |
      | 43       | 41         |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | expires_at          |
      | 1               | 41             | 9999-12-31 23:59:59 |
      | 42              | 41             | 9999-12-31 23:59:59 |
      | 42              | 43             | 9999-12-31 23:59:59 |
      | 43              | 41             | 9999-12-31 23:59:59 |
    And the groups ancestors are computed

  Scenario: Invalid id given
    Given I am the user with id "41"
    When I send a GET request to "/groups/10/11111111111111111111111111111/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: '11111111111111111111111111111', param: 'ids')"

  Scenario: Too many ids given
    Given I am the user with id "41"
    When I send a GET request to "/groups/1/2/3/4/5/6/7/8/9/10/11/breadcrumbs"
    Then the response code should be 400
    And the response error message should contain "No more than 10 ids expected"

  Scenario: Requires all the groups to be public/managed/descendant of managed/joined
    Given I am the user with id "41"
    When I send a GET request to "/groups/1/2/breadcrumbs"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Requires all the groups to exist
    Given I am the user with id "41"
    When I send a GET request to "/groups/1/404/breadcrumbs"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Doesn't allow duplicates
    Given I am the user with id "41"
    When I send a GET request to "/groups/1/1/breadcrumbs"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Doesn't allow ContestParticipants groups
    Given I am the user with id "41"
    When I send a GET request to "/groups/42/41/breadcrumbs"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Doesn't allow ContestParticipants groups via team
    Given I am the user with id "41"
    When I send a GET request to "/groups/42/43/41/breadcrumbs"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
