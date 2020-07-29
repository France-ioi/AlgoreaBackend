Feature: List official sessions for item_id - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name    |
      | 11 | jdoe    |
      | 13 | Group B |
    And the database has the following table 'users':
      | login | group_id | first_name | last_name |
      | jdoe  | 11       | John       | Doe       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 11             |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id  | allows_multiple_attempts | default_language_tag | is_root |
      | 200 | 0                        | fr                   | true    |
      | 210 | 1                        | fr                   | false   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 200     | none                     |
      | 13       | 210     | info                     |
      | 23       | 210     | content_with_descendants |

  Scenario: Wrong item_id
    Given I am the user with id "11"
    When I send a GET request to "/items/abc/official-sessions"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: No access to the item
    Given I am the user with id "11"
    When I send a GET request to "/items/200/official-sessions"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Wrong pagination
    Given I am the user with id "11"
    When I send a GET request to "/items/210/official-sessions?from.id=1234"
    Then the response code should be 400
    And the response error message should contain "Unallowed paging parameters (from.id)"
