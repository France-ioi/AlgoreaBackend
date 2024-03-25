Feature:
  As a group manager,
  I want to be able to strengthen a require_* rule when I update a group,
  and specify what to do with the existing members

  Scenario Outline: >
      Should remove all participants from the group and fill the group_membership_changes when
      a require_* field is strengthened, and approval_change_action is set to "empty"
    Given I am @Teacher
    And there are the following groups:
      | group     | parent | members             | require_personal_info_access_approval       | require_lock_membership_approval_until       | require_watch_approval       |
      | @School   |        | @Teacher            |                                             |                                              |                              |
      | @Class    |        | @Student1,@Student2 | <old_require_personal_info_access_approval> | <old_require_lock_membership_approval_until> | <old_require_watch_approval> |
      | @SubGroup | @Class | @Student3,@Student4 |                                             |                                              |                              |
    And @Teacher is a manager of the group @Class and can manage memberships and group
    And the time now is "2020-01-01T01:00:00Z"
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
    And there should be the following group membership changes:
      | group_id | member_id | action                         | at                  | initiator_id |
      | @Class   | @Student1 | removed_due_to_approval_change | 2020-01-01 01:00:00 | @Teacher     |
      | @Class   | @Student2 | removed_due_to_approval_change | 2020-01-01 01:00:00 | @Teacher     |
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
      | group   | members   | require_personal_info_access_approval       | require_lock_membership_approval_until       | require_watch_approval       |
      | @School | @Teacher  |                                             |                                              |                              |
      | @Class  | @Student1 | <old_require_personal_info_access_approval> | <old_require_lock_membership_approval_until> | <old_require_watch_approval> |
    And @Teacher is a manager of the group @Class and can manage memberships and group
    And the time now is "2020-01-01T01:00:00Z"
    When I send a PUT request to "/groups/@Class" with the following body:
    """
    {
      "<require_field>": <new_value>
    }
    """
    Then the response should be "updated"
    And the field "<require_field>" of the group @Class should be "<new_value_db>"
    Examples:
      | require_field                          | new_value              | new_value_db        | old_require_personal_info_access_approval | old_require_lock_membership_approval_until | old_require_watch_approval |
      | require_personal_info_access_approval  | "none"                 | none                | none                                      |                                            |                            |
      | require_personal_info_access_approval  | "none"                 | none                | view                                      |                                            |                            |
      | require_personal_info_access_approval  | "none"                 | none                | edit                                      |                                            |                            |
      | require_personal_info_access_approval  | "view"                 | view                | view                                      |                                            |                            |
      | require_personal_info_access_approval  | "view"                 | view                | edit                                      |                                            |                            |
      | require_personal_info_access_approval  | "edit"                 | edit                | edit                                      |                                            |                            |
      | require_lock_membership_approval_until | null                   | <null>              |                                           |                                            |                            |
      | require_lock_membership_approval_until | null                   | <null>              |                                           |                                            |                            |
      | require_lock_membership_approval_until | "2020-01-01T12:00:00Z" | 2020-01-01 12:00:00 |                                           | 2020-01-01 12:00:00                        |                            |
      | require_lock_membership_approval_until | "2020-01-01T11:59:59Z" | 2020-01-01 11:59:59 |                                           | 2020-01-01 12:00:00                        |                            |
      | require_lock_membership_approval_until | "2020-01-01T00:59:59Z" | 2020-01-01 00:59:59 |                                           | 2020-01-01 00:59:58                        |                            | # The new value is < NOW()
      | require_lock_membership_approval_until | "2020-01-01T01:00:00Z" | 2020-01-01 01:00:00 |                                           | 2020-01-01 00:59:59                        |                            | # The new valus is == NOW()
      | require_watch_approval                 | false                  | false               |                                           |                                            | false                      |
      | require_watch_approval                 | false                  | false               |                                           |                                            | true                       |
