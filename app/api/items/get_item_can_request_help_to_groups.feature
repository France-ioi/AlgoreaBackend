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
      | item                          | type    |
      | @Chapter1                     | Chapter |
      | @Chapter2                     | Chapter |
      | @Chapter3                     | Chapter |
      | @Chapter4                     | Chapter |
      | @Chapter5                     | Chapter |
      | @Chapter6_Group_IsOwner       | Chapter |
      | @Chapter7_ParentGroup_IsOwner | Chapter |
      | @Item1                        | Task    |
      | @Item2                        | Task    |
      | @Item2_NoPropagation          | Task    |
      | @Item3                        | Task    |
      | @Item4                        | Task    |
      | @Item4_NoPropagation          | Task    |
      | @Item5_IsOwner                | Task    |
      | @Item6                        | Task    |
      | @Item6_NoPropagation          | Task    |
      | @Item7                        | Task    |
      | @Item7_NoPropagation          | Task    |
    And there are the following item relations:
      | item                 | parent                        | request_help_propagation |
      | @Item1               | @Chapter1                     |                          |
      | @Item2               | @Chapter2                     | true                     |
      | @Item2_NoPropagation | @Chapter2                     | false                    |
      | @Item3               | @Chapter3                     |                          |
      | @Item4               | @Chapter4                     | true                     |
      | @Item4_NoPropagation | @Chapter4                     | false                    |
      | @Item5_IsOwner       | @Chapter5                     |                          |
      | @Item6               | @Chapter6_Group_IsOwner       | true                     |
      | @Item6_NoPropagation | @Chapter6_Group_IsOwner       |                          |
      | @Item7               | @Chapter7_ParentGroup_IsOwner | true                     |
      | @Item7_NoPropagation | @Chapter7_ParentGroup_IsOwner |                          |

  Scenario Outline: permissions.can_request_help should be true if there is a can_request_help_to permission
    Given I am @Student
    And there are the following item permissions:
      | item                          | group    | can_view | can_request_help_to | is_owner | can_request_help is defined                                                           |
      | @Item1                        | @Student | solution | @HelperGroup1       |          | Directly on item, current-user                                                        |
      | @Item2                        | @Student | solution |                     |          |                                                                                       |
      | @Item2_NoPropagation          | @Student | solution |                     |          |                                                                                       |
      | @Chapter2                     | @Student |          | @HelperGroup2       |          | On @Item2 and @Item2_NoPropagation ancestor                                           |
      | @Item3                        | @School  | solution | @HelperGroup3       |          | On @Item3, on an ancestor (@School) of current-user                                   |
      | @Chapter4                     | @School  |          | @HelperGroup4       |          | On @Item4 and @Item4_NoPropagation ancestor, on an ancestor (@School) of current-user |
      | @Item4                        | @School  | solution |                     |          |                                                                                       |
      | @Item4_NoPropagation          | @School  | solution |                     |          |                                                                                       |
      | @Item5_IsOwner                | @Student | solution |                     | true     | Because is_owner=1                                                                    |
      | @Chapter6_Group_IsOwner       | @Student |          |                     | true     | On @Item6 and @Item6_NoPropagation ancestor                                           |
      | @Item6                        | @Student | solution |                     |          | Because is_owner=1 on parent chapter                                                  |
      | @Item6_NoPropagation          | @Student | solution |                     |          |                                                                                       |
      | @Chapter7_ParentGroup_IsOwner | @School  |          |                     | true     | On @Item7 and @Item7_NoPropagation ancestor                                           |
      | @Item7                        | @School  | solution |                     |          | Because is_owner=1 on parent chapter and parent group                                 |
      | @Item7_NoPropagation          | @School  | solution |                     |          |                                                                                       |
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
      | @Item5_IsOwner       | true             |
      | @Item6               | true             |
      | @Item6_NoPropagation | false            |
      | @Item7               | true             |
      | @Item7_NoPropagation | false            |

  Scenario Outline: watched_group.permissions.can_request_help should be true if there is a can_request_help_to permission for the watched_group
    Given I am @Teacher
    And there are the following item permissions:
      | item                          | group        | can_view | can_watch | is_owner | can_request_help_to | can_request_help is defined                                                                |
      | @Item1                        | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Item2                        | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Item2_NoPropagation          | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Item3                        | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Item4                        | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Item4_NoPropagation          | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Chapter5                     | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Item5_IsOwner                | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Chapter6_Group_IsOwner       | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Item6                        | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Item6_NoPropagation          | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Chapter7_ParentGroup_IsOwner | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Item7                        | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Item7_NoPropagation          | @Teacher     | solution | answer    |          |                     |                                                                                            |
      | @Item1                        | @Class       |          |           |          | @HelperGroup1       | Directly on @Item1, current-user                                                           |
      | @Chapter2                     | @Class       |          |           |          | @HelperGroup2       | On @Item2 and @Item2_NoPropagation ancestor                                                |
      | @Item3                        | @ClassParent |          |           |          | @HelperGroup3       | On @Item3, on an ancestor (@ClassParent) of current-user                                   |
      | @Chapter4                     | @ClassParent |          |           |          | @HelperGroup4       | On @Item4 and @Item4_NoPropagation ancestor, on an ancestor (@ClassParent) of current-user |
      | @Item5_IsOwner                | @Class       |          |           | true     |                     | Because is_owner=1                                                                         |
      | @Chapter6_Group_IsOwner       | @Class       |          |           | true     |                     | Because is_owner=1 on parent chapter                                                       |
      | @Chapter7_ParentGroup_IsOwner | @ClassParent |          |           | true     |                     | Because is_owner=1 on parent chapter and parent group                                      |
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
      | @Item5_IsOwner       | true             |
      | @Item6               | true             |
      | @Item6_NoPropagation | false            |
      | @Item7               | true             |
      | @Item7_NoPropagation | false            |
