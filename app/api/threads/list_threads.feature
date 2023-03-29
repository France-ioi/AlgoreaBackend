Feature: Get threads
  Background:
    Given there are the following users:
      | name           | @reference      |
      | RichardFeynman | @RichardFeynman |
      | StevenHawking  | @StevenHawking  |
      | DavidBowie     | @DavidBowie     |
      | ClaireStudent  | @ClaireStudent  |
    Given there are the following threads:
      | participant_id  |
      | @RichardFeynman |
      | @ClaireStudent  |
      | @StevenHawking  |
      | @DavidBowie     |

  Scenario: Should have all the fields properly set
    Given I am MarieCurie
    And there is a group Laboratory referenced by @Laboratory
    And I am a manager of the group Laboratory
    And I can watch the group Laboratory
    And there are the following users:
      | @reference      | login          | first_name | last_name |
      | @AlbertEinstein | AlbertEinstein | Albert     | Einstein  |
      | @PaulDirac      | PaulDirac      | Paul       | Dirac     |
      And AlbertEinstein the scientist is a member of the group Laboratory
    And AlbertEinstein has approved access to his personal info for the group Laboratory
    And PaulDirac the scientist is a member of the group Laboratory
    And the database has the following table 'items':
      | id | type | default_language_tag |
      | 1  | Task | fr                   |
      | 2  | Task | en                   |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title      |
      | 1       | en           | Beginning  |
      | 1       | fr           | Debut      |
      | 2       | en           | Experiment |
    And the database has the following table 'threads':
      | item_id | participant_id  | status                  | message_count | latest_update_at    | helper_group_id |
      | 1       | @AlbertEinstein | waiting_for_trainer     | 0             | 2023-01-01 00:00:01 | @Laboratory     |
      | 2       | @PaulDirac      | waiting_for_participant | 1             | 2023-01-01 00:00:02 | @Laboratory     |
    When I send a GET request to "/threads?watched_group_id=@Laboratory"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      [
        {
          "item": {
            "id": "1",
            "language_tag": "fr",
            "title": "Debut",
            "type": "Task"
          },
          "latest_update_at": "2023-01-01T00:00:01Z",
          "message_count": 0,
          "participant": {
            "id": "@AlbertEinstein",
            "login": "AlbertEinstein",
            "first_name": "Albert",
            "last_name": "Einstein"
          },
          "status": "waiting_for_trainer"
        },
        {
          "item": {
            "id": "2",
            "language_tag": "en",
            "title": "Experiment",
            "type": "Task"
          },
          "latest_update_at": "2023-01-01T00:00:02Z",
          "message_count": 1,
          "participant": {
            "id": "@PaulDirac",
            "login": "PaulDirac",
            "first_name": "",
            "last_name": ""
          },
          "status": "waiting_for_participant"
        }
      ]
    """

  Scenario: Should get the threads whose the participant is a descendant (or self) of the watched_group_id
    Given I am RichardFeynman
    And there is a group University referenced by @University
    And I am a manager of the group University
    And I can watch the group University
    And the group FirstYear is a descendant of the group University
    And ClaireStudent the student is a member of the group FirstYear
    And StevenHawking the professor is a member of the group University
    When I send a GET request to "/threads?watched_group_id=@University"
    Then the response code should be 200
    And it should be a JSON array with 2 entries
    And the response should match the following JSONPath:
      | JSONPath            | value          |
      | $[*].participant.id | @ClaireStudent |
      | $[*].participant.id | @StevenHawking |
