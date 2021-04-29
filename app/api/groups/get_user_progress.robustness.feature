Feature: Display the current progress of users on a subset of items (groupUserProgress) - robustness
  Background:
    Given the database has the following users:
      | login | group_id |
      | owner | 21       |
      | user  | 11       |
    And the database has the following table 'groups':
      | id |
      | 13 |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 13       | 11         | false             |
      | 13       | 21         | true              |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id  | type    | default_language_tag |
      | 200 | Course  | fr                   |
      | 210 | Chapter | fr                   |
      | 211 | Task    | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       | can_watch_generated |
      | 11       | 210     | info                     | result              |
      | 20       | 210     | none                     | answer              |
      | 21       | 210     | none                     | answer_with_grant   |
      | 21       | 211     | info                     | none                |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 200            | 210           | 0           |
      | 200            | 220           | 1           |
      | 210            | 211           | 0           |

  Scenario: User is not able to watch group members
    Given I am the user with id "11"
    When I send a GET request to "/groups/13/user-progress?parent_item_ids=210"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Group id is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/abc/user-progress?parent_item_ids=210"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: parent_item_ids is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/user-progress?parent_item_ids=abc,123"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: 'abc', param: 'parent_item_ids')"

  Scenario: Not enough permissions to watch results on parent_item_ids
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/user-progress?parent_item_ids=210,211"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User not found
    Given I am the user with id "404"
    When I send a GET request to "/groups/13/user-progress?parent_item_ids=210"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: sort is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/user-progress?parent_item_ids=210&sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""
