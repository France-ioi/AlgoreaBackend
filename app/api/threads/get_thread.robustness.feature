Feature: Get thread - robustness
  Background:
    Given the database has the following table "groups":
      | id | name       | type  |
      | 10 | Group      | Class |
      | 20 | Help group | Class |
    And the database has the following users:
      | group_id | login   |
      | 1        | john    |
      | 2        | manager |
      | 3        | jack    |
      | 4        | helper  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 20              | 4              |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | default_language_tag |
      | 10 | en                   |
      | 20 | en                   |
      | 40 | en                   |
      | 50 | en                   |
      | 60 | en                   |
      | 70 | en                   |
      | 80 | en                   |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | validated_at        |
      | 0          | 4              | 20      | 2020-01-01 00:00:00 |
      | 0          | 4              | 40      | null                |
      | 0          | 4              | 60      | 2020-01-01 00:00:00 |
      | 0          | 4              | 70      | 2020-01-01 00:00:00 |
      | 0          | 4              | 80      | null                |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 1        | 10      | info               |
      | 2        | 10      | content            |
    And the database has the following table "threads":
      | item_id | participant_id | status                  | helper_group_id | latest_update_at    |
      | 10      | 1              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 |
      | 20      | 3              | closed                  | 20              | 2020-01-05 00:00:00 |
      | 40      | 3              | closed                  | 20              | 2020-01-20 00:00:00 |
      | 50      | 3              | closed                  | 20              | 2020-01-20 00:00:00 |
      | 60      | 3              | closed                  | 10              | 2020-01-20 00:00:00 |
      | 70      | 3              | waiting_for_trainer     | 10              | 2020-01-20 00:00:00 |
      | 80      | 3              | waiting_for_participant | 20              | 2020-01-20 00:00:00 |
    And the DB time now is "2020-01-20 00:00:00"

  Scenario: Should be logged
    When I send a GET request to "/items/10/participant/1/thread"
    Then the response code should be 401
    And the response error message should contain "No access token provided"

  Scenario: The item_id parameter should be an int64
    Given I am the user with id "1"
    When I send a GET request to "/items/aaa/participant/1/thread"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: The participant_id parameter should be an int64
    Given I am the user with id "1"
    When I send a GET request to "/items/10/participant/aaa/thread"
    Then the response code should be 400
    And the response error message should contain "Wrong value for participant_id (should be int64)"

  Scenario: The item should exist
    Given I am the user with id "1"
    When I send a GET request to "/items/404/participant/1/thread"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: When current-user is the thread participant, it should have "can_view >= content" on the item
    Given I am the user with id "1"
    When I send a GET request to "/items/10/participant/1/thread"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: >
      Should be forbidden when
      the current user is a descendant of the thread helper group
      and the current user has a validated result on the item,
      and the current user has can_watch=result permission on the item,
      but the thread has been closed for more than 2 weeks
    Given I am the user with id "4"
    And I have the watch permission set to "result" on the item 20
    When I send a GET request to "/items/20/participant/3/thread"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: >
      Should be forbidden when
      the current user is a descendant of the thread helper group
      and the thread has been closed for less than 2 weeks,
      and the current user has can_watch=result permission on the item,
      but the current user doesn't have a validated result on the item
    Given I am the user with id "4"
    And I have the watch permission set to "result" on the item 40
    When I send a GET request to "/items/40/participant/3/thread"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: >
      Should be forbidden when
      the current user is a descendant of the thread helper group,
      and the thread has been closed for less than 2 weeks,
      and the current user has can_watch=result permission on the item,
      but the current user doesn't have an entry in results for the item
    Given I am the user with id "4"
    And I have the watch permission set to "result" on the item 50
    When I send a GET request to "/items/50/participant/3/thread"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: >
      Should be forbidden when
      the current user has a validated result on the item,
      and the user has can_watch=result permission on the item,
      and the thread has been closed for less than 2 weeks,
      but the user is not a descendant of the thread helper group
    Given I am the user with id "4"
    And I have the watch permission set to "result" on the item 60
    When I send a GET request to "/items/60/participant/3/thread"
      Then the response code should be 403
      And the response error message should contain "Insufficient access rights"

  Scenario: >
      Should be forbidden when
      the thread is open,
      and the current user has a validated result on the item,
      and the current user has can_watch=result permission on the item,
      but the current user is not a descendant of the thread helper group
    Given I am the user with id "4"
    And I have the watch permission set to "result" on the item 70
    When I send a GET request to "/items/70/participant/3/thread"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: >
      Should be forbidden when
      the thread is open,
      and the current user is a descendant of the thread helper group,
      and the current user has can_watch=result permission on the item,
      but the current user doesn't have a validated result on the item
    Given I am the user with id "4"
    And I have the watch permission set to "result" on the item 80
    When I send a GET request to "/items/80/participant/3/thread"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: >
      Should be forbidden when
      the thread is open,
      and the current user is a descendant of the thread helper group,
      and the current user has a validated result on the item
      but the current user has can_watch<result permission on the item
    Given I am the user with id "4"
    And I have a validated result on the item 80
    And I have the watch permission set to "none" on the item 80
    When I send a GET request to "/items/80/participant/3/thread"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
