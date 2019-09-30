Feature: Add item

  Background:
    Given the database has the following table 'users':
      | id | login | temp_user | self_group_id | owned_group_id |
      | 1  | jdoe  | 0         | 11            | 12             |
    And the database has the following table 'groups':
      | id | name       | type      |
      | 11 | jdoe       | UserSelf  |
      | 12 | jdoe-admin | UserAdmin |
    And the database has the following table 'items':
      | id | teams_editable | no_score |
      | 21 | false          | false    |
    And the database has the following table 'groups_items':
      | id | group_id | item_id | cached_manager_access | creator_user_id |
      | 41 | 11       | 21      | true                  | 1               |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 71 | 11                | 11             | 1       |
      | 72 | 12                | 12             | 1       |
    And the database has the following table 'languages':
      | id |
      | 3  |

  Scenario: Valid
    Given I am the user with id "1"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "language_id": "3",
        "title": "my title",
        "image_url":"http://bit.ly/1234",
        "subtitle": "hard task",
        "description": "the goal of this task is ...",
        "parent_item_id": "21",
        "order": 100
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": { "id": "5577006791947779410" }
      }
      """
    And the table "items" at id "5577006791947779410" should be:
      | id                  | type   | url  | default_language_id | teams_editable | no_score | text_id | title_bar_visible | custom_chapter | display_details_in_parent | uses_api | read_only | full_screen | show_difficulty | show_source | hints_allowed | fixed_ranks | validation_type | validation_min | unlocked_item_ids | score_min_unlock | team_mode | teams_editable | qualified_group_id | team_max_members | has_attempts | contest_opens_at | duration | contest_closes_at | show_user_infos | contest_phase | level | no_score | group_code_enter |
      | 5577006791947779410 | Course | null | 3                   | 0              | 0        | null    | 1                 | 0              | 0                         | 1        | 0         | default     | 0               | 0           | 0             | 0           | All             | null           | null              | 100              | null      | 0              | null               | 0                | 0            | null             | null     | null              | 0               | Running       | null  | 0        | 0                |
    And the table "items_strings" should be:
      | id                  | item_id             | language_id | title    | image_url          | subtitle  | description                  |
      | 6129484611666145821 | 5577006791947779410 | 3           | my title | http://bit.ly/1234 | hard task | the goal of this task is ... |
    And the table "items_items" should be:
      | id                  | parent_item_id | child_item_id       | child_order |
      | 4037200794235010051 | 21             | 5577006791947779410 | 100         |
    And the table "items_ancestors" should be:
      | ancestor_item_id | child_item_id       |
      | 21               | 5577006791947779410 |
    And the table "groups_items" at id "8674665223082153551" should be:
      | id                  | group_id | item_id             | creator_user_id | ABS(TIMESTAMPDIFF(SECOND, full_access_since, NOW())) < 3 | owner_access | cached_manager_access | ABS(TIMESTAMPDIFF(SECOND, cached_full_access_since, NOW())) < 3 | cached_full_access |
      | 8674665223082153551 | 11       | 5577006791947779410 | 1               | 1                                                        | 1            | 1                     | 1                                                               | 1                  |

  Scenario: Valid (all the fields are set)
    Given I am the user with id "1"
    And the database has the following table 'groups':
      | id    |
      | 12345 |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 73 | 12                | 12345          | 0       |
    And the database has the following table 'items':
      | id |
      | 12 |
      | 34 |
    And the database has the following table 'groups_items':
      | id | group_id | item_id | cached_manager_access | owner_access | creator_user_id |
      | 42 | 11       | 12      | true                  | false        | 1               |
      | 43 | 11       | 34      | false                 | true         | 1               |
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "url": "http://myurl.com/",
        "text_id": "Task number 1",
        "title_bar_visible": true,
        "custom_chapter": true,
        "display_details_in_parent": true,
        "uses_api": true,
        "read_only": true,
        "full_screen": "forceYes",
        "show_difficulty": true,
        "show_source": true,
        "hints_allowed": true,
        "fixed_ranks": true,
        "validation_type": "AllButOne",
        "validation_min": 1234,
        "unlocked_item_ids": "12,34",
        "score_min_unlock": 34,
        "team_mode": "All",
        "teams_editable": true,
        "qualified_group_id": "12345",
        "team_max_members": 2345,
        "has_attempts": true,
        "contest_opens_at": "2018-01-02T03:04:05Z",
        "duration": "01:02:03",
        "contest_closes_at": "2019-02-03T04:05:06Z",
        "show_user_infos": true,
        "contest_phase": "Analysis",
        "level": 345,
        "no_score": true,
        "group_code_enter": true,
        "language_id": "3",
        "title": "my title",
        "image_url":"http://bit.ly/1234",
        "subtitle": "hard task",
        "description": "the goal of this task is ...",
        "parent_item_id": "21",
        "order": 100,
        "children": [
          {"item_id": "12", "order": 0},
          {"item_id": "34", "order": 1}
        ]
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": { "id": "5577006791947779410" }
      }
      """
    And the table "items" at id "5577006791947779410" should be:
      | id                  | type   | url               | default_language_id | teams_editable | no_score | text_id       | title_bar_visible | custom_chapter | display_details_in_parent | uses_api | read_only | full_screen | show_difficulty | show_source | hints_allowed | fixed_ranks | validation_type | validation_min | unlocked_item_ids | score_min_unlock | team_mode | teams_editable | qualified_group_id | team_max_members | has_attempts | contest_opens_at    | duration | contest_closes_at   | show_user_infos | contest_phase | level | no_score | group_code_enter |
      | 5577006791947779410 | Course | http://myurl.com/ | 3                   | 1              | 1        | Task number 1 | 1                 | 1              | 1                         | 1        | 1         | forceYes    | 1               | 1           | 1             | 1           | AllButOne       | 1234           | 12,34             | 34               | All       | 1              | 12345              | 2345             | 1            | 2018-01-02 03:04:05 | 01:02:03 | 2019-02-03 04:05:06 | 1               | Analysis      | 345   | 1        | 1                |
    And the table "items_strings" should be:
      | id                  | item_id             | language_id | title    | image_url          | subtitle  | description                  |
      | 6129484611666145821 | 5577006791947779410 | 3           | my title | http://bit.ly/1234 | hard task | the goal of this task is ... |
    And the table "items_items" should be:
      | id                  | parent_item_id      | child_item_id       | child_order |
      | 3916589616287113937 | 5577006791947779410 | 12                  | 0           |
      | 4037200794235010051 | 21                  | 5577006791947779410 | 100         |
      | 6334824724549167320 | 5577006791947779410 | 34                  | 1           |
    And the table "items_ancestors" should be:
      | ancestor_item_id    | child_item_id       |
      | 21                  | 12                  |
      | 21                  | 34                  |
      | 21                  | 5577006791947779410 |
      | 5577006791947779410 | 12                  |
      | 5577006791947779410 | 34                  |
    And the table "groups_items" at id "8674665223082153551" should be:
      | id                  | group_id | item_id             | creator_user_id | ABS(TIMESTAMPDIFF(SECOND, full_access_since, NOW())) < 3 | owner_access | cached_manager_access | ABS(TIMESTAMPDIFF(SECOND, cached_full_access_since, NOW())) < 3 | cached_full_access |
      | 8674665223082153551 | 11       | 5577006791947779410 | 1               | 1                                                        | 1            | 1                     | 1                                                               | 1                  |

  Scenario: Valid with empty full_screen
    Given I am the user with id "1"
    When I send a POST request to "/items" with the following body:
    """
    {
      "type": "Course",
      "full_screen": "",
      "language_id": "3",
      "title": "my title",
      "parent_item_id": "21",
      "order": 100
    }
    """
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": { "id": "5577006791947779410" }
    }
    """
    And the table "items" at id "5577006791947779410" should be:
      | id                  | type   | url  | default_language_id | teams_editable | no_score | text_id | title_bar_visible | custom_chapter | display_details_in_parent | uses_api | read_only | full_screen | show_difficulty | show_source | hints_allowed | fixed_ranks | validation_type | validation_min | unlocked_item_ids | score_min_unlock | team_mode | teams_editable | qualified_group_id | team_max_members | has_attempts | contest_opens_at | duration | contest_closes_at | show_user_infos | contest_phase | level | no_score | group_code_enter |
      | 5577006791947779410 | Course | null | 3                   | 0              | 0        | null    | 1                 | 0              | 0                         | 1        | 0         |             | 0               | 0           | 0             | 0           | All             | null           | null              | 100              | null      | 0              | null               | 0                | 0            | null             | null     | null              | 0               | Running       | null  | 0        | 0                |
    And the table "items_strings" should be:
      | id                  | item_id             | language_id | title    | image_url | subtitle | description |
      | 6129484611666145821 | 5577006791947779410 | 3           | my title | null      | null     | null        |
    And the table "items_items" should be:
      | id                  | parent_item_id | child_item_id       | child_order |
      | 4037200794235010051 | 21             | 5577006791947779410 | 100         |
    And the table "items_ancestors" should be:
      | ancestor_item_id | child_item_id       |
      | 21               | 5577006791947779410 |
    And the table "groups_items" at id "8674665223082153551" should be:
      | id                  | group_id | item_id             | creator_user_id | ABS(TIMESTAMPDIFF(SECOND, full_access_since, NOW())) < 3 | owner_access | cached_manager_access | ABS(TIMESTAMPDIFF(SECOND, cached_full_access_since, NOW())) < 3 | cached_full_access |
      | 8674665223082153551 | 11       | 5577006791947779410 | 1               | 1                                                        | 1            | 1                     | 1                                                               | 1                  |
