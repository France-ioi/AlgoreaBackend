Feature: Get qualification state (contestGetQualificationState)
  Background:
    Given the database has the following table 'groups':
      | id | name   | type | team_item_id |
      | 10 | Team 1 | Team | 50           |
      | 11 | Team 2 | Team | 60           |
      | 21 | owner  | User | null         |
      | 31 | john   | User | null         |
      | 41 | jane   | User | null         |
      | 51 | jack   | User | null         |
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
      | ancestor_group_id | child_group_id |
      | 10                | 10             |
      | 10                | 31             |
      | 10                | 41             |
      | 10                | 51             |
      | 11                | 11             |
      | 11                | 31             |
      | 11                | 41             |
      | 11                | 51             |
      | 21                | 21             |
      | 31                | 31             |
      | 41                | 41             |
      | 51                | 51             |

  Scenario Outline: Individual contest without can_enter_from & can_enter_until
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type   | contest_entering_condition | default_language_tag |
      | 50 | 00:00:00 | 1                       | <entry_participant_type> | <entering_condition>       | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/contests/50/qualification-state"
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
    | entry_participant_type | entering_condition | expected_state |
    | User                   | None               | ready          |
    | null                   | All                | not_ready      |
    | User                   | Half               | not_ready      |
    | null                   | One                | not_ready      |

  Scenario Outline: State is ready for an individual contest
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type   | contest_entering_condition | default_language_tag |
      | 50 | 00:00:00 | 1                       | <entry_participant_type> | <entering_condition>       | fr                   |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 31       | 50      | 31              | 1000-01-01 00:00:00 | 9999-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/contests/50/qualification-state"
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
      | entry_participant_type | entering_condition |
      | null                   | None               |
      | User                   | All                |
      | null                   | Half               |
      | User                   | One                |

  Scenario Outline: Team-only contest when no one can enter
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | contest_entering_condition | contest_max_team_size | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entering_condition>       | 3                     | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/contests/60/qualification-state?as_team_id=11"
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
      | id | duration | requires_explicit_entry | entry_participant_type | contest_entering_condition | contest_max_team_size | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entering_condition>       | 3                     | fr                   |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 11       | 60      | 11              | 9999-01-01 10:21:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 41              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
      | 51       | 60      | 51              | 2007-01-01 10:21:21 | 2008-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/contests/60/qualification-state?as_team_id=11"
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
      | id | duration | requires_explicit_entry | entry_participant_type | contest_entering_condition | contest_max_team_size | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entering_condition>       | 3                     | fr                   |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 31       | 60      | 31              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 41              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/contests/60/qualification-state?as_team_id=11"
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
      | id | duration | requires_explicit_entry | entry_participant_type | contest_entering_condition | contest_max_team_size | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entering_condition>       | 3                     | fr                   |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 31       | 60      | 31              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 41              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
      | 51       | 60      | 51              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/contests/60/qualification-state?as_team_id=11"
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
      | id | duration | requires_explicit_entry | entry_participant_type | contest_entering_condition | contest_max_team_size | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entering_condition>       | 2                     | fr                   |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 31       | 60      | 31              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 41              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
      | 51       | 60      | 51              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/contests/60/qualification-state?as_team_id=11"
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
    Given the database table 'groups' has also the following row:
      | id  | type                |
      | 100 | ContestParticipants |
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | contest_entering_condition | contest_max_team_size | contest_participants_group_id | default_language_tag |
      | 50 | 00:00:00 | 1                       | User                   | <entering_condition>       | 0                     | 100                           | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 10       | 50      | content                  |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And the database has the following table 'attempts':
      | group_id | item_id | started_at          | order |
      | 31       | 50      | 2019-05-30 15:00:00 | 1     |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id |
      | 100             | 31             |
    And I am the user with id "31"
    When I send a GET request to "/contests/50/qualification-state"
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
    Given the database table 'groups' has also the following row:
      | id  | type                |
      | 100 | ContestParticipants |
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | contest_entering_condition | contest_max_team_size | contest_participants_group_id | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entering_condition>       | 0                     | 100                           | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And the database has the following table 'attempts':
      | group_id | item_id | started_at          | order |
      | 11       | 60      | 2019-05-30 15:00:00 | 1     |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id |
      | 100             | 11             |
    And I am the user with id "31"
    When I send a GET request to "/contests/60/qualification-state?as_team_id=11"
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

  Scenario: State is not_ready for an individual contest because the participation has expired
    Given the database table 'groups' has also the following row:
      | id  | type                |
      | 100 | ContestParticipants |
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | contest_entering_condition | contest_max_team_size | contest_participants_group_id | default_language_tag |
      | 50 | 00:00:00 | 1                       | User                   | None                       | 0                     | 100                           | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 10       | 50      | content                  |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And the database has the following table 'attempts':
      | group_id | item_id | started_at          | order |
      | 31       | 50      | 2019-05-30 15:00:00 | 1     |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id | expires_at          |
      | 100             | 31             | 2019-05-30 20:00:00 |
    And I am the user with id "31"
    When I send a GET request to "/contests/50/qualification-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entering_condition": "None",
      "other_members": [],
      "state": "not_ready"
    }
    """

  Scenario: State is ready for a team-only contest because the participation has expired
    Given the database table 'groups' has also the following row:
      | id  | type                |
      | 100 | ContestParticipants |
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | allows_multiple_attempts | entry_participant_type | contest_entering_condition | contest_max_team_size | contest_participants_group_id | default_language_tag |
      | 60 | 00:00:00 | 1                       | 1                        | Team                   | None                       | 3                     | 100                           | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And the database has the following table 'attempts':
      | group_id | item_id | started_at          | order |
      | 11       | 60      | 2019-05-30 15:00:00 | 1     |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id | expires_at          |
      | 100             | 11             | 2019-05-30 20:00:00 |
    And I am the user with id "31"
    When I send a GET request to "/contests/60/qualification-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entering_condition": "None",
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
      "state": "ready"
    }
    """

  Scenario Outline: The user cannot enter because of entering_time_min/entering_time_max
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | contest_entering_condition | default_language_tag | entering_time_min   | entering_time_max   |
      | 50 | 00:00:00 | 1                       | User                   | None                       | fr                   | <entering_time_min> | <entering_time_max> |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 11       | 50      | 11              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 10       | 50      | content                  |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/contests/50/qualification-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entering_condition": "None",
      "other_members": [],
      "state": "ready"
    }
    """
  Examples:
    | entering_time_min   | entering_time_max   |
    | 5099-12-31 23:59:59 | 9999-12-31 23:59:59 |
    | 2007-12-31 23:59:59 | 2008-12-31 23:59:59 |
