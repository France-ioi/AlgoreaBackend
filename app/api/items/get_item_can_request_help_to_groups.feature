Feature: Get permissions can_request_help for an item
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
      | item                 | type    |
      | @Chapter1            | Chapter |
      | @Chapter2            | Chapter |
      | @Chapter3            | Chapter |
      | @Chapter4            | Chapter |
      | @Item1               | Task    |
      | @Item2               | Task    |
      | @Item2_NoPropagation | Task    |
      | @Item3               | Task    |
      | @Item4               | Task    |
      | @Item4_NoPropagation | Task    |
    And there are the following item relations:
      | item                 | parent    | request_help_propagation |
      | @Item1               | @Chapter1 |                          |
      | @Item2               | @Chapter2 | true                     |
      | @Item2_NoPropagation | @Chapter2 | false                    |
      | @Item3               | @Chapter3 |                          |
      | @Item4               | @Chapter4 | true                     |
      | @Item4_NoPropagation | @Chapter4 | false                    |

  Scenario Outline: permissions.can_request_help should be true if there is a can_request_help_to permission
    Given I am @Student
    And there are the following item permissions:
      | item                 | group    | can_view | can_request_help_to | can_request_help is defined                                                           |
      | @Item1               | @Student | solution | @HelperGroup1       | Directly on item, current-user                                                        |
      | @Item2               | @Student | solution |                     |                                                                                       |
      | @Item2_NoPropagation | @Student | solution |                     |                                                                                       |
      | @Chapter2            | @Student |          | @HelperGroup2       | On @Item2 and @Item2_NoPropagation ancestor                                           |
      | @Item3               | @School  | solution | @HelperGroup3       | On @Item3, on an ancestor (@School) of current-user                                   |
      | @Chapter4            | @School  |          | @HelperGroup4       | On @Item4 and @Item4_NoPropagation ancestor, on an ancestor (@School) of current-user |
      | @Item4               | @School  | solution |                     |                                                                                       |
      | @Item4_NoPropagation | @School  | solution |                     |                                                                                       |
    When I send a GET request to "/items/<item_id>"
    Then the response code should be 200
    And the response at $.permissions.can_request_help should be "<can_request_help>"
    Examples:
      | item_id              | can_request_help |
      | @Item1               | true             |
      | @Item2               | true             |
      | @Item2_NoPropagation | false            |
      | @Item3               | true             |
      | @Item4               | true             |
      | @Item4_NoPropagation | false            |

  Scenario Outline: watched_group.permissions.can_request_help should be true if there is a can_request_help_to permission for the watched_group
    Given I am @Teacher
    And there are the following item permissions:
      | item                 | group        | can_view | can_watch | can_request_help_to | can_request_help is defined                                                                |
      | @Item1               | @Teacher     | solution | answer    |                     |                                                                                            |
      | @Item2               | @Teacher     | solution | answer    |                     |                                                                                            |
      | @Item2_NoPropagation | @Teacher     | solution | answer    |                     |                                                                                            |
      | @Item3               | @Teacher     | solution | answer    |                     |                                                                                            |
      | @Item4               | @Teacher     | solution | answer    |                     |                                                                                            |
      | @Item4_NoPropagation | @Teacher     | solution | answer    |                     |                                                                                            |
      | @Item1               | @Class       |          |           | @HelperGroup1       | Directly on @Item1, current-user                                                           |
      | @Chapter2            | @Class       |          |           | @HelperGroup2       | On @Item2 and @Item2_NoPropagation ancestor                                                |
      | @Item3               | @ClassParent |          |           | @HelperGroup3       | On @Item3, on an ancestor (@ClassParent) of current-user                                   |
      | @Chapter4            | @ClassParent |          |           | @HelperGroup4       | On @Item4 and @Item4_NoPropagation ancestor, on an ancestor (@ClassParent) of current-user |
    When I send a GET request to "/items/<item_id>?watched_group_id=@Class"
    Then the response code should be 200
    And the response at $.watched_group.permissions.can_request_help should be "<can_request_help>"
    Examples:
      | item_id              | can_request_help |
      | @Item1               | true             |
      | @Item2               | true             |
      | @Item2_NoPropagation | false            |
      | @Item3               | true             |
      | @Item4               | true             |
      | @Item4_NoPropagation | false            |
