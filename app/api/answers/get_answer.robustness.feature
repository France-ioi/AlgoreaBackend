Feature: Get user's answer by id
  Background:
    Given the database has the following table 'groups':
      | id | name | type |
      | 11 | jdoe | User |
      | 13 | team | Team |
      | 14 | jane | User |
      | 15 | bill | User |
      | 16 | jeff | User |
      | 17 | elon | User |
    And the database has the following table 'users':
      | login | group_id |
      | jdoe  | 11       |
      | jane  | 14       |
      | bill  | 15       |
      | jeff  | 16       |
      | elon  | 17       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 14             |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id  | entry_participant_type | default_language_tag |
      | 200 | User                   | fr                   |
      | 210 | Team                   | fr                   |
      | 220 | Team                   | fr                   |
    And the database has the following table 'permissions_generated':
      | item_id | group_id | can_view_generated | can_watch_generated |
      | 200     | 11       | info               | none                |
      | 200     | 14       | none               | none                |
      | 200     | 15       | none               | answer              |
      | 200     | 16       | content            | result              |
      | 200     | 17       | content            | answer              |
      | 220     | 14       | content            | none                |
    And the database has the following table 'attempts':
      | id | participant_id |
      | 1  | 11             |
      | 1  | 13             |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id |
      | 1          | 11             | 200     |
      | 1          | 13             | 210     |
    And the database has the following table 'answers':
      | id  | author_id | participant_id | attempt_id | item_id | type       | state   | answer   | created_at          |
      | 101 | 11        | 11             | 1          | 200     | Submission | Current | print(1) | 2017-05-29 06:38:38 |
      | 102 | 11        | 11             | 1          | 200     | Submission | Current | print(1) | 2017-05-29 06:38:38 |
      | 103 | 11        | 13             | 1          | 200     | Submission | Current | print(1) | 2017-05-29 06:38:38 |
      | 104 | 11        | 11             | 1          | 220     | Submission | Current | print(1) | 2017-05-29 06:38:38 |
    And the database has the following table 'gradings':
      | answer_id | score | graded_at           |
      | 101       | 100   | 2018-05-29 06:38:38 |
      | 102       | 100   | 2019-05-29 06:38:38 |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 13       | 15         | false             |
      | 13       | 16         | true              |
      | 13       | 17         | true              |

  Scenario: Wrong answer_id
    Given I am the user with id "11"
    When I send a GET request to "/answers/abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for answer_id (should be int64)"

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/answers/101"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: User doesn't have sufficient access rights to the answer
    Given I am the user with id "11"
    When I send a GET request to "/answers/101"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access rights to the answer (the user is the participant, can_view<content)
    Given I am the user with id "11"
    When I send a GET request to "/answers/102"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access rights to the answer (the user is a member of the participant group, can_view<content)
    Given I am the user with id "14"
    When I send a GET request to "/answers/103"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access rights to the answer (can_view>=content, but the user is not a member of the participant group)
    Given I am the user with id "14"
    When I send a GET request to "/answers/104"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access rights to the answer (the user is an observer with can_watch>=answer, but without can_watch_members)
    Given I am the user with id "15"
    When I send a GET request to "/answers/103"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access rights to the answer (the user is an observer with can_watch_members, but with can_watch<answer)
    Given I am the user with id "16"
    When I send a GET request to "/answers/103"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access rights to the answer (the user is an observer with can_watch>=answer and can_watch_members, but the participant is a user)
    Given I am the user with id "17"
    When I send a GET request to "/answers/101"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No answers
    Given I am the user with id "11"
    When I send a GET request to "/answers/100"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
