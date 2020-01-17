Feature: Update item

Background:
  Given the database has the following table 'groups':
    | id | name | type     |
    | 11 | jdoe | UserSelf |
  And the database has the following table 'users':
    | login | temp_user | group_id |
    | jdoe  | 0         | 11       |
  And the database has the following table 'items':
    | id | type    | url                  | default_language_tag | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | teams_editable | contest_max_team_size | has_attempts | duration | show_user_infos | group_code_enter | contest_participants_group_id |
    | 21 | Chapter | http://someurl1.com/ | en                   | 1        | Task 1  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | Half                       | 1              | 10                    | 1            | 01:20:30 | 1               | 1                | null                          |
    | 50 | Chapter | http://someurl2.com/ | en                   | 1        | Task 2  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | Half                       | 1              | 10                    | 1            | 01:20:30 | 1               | 1                | null                          |
    | 60 | Chapter | http://someurl2.com/ | en                   | 1        | Task 3  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | Half                       | 1              | 10                    | 1            | 01:20:30 | 1               | 1                | 1234                          |
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
    | group_id | item_id | can_view_generated | can_edit_generated | is_owner_generated |
    | 11       | 21      | solution           | none               | false              |
    | 11       | 50      | solution           | transfer           | true               |
    | 11       | 60      | solution           | transfer           | true               |
  And the database has the following table 'permissions_granted':
    | group_id | item_id | can_view | is_owner | source_group_id | latest_update_on    |
    | 11       | 21      | solution | false    | 11              | 2019-05-30 11:00:00 |
    | 11       | 50      | none     | true     | 11              | 2019-05-30 11:00:00 |
    | 11       | 60      | none     | true     | 11              | 2019-05-30 11:00:00 |
  And the database has the following table 'groups_ancestors':
    | id | ancestor_group_id | child_group_id | is_self |
    | 71 | 11                | 11             | 1       |
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
    | id | type   | url                  | default_language_tag | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | teams_editable | contest_max_team_size | has_attempts | duration | show_user_infos | group_code_enter | contest_participants_group_id |
    | 50 | Course | http://someurl2.com/ | en                   | 1        | Task 2  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | Half                       | 1              | 10                    | 1            | 01:20:30 | 1               | 1                | null                          |
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
      | 11       | 134     | solution           | transfer                 | transfer            | transfer           | true               |
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
        "has_attempts": false,
        "duration": "01:02:03",
        "show_user_infos": false,
        "no_score": false,
        "group_code_enter": false,
        "default_language_tag": "sl",
        "children": [
          {"item_id": "112", "order": 0},
          {"item_id": "134", "order": 1}
        ]
      }
      """
    Then the response should be "updated"
    And the table "items" should stay unchanged but the row with id "50"
    And the table "items" at id "50" should be:
      | id | type   | url               | default_language_tag | teams_editable | no_score | text_id       | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | teams_editable | contest_max_team_size | has_attempts | duration | show_user_infos | group_code_enter | contest_participants_group_id |
      | 50 | Course | http://myurl.com/ | sl                   | 0              | 0        | Task number 1 | 1                 | 0                         | 1        | 0         | forceYes    | 0             | 0           | AllButOne       | All                        | 0              | 2345                  | 0            | 01:02:03 | 0               | 0                | 5577006791947779410           |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should be:
      | parent_item_id | child_item_id | content_view_propagation | upper_view_levels_propagation | grant_view_propagation | watch_propagation | edit_propagation |
      | 21             | 60            | none                     | use_content_view_propagation  | 0                      | 0                 | 0                |
      | 50             | 112           | as_info                  | as_is                         | 0                      | 0                 | 0                |
      | 50             | 134           | as_info                  | as_is                         | 1                      | 1                 | 1                |
    And the table "items_ancestors" should be:
      | ancestor_item_id | child_item_id |
      | 21               | 60            |
      | 50               | 112           |
      | 50               | 134           |
    And the table "groups" should be:
      | id                  | type                | name            |
      | 11                  | UserSelf            | jdoe            |
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
      | 11                  | 50      | solution           | transfer                 | transfer            | transfer           | true               |
      | 11                  | 60      | solution           | transfer                 | transfer            | transfer           | true               |
      | 11                  | 112     | solution           | content                  | answer              | all                | false              |
      | 11                  | 134     | solution           | transfer                 | transfer            | transfer           | true               |
      | 5577006791947779410 | 50      | content            | none                     | none                | none               | false              |
      | 5577006791947779410 | 112     | info               | none                     | none                | none               | false              |
      | 5577006791947779410 | 134     | info               | none                     | none                | none               | false              |
    And the table "attempts" should stay unchanged but the row with item_id "50"
    And the table "attempts" at item_id "50" should be:
      | group_id | item_id | score_computed | order | result_propagation_state |
      | 11       | 50      | 55             | 1     | done                     |

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
      | id | type    | url                  | default_language_tag | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | contest_entering_condition | teams_editable | contest_max_team_size | has_attempts | duration | show_user_infos | group_code_enter | contest_participants_group_id |
      | 50 | Chapter | http://someurl2.com/ | en                   | 1        | Task 2  | 0                 | 1                         | 0        | 1         |             | 1             | 1           | One             | Half                       | 1              | 10                    | 1            | 01:20:30 | 1               | 1                | null                          |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups" should stay unchanged
    And the table "permissions_granted" should stay unchanged

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
    When I send a PUT request to "/items/50" with the following body:
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
      | 21             | 60            |
    And the table "items_ancestors" should be:
      | ancestor_item_id | child_item_id |
      | 21               | 60            |
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
