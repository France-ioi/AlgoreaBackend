Feature: Update thread - robustness
  Background:
    Given the database has the following table 'groups':
      | id  | name           | type  |
      | 1   | john           | User  |
      | 2   | manager        | User  |
      | 3   | jack           | User  |
      | 4   | managernowatch | User  |
      | 5   | jess           | User  |
      | 10  | Class          | Class |
      | 11  | School         | Class |
      | 12  | Region         | Class |
      | 20  | Group          | Class |
      | 40  | Group          | Class |
      | 50  | Group          | Class |
      | 60  | Group          | Class |
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
      | 12              | 11             |
      | 11              | 10             |
      | 10              | 3              |
      | 10              | 4              |
      | 10              | 5              |
      | 20              | 2              |
      | 20              | 3              |
      | 40              | 3              |
      | 40              | 4              |
      | 50              | 4              |
      | 60              | 3              |
      | 100             | 1              |
      | 300             | 1              |
      | 310             | 3              |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 12       | 2          | true              |
      | 12       | 4          | false             |
    And the database has the following table 'items':
      | id   | default_language_tag | type    |
      | 5    | en                   | Task    |
      | 10   | en                   | Task    |
      | 20   | en                   | Task    |
      | 30   | en                   | Task    |
      | 40   | en                   | Task    |
      | 50   | en                   | Task    |
      | 60   | en                   | Task    |
      | 70   | en                   | Task    |
      | 80   | en                   | Task    |
      | 90   | en                   | Task    |
      | 100  | en                   | Task    |
      | 110  | en                   | Task    |
      | 120  | en                   | Task    |
      | 130  | en                   | Task    |
      | 140  | en                   | Task    |
      | 150  | en                   | Task    |
      | 160  | en                   | Task    |
      | 170  | en                   | Task    |
      | 180  | en                   | Task    |
      | 300  | en                   | Task    |
      | 310  | en                   | Task    |
      | 320  | en                   | Task    |
      | 330  | en                   | Task    |
      | 340  | en                   | Task    |
      | 350  | en                   | Task    |
      | 360  | en                   | Task    |
      | 370  | en                   | Task    |
      | 380  | en                   | Task    |
      | 390  | en                   | Task    |
      | 400  | en                   | Task    |
      | 410  | en                   | Task    |
      | 420  | en                   | Task    |
      | 430  | en                   | Task    |
      | 1004 | en                   | Task    |
      | 1005 | en                   | Task    |
      | 1006 | en                   | Task    |
      | 1007 | en                   | Task    |
      | 3000 | en                   | Chapter |
      | 3001 | en                   | Chapter |
      | 3002 | en                   | Chapter |
      | 3003 | en                   | Chapter |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | request_help_propagation | child_order |
      | 3000           | 3001          | 1                        | 1           |
      | 3000           | 1004          | 1                        | 2           |
      | 3001           | 3002          | 1                        | 1           |
      | 3001           | 3003          | 1                        | 2           |
      | 3001           | 1005          | 0                        | 3           |
      | 3002           | 1006          | 0                        | 1           |
      | 3003           | 2004          | 0                        | 1           |
      | 2004           | 1007          | 1                        | 1           |
      | 2004           | 1008          | 0                        | 2           |
    And the database has the following table 'permissions_granted':
      | group_id | source_group_id | item_id | can_request_help_to |
      | 100      | 100             | 240     | 100                 |
      | 100      | 100             | 250     | 100                 |
      | 100      | 100             | 260     | 100                 |
      | 100      | 100             | 270     | 100                 |
      | 12       | 3               | 3000    | 10                  |
      | 12       | 3               | 3001    | 12                  |
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
      | 4        | 160     | none                |
      | 4        | 170     | answer_with_grant   |
      | 4        | 180     | answer              |
      | 2        | 320     | answer              |
      | 4        | 330     | answer              |
      | 4        | 340     | answer_with_grant   |
      | 4        | 350     | answer              |
      | 4        | 360     | answer_with_grant   |
      | 4        | 430     | answer_with_grant   |
      | 4        | 440     | answer              |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | validated_at        |
      | 0          | 1              | 40      | 2020-01-01 00:00:00 |
      | 0          | 1              | 160     | 2020-01-01 00:00:00 |
      | 0          | 4              | 430     | 2020-01-01 00:00:00 |
      | 0          | 4              | 440     | 2020-01-01 00:00:00 |
    And the database has the following table 'threads':
      | item_id | participant_id | status                  | helper_group_id | latest_update_at    |
      | 10      | 1              | waiting_for_trainer     | 1               | 2020-01-01 00:00:00 |
      | 20      | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 |
      | 30      | 3              | waiting_for_participant | 20              | 2020-01-01 00:00:00 |
      | 40      | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 |
      | 50      | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 |
      | 60      | 1              | waiting_for_participant | 1               | 2020-01-01 00:00:00 |
      | 70      | 3              | closed                  | 10              | 2020-01-01 00:00:00 |
      | 80      | 3              | closed                  | 10              | 2020-01-01 00:00:00 |
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
      | 320     | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 |
      | 330     | 3              | waiting_for_participant | 20              | 2020-01-01 00:00:00 |
      | 340     | 3              | waiting_for_trainer     | 20              | 2020-01-01 00:00:00 |
      | 350     | 3              | closed                  | 20              | 2020-01-01 00:00:00 |
      | 360     | 3              | closed                  | 20              | 2020-01-01 00:00:00 |
      | 370     | 3              | waiting_for_participant | 10              | 2020-01-01 00:00:00 |
      | 380     | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 |
      | 390     | 3              | closed                  | 10              | 2020-01-01 00:00:00 |
      | 400     | 3              | closed                  | 10              | 2020-01-01 00:00:00 |
      | 420     | 3              | closed                  | 10              | 2020-01-01 00:00:00 |
      | 430     | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 |
      | 1004    | 3              | closed                  | 10              | 2020-01-01 00:00:00 |
      | 1005    | 3              | closed                  | 10              | 2020-01-01 00:00:00 |
      | 1006    | 3              | closed                  | 10              | 2020-01-01 00:00:00 |

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

  Scenario: The item should exist
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
    Given I am the user with id "1"
    When I send a PUT request to "/items/5/participant/404/thread" with the following body:
      """
      {
        "status": "waiting_for_trainer",
        "helper_group_id": 10
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
      Should return access error when the status is not set and not part of the helper group
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

  Scenario Outline: Participant of a thread should not be able to switch from non-open to open if not allowed to request help on the item
    Given I am the user with id "3"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>",
        "helper_group_id": 11
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
        "helper_group_id": ["the group must be descendant of a group the participant can request help to"]
      }
    }
    """
    Examples:
      | item_id | status                  | comment                                |
      | 70      | waiting_for_trainer     |                                        |
      | 80      | waiting_for_participant |                                        |
      | 90      | waiting_for_trainer     | Thread doesn't exist yet (not_started) |
      | 100     | waiting_for_participant | Thread doesn't exist yet (not_started) |
      | 1004    | waiting_for_participant |                                        |
      | 1005    | waiting_for_participant |                                        |
      | 1006    | waiting_for_participant |                                        |
      | 1007    | waiting_for_participant | Thread doesn't exist yet (not_started) |
      | 1008    | waiting_for_participant | Thread doesn't exist yet (not_started) |

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

  Scenario Outline: Cannot switch between open status if thread open but not a part of the helper group, and cannot watch participant
    Given I am the user with id "1"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>"
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    Examples:
      | item_id | status                  | comment             |
      | 150     | waiting_for_trainer     | can_watch >= answer |
      | 160     | waiting_for_participant | item validated      |

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

  Scenario: A user who can_watch >= answer on the item and can_watch the participant should not be able to close a thread
    Given I am the user with id "2"
    When I send a PUT request to "/items/320/participant/3/thread" with the following body:
      """
      {
        "status": "closed"
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario Outline: A user not part of the helper group with can_watch >= answer on the item cannot switch a thread to an open status
    Given I am the user with id "4"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>",
        "helper_group_id": 10
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    Examples:
      | item_id | status                  | comment                         |
      | 330     | waiting_for_trainer     | Was open already: switch status |
      | 340     | waiting_for_participant | Was open already: switch status |
      | 350     | waiting_for_trainer     | Was closed                      |
      | 360     | waiting_for_participant | Was closed                      |

  Scenario Outline: A user with can_watch_members on the participant cannot switch a thread to an open status
    Given I am the user with id "2"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>",
        "helper_group_id": 10
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    Examples:
      | item_id | status                  | comment                         |
      | 370     | waiting_for_trainer     | Was open already: switch status |
      | 380     | waiting_for_participant | Was open already: switch status |
      | 390     | waiting_for_trainer     | Was closed                      |
      | 400     | waiting_for_participant | Was closed                      |

  Scenario Outline: A user cannot write in a thread that is not open
    Given I am the user with id "3"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "message_count": 1
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    Examples:
      | item_id | comment               |
      | 410     | Thread does not exist |
      | 420     | Thread closed         |

  Scenario: If participant is the user and helper_group_id is given, it must be a descendant or a group he "can_request_help_to"
    Given I am the user with id "3"
    When I send a PUT request to "/items/430/participant/3/thread" with the following body:
      """
      {
        "helper_group_id": 60
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
        "helper_group_id": ["the group must be descendant of a group the participant can request help to"]
      }
    }
    """
