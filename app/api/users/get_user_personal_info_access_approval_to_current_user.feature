Feature: Manager of a user gets `personal_info_access_approval_to_current_user` in the response
  Scenario Outline: Should return the max permission of all the groups the current-user manages and which the user is a descendant of
    Given there are the following groups:
      | group                         | parent  | members        | require_personal_info_access_approval |
      | @AllUsers                     |         | @Manager,@User |                                       |
      | @School                       |         |                | <school_permission>                   |
      | @Class                        | @School | @User          | <class_permission>                    |
      | @OtherManagedGroupWithoutUser |         |                | edit                                  |
      | @OtherManagedGroupWithUser    |         | @User          | <other_permission>                    |
      | @NonManagedGroupWithUser      |         | @User          | edit                                  |
    And I am @Manager
    And I am a manager of the group @School
    And I am a manager of the group @OtherManagedGroupWithUser
    When I send a GET request to "/users/@User"
    Then the response code should be 200
    And the response at $.personal_info_access_approval_to_current_user should be "<result>"
  Examples:
    | school_permission | class_permission | other_permission | result |
    | none              | none             | none             | none   |
    | view              | none             | none             | view   |
    | none              | view             | none             | view   |
    | none              | none             | view             | view   |
    | edit              | none             | view             | edit   |
    | none              | edit             | view             | edit   |
    | view              | none             | edit             | edit   |
