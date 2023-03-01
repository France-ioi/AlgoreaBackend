Feature: Update thread
  Background:
    Given the database has the following table 'groups':
      | id | name    | type  |
      | 1  | john    | User  |
      | 2  | manager | User  |
      | 3  | jack    | User  |
      | 4  | jess    | User  |
      | 10 | Groupe  | Class |
      | 11 | School  | Class |
      | 12 | Region  | Class |
      | 20 | Group   | Class |
      | 30 | Group   | Class |
      | 50 | Group   | Class |
      | 51 | Group   | Class |
    And the database has the following table 'users':
      | login   | group_id |
      | john    | 1        |
      | manager | 2        |
      | jack    | 3        |
      | jess    | 4        |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 10              | 3              |
      | 10              | 4              |
      | 11              | 10             |
      | 12              | 11             |
      | 20              | 2              |
      | 20              | 3              |
      | 30              | 2              |
      | 30              | 3              |
      | 50              | 51             |
      | 51              | 2              |
      | 51              | 3              |
      | 51              | 4              |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 10       | 2          | true              |
    And the database has the following table 'items':
      | id   | default_language_tag | type    |
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
      | 160  | en                   | Task    |
      | 170  | en                   | Task    |
      | 180  | en                   | Task    |
      | 190  | en                   | Task    |
      | 200  | en                   | Task    |
      | 210  | en                   | Task    |
      | 220  | en                   | Task    |
      | 230  | en                   | Task    |
      | 240  | en                   | Task    |
      | 250  | en                   | Task    |
      | 260  | en                   | Task    |
      | 270  | en                   | Task    |
      | 1000 | en                   | Task    |
      | 2000 | en                   | Task    |
      | 2001 | en                   | Task    |
      | 2002 | en                   | Task    |
      | 2003 | en                   | Task    |
      | 3004 | en                   | Chapter |
      | 3005 | en                   | Chapter |
      | 3006 | en                   | Chapter |
      | 2007 | en                   | Task    |
      | 2008 | en                   | Task    |
      | 2009 | en                   | Task    |
      | 2010 | en                   | Task    |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | request_help_propagation | child_order |
      | 3004           | 3005          | 1                        | 1           |
      | 3004           | 2007          | 1                        | 2           |
      | 3005           | 3006          | 1                        | 1           |
      | 3005           | 2008          | 1                        | 2           |
      | 3006           | 2009          | 1                        | 1           |
      | 3006           | 2010          | 1                        | 2           |
    And the database has the following table 'permissions_granted':
      | group_id | source_group_id | item_id | can_request_help_to |
      | 3        | 11              | 270     | 12                  |
      | 3        | 11              | 1000    | 10                   |
      | 3        | 11              | 2000    | 12                  |
      | 3        | 11              | 2001    | 12                  |
      | 11       | 11              | 2002    | 12                  |
      | 11       | 11              | 2003    | 12                  |
      | 12       | 11              | 3004    | 12                  |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_watch_generated |
      | 2        | 20      | answer              |
      | 4        | 30      | answer              |
      | 1        | 100     | none                |
      | 1        | 110     | none                |
      | 2        | 160     | answer              |
      | 2        | 170     | answer_with_grant   |
      | 2        | 180     | answer              |
      | 2        | 190     | answer              |
      | 2        | 200     | answer_with_grant   |
      | 2        | 210     | answer              |
      | 4        | 220     | answer              |
      | 4        | 230     | answer_with_grant   |
      | 2        | 260     | answer              |
      | 3        | 2000    | none                |
      | 3        | 2001    | none                |
      | 11       | 2002    | none                |
      | 11       | 2003    | none                |
      | 11       | 2007    | none                |
      | 11       | 2008    | none                |
      | 11       | 2009    | none                |
      | 11       | 2010    | none                |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | validated_at        |
      | 0          | 4              | 40      | 2020-01-01 00:00:00 |
      | 0          | 4              | 240     | 2020-01-01 00:00:00 |
      | 0          | 4              | 250     | 2020-01-01 00:00:00 |
    And the database has the following table 'threads':
      | item_id | participant_id | status                  | helper_group_id | latest_update_at    | message_count |
      | 10      | 1              | waiting_for_trainer     | 1               | 2020-01-01 00:00:00 | 1             |
      | 20      | 3              | waiting_for_trainer     | 3               | 2020-01-01 00:00:00 | 1             |
      | 30      | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 | 1             |
      | 40      | 3              | waiting_for_trainer     | 10              | 2020-01-01 00:00:00 | 1             |
      | 50      | 1              | waiting_for_trainer     | 1               | 2020-01-01 00:00:00 | 1             |
      | 60      | 1              | waiting_for_trainer     | 1               | 2020-01-01 00:00:00 | 10            |
      | 70      | 1              | waiting_for_trainer     | 1               | 2020-01-01 00:00:00 | 10            |
      | 80      | 1              | waiting_for_trainer     | 1               | 2020-01-01 00:00:00 | 10            |
      | 90      | 1              | waiting_for_trainer     | 1               | 2020-01-01 00:00:00 | 10            |
      | 100     | 3              | waiting_for_trainer     | 3               | 2020-01-01 00:00:00 | 10            |
      | 110     | 3              | waiting_for_participant | 3               | 2020-01-01 00:00:00 | 10            |
      | 160     | 3              | closed                  | 3               | 2020-01-01 00:00:00 | 1             |
      | 170     | 3              | closed                  | 3               | 2020-01-01 00:00:00 | 1             |
      | 180     | 3              | closed                  | 3               | 2020-01-01 00:00:00 | 1             |
      | 190     | 3              | closed                  | 3               | 2020-01-01 00:00:00 | 1             |
      | 220     | 3              | waiting_for_trainer     | 50              | 2020-01-01 00:00:00 | 10            |
      | 230     | 3              | waiting_for_participant | 50              | 2020-01-01 00:00:00 | 10            |
      | 240     | 3              | waiting_for_trainer     | 50              | 2020-01-01 00:00:00 | 10            |
      | 250     | 3              | waiting_for_participant | 50              | 2020-01-01 00:00:00 | 10            |
      | 260     | 3              | waiting_for_participant | 10              | 2020-01-01 00:00:00 | 10            |
      | 2000    | 3              | closed                  | 3               | 2020-01-01 00:00:00 | 10            |
      | 2001    | 3              | closed                  | 11              | 2020-01-01 00:00:00 | 10            |
      | 2007    | 3              | closed                  | 11              | 2020-01-01 00:00:00 | 10            |
      | 2008    | 3              | closed                  | 11              | 2020-01-01 00:00:00 | 10            |
    And the time now is "2022-01-01T00:00:00Z"

  Scenario: Create a thread if it doesn't exist
    Given I am the user with id "3"
    When I send a PUT request to "/items/1000/participant/3/thread" with the following body:
      """
      {
        "status": "waiting_for_trainer",
        "message_count": 1,
        "helper_group_id": 10
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "1000"
    And the table "threads" at item_id "1000" should be:
      | latest_update_at    | message_count | status              | helper_group_id |
      | 2022-01-01 00:00:00 | 1             | waiting_for_trainer | 10               |

  # To write on a thread, a user must fulfill either of those conditions:
  #  (1) be the participant of the thread
  #  (2) have can_watch>=answer permission on the item AND can_watch_members on the participant
  #  (3) be part of the group the participant has requested help to AND either have can_watch>=answer on the item
  #    OR have validated the item.
  Scenario: Can write to thread condition (1) when status is not set
    Given I am the user with id "1"
    When I send a PUT request to "/items/10/participant/1/thread" with the following body:
      """
      {
        "message_count": 1
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "10"
    And the table "threads" at item_id "10" should be:
      | latest_update_at    | message_count |
      | 2022-01-01 00:00:00 | 1             |

  Scenario: Can write to thread condition (2) when status is not set
    Given I am the user with id "2"
    When I send a PUT request to "/items/20/participant/3/thread" with the following body:
      """
      {
        "message_count": 2
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "20"
    And the table "threads" at item_id "20" should be:
      | latest_update_at    | message_count |
      | 2022-01-01 00:00:00 | 2             |

  Scenario: Can write to thread condition (3) with can_watch>=answer on the item, when status is not set
    Given I am the user with id "2"
    When I send a PUT request to "/items/20/participant/3/thread" with the following body:
      """
      {
        "message_count": 3
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "20"
    And the table "threads" at item_id "20" should be:
      | latest_update_at    | message_count |
      | 2022-01-01 00:00:00 | 3             |

  Scenario: Can write to thread test condition (3) with validated item, when status is not set
    Given I am the user with id "4"
    When I send a PUT request to "/items/40/participant/3/thread" with the following body:
      """
      {
        "message_count": 4
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "40"
    And the table "threads" at item_id "40" should be:
      | latest_update_at    | message_count |
      | 2022-01-01 00:00:00 | 4             |

  Scenario: Set message_count to 0
    Given I am the user with id "1"
    When I send a PUT request to "/items/50/participant/1/thread" with the following body:
      """
      {
        "message_count": 0
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "50"
    And the table "threads" at item_id "50" should be:
      | latest_update_at    | message_count |
      | 2022-01-01 00:00:00 | 0             |

  Scenario: Should set message_count to 0 if decrement to a negative value
    Given I am the user with id "1"
    When I send a PUT request to "/items/60/participant/1/thread" with the following body:
      """
      {
        "message_count_increment": -11
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "60"
    And the table "threads" at item_id "60" should be:
      | latest_update_at    | message_count |
      | 2022-01-01 00:00:00 | 0             |

  Scenario Outline: Should increment message_count by message_count_increments
    Given I am the user with id "1"
    When I send a PUT request to "/items/<item_id>/participant/1/thread" with the following body:
      """
      {
        "message_count_increment": <message_count_increment>
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "<item_id>"
    And the table "threads" at item_id "<item_id>" should be:
      | latest_update_at    | message_count   |
      | 2022-01-01 00:00:00 | <message_count> |
    Examples:
      | item_id | message_count_increment | message_count |
      | 70      | 3                       | 13            |
      | 80      | -5                      | 5             |
      | 90      | 0                       | 10            |

  Scenario Outline: Participant of a thread can always switch the thread from open to any other status
    Given I am the user with id "3"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>"
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "<item_id>"
    And the table "threads" at item_id "<item_id>" should be:
      | latest_update_at    | status   |
      | 2022-01-01 00:00:00 | <status> |
    Examples:
      | item_id | status                  |
      | 100     | waiting_for_participant |
      | 100     | closed                  |
      | 110     | waiting_for_trainer     |
      | 110     | closed                  |

  Scenario Outline: Participant of a thread can switch from non-open to open status when allowed to request help on the item
    Given I am the user with id "3"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>",
        "helper_group_id": <helper_group_id>
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "<item_id>"
    And the table "threads" at item_id "<item_id>" should be:
      | latest_update_at    | status   | helper_group_id   |
      | 2022-01-01 00:00:00 | <status> | <helper_group_id> |
    Examples:
      | item_id | status                  | helper_group_id | comment                           |
      | 2000    | waiting_for_trainer     | 11              |                                   |
      | 2001    | waiting_for_participant | 12              |                                   |
      | 2002    | waiting_for_trainer     | 11              | Doesn't exist: Create             |
      | 2003    | waiting_for_participant | 12              | Doesn't exist: Create             |
      | 2007    | waiting_for_participant | 12              | In chapter                        |
      | 2008    | waiting_for_participant | 11              | In chapter                        |
      | 2009    | waiting_for_participant | 12              | In chapter; doesn't exist: Create |
      | 2010    | waiting_for_participant | 11              | In chapter; Doesn't exist: Create |

  Scenario Outline: A user who has can_watch>=answer on the item AND can_watch_members on the participant can always switch to an open status
    Given I am the user with id "2"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>",
        "helper_group_id": <helper_group_id>
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "<item_id>"
    And the table "threads" at item_id "<item_id>" should be:
      | latest_update_at    | status   |
      | 2022-01-01 00:00:00 | <status> |
    Examples:
      | item_id | status                  | helper_group_id | comment                 |
      | 160     | waiting_for_participant | 30              |                         |
      | 170     | waiting_for_trainer     | 30              |                         |
      | 180     | waiting_for_participant | 30              |                         |
      | 190     | waiting_for_trainer     | 30              |                         |
      | 200     | waiting_for_participant | 30              | # Doesn't exist: Create |
      | 210     | waiting_for_trainer     | 30              | # Doesn't exist: Create |

  Scenario Outline: Can switch to open if part of the group the participant has requested help to AND can_watch>=answer on the item
    Given I am the user with id "4"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>",
        "helper_group_id": 50
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "<item_id>"
    And the table "threads" at item_id "<item_id>" should be:
      | latest_update_at    | status   |
      | 2022-01-01 00:00:00 | <status> |
    Examples:
      | item_id | status                  |
      | 220     | waiting_for_participant |
      | 230     | waiting_for_trainer     |

  Scenario Outline: Can switch to open if part of the group the participant has requested help to AND have validated the item
    Given I am the user with id "4"
    When I send a PUT request to "/items/<item_id>/participant/3/thread" with the following body:
      """
      {
        "status": "<status>",
        "helper_group_id": 50
      }
      """
    Then the response should be "updated"
    And the table "threads" should stay unchanged but the row with item_id "<item_id>"
    And the table "threads" at item_id "<item_id>" should be:
      | latest_update_at    | status   |
      | 2022-01-01 00:00:00 | <status> |
    Examples:
      | item_id | status                  |
      | 240     | waiting_for_participant |
      | 250     | waiting_for_trainer     |

  Scenario: If status is open and not provided (no change): update helper_group_id
    Given I am the user with id "2"
    When I send a PUT request to "/items/260/participant/3/thread" with the following body:
      """
      {
        "helper_group_id": 20
      }
      """
    Then the response should be "updated"
  And the table "threads" at item_id "260" should be:
    | latest_update_at    | helper_group_id |
    | 2022-01-01 00:00:00 | 20              |

  Scenario: Participant who can request help on region can request help on class
    Given I am the user with id "3"
    When I send a PUT request to "/items/270/participant/3/thread" with the following body:
      """
      {
        "status": "waiting_for_trainer",
        "helper_group_id": 10
      }
      """
    Then the response should be "updated"
    And the table "threads" at item_id "270" should be:
      | latest_update_at    | helper_group_id |
      | 2022-01-01 00:00:00 | 10              |
