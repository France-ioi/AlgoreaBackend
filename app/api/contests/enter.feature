Feature: Enters a contest as a group (user self or team) (contestEnter)
  Background:
    Given the database has the following table 'groups':
      | id | name         | type                | team_item_id |
      | 11 | Team 2       | Team                | 60           |
      | 21 | owner        | UserSelf            | null         |
      | 31 | john         | UserSelf            | null         |
      | 41 | jane         | UserSelf            | null         |
      | 51 | jack         | UserSelf            | null         |
      | 98 | item60-group | ContestParticipants | null         |
      | 99 | item50-group | ContestParticipants | null         |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | john  | 31       | John        | Doe       |
      | jane  | 41       | Jane        | null      |
      | jack  | 51       | Jack        | Daniel    |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 11              | 31             |
      | 11              | 41             |
      | 11              | 51             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 11                | 31             | 0       |
      | 11                | 41             | 0       |
      | 11                | 51             | 0       |
      | 21                | 21             | 1       |
      | 31                | 31             | 1       |
      | 41                | 41             | 1       |
      | 51                | 51             | 1       |
      | 98                | 98             | 1       |
      | 99                | 99             | 1       |
    And the DB time now is "3019-10-10 10:10:10"

  Scenario: Enter an individual contest
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_participants_group_id | default_language_tag |
      | 50 | 01:01:01 | 0            | None                       | 99                            | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     | additional_time |
      | 11       | 50      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 02:02:02        |
    And I am the user with id "31"
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
    And the table "attempts" should be:
      | group_id | item_id | started_at          | order |
      | 31       | 50      | 3019-10-10 10:10:10 | 1     |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | expires_at          |
      | 11              | 31             | 9999-12-31 23:59:59 |
      | 11              | 41             | 9999-12-31 23:59:59 |
      | 11              | 51             | 9999-12-31 23:59:59 |
      | 99              | 31             | 3019-10-10 13:13:13 |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self | expires_at          |
      | 11                | 11             | 1       | 9999-12-31 23:59:59 |
      | 11                | 31             | 0       | 9999-12-31 23:59:59 |
      | 11                | 41             | 0       | 9999-12-31 23:59:59 |
      | 11                | 51             | 0       | 9999-12-31 23:59:59 |
      | 21                | 21             | 1       | 9999-12-31 23:59:59 |
      | 31                | 31             | 1       | 9999-12-31 23:59:59 |
      | 41                | 41             | 1       | 9999-12-31 23:59:59 |
      | 51                | 51             | 1       | 9999-12-31 23:59:59 |
      | 98                | 98             | 1       | 9999-12-31 23:59:59 |
      | 99                | 31             | 0       | 3019-10-10 13:13:13 |
      | 99                | 99             | 1       | 9999-12-31 23:59:59 |

  Scenario: Enter a team-only contest
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size | contest_participants_group_id | default_language_tag |
      | 60 | 05:05:05 | 1            | Half                       | 3                     | 98                            | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | content                  |
      | 21       | 60      | content_with_descendants |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     | additional_time |
      | 11       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 01:01:01        |
      | 31       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 02:02:02        |
      | 41       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 03:03:03        |
    And I am the user with id "31"
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
    And the table "attempts" should be:
      | group_id | item_id | started_at          | order |
      | 11       | 60      | 3019-10-10 10:10:10 | 1     |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | expires_at          |
      | 11              | 31             | 9999-12-31 23:59:59 |
      | 11              | 41             | 9999-12-31 23:59:59 |
      | 11              | 51             | 9999-12-31 23:59:59 |
      | 98              | 11             | 3019-10-10 16:16:16 |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self | expires_at          |
      | 11                | 11             | 1       | 9999-12-31 23:59:59 |
      | 11                | 31             | 0       | 9999-12-31 23:59:59 |
      | 11                | 41             | 0       | 9999-12-31 23:59:59 |
      | 11                | 51             | 0       | 9999-12-31 23:59:59 |
      | 21                | 21             | 1       | 9999-12-31 23:59:59 |
      | 31                | 31             | 1       | 9999-12-31 23:59:59 |
      | 41                | 41             | 1       | 9999-12-31 23:59:59 |
      | 51                | 51             | 1       | 9999-12-31 23:59:59 |
      | 98                | 11             | 0       | 3019-10-10 16:16:16 |
      | 98                | 31             | 0       | 3019-10-10 16:16:16 |
      | 98                | 41             | 0       | 3019-10-10 16:16:16 |
      | 98                | 51             | 0       | 3019-10-10 16:16:16 |
      | 98                | 98             | 1       | 9999-12-31 23:59:59 |
      | 99                | 99             | 1       | 9999-12-31 23:59:59 |

  Scenario: Reenter a contest as a team
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size | contest_participants_group_id | default_language_tag |
      | 60 | 01:01:01 | 1            | None                       | 10                    | 99                            | fr                   |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id | expires_at          |
      | 99              | 11             | 2019-05-30 11:00:00 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | solution                 |
      | 31       | 60      | content_with_descendants |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     | additional_time |
      | 11       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 02:02:02        |
    And the database has the following table 'attempts':
      | group_id | item_id | started_at          | order |
      | 11       | 60      | 2019-05-29 11:00:00 | 1     |
    And I am the user with id "31"
    When I send a POST request to "/contests/60/groups/11"
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
    And the table "attempts" should be:
      | group_id | item_id | started_at          | order |
      | 11       | 60      | 2019-05-29 11:00:00 | 1     |
      | 11       | 60      | 3019-10-10 10:10:10 | 2     |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | expires_at          |
      | 11              | 31             | 9999-12-31 23:59:59 |
      | 11              | 41             | 9999-12-31 23:59:59 |
      | 11              | 51             | 9999-12-31 23:59:59 |
      | 99              | 11             | 3019-10-10 13:13:13 |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self | expires_at          |
      | 11                | 11             | 1       | 9999-12-31 23:59:59 |
      | 11                | 31             | 0       | 9999-12-31 23:59:59 |
      | 11                | 41             | 0       | 9999-12-31 23:59:59 |
      | 11                | 51             | 0       | 9999-12-31 23:59:59 |
      | 21                | 21             | 1       | 9999-12-31 23:59:59 |
      | 31                | 31             | 1       | 9999-12-31 23:59:59 |
      | 41                | 41             | 1       | 9999-12-31 23:59:59 |
      | 51                | 51             | 1       | 9999-12-31 23:59:59 |
      | 98                | 98             | 1       | 9999-12-31 23:59:59 |
      | 99                | 11             | 0       | 3019-10-10 13:13:13 |
      | 99                | 31             | 0       | 3019-10-10 13:13:13 |
      | 99                | 41             | 0       | 3019-10-10 13:13:13 |
      | 99                | 51             | 0       | 3019-10-10 13:13:13 |
      | 99                | 99             | 1       | 9999-12-31 23:59:59 |

  Scenario: Enter a contest that don't have items.contest_participants_group_id set
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | default_language_tag |
      | 50 | 01:01:01 | 0            | None                       | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     | additional_time |
      | 11       | 50      | 2007-01-01 10:21 | 9999-12-31 23:59:59 | 02:02:02        |
    And I am the user with id "31"
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
    And the table "attempts" should be:
      | group_id | item_id | started_at          | order |
      | 31       | 50      | 3019-10-10 10:10:10 | 1     |
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
    And logs should contain:
      """
      items.contest_participants_group_id is not set for the item with id = 50
      """
