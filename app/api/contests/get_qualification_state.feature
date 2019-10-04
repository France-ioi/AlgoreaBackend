Feature: Get qualification state (contestGetQualificationState)
  Background:
    Given the database has the following table 'users':
      | id | login | self_group_id | owned_group_id | first_name  | last_name |
      | 1  | owner | 21            | 22             | Jean-Michel | Blanquer  |
      | 2  | john  | 31            | 32             | John        | Doe       |
      | 3  | jane  | 41            | 42             | Jane        | null      |
      | 4  | jack  | 51            | 52             | Jack        | Daniel    |
    And the database has the following table 'groups':
      | id | name        | type      | team_item_id |
      | 10 | Team 1      | Team      | 50           |
      | 11 | Team 2      | Team      | 60           |
      | 21 | owner       | UserSelf  | null         |
      | 22 | owner-admin | UserAdmin | null         |
      | 31 | john        | UserSelf  | null         |
      | 32 | john-admin  | UserAdmin | null         |
      | 41 | jane        | UserSelf  | null         |
      | 42 | jane-admin  | UserAdmin | null         |
      | 51 | jack        | UserSelf  | null         |
      | 52 | jack-admin  | UserAdmin | null         |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type               |
      | 10              | 31             | invitationAccepted |
      | 10              | 41             | requestAccepted    |
      | 10              | 51             | joinedByCode       |
      | 11              | 31             | invitationAccepted |
      | 11              | 41             | requestAccepted    |
      | 11              | 51             | joinedByCode       |
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
      | 22                | 22             | 1       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
      | 41                | 41             | 1       |
      | 42                | 42             | 1       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
    And the database has the following table 'groups_items':
      | group_id | item_id | cached_partial_access_since | cached_grayed_access_since | cached_full_access_since | cached_solutions_access_since | creator_user_id |
      | 10       | 50      | 2017-05-29 06:38:38         | null                       | null                     | null                          | 1               |
      | 11       | 50      | null                        | null                       | null                     | null                          | 1               |
      | 11       | 60      | null                        | 2017-05-29 06:38:38        | null                     | null                          | 1               |
      | 21       | 50      | null                        | null                       | null                     | 2018-05-29 06:38:38           | 1               |
      | 21       | 60      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 1               |
      | 31       | 50      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 1               |

  Scenario Outline: Individual contest without can_enter_from & can_enter_until
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition |
      | 50 | 00:00:00 | 0            | <entering_condition>       |
    And I am the user with id "2"
    When I send a GET request to "/contests/50/groups/31/qualification-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entering_condition": {{"<entering_condition>" != "null" ? "\"<entering_condition>\"" : "null"}},
      "other_members": [],
      "state": "<expected_state>"
    }
    """
  Examples:
    | entering_condition | expected_state |
    | None               | ready          |
    | All                | not_ready      |
    | Half               | not_ready      |
    | One                | not_ready      |

  Scenario Outline: State is ready for an individual contest
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition |
      | 50 | 00:00:00 | 0            | <entering_condition>       |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from      | can_enter_until     |
      | 31       | 50      | 1000-01-01 00:00:00 | 9999-12-31 23:59:59 |
    And I am the user with id "2"
    When I send a GET request to "/contests/50/groups/31/qualification-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": true,
      "entering_condition": {{"<entering_condition>" != "null" ? "\"<entering_condition>\"" : "null"}},
      "other_members": [],
      "state": "ready"
    }
    """
    Examples:
      | entering_condition |
      | None               |
      | All                |
      | Half               |
      | One                |

  Scenario Outline: Team-only contest when no one can enter
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size |
      | 60 | 00:00:00 | 1            | <entering_condition>       | 3                     |
    And I am the user with id "2"
    When I send a GET request to "/contests/60/groups/11/qualification-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entering_condition": {{"<entering_condition>" != "null" ? "\"<entering_condition>\"" : "null"}},
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "state": "<expected_state>"
    }
    """
    Examples:
      | entering_condition | expected_state |
      | None               | ready          |
      | All                | not_ready      |
      | Half               | not_ready      |
      | One                | not_ready      |

  Scenario Outline: Team-only contest when one member can enter
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size |
      | 60 | 00:00:00 | 1            | <entering_condition>       | 3                     |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     |
      | 11       | 60      | 9999-01-01 10:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 51       | 60      | 2007-01-01 10:21 | 2008-12-31 23:59:59 |
    And I am the user with id "2"
    When I send a GET request to "/contests/60/groups/11/qualification-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entering_condition": {{"<entering_condition>" != "null" ? "\"<entering_condition>\"" : "null"}},
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": true,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "state": "<expected_state>"
    }
    """
    Examples:
      | entering_condition | expected_state |
      | None               | ready          |
      | All                | not_ready      |
      | Half               | not_ready      |
      | One                | ready          |

  Scenario Outline: Team-only contest when half of members can enter
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size |
      | 60 | 00:00:00 | 1            | <entering_condition>       | 3                     |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     |
      | 31       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
    And I am the user with id "2"
    When I send a GET request to "/contests/60/groups/11/qualification-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": true,
      "entering_condition": {{"<entering_condition>" != "null" ? "\"<entering_condition>\"" : "null"}},
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": true,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "state": "<expected_state>"
    }
    """
    Examples:
      | entering_condition | expected_state |
      | None               | ready          |
      | All                | not_ready      |
      | Half               | ready          |
      | One                | ready          |

  Scenario Outline: Team-only contest when all members can enter
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size |
      | 60 | 00:00:00 | 1            | <entering_condition>       | 3                     |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     |
      | 31       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 51       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
    And I am the user with id "2"
    When I send a GET request to "/contests/60/groups/11/qualification-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": true,
      "entering_condition": {{"<entering_condition>" != "null" ? "\"<entering_condition>\"" : "null"}},
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": true,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": true,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "state": "ready"
    }
    """
    Examples:
      | entering_condition |
      | None               |
      | All                |
      | Half               |
      | One                |

  Scenario Outline: Team-only contest when all members can enter, but the team is too large
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size |
      | 60 | 00:00:00 | 1            | <entering_condition>       | 2                     |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     |
      | 31       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 51       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
    And I am the user with id "2"
    When I send a GET request to "/contests/60/groups/11/qualification-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": true,
      "entering_condition": {{"<entering_condition>" != "null" ? "\"<entering_condition>\"" : "null"}},
      "max_team_size": 2,
      "other_members": [
        {
          "can_enter": true,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": true,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "state": "not_ready"
    }
    """
    Examples:
      | entering_condition |
      | None               |
      | All                |
      | Half               |
      | One                |

  Scenario Outline: State is already_started for an individual contest
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size |
      | 50 | 00:00:00 | 0            | <entering_condition>       | 0                     |
    And the database has the following table 'contest_participations':
      | group_id | item_id | entered_at          |
      | 31       | 50      | 2019-05-30 15:00:00 |
    And I am the user with id "2"
    When I send a GET request to "/contests/50/groups/31/qualification-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entering_condition": {{"<entering_condition>" != "null" ? "\"<entering_condition>\"" : "null"}},
      "other_members": [],
      "state": "already_started"
    }
    """
    Examples:
      | entering_condition |
      | None               |
      | All                |
      | Half               |
      | One                |

  Scenario Outline: State is already_started for a team-only contest
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition | contest_max_team_size |
      | 60 | 00:00:00 | 1            | <entering_condition>       | 0                     |
    And the database has the following table 'contest_participations':
      | group_id | item_id | entered_at          |
      | 11       | 60      | 2019-05-30 15:00:00 |
    And I am the user with id "2"
    When I send a GET request to "/contests/60/groups/11/qualification-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entering_condition": {{"<entering_condition>" != "null" ? "\"<entering_condition>\"" : "null"}},
      "max_team_size": 0,
      "other_members": [
        {
          "can_enter": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "state": "already_started"
    }
    """
    Examples:
      | entering_condition |
      | None               |
      | All                |
      | Half               |
      | One                |
