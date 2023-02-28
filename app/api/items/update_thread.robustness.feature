Feature: Update thread - robustness
  Background:
    Given the database has the following table 'groups':
      | id  | name           | type  |
      | 1   | john           | User  |
      | 2   | manager        | User  |
      | 3   | jack           | User  |
      | 4   | managernowatch | User  |
      | 5   | jess           | User  |
      | 10  | Group          | Class |
      | 20  | Group          | Class |
      | 40  | Group          | Class |
      | 100 | Group          | Class |
      | 300 | Group          | Class |
      | 310 | Group          | Class |
    And the database has the following table 'users':
      | login          | group_id |
      | john           | 1        |
      | manager        | 2        |
      | jack           | 3        |
      | managernowatch | 4        |
      | jess           | 5        |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 10              | 3              |
      | 10              | 5              |
      | 20              | 2              |
      | 20              | 3              |
      | 40              | 3              |
      | 40              | 4              |
      | 100             | 1              |
      | 300             | 1              |
      | 310             | 3              |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 10       | 2          | true              |
      | 10       | 4          | false             |
    And the database has the following table 'items':
      | id  | default_language_tag |
      | 10  | en                   |
      | 20  | en                   |
      | 30  | en                   |
      | 40  | en                   |
      | 50  | en                   |
      | 60  | en                   |
      | 70  | en                   |
      | 80  | en                   |
      | 90  | en                   |
      | 100 | en                   |
      | 110 | en                   |
      | 120 | en                   |
      | 130 | en                   |
      | 140 | en                   |
      | 150 | en                   |
      | 160 | en                   |
      | 170 | en                   |
      | 180 | en                   |
      | 300 | en                   |
      | 310 | en                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_watch_generated |
      | 2        | 20      | result              |
      | 4        | 30      | answer              |
      | 1        | 40      | answer              |
      | 1        | 70      | answer              |
      | 2        | 110     | none                |
      | 2        | 120     | result              |
      | 2        | 130     | result              |
      | 4        | 150     | answer              |
      | 4        | 160     | answer_with_grant   |
      | 4        | 170     | answer_with_grant   |
      | 4        | 180     | answer              |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | validated_at        |
      | 0          | 1              | 40      | 2020-01-01 00:00:00 |
    And the database has the following table 'threads':
      | item_id | participant_id | status                  | helper_group_id | latest_update_at    |
      | 10      | 1              | waiting_for_trainer     | 1               | 2020-01-01 00:00:00 |
      | 20      | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 |
      | 30      | 3              | waiting_for_participant | 10              | 2020-01-01 00:00:00 |
      | 40      | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 |
      | 50      | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 |
      | 60      | 1              | waiting_for_participant | 1               | 2020-01-01 00:00:00 |
      | 70      | 1              | waiting_for_trainer     | 1               | 2020-01-01 00:00:00 |
      | 80      | 1              | waiting_for_trainer     | 1               | 2020-01-01 00:00:00 |
      | 110     | 3              | closed                  | 3               | 2020-01-01 00:00:00 |
      | 120     | 3              | closed                  | 3               | 2020-01-01 00:00:00 |
      | 150     | 3              | waiting_for_participant | 10              | 2020-01-01 00:00:00 |
      | 160     | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 |
      | 200     | 1              | closed                  | 1               | 2020-01-01 00:00:00 |
      | 210     | 1              | closed                  | 1               | 2020-01-01 00:00:00 |
      | 240     | 1              | closed                  | 1               | 2020-01-01 00:00:00 |
      | 250     | 1              | waiting_for_participant | 1               | 2020-01-01 00:00:00 |
      | 260     | 1              | waiting_for_trainer     | 1               | 2020-01-01 00:00:00 |
      | 300     | 3              | waiting_for_trainer     | 3               | 2020-01-01 00:00:00 |
      | 310     | 3              | waiting_for_trainer     | 3               | 2020-01-01 00:00:00 |

  Scenario: Should be logged
    When I send a PUT request to "/items/10/participant/1/thread"
    Then the response code should be 401
    And the response error message should contain "No access token provided"

  Scenario: The item_id parameter should be an int64
    Given I am the user with id "1"
    When I send a PUT request to "/items/aaa/participant/1/thread"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: The participant_id parameter should be an int64
    Given I am the user with id "1"
    When I send a PUT request to "/items/10/participant/aaa/thread"
    Then the response code should be 400
    And the response error message should contain "Wrong value for participant_id (should be int64)"

  Scenario: Either status, helper_group_id, message_count or message_count_increment must be given
      Given I am the user with id "1"
      When I send a PUT request to "/items/10/participant/1/thread" with the following body:
      """
      {}
      """
      Then the response code should be 400
      And the response error message should contain "Either status, helper_group_id, message_count or message_count_increment must be given"

  # TODO: 1) Test should be added for "userCanWrite", that the user have access to item (see doc)
  # TODO: 2) If participant == user, existence should be tested via canRequest Help
  Scenario: TODO: The item should exist
    Given I am the user with id "1"
    When I send a PUT request to "/items/404/participant/1/thread" with the following body:
      """
      {
        "message_count": 42
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The participant should exist

  # To write on a thread, a user must fulfill either of those conditions:
  #  (1) be the participant of the thread
  #  (2) have can_watch>=answer permission on the item AND can_watch_members on the participant
  #  (3) be part of the group the participant has requested help to AND either have can_watch>=answer on the item
  #    OR have validated the item.
  Scenario: >
      Should return access error when the status is not set and
      "can write to thread" condition (2) is not met: can_watch>=answer not met
    Given I am the user with id "2"
    When I send a PUT request to "/items/20/participant/3/thread" with the following body:
      """
      {
        "message_count": 42
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: >
      Should return access error when the status is not set and
      "can write to thread" condition (2) is not met: can_watch_members of the participant
    Given I am the user with id "4"
    When I send a PUT request to "/items/30/participant/3/thread" with the following body:
      """
      {
        "message_count": 42
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: >
      Should return access error when the status is not set and
      "can write to thread" condition (3) is not met: user is not part of the help group
    Given I am the user with id "1"
    When I send a PUT request to "/items/40/participant/3/thread" with the following body:
      """
      {
        "message_count": 42
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: >
      Should return access error when the status is not set and
      "can write to thread" condition (3) is not met: user have neither can_watch>=answer, nor validated the item
    Given I am the user with id "5"
    When I send a PUT request to "/items/50/participant/3/thread" with the following body:
      """
      {
        "message_count": 42
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: message_count should be positive
    Given I am the user with id "1"
    When I send a PUT request to "/items/60/participant/3/thread" with the following body:
      """
      {
        "message_count": -1
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors":{
        "message_count": ["message_count must be 0 or greater"]
      }
    }
    """

  Scenario: Should not contain both message_count and message_count_increment
    Given I am the user with id "1"
    When I send a PUT request to "/items/60/participant/1/thread" with the following body:
      """
      {
        "message_count": 1,
        "message_count_increment": 1
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors":{
        "message_count": ["cannot have both message_count and message_count_increment set"]
      }
    }
    """

    # TODO: Implement when forum permissions are merged
  Scenario Outline: TODO: Participant of a thread should not be able to switch from non-open to open if not allowed to request help on the item
    Given I am the user with id "1"
    When I send a PUT request to "/items/<item_id>/participant/1/thread" with the following body:
      """
      {
        "status": "<status>"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors":{
        "message_count": ["cannot have both message_count and message_count_increment set"]
      }
    }
    """
    Examples:
      | item_id | status                  | comment                                |
      | 70      | waiting_for_trainer     |                                        |
      | 80      | waiting_for_participant |                                        |
      | 90      | waiting_for_trainer     | Thread doesn't exist yet (not_started) |
      | 100     | waiting_for_participant | Thread doesn't exist yet (not_started) |

  Scenario Outline: Should not switch to open if can_watch_members on the participant but can_watch<answer
    Given I am the user with id "2"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>",
        "helper_group_id": 20
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    Examples:
      | item_id | status                  | comment                                |
      | 110     | waiting_for_trainer     |                                        |
      | 120     | waiting_for_participant |                                        |
      | 130     | waiting_for_trainer     | Thread doesn't exist yet (not_started) |
      | 140     | waiting_for_participant | Thread doesn't exist yet (not_started) |

  Scenario Outline: Should not switch to open if can_watch>=answer but cannot watch_members on the participant
    Given I am the user with id "4"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>",
        "helper_group_id": 40
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    Examples:
      | item_id | status                  | comment                                |
      | 110     | waiting_for_trainer     |                                        |
      | 120     | waiting_for_participant |                                        |
      | 130     | waiting_for_trainer     | Thread doesn't exist yet (not_started) |
      | 140     | waiting_for_participant | Thread doesn't exist yet (not_started) |

  Scenario Outline: TODO: Cannot switch between open status if thread open but not a part of the helper group
    Given I am the user with id "1"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors":{
        "message_count": ["cannot have both message_count and message_count_increment set"]
      }
    }
    """
    Examples:
      | item_id | status                  | comment |
      | 150     | waiting_for_trainer     |         |
      | 160     | waiting_for_participant |         |

  Scenario Outline: If switching to an open status from a non-open status, helper_group_id must be given
    Given I am the user with id "1"
    When I send a PUT request to "/items/<item_id>/participant/1/thread" with the following body:
      """
      {
        "status": "<status>"
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors":{
        "status": ["the helper_group_id must be set to switch from a non-open to an open status"]
      }
    }
    """
    Examples:
      | item_id | status                  | comment                                  |
      | 200     | waiting_for_trainer     |                                          |
      | 210     | waiting_for_participant |                                          |
      | 220     | waiting_for_trainer     | # Thread doesn't exist yet (not_started) |
      | 230     | waiting_for_participant | # Thread doesn't exist yet (not_started) |

  Scenario Outline: If status is already "closed" and not changing status OR if switching to status "closed": helper_group_id must not be given
    Given I am the user with id "1"
    When I send a PUT request to "/items/<item_id>/participant/1/thread" with the following body:
      """
      {
        "status": "<status>",
        "helper_group_id": 100
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors":{
        "helper_group_id": ["the helper_group_id must not be given when setting or keeping status to closed"]
      }
    }
    """
    Examples:
      | item_id | status | comment                                  |
      | 240     | closed |                                          |
      | 250     | closed |                                          |
      | 260     | closed |                                          |
      | 270     | closed | # Thread doesn't exist yet (not_started) |

  Scenario: helper_group_id is visible to current-user but not to participant
    Given I am the user with id "1"
    When I send a PUT request to "/items/300/participant/3/thread" with the following body:
      """
      {
        "helper_group_id": 300
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors":{
        "helper_group_id": ["the group must be visible to the current-user and the participant"]
      }
    }
    """

  Scenario: helper_group_id is visible to participant but not to current-user
    Given I am the user with id "1"
    When I send a PUT request to "/items/310/participant/3/thread" with the following body:
      """
      {
        "helper_group_id": 310
      }
      """
    Then the response code should be 400
    And the response body should be, in JSON:
    """
    {
      "success": false,
      "message": "Bad Request",
      "error_text": "Invalid input data",
      "errors":{
        "helper_group_id": ["the group must be visible to the current-user and the participant"]
      }
    }
    """
