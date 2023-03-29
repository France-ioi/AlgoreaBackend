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
      | 10              | 2              |
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
      | 12       | 4          | false             |
    And the database has the following table 'items':
      | id   | default_language_tag | type    |
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
    And the database has the following table 'threads':
      | item_id | participant_id | status | helper_group_id | latest_update_at |

  Scenario: Should be logged in
    When I send a PUT request to "/items/@SomeItem/participant/@RichardFeynman/thread"
    Then the response code should be 401
    And the response error message should contain "No access token provided"

  Scenario: The item_id parameter should be an int64
    Given I am RichardFeynman
    When I send a PUT request to "/items/aaa/participant/@RichardFeynman/thread"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: The participant_id parameter should be an int64
    Given I am RichardFeynman
    When I send a PUT request to "/items/@SomeItem/participant/aaa/thread"
    Then the response code should be 400
    And the response error message should contain "Wrong value for participant_id (should be int64)"

  Scenario: Either status, helper_group_id, message_count or message_count_increment must be given
    Given I am RichardFeynman
    And there is a thread with "item_id=@SomeItem,participant_id=@RichardFeynman"
    When I send a PUT request to "/items/@SomeItem/participant/@RichardFeynman/thread" with the following body:
      """
      {}
      """
    Then the response code should be 400
    And the response error message should contain "Either status, helper_group_id, message_count or message_count_increment must be given"

  Scenario: The item should exist
    Given I am RichardFeynman
    When I send a PUT request to "/items/@NonExistentItem/participant/@RichardFeynman/thread" with the following body:
      """
      {
        "message_count": 42
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The participant should exist
    Given I am RichardFeynman
    When I send a PUT request to "/items/@SomeItem/participant/@NonExistentParticipant/thread" with the following body:
      """
      {
        "status": "waiting_for_trainer",
        "helper_group_id": @SomeItem
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

  Scenario Outline: Cannot set status to a wrong value
    Given I am the user with id "3"
    And there is a thread with "item_id=25,participant_id=3"
    When I send a PUT request to "/items/25/participant/3/thread" with the following body:
      """
      {
        "status": "<status>",
        "message_count": 1
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
            "status": ["status must be one of [waiting_for_participant waiting_for_trainer closed]"]
          }
        }
        """
    Examples:
      | status         |
      |                |
      | not_started    |
      | invalid_status |

  # To write on a thread, a user must fulfill either of those conditions:
  #  (1) be the participant of the thread
  #  (2) have can_watch>=answer permission on the item AND can_watch_members on the participant
  #  (3) be part of the group the participant has requested help to AND either have can_watch>=answer on the item
  #    OR have validated the item.
  Scenario: >
  Should return access error when the status is not set and
  "can write to thread" condition (2) is not met: can_watch>=answer not met
    Given I am the user with id "2"
    And there is a thread with "item_id=20,participant_id=3"
    And I can watch result on item with id "20"
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
    And I can watch answer on item with id "30"
    And there is a thread with "item_id=30,participant_id=3"
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
    Given I am RichardFeynman
    And I have validated the item with id "40"
    And I can watch answer on item with id "40"
    And there is a thread with "item_id=40,participant_id=3"
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
    And there is a thread with "item_id=50,participant_id=3"
    When I send a PUT request to "/items/50/participant/3/thread" with the following body:
      """
      {
        "message_count": 42
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: message_count should be positive
    Given I am RichardFeynman
    And there is a thread with "item_id=60,participant_id=@RichardFeynman"
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
    Given I am RichardFeynman
    And there is a thread with "item_id=60,participant_id=@RichardFeynman"
    When I send a PUT request to "/items/60/participant/@RichardFeynman/thread" with the following body:
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

  Scenario Outline: Participant of a thread should not be able to switch from non-open to open if not allowed to request help on the item when thread exists
    Given I am the user with id "3"
    And I can watch <can_watch> on item with id "<item_id>"
    And there is a thread with "item_id=<item_id>,participant_id=3,status=closed,helper_group_id=10"
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
      | item_id | can_watch | status                  |
      | 70      | answer    | waiting_for_trainer     |
      | 80      | none      | waiting_for_participant |
      | 1004    | none      | waiting_for_participant |
      | 1005    | none      | waiting_for_participant |
      | 1006    | none      | waiting_for_participant |

  Scenario Outline: Participant of a thread should not be able to switch from non-open to open if not allowed to request help on the item when thread doesn't exists
    Given I am the user with id "3"
    And I can watch <can_watch> on item with id "<item_id>"
    And there is no thread with "item_id=<item_id>,participant_id=3"
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
      | item_id | can_watch | status                  |
      | 90      | none      | waiting_for_trainer     |
      | 100     | none      | waiting_for_participant |
      | 1007    | none      | waiting_for_participant |
      | 1008    | none      | waiting_for_participant |

  Scenario Outline: Should not switch to open if can_watch_members on the participant but can_watch<answer when thread exists
    Given I am the user with id "2"
    And I can watch the participant with id "3"
    And I can watch <can_watch> on item with id "<item_id>"
    And there is a thread with "item_id=<item_id>,participant_id=3,status=closed"
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
      | item_id | can_watch | status                  |
      | 110     | none      | waiting_for_trainer     |
      | 120     | result    | waiting_for_participant |

  Scenario Outline: Should not switch to open if can_watch_members on the participant but can_watch<answer when thread doesn't exists
    Given I am the user with id "2"
    And I can watch the participant with id "3"
    And I can watch <can_watch> on item with id "<item_id>"
    And there is no thread with "item_id=<item_id>,participant_id=3"
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
      | item_id | can_watch | status                  |
      | 130     | result    | waiting_for_trainer     |
      | 140     | none      | waiting_for_participant |

  Scenario Outline: Should not switch to open if can_watch>=answer but cannot watch_members on the participant when thread exists
    Given I am the user with id "4"
    And I can watch <can_watch> on item with id "<item_id>"
    And there is a thread with "item_id=<item_id>,participant_id=3,status=closed"
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
      | item_id | can_watch         | status                  |
      | 110     | answer            | waiting_for_trainer     |
      | 120     | answer_with_grant | waiting_for_participant |

  Scenario Outline: Should not switch to open if can_watch>=answer but cannot watch_members on the participant when thread doesn't exists
    Given I am the user with id "4"
    And I can watch <can_watch> on item with id "<item_id>"
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
      | item_id | can_watch         | status                  |
      | 130     | answer            | waiting_for_trainer     |
      | 140     | answer_with_grant | waiting_for_participant |

  Scenario: Cannot switch between open status if only can_watch>answer but not a part of the helper group, and cannot watch participant
    Given I am RichardFeynman
    And I can watch answer on item with id "150"
    And there is a thread with "item_id=150,participant_id=3,status=waiting_for_participant"
    When I send a PUT request to "/items/150/participant/3/thread" with the following body:
      """
      {
        "status": "waiting_for_trainer"
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Cannot switch between open status if only item validated but not a part of the helper group, and cannot watch participant
    Given I am RichardFeynman
    And I have validated the item with id "160"
    And there is a thread with "item_id=160,participant_id=3,status=waiting_for_trainer"
    When I send a PUT request to "/items/160/participant/3/thread" with the following body:
      """
      {
        "status": "waiting_for_participant"
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario Outline: If switching to an open status from a non-open status, helper_group_id must be given when thread exists
    Given I am RichardFeynman
    And there is a thread with "item_id=<item_id>,participant_id=@RichardFeynman,status=closed"
    When I send a PUT request to "/items/<item_id>/participant/@RichardFeynman/thread" with the following body:
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
      | item_id | status                  |
      | 200     | waiting_for_trainer     |
      | 210     | waiting_for_participant |

  Scenario Outline: If switching to an open status from a non-open status, helper_group_id must be given when thread doesn't exists
    Given I am RichardFeynman
    When I send a PUT request to "/items/<item_id>/participant/@RichardFeynman/thread" with the following body:
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
      | item_id | status                  |
      | 220     | waiting_for_trainer     |
      | 230     | waiting_for_participant |

  Scenario Outline: If status is already "closed" and not changing status OR if switching to status "closed": helper_group_id must not be given when thread exists
    Given I am the user with id "1"
    And there is a thread with "item_id=<item_id>,participant_id=1,status=<old_status>"
    When I send a PUT request to "/items/<item_id>/participant/@RichardFeynman/thread" with the following body:
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
      | item_id | old_status              | status |
      | 240     | closed                  | closed |
      | 250     | waiting_for_participant | closed |
      | 260     | waiting_for_trainer     | closed |

  Scenario Outline: If status is already "closed" and not changing status OR if switching to status "closed": helper_group_id must not be given when thread doesn't exists
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
      | item_id | status |
      | 270     | closed |

  Scenario: helper_group_id is visible to current-user but not to participant
    Given I am the user with id "1"
    And there is a thread with "item_id=300,participant_id=3"
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
    Given I am RichardFeynman
    And there is a thread with "item_id=310,participant_id=3"
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
    And I can watch answer on item with id "320"
    And there is a thread with "item_id=320,participant_id=3"
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
    And I can watch <can_watch> on item with id "<item_id>"
    And there is a thread with "item_id=<item_id>,participant_id=3,helper_group_id=20,status=<old_status>"
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
      | item_id | can_watch         | old_status              | status                  |
      | 330     | answer            | waiting_for_participant | waiting_for_trainer     |
      | 340     | answer_with_grant | waiting_for_trainer     | waiting_for_participant |
      | 350     | answer            | closed                  | waiting_for_trainer     |
      | 360     | answer_with_grant | closed                  | waiting_for_participant |

  Scenario Outline: A user with can_watch_members on the participant cannot switch a thread to an open status
    Given I am the user with id "2"
    And I can watch the participant with id "3"
    And there is a thread with "item_id=<item_id>,participant_id=3,status=<old_status>"
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
      | item_id | old_status              | status                  |
      | 370     | waiting_for_participant | waiting_for_trainer     |
      | 380     | waiting_for_trainer     | waiting_for_participant |
      | 390     | closed                  | waiting_for_trainer     |
      | 400     | closed                  | waiting_for_participant |

  Scenario: A user cannot write in a thread that does not exists
    Given I am the user with id "3"
    And there is no thread with "item_id=410,participant_id=3"
    When I send a PUT request to "/items/410/participant/3/thread" with the following body:
      """
      {
        "message_count": 1
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: A user cannot write in a thread that is closed
    Given I am the user with id "3"
    And there is a thread with "item_id=420,participant_id=3,status=closed"
    When I send a PUT request to "/items/420/participant/3/thread" with the following body:
      """
      {
        "message_count": 1
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: If participant is the user and helper_group_id is given, it must be a descendant or a group he "can_request_help_to"
    Given I am the user with id "3"
    And I can watch answer_with_grant on item with id "430"
    And I have validated the item with id "430"
    And there is a thread with "item_id=430,participant_id=3"
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
