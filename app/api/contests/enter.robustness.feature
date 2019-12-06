Feature: Enters a contest as a group (user self or team) (contestEnter) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name         | type                | team_item_id |
      | 10 | Team 1       | Team                | 50           |
      | 11 | Team 2       | Team                | 60           |
      | 21 | owner        | UserSelf            | null         |
      | 31 | john         | UserSelf            | null         |
      | 41 | jane         | UserSelf            | null         |
      | 51 | jack         | UserSelf            | null         |
      | 99 | item50-group | ContestParticipants | null         |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | john  | 31       | John        | Doe       |
      | jane  | 41       | Jane        | null      |
      | jack  | 51       | Jack        | Daniel    |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 10              | 31             |
      | 10              | 41             |
      | 10              | 51             |
      | 11              | 31             |
      | 11              | 41             |
      | 11              | 51             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 10                | 10             | 1       |
      | 10                | 31             | 0       |
      | 10                | 41             | 0       |
      | 10                | 51             | 0       |
      | 11                | 11             | 1       |
      | 11                | 31             | 0       |
      | 11                | 41             | 0       |
      | 11                | 51             | 0       |
      | 21                | 21             | 1       |
      | 31                | 31             | 1       |
      | 41                | 41             | 1       |
      | 51                | 51             | 1       |

  Scenario: Wrong item_id
    Given I am the user with id "31"
    When I send a POST request to "/contests/abc/groups/31"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should be empty

  Scenario: Wrong group_id
    Given I am the user with id "31"
    When I send a POST request to "/contests/50/groups/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should be empty

  Scenario: The item is not visible to the current user
    Given I am the user with id "31"
    When I send a POST request to "/contests/50/groups/21"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should be empty

  Scenario: The item is visible, but it doesn't exist
    Given I am the user with id "31"
    When I send a POST request to "/contests/50/groups/31"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should be empty

  Scenario: The item is not visible (can_view = none)
    Given the database has the following table 'items':
      | id | duration |
      | 50 | 00:00:01 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 31       | 50      | none               |
    And I am the user with id "31"
    When I send a POST request to "/contests/50/groups/31"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should be empty

  Scenario: The item is visible, but it's not a contest
    Given the database has the following table 'items':
      | id |
      | 50 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a POST request to "/contests/50/groups/31"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should be empty

  Scenario: group_id is not a self group of the current user while the item's has_attempts = false
    Given the database has the following table 'items':
      | id | duration | has_attempts |
      | 50 | 00:00:00 | false        |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a POST request to "/contests/50/groups/21"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should be empty

  Scenario: group_id is not a team related to the item while the item's has_attempts = true
    Given the database has the following table 'items':
      | id | duration | has_attempts |
      | 60 | 00:00:00 | true         |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a POST request to "/contests/60/groups/10"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should be empty

  Scenario: group_id is a user self group while the item's has_attempts = true
    Given the database has the following table 'items':
      | id | duration | has_attempts |
      | 60 | 00:00:00 | true         |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
    And I am the user with id "31"
    When I send a POST request to "/contests/60/groups/31"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should be empty

  Scenario: The current user is not a member of group_id while the item's has_attempts = true
    Given the database has the following table 'items':
      | id | duration | has_attempts |
      | 60 | 00:00:00 | true         |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "21"
    When I send a POST request to "/contests/60/groups/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should be empty

  Scenario: The contest is not ready
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size |
      | 60 | 00:00:00 | 1            | All                        | 3                     |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     |
      | 11       | 60      | 9999-01-01 10:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 51       | 60      | 2007-01-01 10:21 | 2008-12-31 23:59:59 |
    And I am the user with id "31"
    When I send a POST request to "/contests/60/groups/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should be empty

  Scenario Outline: Reenter a non-team contest
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_participants_group_id |
      | 50 | 01:01:01 | 0            | None                       | 99                            |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id | expires_at   |
      | 99              | 31             | <expires_at> |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     | additional_time |
      | 31       | 50      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 02:02:02        |
    And the database has the following table 'groups_attempts':
      | group_id | item_id | entered_at          | order |
      | 31       | 50      | 2019-05-29 11:00:00 | 1     |
    And I am the user with id "31"
    When I send a POST request to "/contests/50/groups/31"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should stay unchanged
  Examples:
    | expires_at          |
    | 2019-05-30 11:00:00 |
    | 9999-12-31 23:59:59 |

  Scenario: Reenter an already entered (not expired) contest as a team
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size | contest_participants_group_id |
      | 60 | 01:01:01 | 1            | None                       | 10                    | 99                            |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id |
      | 99              | 11             |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | solution                 |
      | 31       | 60      | content_with_descendants |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     | additional_time |
      | 11       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 02:02:02        |
    And the database has the following table 'groups_attempts':
      | group_id | item_id | entered_at          | order |
      | 11       | 60      | 2019-05-29 11:00:00 | 1     |
    And I am the user with id "31"
    When I send a POST request to "/contests/60/groups/11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "groups_attempts" should stay unchanged
