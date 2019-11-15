Feature: Update item

Background:
  Given the database has the following table 'groups':
    | id | name       | type      |
    | 11 | jdoe       | UserSelf  |
    | 12 | jdoe-admin | UserAdmin |
  And the database has the following table 'users':
    | login | temp_user | group_id | owned_group_id |
    | jdoe  | 0         | 11       | 12             |
  And the database has the following table 'items':
    | id | type    | url                  | default_language_id | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | validation_min | unlocked_item_ids | score_min_unlock | contest_entering_condition | teams_editable | contest_max_team_size | has_attempts | duration | show_user_infos | contest_phase | level | group_code_enter |
    | 21 | Chapter | http://someurl1.com/ | 2                   | 1        | Task 1  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | 12             | 1                 | 99               | Half                       | 1              | 10                    | 1            | 01:20:30 | 1               | Closed        | 3     | 1                |
    | 50 | Chapter | http://someurl2.com/ | 2                   | 1        | Task 2  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | 12             | 1                 | 99               | Half                       | 1              | 10                    | 1            | 01:20:30 | 1               | Closed        | 3     | 1                |
    | 60 | Chapter | http://someurl2.com/ | 2                   | 1        | Task 3  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | 12             | 1                 | 99               | Half                       | 1              | 10                    | 1            | 01:20:30 | 1               | Closed        | 3     | 1                |
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
    | group_id | item_id | can_view | is_owner | giver_group_id |
    | 11       | 21      | solution | false    | 11             |
    | 11       | 50      | none     | true     | 11             |
    | 11       | 60      | none     | true     | 11             |
  And the database has the following table 'groups_ancestors':
    | id | ancestor_group_id | child_group_id | is_self |
    | 71 | 11                | 11             | 1       |
    | 72 | 12                | 12             | 1       |
  And the database has the following table 'languages':
    | id |
    | 2  |
    | 3  |

  Scenario: Valid
    Given I am the user with id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "type": "Course"
      }
      """
    Then the response should be "updated"
    And the table "items" at id "50" should be:
      | id | type   | url                  | default_language_id | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | validation_min | unlocked_item_ids | score_min_unlock | contest_entering_condition | teams_editable | contest_max_team_size | has_attempts | duration | show_user_infos | contest_phase | level | group_code_enter |
      | 50 | Course | http://someurl2.com/ | 2                   | 1        | Task 2  | 0                 | 1                         | 0        | 1         | forceNo     | 1             | 1           | One             | 12             | 1                 | 99               | Half                       | 1              | 10                    | 1            | 01:20:30 | 1               | Closed        | 3     | 1                |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "permissions_granted" should stay unchanged
    And the table "permissions_generated" should be:
      | group_id | item_id | can_view_generated | is_owner_generated |
      | 11       | 21      | solution           | false              |
      | 11       | 50      | solution           | true               |
      | 11       | 60      | solution           | true               |

  Scenario: Valid (all the fields are set)
    Given I am the user with id "11"
    And the database has the following table 'groups':
      | id    |
      | 12345 |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 73 | 12                | 12345          | 0       |
    And the database has the following table 'items':
      | id  |
      | 112 |
      | 134 |
    And the database has the following table 'items_strings':
      | language_id | item_id |
      | 3           | 50      |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated | can_grant_view_generated | is_owner_generated |
      | 11       | 112     | solution           | content                  | false              |
      | 11       | 134     | solution           | transfer                 | true               |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_view | can_grant_view | is_owner | giver_group_id |
      | 11       | 112     | solution | content        | false    | 11             |
      | 11       | 134     | none     | none           | true     | 11             |
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
        "validation_min": 1234,
        "unlocked_item_ids": "112,134",
        "score_min_unlock": 34,
        "contest_entering_condition": "All",
        "teams_editable": false,
        "contest_max_team_size": 2345,
        "has_attempts": false,
        "duration": "01:02:03",
        "show_user_infos": false,
        "contest_phase": "Analysis",
        "level": 345,
        "no_score": false,
        "group_code_enter": false,
        "default_language_id": "3",
        "children": [
          {"item_id": "112", "order": 0},
          {"item_id": "134", "order": 1}
        ]
      }
      """
    Then the response should be "updated"
    And the table "items" at id "50" should be:
      | id | type   | url               | default_language_id | teams_editable | no_score | text_id       | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | validation_min | unlocked_item_ids | score_min_unlock | contest_entering_condition | teams_editable | contest_max_team_size | has_attempts | duration | show_user_infos | contest_phase | level | group_code_enter |
      | 50 | Course | http://myurl.com/ | 3                   | 0              | 0        | Task number 1 | 1                 | 0                         | 1        | 0         | forceYes    | 0             | 0           | AllButOne       | 1234           | 112,134           | 34               | All                        | 0              | 2345                  | 0            | 01:02:03 | 0               | Analysis      | 345   | 0                |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should be:
      | parent_item_id | child_item_id |
      | 21             | 60            |
      | 50             | 112           |
      | 50             | 134           |
    And the table "items_ancestors" should be:
      | ancestor_item_id | child_item_id |
      | 21               | 60            |
      | 50               | 112           |
      | 50               | 134           |
    And the table "permissions_granted" should stay unchanged

  Scenario: Valid with empty full_screen
    Given I am the user with id "11"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "full_screen": ""
      }
      """
    Then the response should be "updated"
    And the table "items" at id "50" should be:
      | id | type    | url                  | default_language_id | no_score | text_id | title_bar_visible | display_details_in_parent | uses_api | read_only | full_screen | hints_allowed | fixed_ranks | validation_type | validation_min | unlocked_item_ids | score_min_unlock | contest_entering_condition | teams_editable | contest_max_team_size | has_attempts | duration | show_user_infos | contest_phase | level | group_code_enter |
      | 50 | Chapter | http://someurl2.com/ | 2                   | 1        | Task 2  | 0                 | 1                         | 0        | 1         |             | 1             | 1           | One             | 12             | 1                 | 99               | Half                       | 1              | 10                    | 1            | 01:20:30 | 1               | Closed        | 3     | 1                |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
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
    And the table "permissions_granted" should stay unchanged
