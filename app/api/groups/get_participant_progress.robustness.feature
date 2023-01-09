Feature: Display the current progress of a participant on children of an item (groupParticipantProgress) - robustness
  Background:
    Given the database has the following users:
      | login | group_id |
      | owner | 21       |
      | user  | 11       |
    And the database has the following table 'groups':
      | id | type |
      | 13 | Base |
      | 14 | Team |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 14              | 21             |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 13       | 11         | false             |
      | 13       | 21         | true              |
      | 14       | 21         | true              |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id  | type    | default_language_tag |
      | 200 | Task    | fr                   |
      | 210 | Chapter | fr                   |
      | 211 | Task    | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       | can_watch_generated |
      | 11       | 210     | content                  | result              |
      | 14       | 200     | info                     | answer              |
      | 20       | 210     | none                     | answer              |
      | 21       | 200     | content                  | answer              |
      | 21       | 210     | content                  | answer_with_grant   |
      | 21       | 211     | info                     | none                |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 200            | 210           | 0           |
      | 200            | 220           | 1           |
      | 210            | 211           | 0           |

  Scenario: User is not able to watch group members
    Given I am the user with id "11"
    When I send a GET request to "/items/210/participant-progress?watched_group_id=13"
    Then the response code should be 403
    And the response error message should contain "No rights to watch for watched_group_id"

  Scenario: watched_group_id is incorrect
    Given I am the user with id "11"
    When I send a GET request to "/items/210/participant-progress?watched_group_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for watched_group_id (should be int64)"

  Scenario: watched_group_id is not User/Team
    Given I am the user with id "21"
    When I send a GET request to "/items/210/participant-progress?watched_group_id=13"
    Then the response code should be 403
    And the response error message should contain "Watched group should be a user or a team"

  Scenario: Both as_team_id and watched_group_id are given
    Given I am the user with id "21"
    When I send a GET request to "/items/210/participant-progress?watched_group_id=13&as_team_id=14"
    Then the response code should be 400
    And the response error message should contain "Only one of as_team_id and watched_group_id can be given"

  Scenario: item_id is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/items/112341234123341234123431241234132412341234312141/participant-progress?watched_group_id=13"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Not enough permissions to watch results on item_id
    Given I am the user with id "21"
    When I send a GET request to "/items/211/participant-progress?watched_group_id=14"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: can_view < content on item_id for a user
    Given I am the user with id "21"
    When I send a GET request to "/items/211/participant-progress"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: can_view < content on item_id for a team
    Given I am the user with id "21"
    When I send a GET request to "/items/200/participant-progress?as_team_id=14"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User not found
    Given I am the user with id "404"
    When I send a GET request to "/items/210/participant-progress?watched_group_id=13"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
