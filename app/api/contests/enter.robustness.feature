Feature: Enters a contest as a group (user self or team) (contestEnter) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name         | type                |
      | 10 | Team 1       | Team                |
      | 11 | Team 2       | Team                |
      | 21 | owner        | User                |
      | 31 | john         | User                |
      | 41 | jane         | User                |
      | 51 | jack         | User                |
      | 99 | item50-group | ContestParticipants |
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
    And the groups ancestors are computed

  Scenario: Wrong item_id
    Given I am the user with id "31"
    When I send a POST request to "/contests/abc/enter"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should be empty

  Scenario: Wrong as_team_id
    Given I am the user with id "31"
    When I send a POST request to "/contests/50/enter?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should be empty

  Scenario: The item is not visible to the team
    Given I am the user with id "31"
    When I send a POST request to "/contests/50/enter?as_team_id=21"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should be empty

  Scenario: The item is visible, but it doesn't exist
    Given I am the user with id "31"
    When I send a POST request to "/contests/50/enter"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should be empty

  Scenario: The item is not visible to the user (can_view = none)
    Given the database has the following table 'items':
      | id | requires_explicit_entry | default_language_tag |
      | 50 | 1                       | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 31       | 50      | none               |
    And I am the user with id "31"
    When I send a POST request to "/contests/50/enter"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should be empty

  Scenario: The item is visible, but it's not a contest
    Given the database has the following table 'items':
      | id | default_language_tag |
      | 50 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a POST request to "/contests/50/enter"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should be empty

  Scenario: as_team_id is given while the item's entry_participant_type = User
    Given the database has the following table 'items':
      | id | requires_explicit_entry | entry_participant_type | default_language_tag |
      | 50 | 1                       | User                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a POST request to "/contests/50/enter?as_team_id=21"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should be empty

  Scenario: as_team_id is not a team related to the item while the item's entry_participant_type = Team
    Given the database has the following table 'items':
      | id | requires_explicit_entry | entry_participant_type | default_language_tag |
      | 60 | 1                       | Team                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a POST request to "/contests/60/enter?as_team_id=10"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should be empty

  Scenario: as_team_id is not given while the item's entry_participant_type = Team
    Given the database has the following table 'items':
      | id | requires_explicit_entry | entry_participant_type | default_language_tag |
      | 60 | 1                       | Team                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
    And I am the user with id "31"
    When I send a POST request to "/contests/60/enter"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should be empty

  Scenario: The current user is not a member of as_team_id while the item's entry_participant_type = Team
    Given the database has the following table 'items':
      | id | requires_explicit_entry | entry_participant_type | default_language_tag |
      | 60 | 1                       | Team                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "21"
    When I send a POST request to "/contests/60/enter?as_team_id=11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should be empty

  Scenario: The contest is not ready
    Given the database has the following table 'items':
      | id | requires_explicit_entry | entry_participant_type | contest_entering_condition | contest_max_team_size | default_language_tag |
      | 60 | 1                       | Team                   | All                        | 3                     | fr                   |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 11       | 60      | 11              | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 41       | 60      | 41              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
      | 51       | 60      | 51              | 2007-01-01 10:21:21 | 2008-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id |
      | 11       | 60      |
      | 41       | 60      |
      | 51       | 60      |
    And I am the user with id "31"
    When I send a POST request to "/contests/60/enter?as_team_id=11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should be empty

  Scenario Outline: Reenter a non-team contest
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | contest_entering_condition | contest_participants_group_id | default_language_tag |
      | 50 | 01:01:01 | 1                       | User                   | None                       | 99                            | fr                   |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id | expires_at   |
      | 99              | 31             | <expires_at> |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 31       | 50      | 31              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | additional_time |
      | 31       | 50      | 02:02:02        |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | creator_id | parent_attempt_id | root_item_id |
      | 1  | 31             | 2019-05-29 11:00:00 | 31         | 0                 | 50           |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          |
      | 1          | 31             | 50      | 2019-05-29 11:00:00 |
    And I am the user with id "31"
    When I send a POST request to "/contests/50/enter"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should stay unchanged
  Examples:
    | expires_at          |
    | 2019-05-30 11:00:00 |
    | 9999-12-31 23:59:59 |

  Scenario: Reenter an already entered (not expired) contest as a team
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | contest_entering_condition | contest_max_team_size | contest_participants_group_id | default_language_tag |
      | 60 | 01:01:01 | 1                       | Team                   | None                       | 10                    | 99                            | fr                   |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id |
      | 99              | 11             |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 11       | 60      | 11              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | solution                 |
      | 31       | 60      | content_with_descendants |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | additional_time |
      | 11       | 60      | 02:02:02        |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | creator_id | parent_attempt_id | root_item_id |
      | 1  | 11             | 2019-05-29 11:00:00 | 31         | 0                 | 60           |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          |
      | 1          | 11             | 60      | 2019-05-29 11:00:00 |
    And I am the user with id "31"
    When I send a POST request to "/contests/60/enter?as_team_id=11"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And the table "attempts" should stay unchanged
