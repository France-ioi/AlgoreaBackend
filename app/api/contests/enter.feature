Feature: Enters a contest as a group (user self or team) (contestEnter)
  Background:
    Given the database has the following table 'groups':
      | id | name         | type      | team_item_id |
      | 11 | Team 2       | Team      | 60           |
      | 21 | owner        | UserSelf  | null         |
      | 22 | owner-admin  | UserAdmin | null         |
      | 31 | john         | UserSelf  | null         |
      | 32 | john-admin   | UserAdmin | null         |
      | 41 | jane         | UserSelf  | null         |
      | 42 | jane-admin   | UserAdmin | null         |
      | 51 | jack         | UserSelf  | null         |
      | 52 | jack-admin   | UserAdmin | null         |
      | 98 | item60-group | Other     | null         |
      | 99 | item50-group | Other     | null         |
    And the database has the following table 'users':
      | login | group_id | owned_group_id | first_name  | last_name |
      | owner | 21       | 22             | Jean-Michel | Blanquer  |
      | john  | 31       | 32             | John        | Doe       |
      | jane  | 41       | 42             | Jane        | null      |
      | jack  | 51       | 52             | Jack        | Daniel    |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type               |
      | 11              | 31             | invitationAccepted |
      | 11              | 41             | requestAccepted    |
      | 11              | 51             | joinedByCode       |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 11                | 31             | 0       |
      | 11                | 41             | 0       |
      | 11                | 51             | 0       |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
      | 41                | 41             | 1       |
      | 42                | 42             | 1       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
      | 98                | 98             | 1       |
      | 99                | 99             | 1       |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_partial_access_since | cached_grayed_access_since | cached_full_access_since | cached_solutions_access_since |
      | 11       | 50      | null                        | null                       | null                     | null                          |
      | 11       | 60      | null                        | 2017-05-29 06:38:38        | null                     | null                          |
      | 21       | 50      | null                        | null                       | null                     | 2018-05-29 06:38:38           |
      | 21       | 60      | null                        | null                       | 2018-05-29 06:38:38      | null                          |
      | 31       | 50      | null                        | null                       | 2018-05-29 06:38:38      | null                          |
    And the DB time now is "3019-10-10 10:10:10"

  Scenario: Enter an individual contest
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_participants_group_id |
      | 50 | 01:01:01 | 0            | None                       | 99                            |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     | additional_time |
      | 11       | 50      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 02:02:02        |
    And I am the user with group_id "31"
    When I send a POST request to "/contests/50/groups/31"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "message": "created",
      "success": true,
      "data": {
        "duration": "01:01:01",
        "entered_at": "3019-10-10T10:10:10Z"
      }
    }
    """
    And the table "contest_participations" should be:
      | group_id | item_id | entered_at          |
      | 31       | 50      | 3019-10-10 10:10:10 |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | type               | expires_at          |
      | 11              | 31             | invitationAccepted | 9999-12-31 23:59:59 |
      | 11              | 41             | requestAccepted    | 9999-12-31 23:59:59 |
      | 11              | 51             | joinedByCode       | 9999-12-31 23:59:59 |
      | 99              | 31             | direct             | 3019-10-10 13:13:13 |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self | expires_at          |
      | 11                | 11             | 1       | 9999-12-31 23:59:59 |
      | 11                | 31             | 0       | 9999-12-31 23:59:59 |
      | 11                | 41             | 0       | 9999-12-31 23:59:59 |
      | 11                | 51             | 0       | 9999-12-31 23:59:59 |
      | 21                | 21             | 1       | 9999-12-31 23:59:59 |
      | 22                | 22             | 1       | 9999-12-31 23:59:59 |
      | 31                | 31             | 1       | 9999-12-31 23:59:59 |
      | 32                | 32             | 1       | 9999-12-31 23:59:59 |
      | 41                | 41             | 1       | 9999-12-31 23:59:59 |
      | 42                | 42             | 1       | 9999-12-31 23:59:59 |
      | 51                | 51             | 1       | 9999-12-31 23:59:59 |
      | 52                | 52             | 1       | 9999-12-31 23:59:59 |
      | 98                | 98             | 1       | 9999-12-31 23:59:59 |
      | 99                | 31             | 0       | 3019-10-10 13:13:13 |
      | 99                | 99             | 1       | 9999-12-31 23:59:59 |

  Scenario: Enter a team-only contest
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size | contest_participants_group_id |
      | 60 | 05:05:05 | 1            | Half                       | 3                     | 98                            |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     | additional_time |
      | 11       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 01:01:01        |
      | 31       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 02:02:02        |
      | 41       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 03:03:03        |
    And I am the user with group_id "31"
    When I send a POST request to "/contests/60/groups/11"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "message": "created",
      "success": true,
      "data": {
        "duration": "05:05:05",
        "entered_at": "3019-10-10T10:10:10Z"
      }
    }
    """
    And the table "contest_participations" should be:
      | group_id | item_id | entered_at          |
      | 11       | 60      | 3019-10-10 10:10:10 |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | type               | expires_at          |
      | 11              | 31             | invitationAccepted | 9999-12-31 23:59:59 |
      | 11              | 41             | requestAccepted    | 9999-12-31 23:59:59 |
      | 11              | 51             | joinedByCode       | 9999-12-31 23:59:59 |
      | 98              | 11             | direct             | 3019-10-10 16:16:16 |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self | expires_at          |
      | 11                | 11             | 1       | 9999-12-31 23:59:59 |
      | 11                | 31             | 0       | 9999-12-31 23:59:59 |
      | 11                | 41             | 0       | 9999-12-31 23:59:59 |
      | 11                | 51             | 0       | 9999-12-31 23:59:59 |
      | 21                | 21             | 1       | 9999-12-31 23:59:59 |
      | 22                | 22             | 1       | 9999-12-31 23:59:59 |
      | 31                | 31             | 1       | 9999-12-31 23:59:59 |
      | 32                | 32             | 1       | 9999-12-31 23:59:59 |
      | 41                | 41             | 1       | 9999-12-31 23:59:59 |
      | 42                | 42             | 1       | 9999-12-31 23:59:59 |
      | 51                | 51             | 1       | 9999-12-31 23:59:59 |
      | 52                | 52             | 1       | 9999-12-31 23:59:59 |
      | 98                | 11             | 0       | 3019-10-10 16:16:16 |
      | 98                | 31             | 0       | 3019-10-10 16:16:16 |
      | 98                | 41             | 0       | 3019-10-10 16:16:16 |
      | 98                | 51             | 0       | 3019-10-10 16:16:16 |
      | 98                | 98             | 1       | 9999-12-31 23:59:59 |
      | 99                | 99             | 1       | 9999-12-31 23:59:59 |

  Scenario: Enter a contest that don't have items.contest_participants_group_id set
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition |
      | 50 | 01:01:01 | 0            | None                       |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     | additional_time |
      | 11       | 50      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 02:02:02        |
    And I am the user with group_id "31"
    When I send a POST request to "/contests/50/groups/31"
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "message": "created",
      "success": true,
      "data": {
        "duration": "01:01:01",
        "entered_at": "3019-10-10T10:10:10Z"
      }
    }
    """
    And the table "contest_participations" should be:
      | group_id | item_id | entered_at          |
      | 31       | 50      | 3019-10-10 10:10:10 |
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And logs should contain:
      """
      items.contest_participants_group_id is not set for the item with id = 50
      """
