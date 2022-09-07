Feature: Create an item dependency
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
    And the database has the following table 'item_dependencies':
      | item_id | dependent_item_id | score | grant_content_view |
      | 100     | 210               | 22    | true               |
      | 100     | 220               | 10    | true               |
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
      | 15       | 210     | content_with_descendants | none                     | none               | none                | false              |
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

  Scenario Outline: Create an item dependency
    Given I am the user with id "11"
    When I send a POST request to "/items/210/prerequisites/200" with the following body:
    """
    {
      <score_in_request>
      "grant_content_view": false
    }
    """
    Then the response should be "created"
    And the table "item_dependencies" should stay unchanged but the rows with dependent_item_id "210"
    And the table "item_dependencies" at dependent_item_id "210" should be:
      | item_id | dependent_item_id | score            | grant_content_view |
      | 100     | 210               | 22               | true               |
      | 200     | 210               | <expected_score> | false              |
  Examples:
    | score_in_request | expected_score |
    | "score": 0,      | 0              |
    |                  | 100            |

  Scenario: Create an item dependency with grant_content_view = true
    Given I am the user with id "11"
    When I send a POST request to "/items/200/prerequisites/210" with the following body:
    """
    {
      "score": 100,
      "grant_content_view": true
    }
    """
    Then the response should be "created"
    And the table "item_dependencies" should stay unchanged but the rows with dependent_item_id "200"
    And the table "item_dependencies" at dependent_item_id "200" should be:
      | item_id | dependent_item_id | score | grant_content_view |
      | 210     | 200               | 100   | true               |

  Scenario: dependent_item_id = prerequisite_item_id
    Given I am the user with id "11"
    When I send a POST request to "/items/210/prerequisites/210" with the following body:
    """
    {
      "score": 70,
      "grant_content_view": false
    }
    """
    Then the response should be "created"
    And the table "item_dependencies" should stay unchanged but the rows with dependent_item_id "210"
    And the table "item_dependencies" at dependent_item_id "210" should be:
      | item_id | dependent_item_id | score | grant_content_view |
      | 100     | 210               | 22    | true               |
      | 210     | 210               | 70    | false              |
