Feature: Get entry state (itemGetEntryState)
  Background:
    Given the database has the following table 'groups':
      | id | name   | type  | frozen_membership |
      | 10 | Team 1 | Team  | 1                 |
      | 11 | Team 2 | Team  | 0                 |
      | 12 | Team 3 | Team  | 0                 |
      | 13 | Club   | Club  | 0                 |
      | 14 | School | Other | 0                 |
      | 21 | owner  | User  | 0                 |
      | 31 | john   | User  | 0                 |
      | 41 | jane   | User  | 0                 |
      | 51 | jack   | User  | 0                 |
      | 61 | mark   | User  | 0                 |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | john  | 31       | John        | Doe       |
      | jane  | 41       | Jane        | null      |
      | jack  | 51       | Jack        | Daniel    |
      | mark  | 61       | Mark        | Moe       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 10              | 31             | null                           |
      | 10              | 41             | null                           |
      | 10              | 51             | null                           |
      | 11              | 31             | null                           |
      | 11              | 41             | null                           |
      | 11              | 51             | null                           |
      | 12              | 31             | null                           |
      | 12              | 61             | null                           |
      | 13              | 41             | 2019-05-30 11:00:00            |
      | 13              | 51             | 2019-05-30 11:00:00            |
      | 13              | 61             | null                           |
      | 14              | 31             | null                           |
      | 14              | 41             | null                           |
      | 14              | 61             | 2019-05-30 11:00:00            |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 14         |

  Scenario Outline: Individual contest without can_enter_from & can_enter_until
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio   | default_language_tag |
      | 50 | 00:00:00 | 1                       | User                   | <entry_min_admitted_members_ratio> | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/50/entry-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": {{"<entry_min_admitted_members_ratio>" != "null" ? "\"<entry_min_admitted_members_ratio>\"" : "null"}},
      "other_members": [],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "<expected_state>"
    }
    """
  Examples:
    | entry_min_admitted_members_ratio | expected_state |
    | None                             | ready          |
    | All                              | not_ready      |
    | Half                             | not_ready      |
    | One                              | not_ready      |

  Scenario Outline: State is ready for an individual contest
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio   | default_language_tag |
      | 50 | 00:00:00 | 1                       | User                   | <entry_min_admitted_members_ratio> | fr                   |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 31       | 50      | 31              | 1000-01-01 00:00:00 | 9999-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/50/entry-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": true,
      "entry_min_admitted_members_ratio": {{"<entry_min_admitted_members_ratio>" != "null" ? "\"<entry_min_admitted_members_ratio>\"" : "null"}},
      "other_members": [],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "ready"
    }
    """
    Examples:
      | entry_min_admitted_members_ratio |
      | None                             |
      | All                              |
      | Half                             |
      | One                              |

  Scenario Outline: Team-only contest when no one can enter
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio   | entry_max_team_size | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entry_min_admitted_members_ratio> | 3                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": {{"<entry_min_admitted_members_ratio>" != "null" ? "\"<entry_min_admitted_members_ratio>\"" : "null"}},
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "<expected_state>"
    }
    """
    Examples:
      | entry_min_admitted_members_ratio | expected_state |
      | None                             | ready          |
      | All                              | not_ready      |
      | Half                             | not_ready      |
      | One                              | not_ready      |

  Scenario Outline: entry_min_admitted_members_ratio is ignored when the team itself can enter the contest (depending on entering_time_* & can_enter_*)
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio | entry_max_team_size | default_language_tag | entering_time_min   | entering_time_max   |
      | 60 | 00:00:00 | 1                       | Team                   | All                              | 3                   | fr                   | <entering_time_min> | <entering_time_max> |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 11       | 60      | 31              | <can_enter_from>    | <can_enter_until>   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": <users_can_enter>,
      "entry_min_admitted_members_ratio": "All",
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": <users_can_enter>,
          "attempts_restriction_violated": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": <users_can_enter>,
          "attempts_restriction_violated": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "<expected_state>"
    }
    """
  Examples:
    | can_enter_from      | can_enter_until     | entering_time_min   | entering_time_max   | expected_state | users_can_enter |
    | 2007-01-01 10:21:21 | 3008-01-01 10:21:21 | 2007-12-31 23:59:59 | 3008-12-31 23:59:59 | ready          | true            |
    | 3007-01-01 10:21:21 | 3008-01-01 10:21:21 | 2007-12-31 23:59:59 | 3008-12-31 23:59:59 | not_ready      | false           |
    | 2007-01-01 10:21:21 | 2008-01-01 10:21:21 | 2007-12-31 23:59:59 | 3008-12-31 23:59:59 | not_ready      | false           |
    | 2007-01-01 10:21:21 | 3008-01-01 10:21:21 | 5099-12-31 23:59:59 | 9999-12-31 23:59:59 | not_ready      | false           |
    | 2007-01-01 10:21:21 | 3008-01-01 10:21:21 | 2007-12-31 23:59:59 | 2008-12-31 23:59:59 | not_ready      | false           |

  Scenario Outline: Team-only contest when one member can enter
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio   | entry_max_team_size | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entry_min_admitted_members_ratio> | 3                   | fr                   |
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
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": {{"<entry_min_admitted_members_ratio>" != "null" ? "\"<entry_min_admitted_members_ratio>\"" : "null"}},
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": true,
          "attempts_restriction_violated": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "<expected_state>"
    }
    """
    Examples:
      | entry_min_admitted_members_ratio | expected_state |
      | None                             | ready          |
      | All                              | not_ready      |
      | Half                             | not_ready      |
      | One                              | ready          |

  Scenario Outline: Team-only contest when half of members can enter
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio   | entry_max_team_size | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entry_min_admitted_members_ratio> | 3                   | fr                   |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 31       | 60      | 31              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 41              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": true,
      "entry_min_admitted_members_ratio": {{"<entry_min_admitted_members_ratio>" != "null" ? "\"<entry_min_admitted_members_ratio>\"" : "null"}},
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": true,
          "attempts_restriction_violated": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "<expected_state>"
    }
    """
    Examples:
      | entry_min_admitted_members_ratio | expected_state |
      | None                             | ready          |
      | All                              | not_ready      |
      | Half                             | ready          |
      | One                              | ready          |

  Scenario Outline: Team-only contest when all members can enter
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio   | entry_max_team_size | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entry_min_admitted_members_ratio> | 3                   | fr                   |
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
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": true,
      "entry_min_admitted_members_ratio": {{"<entry_min_admitted_members_ratio>" != "null" ? "\"<entry_min_admitted_members_ratio>\"" : "null"}},
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": true,
          "attempts_restriction_violated": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": true,
          "attempts_restriction_violated": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "ready"
    }
    """
    Examples:
      | entry_min_admitted_members_ratio |
      | None                             |
      | All                              |
      | Half                             |
      | One                              |

  Scenario Outline: Team-only contest when all members can enter, but the team is too large
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio   | entry_max_team_size | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entry_min_admitted_members_ratio> | 2                   | fr                   |
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
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": true,
      "entry_min_admitted_members_ratio": {{"<entry_min_admitted_members_ratio>" != "null" ? "\"<entry_min_admitted_members_ratio>\"" : "null"}},
      "max_team_size": 2,
      "other_members": [
        {
          "can_enter": true,
          "attempts_restriction_violated": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": true,
          "attempts_restriction_violated": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "not_ready"
    }
    """
    Examples:
      | entry_min_admitted_members_ratio |
      | None                             |
      | All                              |
      | Half                             |
      | One                              |

  Scenario Outline: Team-only contest ignores size of the team when the team itself can enter the contest (depending on entering_time_* & can_enter_*)
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio | entry_max_team_size | default_language_tag | entering_time_min   | entering_time_max   |
      | 60 | 00:00:00 | 1                       | Team                   | None                             | 2                   | fr                   | <entering_time_min> | <entering_time_max> |
    And the database table 'permissions_granted' has also the following row:
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 11       | 60      | 31              | <can_enter_from>    | <can_enter_until>   |
      | 31       | 60      | 31              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
      | 41       | 60      | 41              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
      | 51       | 60      | 51              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": <users_can_enter>,
      "entry_min_admitted_members_ratio": "None",
      "max_team_size": 2,
      "other_members": [
        {
          "can_enter": <users_can_enter>,
          "attempts_restriction_violated": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": <users_can_enter>,
          "attempts_restriction_violated": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "<expected_state>"
    }
    """
  Examples:
    | can_enter_from      | can_enter_until     | entering_time_min   | entering_time_max   | expected_state | users_can_enter |
    | 2007-01-01 10:21:21 | 3008-01-01 10:21:21 | 2007-12-31 23:59:59 | 3008-12-31 23:59:59 | ready          | true            |
    | 3007-01-01 10:21:21 | 3008-01-01 10:21:21 | 2007-12-31 23:59:59 | 3008-12-31 23:59:59 | not_ready      | true            |
    | 2007-01-01 10:21:21 | 2008-01-01 10:21:21 | 2007-12-31 23:59:59 | 3008-12-31 23:59:59 | not_ready      | true            |
    | 2007-01-01 10:21:21 | 3008-01-01 10:21:21 | 5099-12-31 23:59:59 | 9999-12-31 23:59:59 | not_ready      | false           |
    | 2007-01-01 10:21:21 | 3008-01-01 10:21:21 | 2007-12-31 23:59:59 | 2008-12-31 23:59:59 | not_ready      | false           |

  Scenario Outline: State is already_started for an individual contest
    Given the database table 'groups' has also the following row:
      | id  | type                |
      | 100 | ContestParticipants |
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio   | entry_max_team_size | participants_group_id | default_language_tag |
      | 50 | 00:00:00 | 1                       | User                   | <entry_min_admitted_members_ratio> | 0                   | 100                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 10       | 50      | content                  |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | creator_id | parent_attempt_id | root_item_id |
      | 1  | 31             | 2019-05-30 15:00:00 | 31         | 0                 | 50           |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id |
      | 100             | 31             |
    And I am the user with id "31"
    When I send a GET request to "/items/50/entry-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": {{"<entry_min_admitted_members_ratio>" != "null" ? "\"<entry_min_admitted_members_ratio>\"" : "null"}},
      "other_members": [],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "already_started"
    }
    """
    Examples:
      | entry_min_admitted_members_ratio |
      | None                             |
      | All                              |
      | Half                             |
      | One                              |

  Scenario Outline: State is already_started for a team-only contest
    Given the database table 'groups' has also the following row:
      | id  | type                |
      | 100 | ContestParticipants |
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio   | entry_max_team_size | participants_group_id | default_language_tag |
      | 60 | 00:00:00 | 1                       | Team                   | <entry_min_admitted_members_ratio> | 0                   | 100                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | creator_id | parent_attempt_id | root_item_id |
      | 1  | 11             | 2019-05-30 15:00:00 | 31         | 0                 | 60           |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id |
      | 100             | 11             |
    And I am the user with id "31"
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": {{"<entry_min_admitted_members_ratio>" != "null" ? "\"<entry_min_admitted_members_ratio>\"" : "null"}},
      "max_team_size": 0,
      "other_members": [
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "already_started"
    }
    """
    Examples:
      | entry_min_admitted_members_ratio |
      | None                             |
      | All                              |
      | Half                             |
      | One                              |

  Scenario: State is not_ready for an individual contest because the participation has expired
    Given the database table 'groups' has also the following row:
      | id  | type                |
      | 100 | ContestParticipants |
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio | entry_max_team_size | participants_group_id | default_language_tag |
      | 50 | 00:00:00 | 1                       | User                   | None                             | 0                   | 100                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 10       | 50      | content                  |
      | 11       | 50      | none                     |
      | 21       | 50      | solution                 |
      | 31       | 50      | content_with_descendants |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | creator_id | parent_attempt_id | root_item_id | allows_submissions_until |
      | 1  | 31             | 2019-05-30 15:00:00 | 31         | 0                 | 50           | 2019-05-30 20:00:00      |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id | expires_at          |
      | 100             | 31             | 2019-05-30 20:00:00 |
    And I am the user with id "31"
    When I send a GET request to "/items/50/entry-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": "None",
      "other_members": [],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "not_ready"
    }
    """

  Scenario: State is ready for a team-only contest because the participation has expired
    Given the database table 'groups' has also the following row:
      | id  | type                |
      | 100 | ContestParticipants |
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | allows_multiple_attempts | entry_participant_type | entry_min_admitted_members_ratio | entry_max_team_size | participants_group_id | default_language_tag |
      | 60 | 00:00:00 | 1                       | 1                        | Team                   | None                             | 3                   | 100                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | creator_id | parent_attempt_id | root_item_id | allows_submissions_until |
      | 1  | 11             | 2019-05-30 15:00:00 | 31         | 0                 | 60           | 2019-05-30 20:00:00      |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id | expires_at          |
      | 100             | 11             | 2019-05-30 20:00:00 |
    And I am the user with id "31"
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": "None",
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "ready"
    }
    """

  Scenario: State is ready for a team-only contest because the all its started attempts has expired
    Given the database table 'groups' has also the following row:
      | id  | type                |
      | 100 | ContestParticipants |
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | allows_multiple_attempts | entry_participant_type | entry_min_admitted_members_ratio | entry_max_team_size | participants_group_id | default_language_tag |
      | 60 | 00:00:00 | 1                       | 1                        | Team                   | None                             | 3                   | 100                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | creator_id | parent_attempt_id | root_item_id | allows_submissions_until |
      | 1  | 11             | 2019-05-30 15:00:00 | 31         | 0                 | 60           | 2019-05-30 20:00:00      |
      | 2  | 11             | 2019-05-30 15:00:00 | 31         | 0                 | 60           | 2019-05-30 20:00:00      |
      | 3  | 11             | 2019-05-30 15:00:00 | 31         | 0                 | 60           | 2019-05-30 20:00:00      |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id |
      | 100             | 11             |
    And I am the user with id "31"
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": "None",
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "ready"
    }
    """

  Scenario: State is not ready for a team-only contest because the team's users have a conflicting participation
    Given the database table 'groups' has also the following row:
      | id  | type                |
      | 100 | ContestParticipants |
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | allows_multiple_attempts | entry_participant_type | entry_min_admitted_members_ratio | entry_max_team_size | participants_group_id | default_language_tag |
      | 60 | 00:00:00 | 1                       | 1                        | Team                   | None                             | 3                   | 100                   | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 60      | info                     |
      | 21       | 60      | content_with_descendants |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | creator_id | parent_attempt_id | root_item_id | allows_submissions_until |
      | 1  | 11             | 2019-05-30 15:00:00 | 31         | 0                 | 60           | 2019-05-30 20:00:00      |
      | 2  | 11             | 2019-05-30 15:00:00 | 31         | 0                 | 60           | 2019-05-30 20:00:00      |
      | 3  | 11             | 2019-05-30 15:00:00 | 31         | 0                 | 60           | 2019-05-30 20:00:00      |
    And the database table 'groups_groups' has also the following row:
      | parent_group_id | child_group_id |
      | 100             | 11             |
    And the database table 'attempts' has also the following row:
      | participant_id | id | root_item_id |
      | 10             | 1  | 60           |
    And I am the user with id "31"
    When I send a GET request to "/items/60/entry-state?as_team_id=11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": "None",
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": false,
          "attempts_restriction_violated": true,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "attempts_restriction_violated": true,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "not_ready"
    }
    """

  Scenario Outline: The user cannot enter because of entering_time_min/entering_time_max
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio | default_language_tag | entering_time_min   | entering_time_max   |
      | 50 | 00:00:00 | 1                       | User                   | None                             | fr                   | <entering_time_min> | <entering_time_max> |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | source_group_id | can_enter_from      | can_enter_until     |
      | 11       | 50      | 11              | 2007-01-01 10:21:21 | 9999-12-31 23:59:59 |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/50/entry-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": "None",
      "other_members": [],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "ready"
    }
    """
  Examples:
    | entering_time_min   | entering_time_max   |
    | 5099-12-31 23:59:59 | 9999-12-31 23:59:59 |
    | 2007-12-31 23:59:59 | 2008-12-31 23:59:59 |

  Scenario: entry_frozen_teams is ignored for non-team contests
    Given the database has the following table 'items':
      | id | duration | requires_explicit_entry | entry_participant_type | entry_min_admitted_members_ratio | default_language_tag | entry_frozen_teams |
      | 50 | 00:00:00 | 1                       | User                   | None                             | fr                   | 1                  |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 31       | 50      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/50/entry-state"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": "None",
      "other_members": [],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "ready"
    }
    """

  Scenario Outline: State depends on frozen_membership when items.entry_frozen_teams = 1
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | allows_multiple_attempts | entry_participant_type | entry_min_admitted_members_ratio | entry_max_team_size | default_language_tag | entry_frozen_teams |
      | 60 | 00:00:00 | 1                       | 1                        | Team                   | None                             | 3                   | fr                   | 1                  |
    And the database has the following table 'permissions_generated':
      | group_id  | item_id | can_view_generated       |
      | <team_id> | 60      | info                     |
      | 21        | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/60/entry-state?as_team_id=<team_id>"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": "None",
      "max_team_size": 3,
      "other_members": [
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jane",
          "group_id": "41",
          "last_name": null,
          "login": "jane"
        },
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "first_name": "Jack",
          "group_id": "51",
          "last_name": "Daniel",
          "login": "jack"
        }
      ],
      "current_team_is_frozen": <current_team_is_frozen>,
      "frozen_teams_required": true,
      "state": "<state>"
    }
    """
  Examples:
    | team_id | current_team_is_frozen | state     |
    | 10      | true                   | ready     |
    | 11      | false                  | not_ready |

  Scenario: Hides personal info for users who haven't give the approval
    And the database has the following table 'items':
      | id | duration | requires_explicit_entry | allows_multiple_attempts | entry_participant_type | entry_min_admitted_members_ratio | default_language_tag |
      | 60 | 00:00:00 | 1                       | 1                        | Team                   | None                             | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id  | item_id | can_view_generated       |
      | 12        | 60      | info                     |
      | 21        | 60      | content_with_descendants |
    And I am the user with id "31"
    When I send a GET request to "/items/60/entry-state?as_team_id=12"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "current_user_can_enter": false,
      "entry_min_admitted_members_ratio": "None",
      "max_team_size": 0,
      "other_members": [
        {
          "can_enter": false,
          "attempts_restriction_violated": false,
          "group_id": "61",
          "login": "mark"
        }
      ],
      "current_team_is_frozen": false,
      "frozen_teams_required": false,
      "state": "not_ready"
    }
    """
