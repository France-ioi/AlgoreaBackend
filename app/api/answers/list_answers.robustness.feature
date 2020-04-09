Feature: List answers - robustness
Background:
  Given the database has the following table 'groups':
    | id | name | type |
    | 1  | jdoe | User |
  And the database has the following table 'users':
    | login | temp_user | group_id |
    | jdoe  | 0         | 1        |
  And the database has the following table 'groups_ancestors':
    | ancestor_group_id | child_group_id |
    | 1                 | 1              |
  And the database has the following table 'items':
    | id  | type    | teams_editable | no_score | default_language_tag |
    | 190 | Chapter | false          | false    | fr                   |
    | 200 | Chapter | false          | false    | fr                   |
  And the database has the following table 'permissions_generated':
    | group_id | item_id | can_view_generated       |
    | 1        | 190     | content_with_descendants |
    | 1        | 200     | none                     |

  Scenario: Should fail when item_id is invalid
    Given I am the user with id "1"
    When I send a GET request to "/items/abc/answers"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Should fail when only item_id is present
    Given I am the user with id "1"
    When I send a GET request to "/items/190/answers"
    Then the response code should be 400
    And the response error message should contain "Either author_id or attempt_id must be present"

  Scenario: Should fail when no access to item_id
    Given I am the user with id "1"
    When I send a GET request to "/items/200/answers"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
