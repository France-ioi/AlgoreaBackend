Feature: List threads
  Background:
    And there are the following groups:
      | group         | parent        | members                                               |
      | @Consortium   |               | @ConsortiumMember                                     |
      | @A_University | @Consortium   | @A_UniversityMember,@A_UniversityManagerCanWatch      |
      | @B_University | @Consortium   | @B_UniversityMember                                   |
      | @A_Section    | @A_University |                                                       |
      | @B_Section    | @B_University | @B_SectionMember1,@B_SectionMember2,@B_SectionMember3 |
      | @A_Class      | @A_Section    | @A_ClassMember1,@A_ClassMember2                       |
      | @B_Class      | @B_Section    |                                                       |
      | @OtherGroup   |               | @OtherGroupMember                                     |
    And the group @A_UniversityManagerCanWatch is a manager of the group @A_UniversityParent and can watch for submissions from the group and its descendants
    And the group @A_University is a child of the group @A_UniversityParent
    And there are the following tasks:
      | item                                            |
      | @B_SectionMember2_CanViewInfo                   |
      | @B_SectionMember2_CanViewContent1               |
      | @B_SectionMember2_CanViewContent2               |
      | @B_SectionMember2_CanViewContentWithDescendants |
      | @B_UniversityMember_HasValidated1               |
      | @B_UniversityMember_HasValidated2               |
      | @B_UniversityMember_HasValidated3               |
      | @B_UniversityMember_HasValidated4               |
      | @B_UniversityMember_HasValidated5               |
      | @B_UniversityMember_HasValidated6               |
      | @B_UniversityMember_HasNotValidated             |
      | @B_UniversityMember_CanWatchAnswer1             |
      | @B_UniversityMember_CanWatchAnswer2             |
      | @B_UniversityMember_CanWatchAnswer3             |
      | @B_UniversityMember_CanWatchAnswer4             |
      | @B_UniversityMember_CanWatchAnswer5             |
      | @B_UniversityMember_CanWatchAnswer6             |
      | @Item1                                          |
      | @A_UniversityManagerCanWatch_CanViewContent1    |
      | @A_UniversityManagerCanWatch_CanViewContent2    |
      | @Item2                                          |
      | @A_UniversityManagerCanWatch_CanViewInfo        |
    And there are the following item permissions:
      | item                                            | group                        | can_view                 | can_watch |
      | @B_SectionMember2_CanViewInfo                   | @B_SectionMember2            | info                     | answer    |
      | @B_SectionMember2_CanViewContent1               | @B_SectionMember2            | content                  | answer    |
      | @B_SectionMember2_CanViewContent2               | @B_SectionMember2            | content                  | answer    |
      | @B_SectionMember2_CanViewContentWithDescendants | @B_SectionMember2            | content_with_descendants | answer    |
      | @B_UniversityMember_HasNotValidated             | @B_UniversityMember          | content                  |           |
      | @B_UniversityMember_HasValidated1               | @B_UniversityMember          | content                  |           |
      | @B_UniversityMember_HasValidated2               | @B_UniversityMember          | content                  |           |
      | @B_UniversityMember_HasValidated3               | @B_UniversityMember          | content                  | result    |
      | @B_UniversityMember_HasValidated4               | @B_University                | content                  | result    |
      | @B_UniversityMember_HasValidated5               | @B_UniversityMember          | content                  | result    |
      | @B_UniversityMember_HasValidated6               | @B_UniversityMember          | content                  | result    |
      | @B_UniversityMember_CanWatchAnswer1             | @B_UniversityMember          | content                  | answer    |
      | @B_UniversityMember_CanWatchAnswer2             | @B_UniversityMember          | content                  | answer    |
      | @B_UniversityMember_CanWatchAnswer3             | @B_UniversityMember          | content                  | answer    |
      | @B_UniversityMember_CanWatchAnswer4             | @B_UniversityMember          | content                  | answer    |
      | @B_UniversityMember_CanWatchAnswer5             | @B_UniversityMember          | content                  | answer    |
      | @B_UniversityMember_CanWatchAnswer6             | @B_UniversityMember          | none                     | answer    |
      | @A_UniversityManagerCanWatch_CanViewContent1    | @A_UniversityManagerCanWatch | content                  | answer    |
      | @A_UniversityManagerCanWatch_CanViewContent2    | @A_UniversityManagerCanWatch | content                  | answer    |
      | @A_UniversityManagerCanWatch_CanViewInfo        | @A_UniversityManagerCanWatch | info                     | answer    |
      | @Item1                                          | @A_UniversityManagerCanWatch | content                  | answer    |
      | @Item2                                          | @A_UniversityManagerCanWatch | content                  | answer    |
    And there are the following validated results:
      | item                              | participant         | validated |
      | @B_UniversityMember_HasValidated1 | @B_UniversityMember | 1         |
      | @B_UniversityMember_HasValidated2 | @B_UniversityMember | 1         |
      | @B_UniversityMember_HasValidated3 | @B_UniversityMember | 1         |
      | @B_UniversityMember_HasValidated4 | @B_UniversityMember | 1         |
      | @B_UniversityMember_HasValidated5 | @B_UniversityMember | 1         |
      | @B_UniversityMember_HasValidated6 | @B_UniversityMember | 1         |
    Given there are the following threads:
      | participant         | item                                            | helper_group  | status                  | latest_update_at    | message_count | comment                                                                                                                       |
      | @ConsortiumMember   | @Item1                                          |               |                         |                     | 0             |                                                                                                                               |
      | @A_UniversityMember | @A_UniversityManagerCanWatch_CanViewContent1    |               |                         |                     | 1             |                                                                                                                               |
      | @A_ClassMember1     | @A_UniversityManagerCanWatch_CanViewContent2    |               |                         |                     | 2             | @A_UniversityManagerCanWatch watched_group_id -> ok, manager can_view=content                                                 |
      | @A_ClassMember2     | @A_UniversityManagerCanWatch_CanViewInfo        |               |                         |                     | 2             | @A_UniversityManagerCanWatch watched_group_id -> notok, manager can_view=info only                                            |
      | @B_UniversityMember | @B_UniversityMember_CanWatchAnswer1             |               |                         |                     | 3             | @B_UniversityMember is_mine=0 -> notok: must not be the participant                                                           |
      | @B_SectionMember1   | @B_UniversityMember_HasValidated1               | @B_Section    | waiting_for_trainer     |                     | 4             | @B_UniversityMember is_mine=0 -> List thread notok: not part of helper group                                                  |
      | @B_SectionMember1   | @B_UniversityMember_HasNotValidated             | @B_University | waiting_for_trainer     |                     | 5             | @B_UniversityMember is_mine=0 -> List thread notok: Has not validated                                                         |
      | @B_SectionMember1   | @B_UniversityMember_HasValidated2               | @B_University | waiting_for_trainer     |                     | 6             | @B_UniversityMember is_mine=0 -> List thread notok: part of helper group, open thread and validated item, can_watch < result  |
      | @B_SectionMember1   | @B_UniversityMember_HasValidated3               | @Consortium   | waiting_for_trainer     |                     | 7             | @B_UniversityMember is_mine=0 -> List thread ok: part of helper group, open thread and validated item                         |
      | @B_SectionMember1   | @B_UniversityMember_CanWatchAnswer2             | @B_University | waiting_for_trainer     |                     | 8             | @B_UniversityMember is_mine=0 -> List thread ok: can_watch >= answer                                                          |
      | @B_SectionMember1   | @B_UniversityMember_HasValidated4               | @B_University | waiting_for_participant |                     | 9             | @B_UniversityMember is_mine=0 -> List thread ok: part of helper group, open thread and validated item                         |
      | @B_SectionMember1   | @B_UniversityMember_CanWatchAnswer3             | @B_University | waiting_for_participant |                     | 10            | @B_UniversityMember is_mine=0 -> List thread ok: can_watch >= answer                                                          |
      | @B_SectionMember1   | @B_UniversityMember_CanWatchAnswer6             | @B_University | waiting_for_participant |                     | 11            | @B_UniversityMember is_mine=0 -> List thread notok: cannot view the item                                                      |
      | @B_SectionMember1   | @B_UniversityMember_HasValidated5               | @B_University | closed                  | 2021-12-20 00:00:00 | 12            | @B_UniversityMember is_mine=0 -> List thread ok: part of helper group, closed thread for less than 2 weeks and validated item |
      | @B_SectionMember1   | @B_UniversityMember_CanWatchAnswer4             | @B_University | closed                  | 2021-12-20 00:00:00 | 13            | @B_UniversityMember is_mine=0 -> List thread ok: can_watch >= answer                                                          |
      | @B_SectionMember1   | @B_UniversityMember_HasValidated6               | @B_University | closed                  | 2021-11-01 00:00:00 | 14            | @B_UniversityMember is_mine=0 -> List thread notok: closed for more than 2 weeks                                              |
      | @B_SectionMember1   | @B_UniversityMember_CanWatchAnswer5             | @B_University | closed                  | 2021-11-01 00:00:00 | 15            | @B_UniversityMember is_mine=0 -> List thread ok: can_watch >= answer                                                          |
      | @B_SectionMember2   | @B_SectionMember2_CanViewInfo                   |               |                         |                     | 16            | @B_SectionMember2 is_mine=1 -> notok: can_view < content                                                                      |
      | @B_SectionMember2   | @B_SectionMember2_CanViewContent1               |               |                         |                     | 17            | @B_SectionMember2 is_mine=1 -> ok: can_view >= content                                                                        |
      | @B_SectionMember3   | @B_SectionMember2_CanViewContent2               |               |                         |                     | 18            | @B_SectionMember2 is_mine=1 -> notok: not the participant                                                                     |
      | @B_SectionMember2   | @B_SectionMember2_CanViewContentWithDescendants |               |                         |                     | 19            | @B_SectionMember2 is_mine=1 -> ok: can_view >= content                                                                        |
      | @OtherGroupMember   | @Item2                                          |               |                         |                     | 20            |                                                                                                                               |
    And the DB time now is "2022-01-01 00:00:00"

  Scenario: Should have all the fields properly set, including first_name and last_name when the access is approved
    Given I am @LaboratoryManagerCanWatch
    And I am a manager of the group @LaboratoryParent and can watch for submissions from the group and its descendants
    And the group @Laboratory is a child of the group @LaboratoryParent
    And there are the following users:
      | user                                                 | first_name            | last_name            |
      | @LaboratoryMember_WithApprovedAccessPersonalInfo     | FirstName_Approved    | LastName_Approved    |
      | @LaboratoryMember_WithApprovedAccessPersonalInfoNull | null                  | null                 |
      | @LaboratoryMember_WithoutApprovedAccessPersonalInfo  | FirstName_NotApproved | LastName_NotApproved |
    And @LaboratoryMember_WithApprovedAccessPersonalInfo is a member of the group @Laboratory who has approved access to his personal info
    And @LaboratoryMember_WithoutApprovedAccessPersonalInfo is a member of the group @Laboratory
    And @LaboratoryMember_WithApprovedAccessPersonalInfoNull is a member of the group @Laboratory who has approved access to his personal info
    And the database has the following table "items":
      | id | type | default_language_tag |
      | 1  | Task | fr                   |
      | 2  | Task | en                   |
    And the database has the following table "permissions_generated":
      | group_id                   | item_id | can_view_generated | can_watch_generated |
      | @LaboratoryManagerCanWatch | 1       | content            | answer              |
      | @LaboratoryManagerCanWatch | 2       | content            | answer              |
    And the database has the following table "items_strings":
      | item_id | language_tag | title      |
      | 1       | en           | Beginning  |
      | 1       | fr           | Debut      |
      | 2       | en           | Experiment |
    And the database has the following table "threads":
      | item_id | participant_id                                       | status                  | message_count | latest_update_at    | helper_group_id |
      | 1       | @LaboratoryMember_WithApprovedAccessPersonalInfo     | waiting_for_trainer     | 0             | 2023-01-01 00:00:01 | @Laboratory     |
      | 2       | @LaboratoryMember_WithApprovedAccessPersonalInfoNull | waiting_for_participant | 1             | 2023-01-01 00:00:02 | @Laboratory     |
      | 2       | @LaboratoryMember_WithoutApprovedAccessPersonalInfo  | waiting_for_participant | 1             | 2023-01-01 00:00:02 | @Laboratory     |
    When I send a GET request to "/threads?watched_group_id=@Laboratory"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      [
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
            "id": "@LaboratoryMember_WithApprovedAccessPersonalInfoNull",
            "login": "LaboratoryMember_WithApprovedAccessPersonalInfoNull",
            "first_name": null,
            "last_name": null
          },
          "status": "waiting_for_participant"
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
            "id": "@LaboratoryMember_WithoutApprovedAccessPersonalInfo",
            "login": "LaboratoryMember_WithoutApprovedAccessPersonalInfo"
          },
          "status": "waiting_for_participant"
        },
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
            "id": "@LaboratoryMember_WithApprovedAccessPersonalInfo",
            "login": "LaboratoryMember_WithApprovedAccessPersonalInfo",
            "first_name": "FirstName_Approved",
            "last_name": "LastName_Approved"
          },
          "status": "waiting_for_trainer"
        }
      ]
      """

  Scenario: Should get the threads whose participant is a descendant of the watched_group_id
    Given I am @A_UniversityManagerCanWatch
    When I send a GET request to "/threads?watched_group_id=@A_University"
    And the response at $[*].participant.id should be:
      | @A_UniversityMember |
      | @A_ClassMember1     |

  Scenario: Should get the threads whose participant is equal to the watched_group_id
    Given I am @A_UniversityManagerCanWatch
    When I send a GET request to "/threads?watched_group_id=@A_ClassMember1"
    And the response should be a JSON array with 1 entry
    And the response at $[0].participant.id should be "@A_ClassMember1"

  Scenario: Should return only threads having the current user as the participant and a visible item when is_mine=1
    Given I am @B_SectionMember2
    When I send a GET request to "/threads?is_mine=1"
    Then the response code should be 200
    And the response at $[*].item.id should be:
      | @B_SectionMember2_CanViewContent1               |
      | @B_SectionMember2_CanViewContentWithDescendants |

  Scenario: Should return only threads the current user can list excluding those having the current user as the participant when is_mine=0
    Given I am @B_UniversityMember
    When I send a GET request to "/threads?is_mine=0"
    Then the response code should be 200
    And the response at $[*].item.id should be:
      | @B_UniversityMember_HasValidated3   |
      | @B_UniversityMember_CanWatchAnswer2 |
      | @B_UniversityMember_HasValidated4   |
      | @B_UniversityMember_CanWatchAnswer3 |
      | @B_UniversityMember_HasValidated5   |
      | @B_UniversityMember_CanWatchAnswer4 |
      | @B_UniversityMember_CanWatchAnswer5 |

  Scenario: Should return only threads the current user can list and whose participant is a descendant of the watched_group_id
    Given I am @B_UniversityMember
    And the group @B_UniversityMember is a manager of the group @B_University and can watch for submissions from the group and its descendants
    When I send a GET request to "/threads?watched_group_id=@B_University"
    Then the response code should be 200
    And the response at $[*].item.id should be:
      | @B_UniversityMember_HasValidated3   |
      | @B_UniversityMember_CanWatchAnswer1 |
      | @B_UniversityMember_CanWatchAnswer2 |
      | @B_UniversityMember_HasValidated4   |
      | @B_UniversityMember_CanWatchAnswer3 |
      | @B_UniversityMember_HasValidated5   |
      | @B_UniversityMember_CanWatchAnswer4 |
      | @B_UniversityMember_CanWatchAnswer5 |

  Scenario: Should return only threads related to the item_id or its descendants when item_id is given
    Given I am @John
    And there are the following items:
      | item              | parent                          | type    |
      | @Root_Task        |                                 | Task    |
      | @Chapter1         |                                 | Chapter |
      | @Chapter1_Task    | @Chapter1                       | Task    |
      | @ChapterRoot_2A   |                                 | Chapter |
      | @ChapterRoot_2B   |                                 | Chapter |
      | @Chapter2         | @Chapter_Root_2,@ChapterRoot_2B | Chapter |
      | @Chapter2_Task    | @Chapter2                       | Task    |
      | @Chapter2_1       | @Chapter2                       | Chapter |
      | @Chapter2_1_Task1 | @Chapter2_1                     | Task    |
      | @Chapter2_1_Task2 | @Chapter2_1                     | Task    |
      | @Chapter3         |                                 | Chapter |
    And there are the following threads:
      | participant | item              | visible_by_participant | message_count |
      | @John       | @Root_Task        | 1                      | 100           |
      | @John       | @Chapter1         | 1                      | 101           |
      | @John       | @Chapter1_Task    | 1                      | 102           |
      | @John       | @Chapter2         | 1                      | 103           |
      | @John       | @Chapter2_Task    | 1                      | 104           |
      | @John       | @Chapter3         | 1                      | 105           |
      | @John       | @Chapter2_1       | 1                      | 106           |
      | @John       | @Chapter2_1_Task1 | 1                      | 107           |
      | @John       | @Chapter2_1_Task2 | 1                      | 108           |
    When I send a GET request to "/threads?is_mine=1&item_id=@Chapter2"
    Then the response code should be 200
    And the response at $[*].item.id should be:
      | @Chapter2         |
      | @Chapter2_Task    |
      | @Chapter2_1       |
      | @Chapter2_1_Task1 |
      | @Chapter2_1_Task2 |

  Scenario Outline: Should support sort and limit parameters
    Given I am @John
    And there are the following items:
      | item   | type |
      | @Task1 | Task |
      | @Task2 | Task |
      | @Task3 | Task |
      | @Task4 | Task |
    And there are the following threads:
      | participant | item                   | visible_by_participant | message_count | latest_update_at    |
      | @John       | @TaskSecondMaxUpdateAt | 1                      | 100           | 2023-01-01 00:00:10 |
      | @John       | @TaskMinUpdateAt       | 1                      | 101           | 2023-01-01 00:00:01 |
      | @John       | @TaskMaxUpdateAt       | 1                      | 102           | 2023-01-01 00:00:11 |
      | @John       | @TaskSecondMinUpdateAt | 1                      | 103           | 2023-01-01 00:00:02 |
    When I send a GET request to "/threads?is_mine=1&limit=<limit>&sort=<sort>"
    Then the response code should be 200
    And the response should be a JSON array with <nb_results> entries
    And the response at $[<result_item_index>].item.id should be "<result_item>"
    Examples:
      | sort              | limit | nb_results | result_item_index | result_item            |
      | latest_update_at  | 1     | 1          | 0                 | @TaskMinUpdateAt       |
      | -latest_update_at | 1     | 1          | 0                 | @TaskMaxUpdateAt       |
      | -latest_update_at | 2     | 2          | 0                 | @TaskMaxUpdateAt       |
      | -latest_update_at | 2     | 2          | 1                 | @TaskSecondMaxUpdateAt |

  Scenario: Should support pagination parameters with results
    Given I am @John
    And there are the following items:
      | item             | type |
      | @TaskMinUpdateAt | Task |
      | @TaskMaxUpdateAt | Task |
    And there are the following threads:
      | participant | item             | visible_by_participant | latest_update_at    |
      | @John       | @TaskMinUpdateAt | 1                      | 2023-01-01 00:00:01 |
      | @John       | @TaskMaxUpdateAt | 1                      | 2023-01-01 00:00:02 |
    When I send a GET request to "/threads?is_mine=1&limit=1&sort=latest_update_at&from.item_id=@TaskMinUpdateAt&from.participant_id=@John"
    Then the response code should be 200
    And the response should be a JSON array with 1 entries
    And the response at $[0].item.id should be "@TaskMaxUpdateAt"

  Scenario: Should support pagination parameters with no results
    Given I am @John
    And there are the following items:
      | item             | type |
      | @TaskMinUpdateAt | Task |
      | @TaskMaxUpdateAt | Task |
    And there are the following threads:
      | participant | item             | visible_by_participant | latest_update_at    |
      | @John       | @TaskMinUpdateAt | 1                      | 2023-01-01 00:00:01 |
      | @John       | @TaskMaxUpdateAt | 1                      | 2023-01-01 00:00:02 |
    When I send a GET request to "/threads?is_mine=1&limit=1&sort=latest_update_at&from.item_id=@TaskMaxUpdateAt&from.participant_id=@John"
    Then the response code should be 200
    And the response should be a JSON array with 0 entries

  Scenario Outline: Should filter by status if parameter status is given
    Given I am @John
    And there are the following items:
      | item                             | type |
      | @TaskWaitingForParticipantThread | Task |
      | @TaskWaitingForTrainerThread     | Task |
      | @TaskClosedThread                | Task |
    And there are the following threads:
      | participant | item                             | status                  | visible_by_participant |
      | @John       | @TaskWaitingForParticipantThread | waiting_for_participant | 1                      |
      | @John       | @TaskWaitingForTrainerThread     | waiting_for_trainer     | 1                      |
      | @John       | @TaskClosedThread                | closed                  | 1                      |
    When I send a GET request to "/threads?is_mine=1&status=<status>"
    Then the response code should be 200
    And the response should be a JSON array with 1 entries
    And the response at $[0].item.id should be "<result_item>"
    Examples:
      | status              | result_item                  |
      | waiting_for_trainer | @TaskWaitingForTrainerThread |
      | closed              | @TaskClosedThread            |

  Scenario Outline: Should return only threads with latest_update_at>latest_update_gt if parameter latest_update_gt is given
    Given I am @John
    And there are the following items:
      | item   | type |
      | @Task1 | Task |
      | @Task2 | Task |
      | @Task3 | Task |
    And there are the following threads:
      | participant | item   | visible_by_participant | latest_update_at    |
      | @John       | @Task1 | 1                      | 2023-01-01 00:00:01 |
      | @John       | @Task2 | 1                      | 2023-01-01 00:00:02 |
      | @John       | @Task3 | 1                      | 2023-01-01 00:00:03 |
    When I send a GET request to "/threads?is_mine=1&latest_update_gt=<latest_update_gt>&sort=latest_update_at"
    Then the response code should be 200
    And the response should be a JSON array with <nb_results> entries
    And the response at $[0].item.id should be "<first_result_item>"
    Examples:
      | latest_update_gt     | first_result_item | nb_results |
      | 2023-01-01T00:00:00Z | @Task1            | 3          |
      | 2023-01-01T00:00:02Z | @Task3            | 1          |

  Scenario: Should return no results when latest_update_gt is given but no entries are greater than latest_update_gt
    Given I am @John
    And there are the following items:
      | item   | type |
      | @Task1 | Task |
      | @Task2 | Task |
      | @Task3 | Task |
    And there are the following threads:
      | participant | item   | visible_by_participant | latest_update_at    |
      | @John       | @Task1 | 1                      | 2023-01-01 00:00:01 |
      | @John       | @Task2 | 1                      | 2023-01-01 00:00:02 |
      | @John       | @Task3 | 1                      | 2023-01-01 00:00:03 |
    When I send a GET request to "/threads?is_mine=1&latest_update_gt=2023-01-01T00:00:03Z&sort=latest_update_at"
    Then the response code should be 200
    And the response should be a JSON array with 0 entries
