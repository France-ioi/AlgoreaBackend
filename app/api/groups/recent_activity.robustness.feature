Feature: Get recent activity for group_id and item_id - robustness
  Background:
    Given the database has the following table 'users':
      | id | login   | temp_user | self_group_id | owned_group_id | first_name  | last_name |
      | 1  | someone | 0         | 21            | 22             | Bill        | Clinton   |
      | 2  | user    | 0         | 11            | 12             | John        | Doe       |
      | 3  | owner   | 0         | 23            | 24             | Jean-Michel | Blanquer  |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self | version |
      | 75 | 24                | 13             | 0       | 0       |
      | 76 | 13                | 11             | 0       | 0       |
      | 77 | 22                | 11             | 0       | 0       |
      | 78 | 21                | 21             | 1       | 0       |
      | 79 | 23                | 23             | 1       | 0       |
    And the database has the following table 'users_answers':
      | id | user_id | item_id | attempt_id | name             | type       | state   | lang_prog | submission_date     | score | validated |
      | 1  | 2       | 200     | 100        | My answer        | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | true      |
      | 2  | 2       | 200     | 101        | My second anwser | Submission | Current | python    | 2017-05-29 06:38:38 | 100   | true      |
    And the database has the following table 'items':
      | id  | type     | teams_editable | no_score | unlocked_item_ids | transparent_folder | version |
      | 200 | Category | false          | false    | 1234,2345         | true               | 0       |
    And the database has the following table 'groups_items':
      | id | group_id | item_id | cached_full_access_date | cached_partial_access_date | cached_grayed_access_date | creator_user_id | version |
      | 43 | 21       | 200     | 2017-05-29 06:38:38     | 2017-05-29 06:38:38        | 2017-05-29 06:38:38       | 0               | 0       |
      | 44 | 23       | 200     | 2037-05-29 06:38:38     | 2037-05-29 06:38:38        | 2037-05-29 06:38:38       | 0               | 0       |
    And the database has the following table 'items_ancestors':
      | id | ancestor_item_id | child_item_id | version |
      | 1  | 200              | 200           | 0       |

  Scenario: Wrong group
    Given I am the user with id "3"
    When I send a GET request to "/groups/abc/recent_activity?item_id=200"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Wrong item
    Given I am the user with id "3"
    When I send a GET request to "/groups/13/recent_activity?item_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Should fail when user is not an admin of the group
    Given I am the user with id "1"
    When I send a GET request to "/groups/13/recent_activity?item_id=200"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail when user doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/groups/13/recent_activity?item_id=200"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Should return empty array when user is an admin of the group, but has no access rights to the item
    Given I am the user with id "3"
    When I send a GET request to "/groups/13/recent_activity?item_id=200"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: Should fail when from.id is given, but from.submission_date is not
    Given I am the user with id "3"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&from.id=1"
    Then the response code should be 400
    And the response error message should contain "All 'from' parameters (from.submission_date, from.id) or none of them must be present"

  Scenario: Should fail when from.submission_date is given, but from.id is not
    Given I am the user with id "3"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&from.submission_date=2017-05-30T06:38:38Z"
    Then the response code should be 400
    And the response error message should contain "All 'from' parameters (from.submission_date, from.id) or none of them must be present"
