Feature: Get threads
  Background:
    Given there are the following threads:
      | participant_id  |
      | @RichardFeynman |
      | @University     |
      | @FirstYear      |
      | @Tezos          |

  Scenario: Should get the threads whose the participant is a descendant (or self) of the watched_group_id
    Given I am RichardFeynman
    And I am a manager of the group University
    And I can watch the group University
    And the group FirstYear is a descendant of the group University
    When I send a GET request to "/threads?watched_group_id=@University"
    Then the response code should be 200
    And it should be a JSON array with 2 entries
    And the response should match the following JSONPath:
      | JSONPath            | value       |
      | $[*].participant.id | @University |
      | $[*].participant.id | @FirstYear  |
