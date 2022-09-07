Feature: Apply an item dependency
  Background:
    Given the database has the following table 'groups':
      | id | name       | grade | type  |
      | 11 | jdoe       | -2    | User  |
      | 13 | Group B    | -2    | Team  |
      | 14 | nosolution | -2    | User  |
      | 15 | Group C    | -2    | Class |
      | 17 | fr         | -2    | User  |
      | 22 | info       | -2    | User  |
      | 23 | jane       | -2    | User  |
      | 26 | team       | -2    | Team  |
    And the database has the following table 'users':
      | login      | temp_user | group_id | default_language |
      | jdoe       | 0         | 11       |                  |
      | nosolution | 0         | 14       |                  |
      | fr         | 0         | 17       | fr               |
      | info       | 0         | 22       |                  |
      | jane       | 0         | 23       |                  |
    And the database has the following table 'items':
      | id  | type    | default_language_tag | requires_explicit_entry |
      | 100 | Course  | en                   | true                    |
      | 200 | Course  | en                   | true                    |
      | 210 | Chapter | en                   | false                   |
      | 220 | Chapter | en                   | false                   |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 11             |
      | 13              | 17             |
      | 15              | 11             |
      | 15              | 14             |
      | 15              | 17             |
      | 26              | 11             |
      | 26              | 22             |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | manager_id | group_id | can_watch_members |
      | 22         | 15       | true              |
    And the database has the following table 'item_dependencies':
      | item_id | dependent_item_id | score | grant_content_view |
      | 100     | 200               | 22    | true               |
      | 100     | 220               | 10    | true               |
      | 200     | 210               | 20    | true               |
      | 200     | 220               | 30    | false              |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | source_group_id | origin         | latest_update_at    | can_view                 | can_enter_from      | can_enter_until     | can_grant_view | can_watch | can_edit | can_make_session_official | is_owner |
      | 22       | 200     | 22              | item_unlocking | 2019-05-30 11:00:00 | info                     | 3019-12-31 23:59:59 | 2020-01-31 23:59:59 | none           | none      | none     | false                     | false    |
      | 22       | 210     | 22              | item_unlocking | 2019-05-30 11:00:00 | info                     | 2019-12-31 23:59:59 | 2020-01-31 23:59:59 | none           | none      | none     | false                     | false    |
      | 26       | 210     | 26              | item_unlocking | 2019-05-30 11:00:00 | content_with_descendants | 2019-12-31 23:59:59 | 2020-01-31 23:59:59 | none           | none      | none     | false                     | false    |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_edit_generated | can_watch_generated | is_owner_generated |
      | 11       | 200     | solution                 | content                  | children           | result              | true               |
      | 11       | 210     | solution                 | none                     | all                | none                | true               |
      | 11       | 220     | solution                 | none                     | none               | none                | false              |
      | 13       | 200     | solution                 | none                     | none               | none                | false              |
      | 13       | 210     | solution                 | none                     | none               | none                | false              |
      | 13       | 220     | solution                 | none                     | none               | none                | false              |
      | 15       | 200     | none                     | none                     | all                | none                | false              |
      | 15       | 210     | content_with_descendants | content                  | none               | none                | false              |
      | 17       | 200     | solution                 | none                     | none               | none                | false              |
      | 17       | 210     | solution                 | none                     | none               | none                | false              |
      | 17       | 220     | solution                 | none                     | none               | none                | false              |
      | 22       | 200     | solution                 | none                     | none               | none                | false              |
      | 22       | 210     | info                     | none                     | none               | result              | false              |
      | 22       | 220     | info                     | none                     | none               | none                | false              |
      | 23       | 200     | info                     | none                     | none               | none                | false              |
      | 26       | 200     | solution                 | none                     | none               | none                | false              |
      | 26       | 210     | content_with_descendants | none                     | none               | none                | false              |
      | 26       | 220     | info                     | none                     | none               | none                | false              |
    And the database has the following table 'languages':
      | tag |
      | fr  |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          | root_item_id | parent_attempt_id |
      | 0  | 11             | 2019-05-30 10:00:00 | null         | null              |
      | 0  | 13             | 2019-05-30 10:00:00 | null         | null              |
      | 0  | 17             | 2019-05-30 10:00:00 | null         | null              |
      | 0  | 22             | 2019-05-30 10:00:00 | null         | null              |
      | 1  | 11             | 2019-05-30 11:00:00 | null         | null              |
      | 1  | 13             | 2019-05-30 11:00:00 | null         | null              |
      | 1  | 17             | 2019-05-30 10:00:00 | 200          | 0                 |
      | 1  | 26             | 2019-05-30 10:00:00 | null         | null              |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          | latest_activity_at  | score_computed | validated_at        |
      | 0          | 11             | 100     | null                | 2019-05-30 11:00:01 | 22             | null                |
      | 0          | 11             | 200     | null                | 2019-05-30 11:00:01 | 11.1           | null                |
      | 0          | 11             | 210     | null                | 2018-05-30 11:00:01 | 12.2           | null                |
      | 0          | 11             | 220     | 2019-05-30 11:00:00 | 2019-05-30 11:00:02 | 13.3           | null                |
      | 0          | 13             | 200     | 2019-05-30 11:00:00 | 2019-05-30 11:00:03 | 0              | null                |
      | 0          | 13             | 210     | 2019-05-30 11:00:00 | 2019-05-30 11:00:03 | 14.4           | null                |
      | 0          | 13             | 220     | null                | 2018-05-30 11:00:02 | 15.5           | null                |
      | 0          | 17             | 200     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 0              | null                |
      | 0          | 17             | 210     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 10             | 2019-05-30 11:00:01 |
      | 0          | 22             | 100     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 45             | null                |
      | 0          | 22             | 200     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 30             | null                |
      | 0          | 26             | 200     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 0              | null                |
      | 1          | 11             | 200     | null                | 2019-05-30 12:00:01 | 21.1           | null                |
      | 1          | 11             | 210     | null                | 2018-05-30 12:00:01 | 22.2           | null                |
      | 1          | 11             | 220     | 2019-05-30 12:00:00 | 2019-05-30 12:00:02 | 3.3            | null                |
      | 1          | 13             | 210     | 2019-05-30 12:00:00 | 2019-05-30 12:00:03 | 24.4           | null                |
      | 1          | 13             | 220     | null                | 2018-05-30 12:00:02 | 5.5            | null                |
      | 1          | 17             | 200     | 2018-05-30 11:00:00 | 2018-05-30 11:00:01 | 10             | 2018-05-30 11:00:01 |
      | 1          | 17             | 210     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 20             | 2019-05-30 11:00:01 |
      | 1          | 26             | 200     | 2019-05-30 11:00:00 | 2019-05-30 11:00:01 | 20             | 2019-05-30 11:00:01 |

  Scenario: Apply an item dependency for dependent item not requiring explicit entry
    Given I am the user with id "11"
    And the DB time now is "2020-05-30 11:00:00"
    When I send a POST request to "/items/210/prerequisites/200/apply"
    Then the response should be "updated"
    And the table "results" should stay unchanged
    And the table "permissions_granted" should be:
      | group_id | item_id | source_group_id | origin         | latest_update_at    | can_view                 | can_enter_from      | can_enter_until     | can_grant_view | can_watch | can_edit | can_make_session_official | is_owner |
      | 11       | 210     | 11              | item_unlocking | 2020-05-30 11:00:00 | content                  | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 | none           | none      | none     | false                     | false    |
      | 22       | 200     | 22              | item_unlocking | 2019-05-30 11:00:00 | info                     | 3019-12-31 23:59:59 | 2020-01-31 23:59:59 | none           | none      | none     | false                     | false    |
      | 22       | 210     | 22              | item_unlocking | 2020-05-30 11:00:00 | content                  | 2019-12-31 23:59:59 | 2020-01-31 23:59:59 | none           | none      | none     | false                     | false    |
      | 26       | 210     | 26              | item_unlocking | 2019-05-30 11:00:00 | content_with_descendants | 2019-12-31 23:59:59 | 2020-01-31 23:59:59 | none           | none      | none     | false                     | false    |
    And the table "permissions_generated" should stay unchanged but the rows with group_id "11,22"
    And the table "permissions_generated" at group_id "11" should be:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_edit_generated | can_watch_generated | is_owner_generated |
      | 11       | 200     | solution           | content                  | children           | result              | true               |
      | 11       | 210     | content            | none                     | none               | none                | false              |
      | 11       | 220     | solution           | none                     | none               | none                | false              |
    And the table "permissions_generated" at group_id "22" should be:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_edit_generated | can_watch_generated | is_owner_generated |
      | 22       | 200     | info               | none                     | none               | none                | false              |
      | 22       | 210     | content            | none                     | none               | none                | false              |
      | 22       | 220     | info               | none                     | none               | none                | false              |

  Scenario: Apply an item dependency for dependent item requiring explicit entry
    Given I am the user with id "11"
    And the DB time now is "2020-05-30 11:00:00"
    When I send a POST request to "/items/200/prerequisites/100/apply"
    Then the response should be "updated"
    And the table "results" should stay unchanged
    And the table "permissions_granted" should be:
      | group_id | item_id | source_group_id | origin         | latest_update_at    | can_view                 | can_enter_from      | can_enter_until     | can_grant_view | can_watch | can_edit | can_make_session_official | is_owner |
      | 11       | 200     | 11              | item_unlocking | 2020-05-30 11:00:00 | none                     | 2020-05-30 11:00:00 | 9999-12-31 23:59:59 | none           | none      | none     | false                     | false    |
      | 22       | 200     | 22              | item_unlocking | 2020-05-30 11:00:00 | info                     | 2020-05-30 11:00:00 | 2020-01-31 23:59:59 | none           | none      | none     | false                     | false    |
      | 22       | 210     | 22              | item_unlocking | 2019-05-30 11:00:00 | info                     | 2019-12-31 23:59:59 | 2020-01-31 23:59:59 | none           | none      | none     | false                     | false    |
      | 26       | 210     | 26              | item_unlocking | 2019-05-30 11:00:00 | content_with_descendants | 2019-12-31 23:59:59 | 2020-01-31 23:59:59 | none           | none      | none     | false                     | false    |
    And the table "permissions_generated" should stay unchanged but the rows with group_id "11,22"
    And the table "permissions_generated" at group_id "22" should be:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_edit_generated | can_watch_generated | is_owner_generated |
      | 22       | 200     | info               | none                     | none               | none                | false              |
      | 22       | 210     | info               | none                     | none               | none                | false              |
      | 22       | 220     | info               | none                     | none               | none                | false              |
