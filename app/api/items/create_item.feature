Feature: Create item

  Background:
    Given the database has the following table 'groups':
      | id | name    | type    |
      | 10 | Friends | Friends |
      | 11 | jdoe    | User    |
    And the database has the following table 'users':
      | login | temp_user | group_id |
      | jdoe  | 0         | 11       |
    And the database has the following table 'items':
      | id | teams_editable | no_score | default_language_tag |
      | 21 | false          | false    | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_edit_generated |
      | 11       | 21      | solution           | children           |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_view | can_edit | source_group_id | latest_update_on    |
      | 11       | 21      | solution | children | 11              | 2019-05-30 11:00:00 |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 10                | 10             |
      | 10                | 11             |
      | 11                | 11             |
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 11             |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | result_propagation_state |
      | 0          | 11             | 21      | done                     |
    And the database has the following table 'languages':
      | tag |
      | sl  |

  Scenario: Valid
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "language_tag": "sl",
        "title": "my title",
        "image_url":"http://bit.ly/1234",
        "subtitle": "hard task",
        "description": "the goal of this task is ...",
        "parent_item_id": "21"
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
      | id                  | type   | url  | default_language_tag | teams_editable | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | teams_editable | contest_max_team_size | allows_multiple_attempts | duration | show_user_infos | no_score | prompt_to_join_group_by_code | entering_time_min   | entering_time_max   | contest_participants_group_id |
      | 5577006791947779410 | Course | null | sl                   | 0              | 0        | null    | 1                 | 0                         | 1        | 0         | default     | 0             | 0           | All             | None                       | 0              | 0                     | 0                        | null     | 0               | 0        | 0                            | 1000-01-01 00:00:00 | 9999-12-31 23:59:59 | null                          |
    And the table "items_strings" should be:
      | item_id             | language_tag | title    | image_url          | subtitle  | description                  |
      | 5577006791947779410 | sl           | my title | http://bit.ly/1234 | hard task | the goal of this task is ... |
    And the table "items_items" should be:
      | parent_item_id | child_item_id       | child_order | content_view_propagation | upper_view_levels_propagation | grant_view_propagation | watch_propagation | edit_propagation | category  | score_weight |
      | 21             | 5577006791947779410 | 1           | as_info                  | as_is                         | 1                      | 1                 | 1                | Undefined | 1            |
    And the table "items_ancestors" should be:
      | ancestor_item_id | child_item_id       |
      | 21               | 5577006791947779410 |
    And the table "permissions_granted" at group_id "11" should be:
      | group_id | item_id             | source_group_id | origin           | can_view | can_grant_view | can_watch | can_edit | is_owner | ABS(TIMESTAMPDIFF(SECOND, latest_update_on, NOW())) < 3 |
      | 11       | 21                  | 11              | group_membership | solution | none           | none      | children | 0        | 0                                                       |
      | 11       | 5577006791947779410 | 11              | self             | none     | none           | none      | none     | 1        | 1                                                       |
    And the table "permissions_generated" should be:
      | group_id | item_id             | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 11       | 21                  | solution           | none                     | none                | children           | 0                  |
      | 11       | 5577006791947779410 | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | 1                  |
    And the table "groups" should stay unchanged
    And the table "attempts" should stay unchanged
    And the table "results" should stay unchanged

  Scenario: Valid (all the fields are set)
    Given I am the user with id "11"
    And the database table 'items' has also the following rows:
      | id | default_language_tag |
      | 12 | fr                   |
      | 34 | fr                   |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 10       | 21      | none                     | content                  | none                | none               | 0                  |
      | 11       | 12      | content_with_descendants | solution                 | answer              | all                | 0                  |
      | 11       | 34      | solution                 | solution_with_grant      | answer_with_grant   | all_with_grant     | 0                  |
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | can_view                 | can_grant_view      | can_watch         | can_edit       | is_owner | source_group_id | latest_update_on    |
      | 11       | 12      | content_with_descendants | solution            | answer            | all            | 0        | 11              | 2019-05-30 11:00:00 |
      | 11       | 34      | solution                 | solution_with_grant | answer_with_grant | all_with_grant | 0        | 11              | 2019-05-30 11:00:00 |
    And the database table 'results' has also the following rows:
      | attempt_id | participant_id | item_id | result_propagation_state |
      | 0          | 11             | 12      | done                     |
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Chapter",
        "url": "http://myurl.com/",
        "text_id": "Task number 1",
        "title_bar_visible": true,
        "display_details_in_parent": true,
        "uses_api": true,
        "read_only": true,
        "full_screen": "forceYes",
        "hints_allowed": true,
        "fixed_ranks": true,
        "validation_type": "AllButOne",
        "contest_entering_condition": "All",
        "entering_time_min": "2007-01-01T01:02:03Z",
        "entering_time_max": "3007-01-01T01:02:03Z",
        "teams_editable": true,
        "contest_max_team_size": 2345,
        "allows_multiple_attempts": true,
        "entry_participant_type": "Team",
        "duration": "01:02:03",
        "show_user_infos": true,
        "no_score": true,
        "prompt_to_join_group_by_code": true,
        "language_tag": "sl",
        "title": "my title",
        "image_url":"http://bit.ly/1234",
        "subtitle": "hard task",
        "description": "the goal of this task is ...",
        "parent_item_id": "21",
        "category": "Challenge",
        "score_weight": 3,
        "content_view_propagation": "as_content",
        "upper_view_levels_propagation": "use_content_view_propagation",
        "children": [
          {"item_id": "12", "order": 0},
          {"item_id": "34", "order": 1, "category": "Application", "score_weight": 2}
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
      | id                  | type    | url               | default_language_tag | teams_editable | no_score | text_id       | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | teams_editable | contest_max_team_size | allows_multiple_attempts | entry_participant_type | duration | show_user_infos | no_score | prompt_to_join_group_by_code | entering_time_min   | entering_time_max   | contest_participants_group_id |
      | 5577006791947779410 | Chapter | http://myurl.com/ | sl                   | 1              | 1        | Task number 1 | 1                 | 1                         | 1        | 1         | forceYes    | 1             | 1           | AllButOne       | All                        | 1              | 2345                  | 1                        | Team                   | 01:02:03 | 1               | 1        | 1                            | 2007-01-01 01:02:03 | 3007-01-01 01:02:03 | 8674665223082153551           |
    And the table "items_strings" should be:
      | item_id             | language_tag | title    | image_url          | subtitle  | description                  |
      | 5577006791947779410 | sl           | my title | http://bit.ly/1234 | hard task | the goal of this task is ... |
    And the table "items_items" should be:
      | parent_item_id      | child_item_id       | child_order | content_view_propagation | upper_view_levels_propagation | grant_view_propagation | watch_propagation | edit_propagation | category    | score_weight |
      | 21                  | 5577006791947779410 | 1           | as_info                  | as_is                         | 1                      | 1                 | 1                | Challenge   | 3            |
      | 5577006791947779410 | 12                  | 0           | as_info                  | as_is                         | 0                      | 0                 | 0                | Undefined   | 1            |
      | 5577006791947779410 | 34                  | 1           | as_info                  | as_is                         | 1                      | 1                 | 1                | Application | 2            |
    And the table "items_ancestors" should be:
      | ancestor_item_id    | child_item_id       |
      | 21                  | 12                  |
      | 21                  | 34                  |
      | 21                  | 5577006791947779410 |
      | 5577006791947779410 | 12                  |
      | 5577006791947779410 | 34                  |
    And the table "groups" should be:
      | id                  | type                | name                             |
      | 10                  | Friends             | Friends                          |
      | 11                  | User                | jdoe                             |
      | 8674665223082153551 | ContestParticipants | 5577006791947779410-participants |
    And the table "permissions_granted" should be:
      | group_id            | item_id             | source_group_id     | origin           | can_view                 | can_grant_view      | can_watch         | can_edit       | is_owner | ABS(TIMESTAMPDIFF(SECOND, latest_update_on, NOW())) < 3 |
      | 11                  | 12                  | 11                  | group_membership | content_with_descendants | solution            | answer            | all            | 0        | 0                                                       |
      | 11                  | 21                  | 11                  | group_membership | solution                 | none                | none              | children       | 0        | 0                                                       |
      | 11                  | 34                  | 11                  | group_membership | solution                 | solution_with_grant | answer_with_grant | all_with_grant | 0        | 0                                                       |
      | 11                  | 5577006791947779410 | 11                  | self             | none                     | none                | none              | none           | 1        | 1                                                       |
      | 8674665223082153551 | 5577006791947779410 | 8674665223082153551 | group_membership | content                  | none                | none              | none           | 0        | 1                                                       |
    And the table "permissions_generated" should be:
      | group_id            | item_id             | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 10                  | 12                  | none               | none                     | none                | none               | 0                  |
      | 10                  | 21                  | none               | content                  | none                | none               | 0                  |
      | 10                  | 34                  | none               | content                  | none                | none               | 0                  |
      | 10                  | 5577006791947779410 | none               | content                  | none                | none               | 0                  |
      | 11                  | 12                  | solution           | solution                 | answer              | all                | 0                  |
      | 11                  | 21                  | solution           | none                     | none                | children           | 0                  |
      | 11                  | 34                  | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | 0                  |
      | 11                  | 5577006791947779410 | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | 1                  |
      | 8674665223082153551 | 12                  | info               | none                     | none                | none               | 0                  |
      | 8674665223082153551 | 34                  | info               | none                     | none                | none               | 0                  |
      | 8674665223082153551 | 5577006791947779410 | content            | none                     | none                | none               | 0                  |
    And the table "attempts" should stay unchanged
    And the table "results" should be:
      | attempt_id | participant_id | item_id             | result_propagation_state |
      | 0          | 11             | 12                  | done                     |
      | 0          | 11             | 21                  | done                     |
      | 0          | 11             | 5577006791947779410 | done                     |

  Scenario: Valid with empty full_screen
    Given I am the user with id "11"
    When I send a POST request to "/items" with the following body:
    """
    {
      "type": "Course",
      "full_screen": "",
      "language_tag": "sl",
      "title": "my title",
      "parent_item_id": "21"
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
      | id                  | type   | url  | default_language_tag | teams_editable | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | teams_editable | contest_max_team_size | allows_multiple_attempts | duration | show_user_infos | no_score | prompt_to_join_group_by_code | contest_participants_group_id |
      | 5577006791947779410 | Course | null | sl                   | 0              | 0        | null    | 1                 | 0                         | 1        | 0         |             | 0             | 0           | All             | None                       | 0              | 0                     | 0                        | null     | 0               | 0        | 0                            | null                          |
    And the table "items_strings" should be:
      | item_id             | language_tag | title    | image_url | subtitle | description |
      | 5577006791947779410 | sl           | my title | null      | null     | null        |
    And the table "items_items" should be:
      | parent_item_id | child_item_id       | child_order | content_view_propagation | upper_view_levels_propagation | grant_view_propagation | watch_propagation | edit_propagation | category  | score_weight |
      | 21             | 5577006791947779410 | 1           | as_info                  | as_is                         | 1                      | 1                 | 1                | Undefined | 1            |
    And the table "items_ancestors" should be:
      | ancestor_item_id | child_item_id       |
      | 21               | 5577006791947779410 |
    And the table "permissions_granted" at group_id "11" should be:
      | group_id | item_id             | source_group_id | origin           | can_view | can_grant_view | can_watch | can_edit | is_owner | ABS(TIMESTAMPDIFF(SECOND, latest_update_on, NOW())) < 3 |
      | 11       | 21                  | 11              | group_membership | solution | none           | none      | children | 0        | 0                                                       |
      | 11       | 5577006791947779410 | 11              | self             | none     | none           | none      | none     | 1        | 1                                                       |
    And the table "permissions_generated" should be:
      | group_id | item_id             | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 11       | 21                  | solution           | none                     | none                | children           | 0                  |
      | 11       | 5577006791947779410 | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | 1                  |
    And the table "groups" should stay unchanged
    And the table "attempts" should stay unchanged


  Scenario: Valid when type=Skill
    Given I am the user with id "11"
    And the database table 'items' has also the following rows:
      | id | default_language_tag | type    |
      | 12 | fr                   | Skill   |
      | 34 | fr                   | Chapter |
      | 50 | fr                   | Skill   |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 10       | 21      | none                     | content                  | none                | none               | 0                  |
      | 11       | 12      | content_with_descendants | solution                 | answer              | all                | 0                  |
      | 11       | 34      | solution                 | solution_with_grant      | answer_with_grant   | all_with_grant     | 0                  |
      | 11       | 50      | solution                 | solution_with_grant      | answer_with_grant   | all_with_grant     | 0                  |
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | can_view                 | can_grant_view      | can_watch         | can_edit       | is_owner | source_group_id | latest_update_on    |
      | 11       | 12      | content_with_descendants | solution            | answer            | all            | 0        | 11              | 2019-05-30 11:00:00 |
      | 11       | 34      | solution                 | solution_with_grant | answer_with_grant | all_with_grant | 0        | 11              | 2019-05-30 11:00:00 |
      | 11       | 50      | solution                 | solution_with_grant | answer_with_grant | all_with_grant | 0        | 11              | 2019-05-30 11:00:00 |
    When I send a POST request to "/items" with the following body:
    """
    {
      "type": "Skill",
      "language_tag": "sl",
      "title": "my skill",
      "parent_item_id": "50",
      "duration": "01:02:03",
      "children": [
        {"item_id": "12", "order": 0, "category": "Application", "score_weight": 2},
        {"item_id": "34", "order": 1, "category": "Application", "score_weight": 2}
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
      | id                  | type  | url  | default_language_tag | teams_editable | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | teams_editable | contest_max_team_size | allows_multiple_attempts | duration | show_user_infos | no_score | prompt_to_join_group_by_code | contest_participants_group_id |
      | 5577006791947779410 | Skill | null | sl                   | 0              | 0        | null    | 1                 | 0                         | 1        | 0         | default     | 0             | 0           | All             | None                       | 0              | 0                     | 0                        | 01:02:03 | 0               | 0        | 0                            | 8674665223082153551           |
    And the table "items_strings" should be:
      | item_id             | language_tag | title    | image_url | subtitle | description |
      | 5577006791947779410 | sl           | my skill | null      | null     | null        |
    And the table "items_items" should be:
      | parent_item_id      | child_item_id       | child_order | content_view_propagation | upper_view_levels_propagation | grant_view_propagation | watch_propagation | edit_propagation | category    | score_weight |
      | 50                  | 5577006791947779410 | 1           | as_info                  | as_is                         | 1                      | 1                 | 1                | Undefined   | 1            |
      | 5577006791947779410 | 12                  | 0           | as_info                  | as_is                         | 0                      | 0                 | 0                | Application | 2            |
      | 5577006791947779410 | 34                  | 1           | as_info                  | as_is                         | 1                      | 1                 | 1                | Application | 2            |
    And the table "items_ancestors" should be:
      | ancestor_item_id    | child_item_id       |
      | 50                  | 12                  |
      | 50                  | 34                  |
      | 50                  | 5577006791947779410 |
      | 5577006791947779410 | 12                  |
      | 5577006791947779410 | 34                  |
    And the table "groups" should be:
      | id                  | type                | name                             |
      | 10                  | Friends             | Friends                          |
      | 11                  | User                | jdoe                             |
      | 8674665223082153551 | ContestParticipants | 5577006791947779410-participants |
    And the table "permissions_granted" should be:
      | group_id            | item_id             | source_group_id     | origin           | can_view                 | can_grant_view      | can_watch         | can_edit       | is_owner | ABS(TIMESTAMPDIFF(SECOND, latest_update_on, NOW())) < 3 |
      | 11                  | 12                  | 11                  | group_membership | content_with_descendants | solution            | answer            | all            | 0        | 0                                                       |
      | 11                  | 21                  | 11                  | group_membership | solution                 | none                | none              | children       | 0        | 0                                                       |
      | 11                  | 34                  | 11                  | group_membership | solution                 | solution_with_grant | answer_with_grant | all_with_grant | 0        | 0                                                       |
      | 11                  | 50                  | 11                  | group_membership | solution                 | solution_with_grant | answer_with_grant | all_with_grant | 0        | 0                                                       |
      | 11                  | 5577006791947779410 | 11                  | self             | none                     | none                | none              | none           | 1        | 1                                                       |
      | 8674665223082153551 | 5577006791947779410 | 8674665223082153551 | group_membership | content                  | none                | none              | none           | 0        | 1                                                       |
    And the table "permissions_generated" should be:
      | group_id            | item_id             | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 10                  | 21                  | none               | content                  | none                | none               | 0                  |
      | 11                  | 12                  | solution           | solution                 | answer              | all                | 0                  |
      | 11                  | 21                  | solution           | none                     | none                | children           | 0                  |
      | 11                  | 34                  | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | 0                  |
      | 11                  | 50                  | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | 0                  |
      | 11                  | 5577006791947779410 | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | 1                  |
      | 8674665223082153551 | 12                  | info               | none                     | none                | none               | 0                  |
      | 8674665223082153551 | 34                  | info               | none                     | none                | none               | 0                  |
      | 8674665223082153551 | 5577006791947779410 | content            | none                     | none                | none               | 0                  |
