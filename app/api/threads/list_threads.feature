Feature: List threads
  Background:
    And there are the following groups:
      | name                 | parent               | members                            |
      | UniversityConsortium |                      | ConsortiumPresident                |
      | University           | UniversityConsortium | UniversityPresident,RichardFeynman |
      | FirstYear            | University           |                                    |
      | Classroom            | FirstYear            | Mary                               |
      | Université           | UniversityConsortium | EtienneKlein                       |
      | PremièreAnnée        | Université           | Charlotte,Baptiste,Thibaut         |
      | Classe               | PremièreAnnée        |                                    |
      | Superstar            |                      | DavidBowie                         |
    And there are the following items with permissions:
      | item                                  | group        | has_validated | can_view                 | can_watch |
      | BaptisteCanViewInfo                   | Baptiste     |               | info                     |           |
      | BaptisteCanViewContent1               | Baptiste     |               | content                  |           |
      | BaptisteCanViewContent2               | Baptiste     |               | content                  |           |
      | BaptisteCanViewContentWithDescendants | Baptiste     |               | content_with_descendants |           |
      | EtienneKleinHasValidated1             | EtienneKlein | 1             |                          |           |
      | EtienneKleinHasValidated2             | EtienneKlein | 1             |                          |           |
      | EtienneKleinHasValidated3             | EtienneKlein | 1             |                          |           |
      | EtienneKleinHasValidated4             | EtienneKlein | 1             |                          |           |
      | EtienneKleinHasValidated5             | EtienneKlein | 1             |                          |           |
      | EtienneKleinHasValidated6             | EtienneKlein | 1             |                          |           |
      | EtienneKleinHasNotValidated           | EtienneKlein |               |                          | answer    |
      | EtienneKleinCanWatchAnswer1           | EtienneKlein |               |                          | answer    |
      | EtienneKleinCanWatchAnswer2           | EtienneKlein |               |                          | answer    |
      | EtienneKleinCanWatchAnswer3           | EtienneKlein |               |                          | answer    |
      | EtienneKleinCanWatchAnswer4           | EtienneKlein |               |                          | answer    |
    Given there are the following threads:
      | participant         | item                                  | helper_group         | status                  | latest_update_at     | comment                                                                                                                |
      | ConsortiumPresident |                                       |                      |                         |                      |                                                                                                                        |
      | UniversityPresident |                                       |                      |                         |                      |                                                                                                                        |
      | Mary                |                                       |                      |                         |                      |                                                                                                                        |
      | EtienneKlein        | EtienneKleinCanWatchAnswer1           |                      |                         |                      | EtienneKlein is_mine=0 -> notok: must not be the participant                                                           |
      | Charlotte           | EtienneKleinHasValidated1             | PremièreAnnée        | waiting_for_trainer     |                      | EtienneKlein is_mine=0 -> List thread notok: not part of helper group                                                  |
      | Charlotte           | EtienneKleinHasNotValidated           | Université           | waiting_for_trainer     |                      | EtienneKlein is_mine=0 -> List thread notok: Has not validated                                                         |
      | Charlotte           | EtienneKleinHasValidated2             | Université           | waiting_for_trainer     |                      | EtienneKlein is_mine=0 -> List thread ok: part of helper group, open thread and validated item                         |
      | Charlotte           | EtienneKleinHasValidated3             | UniversityConsortium | waiting_for_trainer     |                      | EtienneKlein is_mine=0 -> List thread ok: part of helper group, open thread and validated item                         |
      | Charlotte           | EtienneKleinCanWatchAnswer2           | Université           | waiting_for_trainer     |                      | EtienneKlein is_mine=0 -> List thread ok: can_watch >= answer                                                          |
      | Charlotte           | EtienneKleinHasValidated4             | Université           | waiting_for_participant |                      | EtienneKlein is_mine=0 -> List thread ok: part of helper group, open thread and validated item                         |
      | Charlotte           | EtienneKleinCanWatchAnswer3           | Université           | waiting_for_participant |                      | EtienneKlein is_mine=0 -> List thread ok: can_watch >= answer                                                          |
      | Charlotte           | EtienneKleinHasValidated5             | Université           | closed                  | 2021-12-20T00:00:00Z | EtienneKlein is_mine=0 -> List thread ok: part of helper group, closed thread for less than 2 weeks and validated item |
      | Charlotte           | EtienneKleinCanWatchAnswer4           | Université           | closed                  | 2021-12-20T00:00:00Z | EtienneKlein is_mine=0 -> List thread ok: can_watch >= answer                                                          |
      | Charlotte           | EtienneKleinHasValidated6             | Université           | closed                  | 2021-11-00T00:00:00Z | EtienneKlein is_mine=0 -> List thread notok: closed for more than 2 weeks                                              |
      | Charlotte           | EtienneKleinCanWatchAnswer5           | Université           | closed                  | 2021-11-00T00:00:00Z | EtienneKlein is_mine=0 -> List thread ok: can_watch >= answer                                                          |
      | Baptiste            | BaptisteCanViewInfo                   |                      |                         |                      | Baptiste is_mine=1 -> notok: can_view < content                                                                        |
      | Baptiste            | BaptisteCanViewContent1               |                      |                         |                      | Baptiste is_mine=1 -> ok: can_view >= content                                                                          |
      | Thibaut             | BaptisteCanViewContent2               |                      |                         |                      | Baptiste is_mine=1 -> notok: not the participant                                                                       |
      | Baptiste            | BaptisteCanViewContentWithDescendants |                      |                         |                      | Baptiste is_mine=1 -> ok: can_view >= content                                                                          |
      | DavidBowie          |                                       |                      |                         |                      |                                                                                                                        |
    And the time now is "2022-01-01T00:00:00Z"

  Scenario: Should have all the fields properly set, including first_name and last_name when the access is approved
    Given I am MarieCurie
    And there is a group Laboratory referenced by @Laboratory
    And I am a manager of the group Laboratory and can watch its members
    And there are the following users:
      | @reference      | login          | first_name | last_name |
      | @AlbertEinstein | AlbertEinstein | Albert     | Einstein  |
      | @PaulDirac      | PaulDirac      | Paul       | Dirac     |
    And AlbertEinstein is a member of the group Laboratory who has approved access to his personal info
    And PaulDirac is a member of the group Laboratory
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

  Scenario: Should get the threads whose the participant is a descendant of the watched_group_id
    Given I am RichardFeynman
    And I can watch the group University
    And the user Mary is referenced by @Mary
    And the user UniversityPresident is referenced by @UniversityPresident
    And the group University is referenced by @University
    When I send a GET request to "/threads?watched_group_id=@University"
    And it should be a JSON array with 2 entries
    And the response at $[*].participant.id should be:
      | @UniversityPresident |
      | @Mary                |

  Scenario: Should get the threads whose the participant is equal to the watched_group_id
    Given I am RichardFeynman
    And I can watch the group University
    And the user Mary is referenced by @Mary
    When I send a GET request to "/threads?watched_group_id=@Mary"
    And it should be a JSON array with 1 entries
    And the response at $[0].participant.id should be "@Mary"

  Scenario: Should return only the threads in which the participant is the current user and the item is visible when is_mine=1
    # Baptiste can see BaptisteCanViewContent and BaptisteCanViewContentWithDescendants
    # Waiting for implementation of is_mine

  Scenario: Should return only the threads that the current-user can list and in which he is not the participant when is_mine=0
    # EtienneKlein can see EtienneKleinHasValidated2, EtienneKleinHasValidated3, EtienneKleinCanWatchAnswer2, EtienneKleinHasValidated4, EtienneKleinCanWatchAnswer3, EtienneKleinHasValidated5, EtienneKleinCanWatchAnswer4, EtienneKleinCanWatchAnswer5
    # Waiting for implementation of is_mine
