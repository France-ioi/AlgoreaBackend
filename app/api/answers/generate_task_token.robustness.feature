Feature: Generate a read-only task token for an item from an answer - robustness
  Background:
    Given the database has the following table "groups":
      | id  | name     | type  |
      | 102 | team     | Team  |
      | 106 | Groupe A | Class |
    And the database has the following users:
      | group_id | login |
      | 101      | john  |
      | 103      | jack  |
      | 104      | jess  |
      | 105      | jim   |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 102             | 101            |
      | 106             | 101            |
      | 106             | 103            |
      | 106             | 104            |
      | 106             | 105            |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 106      | 103        | 0                 |
      | 106      | 104        | 1                 |
      | 106      | 105        | 1                 |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id | url                                                                     | type    | entry_participant_type | default_language_tag | text_id |
      | 10 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Chapter | Team                   | fr                   | task10  |
      | 20 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | User                   | fr                   | task20  |
      | 30 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | Team                   | fr                   | task30  |
      | 40 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | User                   | fr                   | task40  |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 101      | 10      | content            | none                |
      | 101      | 20      | info               | none                |
      | 101      | 30      | info               | none                |
      | 102      | 30      | info               | none                |
      | 103      | 40      | content            | answer              |
      | 104      | 40      | content            | none                |
      | 105      | 40      | content            | answer              |
      | 106      | 40      | content            | none                |
    And the database has the following table "attempts":
      | id | participant_id |
      | 0  | 101            |
      | 0  | 102            |
      | 0  | 103            |
      | 0  | 104            |
      | 0  | 105            |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | started_at          |
      | 0          | 101            | 10      | 2020-01-01 01:01:01 |
      | 0          | 101            | 20      | 2020-01-01 01:01:01 |
      | 0          | 101            | 30      | 2020-01-01 01:01:01 |
      | 0          | 101            | 40      | 2020-01-01 01:01:01 |
      | 0          | 102            | 10      | 2020-01-01 01:01:01 |
      | 0          | 102            | 20      | 2020-01-01 01:01:01 |
      | 0          | 102            | 30      | 2020-01-01 01:01:01 |
      | 0          | 102            | 40      | 2020-01-01 01:01:01 |
      | 0          | 103            | 40      | 2020-01-01 01:01:01 |
      | 0          | 104            | 40      | 2020-01-01 01:01:01 |
      | 0          | 105            | 40      | null                |
    And the database has the following table "answers":
      | id | participant_id | attempt_id | item_id | author_id  | created_at          |
      | 1  | 101            | 0          | 10      | 101        | 2020-01-01 01:01:01 |
      | 2  | 101            | 0          | 20      | 101        | 2020-01-01 01:01:01 |
      | 3  | 102            | 0          | 30      | 102        | 2020-01-01 01:01:01 |
      | 4  | 101            | 0          | 40      | 101        | 2020-01-01 01:01:01 |

  Scenario: Invalid answer_id
    Given I am the user with id "101"
    When I send a POST request to "/answers/1111111111111111111111111111/generate-task-token"
    Then the response code should be 400
    And the response error message should contain "Wrong value for answer_id (should be int64)"

  Scenario: User not found
    Given I am the user with id "404"
    When I send a POST request to "/answers/1/generate-task-token"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: The answer doesn't exists
    Given I am the user with id "101"
    When I send a POST request to "/answers/404/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item type is not "Task"
    Given I am the user with id "101"
    When I send a POST request to "/answers/1/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The participant is the current user but is not allowed to "view" >= "content" on the item
    Given I am the user with id "101"
    When I send a POST request to "/answers/2/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The participant is a team which the current user is a member of but is not allowed to "view" >= "content" on the item
    Given I am the user with id "101"
    When I send a POST request to "/answers/3/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is someone else and not allowed to "watch" the participant of the answer
    Given I am the user with id "103"
    When I send a POST request to "/answers/4/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is someone else and not allowed to "watch answer" of the item
    Given I am the user with id "104"
    When I send a POST request to "/answers/4/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user doesn't have a started result on the item
    Given I am the user with id "105"
    When I send a POST request to "/answers/4/generate-task-token"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
