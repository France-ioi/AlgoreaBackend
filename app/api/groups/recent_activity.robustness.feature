Feature: Get recent activity for group_id and item_id - robustness
  Background:
    Given the database has the following users:
      | login   | temp_user | group_id | first_name  | last_name |
      | someone | 0         | 21       | Bill        | Clinton   |
      | user    | 0         | 11       | John        | Doe       |
      | owner   | 0         | 23       | Jean-Michel | Blanquer  |
    And the database has the following table 'groups':
      | id |
      | 13 |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 23         |
    And the groups ancestors are computed
    And the database has the following table 'attempts':
      | id | participant_id |
      | 0  | 11             |
      | 1  | 11             |
    And the database has the following table 'results':
      | attempt_id | item_id | participant_id |
      | 0          | 200     | 11             |
      | 1          | 200     | 11             |
    And the database has the following table 'answers':
      | id | author_id | participant_id | attempt_id | item_id | type       | state   | created_at          |
      | 1  | 11        | 11             | 0          | 200     | Submission | Current | 2017-05-29 06:38:38 |
      | 2  | 11        | 11             | 1          | 200     | Submission | Current | 2017-05-29 06:38:38 |
    And the database has the following table 'gradings':
      | answer_id | graded_at           | score |
      | 1         | 2017-05-29 06:38:38 | 100   |
      | 2         | 2017-05-29 06:38:38 | 100   |
    And the database has the following table 'items':
      | id  | type    | teams_editable | no_score | default_language_tag |
      | 200 | Chapter | false          | false    | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 21       | 200     | content_with_descendants |
      | 23       | 200     | none                     |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 200              | 200           |

  Scenario: Wrong group
    Given I am the user with id "23"
    When I send a GET request to "/groups/abc/recent_activity?item_id=200"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Wrong item
    Given I am the user with id "23"
    When I send a GET request to "/groups/13/recent_activity?item_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Should fail when user is not a manager of the group
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/recent_activity?item_id=200"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/groups/13/recent_activity?item_id=200"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should return empty array when user is an admin of the group, but has no access rights to the item
    Given I am the user with id "23"
    When I send a GET request to "/groups/13/recent_activity?item_id=200"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: Should fail when from.id is given, but from.created_at is not
    Given I am the user with id "23"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&from.id=1"
    Then the response code should be 400
    And the response error message should contain "All 'from' parameters (from.created_at, from.id) or none of them must be present"

  Scenario: Should fail when from.created_at is given, but from.id is not
    Given I am the user with id "23"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&from.created_at=2017-05-30T06:38:38Z"
    Then the response code should be 400
    And the response error message should contain "All 'from' parameters (from.created_at, from.id) or none of them must be present"
