Feature: Update item

Background:
  Given the database has the following table 'users':
    | id | login | temp_user | self_group_id | owned_group_id |
    | 1  | jdoe  | 0         | 11            | 12             |
  And the database has the following table 'groups':
    | id | name       | type      |
    | 11 | jdoe       | UserSelf  |
    | 12 | jdoe-admin | UserAdmin |
  And the database has the following table 'items':
    | id | type    | url                  | default_language_id | no_score | text_id | title_bar_visible | custom_chapter | display_details_in_parent | uses_api | read_only | full_screen | show_difficulty | show_source | hints_allowed | fixed_ranks | validation_type | validation_min | unlocked_item_ids | score_min_unlock | team_mode | teams_editable | qualified_group_id | team_max_members | has_attempts | access_open_date    | duration | end_contest_date    | show_user_infos | contest_phase | level | group_code_enter |
    | 21 | Chapter | http://someurl1.com/ | 2                   | 1        | Task 1  | 0                 | 1              | 1                         | 0        | 1         | forceNo     | 1               | 1           | 1             | 1           | One             | 12             | 1                 | 99               | Half      | 1              | 2                  | 10               | 1            | 2016-01-02 03:04:05 | 01:20:30 | 2017-01-02 03:04:05 | 1               | Closed        | 3     | 1                |
    | 50 | Chapter | http://someurl2.com/ | 2                   | 1        | Task 2  | 0                 | 1              | 1                         | 0        | 1         | forceNo     | 1               | 1           | 1             | 1           | One             | 12             | 1                 | 99               | Half      | 1              | 2                  | 10               | 1            | 2016-01-02 03:04:05 | 01:20:30 | 2017-01-02 03:04:05 | 1               | Closed        | 3     | 1                |
    | 60 | Chapter | http://someurl2.com/ | 2                   | 1        | Task 3  | 0                 | 1              | 1                         | 0        | 1         | forceNo     | 1               | 1           | 1             | 1           | One             | 12             | 1                 | 99               | Half      | 1              | 2                  | 10               | 1            | 2016-01-02 03:04:05 | 01:20:30 | 2017-01-02 03:04:05 | 1               | Closed        | 3     | 1                |
  And the database has the following table 'items_items':
    | parent_item_id | child_item_id | child_order |
    | 21             | 60            | 0           |
    | 50             | 21            | 0           |
  And the database has the following table 'items_ancestors':
    | ancestor_item_id | child_item_id |
    | 21               | 60            |
    | 50               | 21            |
    | 50               | 60            |
  And the database has the following table 'groups_items':
    | id | group_id | item_id | manager_access | cached_manager_access | owner_access | creator_user_id |
    | 40 | 11       | 50      | false          | false                 | true         | 1               |
    | 41 | 11       | 21      | true           | true                  | false        | 1               |
    | 42 | 11       | 60      | false          | false                 | true         | 1               |
  And the database has the following table 'groups_ancestors':
    | id | ancestor_group_id | child_group_id | is_self |
    | 71 | 11                | 11             | 1       |
    | 72 | 12                | 12             | 1       |
  And the database has the following table 'languages':
    | id |
    | 2  |
    | 3  |

