Feature:
  As a group manager,
  I want to be able to strengthen a require_* rule when I update a group,
  and specify what to do with the existing members

  Scenario Outline: >
      Should remove all participants from the group and fill the group_membership_changes when
      a require_* field is strengthened, and approval_change_action is set to "empty"
    Given I am @Teacher
    And there are the following groups:
      | group     | parent       | members             | require_personal_info_access_approval       | require_lock_membership_approval_until       | require_watch_approval       |
      | @School   |              | @Teacher            |                                             |                                              |                              |
      | @Class    | @ClassParent | @Student1,@Student2 | <old_require_personal_info_access_approval> | <old_require_lock_membership_approval_until> | <old_require_watch_approval> |
      | @SubGroup | @Class       | @Student3,@Student4 |                                             |                                              |                              |
    And the group @Teacher is a manager of the group @ClassParent and can manage memberships and the group
    And the time now is "2020-01-01T01:00:00.001Z"
    When I send a PUT request to "/groups/@Class" with the following body:
    """
    {
      "<require_field>": <new_value>,
      "approval_change_action": "empty"
    }
    """
    Then the response should be "updated"
    And the field "<require_field>" of the group @Class should be "<new_value_db>"
    And @Student1 should not be a member of the group @Class
    And @Student2 should not be a member of the group @Class
    # The subgroups should not be affected.
    And @SubGroup should be a member of the group @Class
    And @Student3 should be a member of the group @SubGroup
    And @Student4 should be a member of the group @SubGroup
    And the table "group_membership_changes" should be:
      | group_id | member_id | action                         | at                    | initiator_id |
      | @Class   | @Student1 | removed_due_to_approval_change | {{currentTimeDBMs()}} | @Teacher     |
      | @Class   | @Student2 | removed_due_to_approval_change | {{currentTimeDBMs()}} | @Teacher     |
    Examples:
      | require_field                          | new_value              | new_value_db        | old_require_personal_info_access_approval | old_require_lock_membership_approval_until | old_require_watch_approval |
      | require_personal_info_access_approval  | "view"                 | view                | none                                      |                                            |                            |
      | require_personal_info_access_approval  | "edit"                 | edit                | none                                      |                                            |                            |
      | require_personal_info_access_approval  | "edit"                 | edit                | view                                      |                                            |                            |
      | require_lock_membership_approval_until | "2020-01-01T12:00:00Z" | 2020-01-01 12:00:00 |                                           | null                                       |                            |
      | require_lock_membership_approval_until | "2020-01-01T12:00:01Z" | 2020-01-01 12:00:01 |                                           | 2020-01-01 12:00:00                        |                            |
      | require_watch_approval                 | true                   | 1                   |                                           |                                            | false                      |

  Scenario Outline: >
      Should be able to update the require_* fields without approval_change_action when they are not strengthened
    Given I am @Teacher
    And there are the following groups:
      | group   | parent       | members         | require_personal_info_access_approval       | require_lock_membership_approval_until       | require_watch_approval       |
      | @School |              | @Teacher        |                                             |                                              |                              |
      | @Class  | @ClassParent | <group_members> | <old_require_personal_info_access_approval> | <old_require_lock_membership_approval_until> | <old_require_watch_approval> |
    And the group @Teacher is a manager of the group @ClassParent and can manage memberships and the group
    And the server time now is "2020-01-01T01:00:00.001Z"
    When I send a PUT request to "/groups/@Class" with the following body:
    """
    {
      "<require_field>": <new_value>
    }
    """
    Then the response should be "updated"
    And the field "<require_field>" of the group @Class should be "<new_value_db>"
    # If the group doesn't have any members, the fields are never considered strengthened.
    Examples:
      | require_field                          | new_value              | new_value_db        | group_members | old_require_personal_info_access_approval | old_require_lock_membership_approval_until | old_require_watch_approval |
      | require_personal_info_access_approval  | "none"                 | none                | @Student1     | none                                      |                                            |                            |
      | require_personal_info_access_approval  | "view"                 | view                |               | none                                      |                                            |                            |
      | require_personal_info_access_approval  | "edit"                 | edit                |               | none                                      |                                            |                            |
      | require_personal_info_access_approval  | "none"                 | none                | @Student1     | view                                      |                                            |                            |
      | require_personal_info_access_approval  | "view"                 | view                | @Student1     | view                                      |                                            |                            |
      | require_personal_info_access_approval  | "edit"                 | edit                |               | view                                      |                                            |                            |
      | require_personal_info_access_approval  | "none"                 | none                | @Student1     | edit                                      |                                            |                            |
      | require_personal_info_access_approval  | "view"                 | view                | @Student1     | edit                                      |                                            |                            |
      | require_personal_info_access_approval  | "edit"                 | edit                | @Student1     | edit                                      |                                            |                            |
      | require_lock_membership_approval_until | null                   | <null>              | @Student1     |                                           |                                            |                            |
      | require_lock_membership_approval_until | null                   | <null>              | @Student1     |                                           | 2020-01-01 12:00:00                        |                            |
      | require_lock_membership_approval_until | "2020-01-01T12:00:00Z" | 2020-01-01 12:00:00 | @Student1     |                                           | 2020-01-01 12:00:00                        |                            |
      | require_lock_membership_approval_until | "2020-01-01T11:59:59Z" | 2020-01-01 11:59:59 | @Student1     |                                           | 2020-01-01 12:00:00                        |                            |
      | require_lock_membership_approval_until | "2020-01-01T12:00:01Z" | 2020-01-01 12:00:01 |               |                                           | 2020-01-01 12:00:00                        |                            |
      | require_lock_membership_approval_until | "2020-01-01T00:59:59Z" | 2020-01-01 00:59:59 | @Student1     |                                           | 2020-01-01 00:59:58                        |                            | # The new value is < NOW()
      | require_lock_membership_approval_until | "2020-01-01T01:00:00Z" | 2020-01-01 01:00:00 | @Student1     |                                           | 2020-01-01 00:59:59                        |                            | # The new valus is == NOW()
      | require_lock_membership_approval_until | "2020-01-01T00:59:59Z" | 2020-01-01 00:59:59 | @Student1     |                                           |                                            |                            | # The new valus is < NOW()
      | require_lock_membership_approval_until | "2020-01-01T01:00:00Z" | 2020-01-01 01:00:00 | @Student1     |                                           |                                            |                            | # The new valus is == NOW()
      | require_watch_approval                 | false                  | false               | @Student1     |                                           |                                            | false                      |
      | require_watch_approval                 | true                   | true                |               |                                           |                                            | false                      |
      | require_watch_approval                 | false                  | false               | @Student1     |                                           |                                            | true                       |

  Scenario: Should be able to set require_lock_membership_approval_until to null when it is already set
    Given I am @Teacher
    And there are the following groups:
      | group   | parent       | members   | require_lock_membership_approval_until |
      | @School |              | @Teacher  |                                        |
      | @Class  | @ClassParent | @Student1 | 2020-01-01 12:00:00                    |
    And the group @Teacher is a manager of the group @ClassParent and can manage memberships and the group
    When I send a PUT request to "/groups/@Class" with the following body:
    """
    {
      "require_lock_membership_approval_until": null
    }
    """
    Then the response should be "updated"
    And the field "require_lock_membership_approval_until" of the group @Class should be "<null>"

  Scenario: Should reject all pending requests when approval_change_action = 'empty'
    Given I am @Teacher
    And the time now is "2020-01-01T01:00:00.001Z"
    And there are the following groups:
      | group   | parent       | members                       | require_watch_approval |
      | @School |              | @Teacher                      |                        |
      | @Class  | @ClassParent | @Student1,@Student2           | false                  |
      | @Other  |              | @Student3,@Student4,@Student5 |                        |
    And the database has the following table "group_pending_requests":
      | group_id | member_id | type          | at                      |
      | @Class   | @Student1 | leave_request | 2020-01-01 00:00:01.000 |
      | @Class   | @Student3 | join_request  | 2020-01-01 00:00:03.000 |
      | @Class   | @Student4 | join_request  | 2020-01-01 00:00:04.000 |
      | @Class   | @Student5 | invitation    | 2020-01-01 00:00:05.000 |
      | @Other   | @Student5 | join_request  | 2020-01-01 00:00:15.000 |
    And the group @Teacher is a manager of the group @ClassParent and can manage memberships and the group
    When I send a PUT request to "/groups/@Class" with the following body:
    """
    {
      "require_watch_approval": true,
      "approval_change_action": "empty"
    }
    """
    Then the response should be "updated"
    And there should be no group pending requests for the group @Class with the type "join_request"
    And the table "group_pending_requests" should be:
      | group_id | member_id | type         | at                      |
      | @Class   | @Student5 | invitation   | 2020-01-01 00:00:05.000 |
      | @Other   | @Student5 | join_request | 2020-01-01 00:00:15.000 |
    And the table "group_membership_changes" should be:
      | group_id | member_id | at                    | action                         | initiator_id |
      | @Class   | @Student1 | {{currentTimeDBMs()}} | removed_due_to_approval_change | @Teacher     |
      | @Class   | @Student2 | {{currentTimeDBMs()}} | removed_due_to_approval_change | @Teacher     |
      | @Class   | @Student3 | {{currentTimeDBMs()}} | join_request_refused           | @Teacher     |
      | @Class   | @Student4 | {{currentTimeDBMs()}} | join_request_refused           | @Teacher     |

  Scenario: Should reject all pending requests and send invitations to the past members when approval_change_action = 'reinvite'
    Given I am @Teacher
    And there are the following groups:
      | group   | parent       | members                       | require_watch_approval |
      | @School |              | @Teacher                      |                        |
      | @Class  | @ClassParent | @Student1,@Student2           | false                  |
      | @Other  |              | @Student3,@Student4,@Student5 |                        |
    And the database has the following table "group_pending_requests":
      | group_id | member_id | type          | at                      |
      | @Class   | @Student1 | leave_request | 2020-01-01 00:00:01.000 |
      | @Class   | @Student3 | join_request  | 2020-01-01 00:00:03.000 |
      | @Class   | @Student4 | join_request  | 2020-01-01 00:00:04.000 |
      | @Class   | @Student5 | invitation    | 2020-01-01 00:00:05.000 |
      | @Other   | @Student5 | join_request  | 2020-01-01 00:00:15.000 |
    And the group @Teacher is a manager of the group @ClassParent and can manage memberships and the group
    And the time now is "2020-01-01T01:00:00.001Z"
    When I send a PUT request to "/groups/@Class" with the following body:
    """
    {
      "require_watch_approval": true,
      "approval_change_action": "reinvite"
    }
    """
    Then the response should be "updated"
    And the table "group_pending_requests" should be:
      | group_id | member_id | type         | at                      |
      | @Class   | @Student2 | invitation   | {{currentTimeDB()}}     |
      | @Class   | @Student5 | invitation   | 2020-01-01 00:00:05.000 |
      | @Other   | @Student5 | join_request | 2020-01-01 00:00:15.000 |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action                         | at                  | initiator_id |
      | @Class   | @Student1 | removed_due_to_approval_change | {{currentTimeDB()}} | @Teacher     |
      | @Class   | @Student2 | invitation_created             | {{currentTimeDB()}} | @Teacher     |
      | @Class   | @Student3 | join_request_refused           | {{currentTimeDB()}} | @Teacher     |
      | @Class   | @Student4 | join_request_refused           | {{currentTimeDB()}} | @Teacher     |

  Scenario: Should empty the group when approval_change_action = "reinvite"
    Given I am @Teacher
    And there are the following groups:
      | group     | parent       | members             | require_watch_approval |
      | @School   |              | @Teacher            |                        |
      | @Class    | @ClassParent | @Student1,@Student2 | false                  |
      | @SubGroup | @Class       | @Student3,@Student4 |                        |
    And the group @Teacher is a manager of the group @ClassParent and can manage memberships and the group
    When I send a PUT request to "/groups/@Class" with the following body:
    """
    {
      "require_watch_approval": true,
      "approval_change_action": "reinvite"
    }
    """
    Then the response should be "updated"
    And @Student1 should not be a member of the group @Class
    And @Student2 should not be a member of the group @Class
    And @Student3 should be a member of the group @SubGroup

  # If approval_change_action = "reinvite", the leave requests are transformed into invitations,
  # because the primary key of the group_pending_requests table is (group_id, member_id).
  # This was already tested in a test above.


  Scenario: Should reject all pending leave requests when require_lock_membership_approval_until is strengthened and approval_change_action = 'empty'
    Given I am @Teacher
    And the time now is "2020-01-01T01:00:00.001Z"
    And there are the following groups:
      | group   | parent       | members                       | require_lock_membership_approval_until |
      | @School |              | @Teacher                      |                                        |
      | @Class  | @ClassParent | @Student1,@Student2,@Student3 | 2020-01-01 12:00:00                    |
      | @Other  |              | @Student4,@Student5,@Student6 |                                        |
    And the database has the following table "group_pending_requests":
      | group_id | member_id | type          | at                      |
      | @Class   | @Student1 | leave_request | 2020-01-01 00:00:01.000 |
      | @Class   | @Student2 | leave_request | 2020-01-01 00:00:02.000 |
      | @Class   | @Student4 | join_request  | 2020-01-01 00:00:04.000 |
      | @Class   | @Student5 | join_request  | 2020-01-01 00:00:05.000 |
      | @Class   | @Student6 | invitation    | 2020-01-01 00:00:06.000 |
      | @Other   | @Student5 | leave_request | 2020-01-01 00:00:15.000 |
      | @Other   | @Student6 | join_request  | 2020-01-01 00:00:16.000 |
    And the group @Teacher is a manager of the group @ClassParent and can manage memberships and the group
    When I send a PUT request to "/groups/@Class" with the following body:
    """
    {
      "require_lock_membership_approval_until": "2020-01-01T12:00:01Z",
      "approval_change_action": "empty"
    }
    """
    Then the response should be "updated"
    And the table "group_pending_requests" should be:
      | group_id | member_id | type          | at                      |
      | @Class   | @Student6 | invitation    | 2020-01-01 00:00:06.000 |
      | @Other   | @Student5 | leave_request | 2020-01-01 00:00:15.000 |
      | @Other   | @Student6 | join_request  | 2020-01-01 00:00:16.000 |
    And the table "group_membership_changes" should be:
      | group_id | member_id | at                    | action                         | initiator_id |
      | @Class   | @Student1 | {{currentTimeDBMs()}} | removed_due_to_approval_change | @Teacher     |
      | @Class   | @Student2 | {{currentTimeDBMs()}} | removed_due_to_approval_change | @Teacher     |
      | @Class   | @Student3 | {{currentTimeDBMs()}} | removed_due_to_approval_change | @Teacher     |
      | @Class   | @Student4 | {{currentTimeDBMs()}} | join_request_refused           | @Teacher     |
      | @Class   | @Student5 | {{currentTimeDBMs()}} | join_request_refused           | @Teacher     |
