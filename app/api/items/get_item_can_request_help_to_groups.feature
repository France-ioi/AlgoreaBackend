Feature: Get permissions has_can_request_help_to for an item
  Background:
    Given there are the following groups:
      | group               | parent  | members               |
      | @InaccessibleSchool |         | @InaccessibleGroup    |
      | @School             |         | @TeacherGroup         |
      | @TeacherGroup       |         | @Teacher              |
      | @ClassParent        |         | @Class                |
      | @Class              | @School | @Student,@HelperGroup |
    And @Teacher is a manager of the group @Class and can watch its members
    And there are the following items:
      | item                 | parent    | type    | request_help_propagation |
      | @Chapter1            |           | Chapter |                          |
      | @Chapter2            |           | Chapter |                          |
      | @Chapter3            |           | Chapter |                          |
      | @Chapter4            |           | Chapter |                          |
      | @Item1               | @Chapter1 | Task    |                          |
      | @Item2               | @Chapter2 | Task    | true                     |
      | @Item2_NoPropagation | @Chapter2 | Task    | false                    |
      | @Item3               | @Chapter3 | Task    |                          |
      | @Item4               | @Chapter4 | Task    | true                     |

  Scenario Outline: permissions.has_can_request_help_to should be true if there is a can_request_help_to permission
    Given I am @Student
    And there are the following item permissions:
      | item                 | group    | can_view | can_request_help_to | can_request_help is defined                        |
      | @Item1               | @Student | solution | @HelperGroup1       | Directly on item, current-user                     |
      | @Item2               | @Student | solution |                     |                                                    |
      | @Item2_NoPropagation | @Student | solution |                     |                                                    |
      | @Chapter2            | @Student |          | @HelperGroup2       | On item's ancestor                                 |
      | @Item3               | @School  | solution | @HelperGroup3       | On item, on an ancestor of current-user            |
      | @Chapter4            | @School  |          | @HelperGroup4       | On item's ancestor, on an ancestor of current-user |
      | @Item4               | @School  | solution |                     |                                                    |
    When I send a GET request to "/items/<item_id>"
    Then the response code should be 200
    And the response at $.permissions.has_can_request_help_to should be "<has_can_request_help_to>"
    Examples:
      | item_id              | has_can_request_help_to |
      | @Item1               | true                    |
      | @Item2               | true                    |
      | @Item2_NoPropagation | false                   |
      | @Item3               | true                    |
      | @Item4               | true                    |

  Scenario Outline: watched_group.permissions.has_can_request_help_to should be true if there is a can_request_help_to permission for the watched_group
    Given I am @Teacher
    And there are the following item permissions:
      | item                 | group        | can_view | can_watch | can_request_help_to | can_request_help is defined                        |
      | @Item1               | @Teacher     | solution | answer    |                     |                                                    |
      | @Item2               | @Teacher     | solution | answer    |                     |                                                    |
      | @Item2_NoPropagation | @Teacher     | solution | answer    |                     |                                                    |
      | @Item3               | @Teacher     | solution | answer    |                     |                                                    |
      | @Item4               | @Teacher     | solution | answer    |                     |                                                    |
      | @Item1               | @Class       |          |           | @HelperGroup1       | Directly on item, current-user                     |
      | @Chapter2            | @Class       |          |           | @HelperGroup2       | On item's ancestor                                 |
      | @Item3               | @ClassParent |          |           | @HelperGroup3       | On item, on an ancestor of current-user            |
      | @Chapter4            | @ClassParent |          |           | @HelperGroup4       | On item's ancestor, on an ancestor of current-user |
  When I send a GET request to "/items/<item_id>?watched_group_id=@Class"
    Then the response code should be 200
    And the response at $.watched_group.permissions.has_can_request_help_to should be "<has_can_request_help_to>"
    Examples:
      | item_id              | has_can_request_help_to |
      | @Item1               | true                    |
      | @Item2               | true                    |
      | @Item2_NoPropagation | false                   |
      | @Item3               | true                    |
      | @Item4               | true                    |