Scenario: Valid
  Given I am the user with id "1"
  When I send a PUT request to "/items/50" with the following body:
    """
    {
      "type": "Course"
    }
    """
  Then the response should be "updated"
  And the table "items" at id "50" should be:
    | id | type   | url                  | default_language_id | no_score | text_id | title_bar_visible | custom_chapter | display_details_in_parent | uses_api | read_only | full_screen | show_difficulty | show_source | hints_allowed | fixed_ranks | validation_type | validation_min | unlocked_item_ids | score_min_unlock | team_mode | teams_editable | qualified_group_id | team_max_members | has_attempts | access_open_date    | duration | end_contest_date    | show_user_infos | contest_phase | level | group_code_enter |
    | 50 | Course | http://someurl2.com/ | 2                   | 1        | Task 2  | 0                 | 1              | 1                         | 0        | 1         | forceNo     | 1               | 1           | 1             | 1           | One             | 12             | 1                 | 99               | Half      | 1              | 2                  | 10               | 1            | 2016-01-02 03:04:05 | 01:20:30 | 2017-01-02 03:04:05 | 1               | Closed        | 3     | 1                |
  And the table "items_strings" should stay unchanged
  And the table "items_items" should stay unchanged
  And the table "items_ancestors" should stay unchanged
  And the table "groups_items" should be:
    | group_id | item_id | manager_access | cached_manager_access | owner_access |
    | 11       | 21      | true           | true                  | false        |
    | 11       | 50      | false          | false                 | true         |
    | 11       | 60      | false          | false                 | true         |

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
    And the database has the following table 'items_strings':
      | language_id | item_id |
      | 3           | 50      |
    And the database has the following table 'groups_items':
      | id | group_id | item_id | manager_access | cached_manager_access | owner_access | creator_user_id |
      | 43 | 11       | 12      | true           | true                  | false        | 1               |
      | 44 | 11       | 34      | false          | false                 | true         | 1               |
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "type": "Course",
        "url": "http://myurl.com/",
        "text_id": "Task number 1",
        "title_bar_visible": true,
        "custom_chapter": false,
        "display_details_in_parent": false,
        "uses_api": true,
        "read_only": false,
        "full_screen": "forceYes",
        "show_difficulty": false,
        "show_source": false,
        "hints_allowed": false,
        "fixed_ranks": false,
        "validation_type": "AllButOne",
        "validation_min": 1234,
        "unlocked_item_ids": "12,34",
        "score_min_unlock": 34,
        "team_mode": "All",
        "teams_editable": false,
        "qualified_group_id": "12345",
        "team_max_members": 2345,
        "has_attempts": false,
        "access_open_date": "2018-01-02T03:04:05Z",
        "duration": "01:02:03",
        "end_contest_date": "2019-02-03T04:05:06Z",
        "show_user_infos": false,
        "contest_phase": "Analysis",
        "level": 345,
        "no_score": false,
        "group_code_enter": false,
        "default_language_id": "3",
        "children": [
          {"item_id": "12", "order": 0},
          {"item_id": "34", "order": 1}
        ]
      }
      """
    Then the response should be "updated"
    And the table "items" at id "50" should be:
      | id | type   | url               | default_language_id | teams_editable | no_score | text_id       | title_bar_visible | custom_chapter | display_details_in_parent | uses_api | read_only | full_screen | show_difficulty | show_source | hints_allowed | fixed_ranks | validation_type | validation_min | unlocked_item_ids | score_min_unlock | team_mode | teams_editable | qualified_group_id | team_max_members | has_attempts | access_open_date    | duration | end_contest_date    | show_user_infos | contest_phase | level | group_code_enter |
      | 50 | Course | http://myurl.com/ | 3                   | 0              | 0        | Task number 1 | 1                 | 0              | 0                         | 1        | 0         | forceYes    | 0               | 0           | 0             | 0           | AllButOne       | 1234           | 12,34             | 34               | All       | 0              | 12345              | 2345             | 0            | 2018-01-02 03:04:05 | 01:02:03 | 2019-02-03 04:05:06 | 0               | Analysis      | 345   | 0                |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should be:
      | parent_item_id | child_item_id |
      | 21             | 60            |
      | 50             | 12            |
      | 50             | 34            |
    And the table "items_ancestors" should be:
      | ancestor_item_id | child_item_id |
      | 21               | 60            |
      | 50               | 12            |
      | 50               | 34            |
    And the table "groups_items" should stay unchanged

  Scenario: Valid with empty full_screen
    Given I am the user with id "1"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "full_screen": ""
      }
      """
    Then the response should be "updated"
    And the table "items" at id "50" should be:
      | id | type    | url                  | default_language_id | no_score | text_id | title_bar_visible | custom_chapter | display_details_in_parent | uses_api | read_only | full_screen | show_difficulty | show_source | hints_allowed | fixed_ranks | validation_type | validation_min | unlocked_item_ids | score_min_unlock | team_mode | teams_editable | qualified_group_id | team_max_members | has_attempts | access_open_date    | duration | end_contest_date    | show_user_infos | contest_phase | level | group_code_enter |
      | 50 | Chapter | http://someurl2.com/ | 2                   | 1        | Task 2  | 0                 | 1              | 1                         | 0        | 1         |             | 1               | 1           | 1             | 1           | One             | 12             | 1                 | 99               | Half      | 1              | 2                  | 10               | 1            | 2016-01-02 03:04:05 | 01:20:30 | 2017-01-02 03:04:05 | 1               | Closed        | 3     | 1                |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Valid without any fields
    Given I am the user with id "1"
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
    And the table "groups_items" should stay unchanged

  Scenario: Valid with empty children array
    Given I am the user with id "1"
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
    And the table "groups_items" should stay unchanged
