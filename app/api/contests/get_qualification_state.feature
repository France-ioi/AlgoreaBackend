Feature: Get qualification state (contestGetQualificationState)
  Background:
    Given the database has the following table 'groups':
      | id | name   | type     | team_item_id |
      | 10 | Team 1 | Team     | 50           |
      | 11 | Team 2 | Team     | 60           |
      | 21 | owner  | UserSelf | null         |
      | 31 | john   | UserSelf | null         |
      | 41 | jane   | UserSelf | null         |
      | 51 | jack   | UserSelf | null         |
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

  Scenario Outline: Individual contest without can_enter_from & can_enter_until
    Given the database has the following table 'items':
      | id | duration | has_attempts | contest_entering_condition |
      | 50 | 00:00:00 | 0            | <entering_condition>       |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
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
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 31       | 50      | content_with_descendants |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from      | can_enter_until     |
      | 31       | 50      | 1000-01-01 00:00:00 | 9999-12-31 23:59:59 |
    And I am the user with id "31"
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
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
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
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     |
      | 31       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
    And I am the user with id "31"
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
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     |
      | 31       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 51       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
    And I am the user with id "31"
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
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    Given the database has the following table 'groups_contest_items':
      | group_id | item_id | can_enter_from   | can_enter_until     |
      | 31       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
      | 51       | 60      | 2007-01-01 10:21 | 9999-12-31 23:59:59 |
    And I am the user with id "31"
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
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 10       | 50      | content                  |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And the database has the following table 'groups_attempts':
      | group_id | item_id | entered_at          | order |
      | 31       | 50      | 2019-05-30 15:00:00 | 1     |
    And I am the user with id "31"
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
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And the database has the following table 'groups_attempts':
      | group_id | item_id | entered_at          | order |
      | 11       | 60      | 2019-05-30 15:00:00 | 1     |
    And I am the user with id "31"
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
