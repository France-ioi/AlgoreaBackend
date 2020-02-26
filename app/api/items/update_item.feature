Feature: Update item

Background:
  Given the database has the following table 'groups':
    | id | name | type |
    | 11 | jdoe | User |
  And the database has the following table 'users':
    | login | temp_user | group_id |
    | jdoe  | 0         | 11       |
  And the database has the following table 'items':
    | id | type    | url                  | default_language_tag | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | teams_editable | contest_max_team_size | allows_multiple_attempts | duration | show_user_infos | prompt_to_join_group_by_code | entering_time_min   | entering_time_max   | contest_participants_group_id |
    | 21 | Chapter | http://someurl1.com/ | en                   | 1        | Task 1  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | Half                       | 1              | 10                    | 1                        | 01:20:30 | 1               | 1                            | 2007-01-01 01:02:03 | 3007-01-01 01:02:03 | null                          |
    | 50 | Chapter | http://someurl2.com/ | en                   | 1        | Task 2  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | Half                       | 1              | 10                    | 1                        | 01:20:30 | 1               | 1                            | 2007-01-01 01:02:03 | 3007-01-01 01:02:03 | null                          |
    | 60 | Chapter | http://someurl2.com/ | en                   | 1        | Task 3  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | Half                       | 1              | 10                    | 1                        | 01:20:30 | 1               | 1                            | 2007-01-01 01:02:03 | 3007-01-01 01:02:03 | 1234                          |
  And the database has the following table 'items_items':
    | parent_item_id | child_item_id | child_order |
    | 21             | 60            | 0           |
    | 50             | 21            | 0           |
  And the database has the following table 'items_ancestors':
    | ancestor_item_id | child_item_id |
    | 21               | 60            |
    | 50               | 21            |
    | 50               | 60            |
  And the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated | can_grant_view_generated | can_edit_generated | is_owner_generated |
    | 11       | 21      | solution           | none                     | children           | false              |
    | 11       | 50      | solution           | solution_with_grant      | all                | true               |
    | 11       | 60      | solution           | solution_with_grant      | all_with_grant     | true               |
  And the database has the following table 'permissions_granted':
    | group_id | item_id | can_view | is_owner | source_group_id | latest_update_on    |
    | 11       | 21      | solution | false    | 11              | 2019-05-30 11:00:00 |
    | 11       | 50      | none     | true     | 11              | 2019-05-30 11:00:00 |
    | 11       | 60      | none     | true     | 11              | 2019-05-30 11:00:00 |
  And the database has the following table 'groups_ancestors':
    | ancestor_group_id | child_group_id |
    | 11                | 11             |
  And the database has the following table 'attempts':
    | group_id | item_id | score_computed | order | result_propagation_state |
    | 11       | 21      | 0              | 1     | done                     |
    | 11       | 50      | 10             | 1     | done                     |
  And the database has the following table 'languages':
    | tag |
    | en  |
    | sl  |

  Scenario: Valid
    Given I am the user with id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "type": "Course"
      }
      """
    Then the response should be "updated"
    And the table "items" should stay unchanged but the row with id "50"
    And the table "items" at id "50" should be:
    | id | type   | url                  | default_language_tag | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | teams_editable | contest_max_team_size | allows_multiple_attempts | duration | show_user_infos | prompt_to_join_group_by_code | entering_time_min   | entering_time_max   | contest_participants_group_id |
    | 50 | Course | http://someurl2.com/ | en                   | 1        | Task 2  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | Half                       | 1              | 10                    | 1                        | 01:20:30 | 1               | 1                            | 2007-01-01 01:02:03 | 3007-01-01 01:02:03 | null                          |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "attempts" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should be:
      | group_id | item_id | can_view_generated | is_owner_generated |
      | 11       | 21      | solution           | false              |
      | 11       | 50      | solution           | true               |
      | 11       | 60      | solution           | true               |

  Scenario: Valid (all the fields are set)
    Given I am the user with id "11"
    And the database has the following table 'items':
      | id  | default_language_tag |
      | 112 | fr                   |
      | 134 | fr                   |
    And the database has the following table 'items_strings':
      | language_tag | item_id |
      | sl           | 50      |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 11       | 112     | solution           | content                  | answer              | all                | false              |
      | 11       | 134     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_view | can_grant_view | can_watch | can_edit | is_owner | source_group_id | latest_update_on    |
      | 11       | 112     | solution | content        | answer    | all      | false    | 11              | 2019-05-30 11:00:00 |
      | 11       | 134     | none     | none           | none      | none     | true     | 11              | 2019-05-30 11:00:00 |
    And the database table 'attempts' has also the following rows:
      | group_id | item_id | order | score_computed | result_propagation_state |
      | 11       | 112     | 1     | 50             | done                     |
      | 11       | 134     | 1     | 60             | done                     |
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "type": "Course",
        "url": "http://myurl.com/",
        "text_id": "Task number 1",
        "title_bar_visible": true,
        "display_details_in_parent": false,
        "uses_api": true,
        "read_only": false,
        "full_screen": "forceYes",
        "hints_allowed": false,
        "fixed_ranks": false,
        "validation_type": "AllButOne",
        "contest_entering_condition": "All",
        "teams_editable": false,
        "contest_max_team_size": 2345,
        "allows_multiple_attempts": false,
        "duration": "01:02:03",
        "show_user_infos": false,
        "no_score": false,
        "prompt_to_join_group_by_code": false,
        "default_language_tag": "sl",
        "children": [
          {"item_id": "112", "order": 0, "category": "Discovery", "score_weight": 1},
          {"item_id": "134", "order": 1, "category": "Application", "score_weight": 2}
        ]
      }
      """
    Then the response should be "updated"
    And the table "items" should stay unchanged but the row with id "50"
    And the table "items" at id "50" should be:
      | id | type   | url               | default_language_tag | teams_editable | no_score | text_id       | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | teams_editable | contest_max_team_size | allows_multiple_attempts | duration | show_user_infos | prompt_to_join_group_by_code | contest_participants_group_id |
      | 50 | Course | http://myurl.com/ | sl                   | 0              | 0        | Task number 1 | 1                 | 0                         | 1        | 0         | forceYes    | 0             | 0           | AllButOne       | All                        | 0              | 2345                  | 0                        | 01:02:03 | 0               | 0                            | 5577006791947779410           |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should be:
      | parent_item_id | child_item_id | category    | score_weight | content_view_propagation | upper_view_levels_propagation | grant_view_propagation | watch_propagation | edit_propagation |
      | 21             | 60            | Undefined   | 1            | none                     | use_content_view_propagation  | 0                      | 0                 | 0                |
      | 50             | 112           | Discovery   | 1            | as_info                  | use_content_view_propagation  | 0                      | 0                 | 0                |
      | 50             | 134           | Application | 2            | as_info                  | as_is                         | 1                      | 1                 | 1                |
    And the table "items_ancestors" should be:
      | ancestor_item_id | child_item_id |
      | 21               | 60            |
      | 50               | 112           |
      | 50               | 134           |
    And the table "groups" should be:
      | id                  | type                | name            |
      | 11                  | User                | jdoe            |
      | 5577006791947779410 | ContestParticipants | 50-participants |
    And the table "permissions_granted" should be:
      | group_id            | item_id | can_view | can_grant_view | can_watch | can_edit | is_owner | source_group_id     | ABS(TIMESTAMPDIFF(SECOND, latest_update_on, NOW())) < 3 |
      | 11                  | 21      | solution | none           | none      | none     | false    | 11                  | 0                                                       |
      | 11                  | 50      | none     | none           | none      | none     | true     | 11                  | 0                                                       |
      | 11                  | 60      | none     | none           | none      | none     | true     | 11                  | 0                                                       |
      | 11                  | 112     | solution | content        | answer    | all      | false    | 11                  | 0                                                       |
      | 11                  | 134     | none     | none           | none      | none     | true     | 11                  | 0                                                       |
      | 5577006791947779410 | 50      | content  | none           | none      | none     | false    | 5577006791947779410 | 1                                                       |
    And the table "permissions_generated" should be:
      | group_id            | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 11                  | 21      | solution           | none                     | none                | none               | false              |
      | 11                  | 50      | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 11                  | 60      | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 11                  | 112     | solution           | content                  | answer              | all                | false              |
      | 11                  | 134     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 5577006791947779410 | 50      | content            | none                     | none                | none               | false              |
      | 5577006791947779410 | 112     | info               | none                     | none                | none               | false              |
      | 5577006791947779410 | 134     | info               | none                     | none                | none               | false              |
    And the table "attempts" should stay unchanged but the row with item_id "50"
    And the table "attempts" at item_id "50" should be:
      | group_id | item_id | score_computed | order | result_propagation_state |
      | 11       | 50      | 56.666668      | 1     | done                     |

  Scenario: Valid with empty full_screen
    Given I am the user with id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "full_screen": ""
      }
      """
    Then the response should be "updated"
    And the table "items" should stay unchanged but the row with id "50"
    And the table "items" at id "50" should be:
      | id | type    | url                  | default_language_tag | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | entering_time_min   | entering_time_max   | teams_editable | contest_max_team_size | allows_multiple_attempts | duration | show_user_infos | prompt_to_join_group_by_code | contest_participants_group_id |
      | 50 | Chapter | http://someurl2.com/ | en                   | 1        | Task 2  | 0                 | 1                         | 0        | 1         |             | 1             | 1           | One             | Half                       | 2007-01-01 01:02:03 | 3007-01-01 01:02:03 | 1              | 10                    | 1                        | 01:20:30 | 1               | 1                            | null                          |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "permissions_granted" should stay unchanged

  Scenario: Should set content_view_propagation to 'none' by default if can_grant_view = 'none' for the parent item
    Given I am the user with id "11"
    And the database has the following table 'items':
      | id  | default_language_tag |
      | 112 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 11       | 112     | solution           | content                  | answer              | all                | false              |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_view | can_grant_view | can_watch | can_edit | is_owner | source_group_id | latest_update_on    |
      | 11       | 112     | solution | content        | answer    | all      | false    | 11              | 2019-05-30 11:00:00 |
    When I send a PUT request to "/items/21" with the following body:
      """
      {
        "children": [
          {"item_id": "112", "order": 0}
        ]
      }
      """
    Then the response should be "updated"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should be:
      | parent_item_id | child_item_id | category  | score_weight | content_view_propagation | upper_view_levels_propagation | grant_view_propagation | watch_propagation | edit_propagation |
      | 21             | 112           | Undefined | 1            | as_info                  | use_content_view_propagation  | 0                      | 0                 | 0                |
      | 50             | 21            | Undefined | 1            | none                     | use_content_view_propagation  | 0                      | 0                 | 0                |
    And the table "items_ancestors" should be:
      | ancestor_item_id | child_item_id |
      | 21               | 112           |
      | 50               | 21            |
      | 50               | 112           |
    And the table "groups" should be:
      | id                  | type                | name            |
      | 11                  | User                | jdoe            |
    And the table "permissions_granted" should be:
      | group_id            | item_id | can_view | can_grant_view | can_watch | can_edit | is_owner | source_group_id     | ABS(TIMESTAMPDIFF(SECOND, latest_update_on, NOW())) < 3 |
      | 11                  | 21      | solution | none           | none      | none     | false    | 11                  | 0                                                       |
      | 11                  | 50      | none     | none           | none      | none     | true     | 11                  | 0                                                       |
      | 11                  | 60      | none     | none           | none      | none     | true     | 11                  | 0                                                       |
      | 11                  | 112     | solution | content        | answer    | all      | false    | 11                  | 0                                                       |
    And the table "permissions_generated" should be:
      | group_id            | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 11                  | 21      | solution           | none                     | none                | none               | false              |
      | 11                  | 50      | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 11                  | 60      | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 11                  | 112     | solution           | content                  | answer              | all                | false              |
    And the table "attempts" should stay unchanged but the row with item_id "50"
    And the table "attempts" at item_id "50" should be:
      | group_id | item_id | score_computed | order | result_propagation_state |
      | 11       | 50      | 0              | 1     | done                     |

  Scenario: Valid without any fields
    Given I am the user with id "11"
    When I send a PUT request to "/items/50" with the following body:
    """
    {
    }
    """
    Then the response should be "updated"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "permissions_granted" should stay unchanged

  Scenario: Valid with empty children array
    Given I am the user with id "11"
    When I send a PUT request to "/items/21" with the following body:
    """
    {
      "children": []
    }
    """
    Then the response should be "updated"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should be:
      | parent_item_id | child_item_id |
      | 50             | 21            |
    And the table "items_ancestors" should be:
      | ancestor_item_id | child_item_id |
      | 50               | 21            |
    And the table "groups" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "attempts" should stay unchanged but the row with item_id "50"
    And the table "attempts" at item_id "50" should be:
      | group_id | item_id | score_computed | order | result_propagation_state |
      | 11       | 50      | 0              | 1     | done                     |

  Scenario: Keep existing contest participants group
    Given I am the user with id "11"
    When I send a PUT request to "/items/60" with the following body:
    """
    {
      "duration": null
    }
    """
    Then the response should be "updated"
    And the table "items" should stay unchanged but the row with id "60"
    And the table "items" at id "60" should be:
      | id | duration | contest_participants_group_id |
      | 60 | null     | 1234                          |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    When I send a PUT request to "/items/60" with the following body:
    """
    {
      "duration": "12:34:56"
    }
    """
    Then the response should be "updated"
    And the table "items" should stay unchanged but the row with id "60"
    And the table "items" at id "60" should be:
      | id | duration | contest_participants_group_id |
      | 60 | 12:34:56 | 1234                          |
    And the table "groups" should stay unchanged

  Scenario: Recomputes attempts if no_score is given
    Given I am the user with id "11"
    When I send a PUT request to "/items/50" with the following body:
    """
    {
      "no_score": false
    }
    """
    Then the response should be "updated"
    And the table "items" should stay unchanged but the row with id "50"
    And the table "items" at id "50" should be:
      | id | no_score |
      | 50 | false    |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "attempts" should stay unchanged but the row with item_id "50"
    And the table "attempts" at item_id "50" should be:
      | group_id | item_id | score_computed | order | result_propagation_state |
      | 11       | 50      | 0              | 1     | done                     |

  Scenario: Recomputes attempts if validation_type is given
    Given I am the user with id "11"
    When I send a PUT request to "/items/50" with the following body:
    """
    {
      "validation_type": "All"
    }
    """
    Then the response should be "updated"
    And the table "items" should stay unchanged but the row with id "50"
    And the table "items" at id "50" should be:
      | id | validation_type |
      | 50 | All             |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "attempts" should stay unchanged but the row with item_id "50"
    And the table "attempts" at item_id "50" should be:
      | group_id | item_id | score_computed | order | result_propagation_state |
      | 11       | 50      | 0              | 1     | done                     |


  Scenario Outline: Sets default values of items_items.content_view_propagation/upper_view_levels_propagation/grant_view_propagation correctly for each can_grant_view
    Given I am the user with id "11"
    And the database has the following table 'items':
      | id  | default_language_tag |
      | 112 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_grant_view_generated |
      | 11       | 112     | info               | <can_grant_view>         |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_view | can_grant_view   | source_group_id |
      | 11       | 112     | info     | <can_grant_view> | 11              |
    When I send a PUT request to "/items/21" with the following body:
      """
      {
        "children": [{"item_id": 112, "order": 1}]
      }
      """
    Then the response should be "updated"
    And the table "items_items" should be:
      | parent_item_id | child_item_id | child_order | content_view_propagation   | upper_view_levels_propagation   | grant_view_propagation   |
      | 21             | 112           | 1           | <content_view_propagation> | <upper_view_levels_propagation> | <grant_view_propagation> |
      | 50             | 21            | 0           | none                       | use_content_view_propagation    | false                    |
    Examples:
      | can_grant_view           | content_view_propagation | upper_view_levels_propagation | grant_view_propagation |
      | solution_with_grant      | as_info                  | as_is                         | true                   |
      | solution                 | as_info                  | as_is                         | false                  |
      | content_with_descendants | as_info                  | as_content_with_descendants   | false                  |
      | content                  | as_info                  | use_content_view_propagation  | false                  |
      | none                     | none                     | use_content_view_propagation  | false                  |

  Scenario Outline: Sets default values of items_items.watch_propagation/edit_propagation correctly
    Given I am the user with id "11"
    And the database has the following table 'items':
      | id  | default_language_tag |
      | 112 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | <parent_permission_column> |
      | 11       | 112     | info               | <parent_permission_value>  |
    When I send a PUT request to "/items/21" with the following body:
      """
      {
        "children": [{"item_id": 112, "order": 1}]
      }
      """
    Then the response should be "updated"
    And the table "items_items" at parent_item_id "21" should be:
      | parent_item_id | child_item_id | child_order | <propagation_column> |
      | 21             | 112           | 1           | <propagation_value>  |
    Examples:
      | parent_permission_column | parent_permission_value | propagation_column | propagation_value |
      | can_watch_generated      | answer_with_grant       | watch_propagation  | true              |
      | can_watch_generated      | answer                  | watch_propagation  | false             |
      | can_watch_generated      | result                  | watch_propagation  | false             |
      | can_watch_generated      | none                    | watch_propagation  | false             |
      | can_edit_generated       | all_with_grant          | edit_propagation   | true              |
      | can_edit_generated       | all                     | edit_propagation   | false             |
      | can_edit_generated       | children                | edit_propagation   | false             |
      | can_edit_generated       | none                    | edit_propagation   | false             |

  Scenario Outline: Sets items_items.content_view_propagation/upper_view_levels_propagation/grant_view_propagation correctly
    Given I am the user with id "11"
    And the database has the following table 'items':
      | id  | default_language_tag |
      | 112 | fr                   |
    And the database table 'permissions_generated' has also the following row:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 11       | 112     | info               | <can_grant_view>         | none                | none               | 0                  |
    When I send a PUT request to "/items/21" with the following body:
      """
      {
        "children": [{
          "item_id": 112,
          "order": 1,
          "<field_name>": {{"<value>" != "true" && "<value>" != "false" ? "\"<value>\"" : <value>}}
        }]
      }
      """
    Then the response should be "updated"
    And the table "items_items" at parent_item_id "21" should be:
      | parent_item_id | child_item_id | child_order | <field_name> |
      | 21             | 112           | 1           | <value>      |
    Examples:
      | can_grant_view           | field_name                    | value                        |
      | solution_with_grant      | content_view_propagation      | as_content                   |
      | solution                 | content_view_propagation      | as_content                   |
      | content_with_descendants | content_view_propagation      | as_content                   |
      | content                  | content_view_propagation      | as_content                   |
      | solution_with_grant      | content_view_propagation      | as_info                      |
      | solution                 | content_view_propagation      | as_info                      |
      | content_with_descendants | content_view_propagation      | as_info                      |
      | content                  | content_view_propagation      | as_info                      |
      | enter                    | content_view_propagation      | as_info                      |
      | solution_with_grant      | content_view_propagation      | none                         |
      | solution                 | content_view_propagation      | none                         |
      | content_with_descendants | content_view_propagation      | none                         |
      | content                  | content_view_propagation      | none                         |
      | enter                    | content_view_propagation      | none                         |
      | none                     | content_view_propagation      | none                         |
      | solution_with_grant      | upper_view_levels_propagation | as_is                        |
      | solution                 | upper_view_levels_propagation | as_is                        |
      | solution_with_grant      | upper_view_levels_propagation | as_content_with_descendants  |
      | solution                 | upper_view_levels_propagation | as_content_with_descendants  |
      | content_with_descendants | upper_view_levels_propagation | as_content_with_descendants  |
      | content                  | upper_view_levels_propagation | use_content_view_propagation |
      | enter                    | upper_view_levels_propagation | use_content_view_propagation |
      | none                     | upper_view_levels_propagation | use_content_view_propagation |
      | solution_with_grant      | grant_view_propagation        | true                         |
      | solution_with_grant      | grant_view_propagation        | false                        |
      | solution                 | grant_view_propagation        | false                        |
      | content_with_descendants | grant_view_propagation        | false                        |
      | content                  | grant_view_propagation        | false                        |
      | enter                    | grant_view_propagation        | false                        |
      | none                     | grant_view_propagation        | false                        |

  Scenario Outline: Sets items_items.watch_propagation/edit_propagation correctly
    Given I am the user with id "11"
    And the database has the following table 'items':
      | id  | default_language_tag |
      | 112 | fr                   |
    And the database table 'permissions_generated' has also the following row:
      | group_id | item_id | can_view_generated | <parent_permission_column> |
      | 11       | 112     | info               | <parent_permission_value>  |
    When I send a PUT request to "/items/21" with the following body:
      """
      {
        "children": [{
          "item_id": 112,
          "order": 1,
          "<field_name>": {{"<value>" != "true" && "<value>" != "false" ? "\"<value>\"" : <value>}}
        }]
      }
      """
    Then the response should be "updated"
    And the table "items_items" at parent_item_id "21" should be:
      | parent_item_id | child_item_id | child_order | <field_name> |
      | 21             | 112           | 1           | <value>      |
    Examples:
      | parent_permission_column | parent_permission_value | field_name        | value |
      | can_watch_generated      | answer_with_grant       | watch_propagation | true  |
      | can_watch_generated      | answer_with_grant       | watch_propagation | false |
      | can_watch_generated      | answer                  | watch_propagation | false |
      | can_watch_generated      | result                  | watch_propagation | false |
      | can_watch_generated      | none                    | watch_propagation | false |
      | can_edit_generated       | all_with_grant          | edit_propagation  | true  |
      | can_edit_generated       | all_with_grant          | edit_propagation  | false |
      | can_edit_generated       | all                     | edit_propagation  | false |
      | can_edit_generated       | children                | edit_propagation  | false |
      | can_edit_generated       | none                    | edit_propagation  | false |
