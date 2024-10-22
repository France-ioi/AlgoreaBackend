Feature: Check if the group code is valid

  Background:
    Given the database has the following table "groups":
      | id | type  | code       | code_expires_at     | code_lifetime | frozen_membership | name           | require_lock_membership_approval_until | require_personal_info_access_approval | require_watch_approval | root_activity_id | root_skill_id |
      | 3  | Base  | null       | null                | null          | false             | Base Group     | null                                   | none                                  | false                  | null             | null          |
      | 11 | Team  | 3456789abc | 2037-05-29 06:38:38 | 3723          | false             | Our Team       | null                                   | edit                                  | false                  | 1234             | null          |
      | 12 | Team  | abc3456789 | null                | 45296         | false             | Their Team     | null                                   | none                                  | false                  | null             | null          |
      | 13 | Team  | 456789abcd | 2017-05-29 06:38:38 | 3723          | false             | Some Team      | null                                   | none                                  | false                  | null             | null          |
      | 14 | Team  | cba9876543 | null                | null          | false             | Another Team   | null                                   | none                                  | false                  | null             | null          |
      | 15 | Team  | 987654321a | null                | null          | false             | Someone's Team | null                                   | none                                  | false                  | null             | null          |
      | 16 | Class | dcef123492 | null                | null          | false             | Our Class      | 2037-01-02 12:30:55                    | none                                  | true                   | null             | 5678          |
      | 18 | Team  | 5987654abc | null                | null          | false             | One More Team  | null                                   | none                                  | false                  | null             | null          |
      | 19 | Team  | 87654abcde | null                | null          | true              | Somewhat Team  | null                                   | none                                  | false                  | null             | null          |
      | 21 | User  | null       | null                | null          | false             | john           | null                                   | none                                  | false                  | null             | null          |
      | 22 | User  | 3333333333 | null                | null          | false             | tmp            | null                                   | none                                  | false                  | null             | null          |
      | 23 | User  | null       | null                | null          | false             | jane           | null                                   | none                                  | false                  | null             | null          |
    And the database has the following table "users":
      | group_id | login | temp_user | first_name | last_name |
      | 21       | john  | false     | null       | null      |
      | 22       | tmp   | true      | null       | null      |
      | 23       | jane  | false     | Jane       | Doe       |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 14              | 21             |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id   | default_language_tag |
      | 1234 | fr                   |
    And the database has the following table "attempts":
      | id | participant_id | root_item_id |
      | 0  | 21             | null         |
      | 2  | 14             | 1234         |
      | 2  | 18             | 1234         |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id |
      | 0          | 21             | 30      |
    And the database has the following table "group_managers":
      | group_id | manager_id |
      | 11       | 21         |
      | 11       | 23         |

  Scenario Outline: The code is valid for a normal user
    Given I am the user with id "21"
    When I send a GET request to "/groups/is-code-valid?code=<code>"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": true, "group": <group_info>}
    """
    Examples:
      | code       | group_info                                                                                                                                                                                                                                                                                                                                                             |
      | 3456789abc | {"name": "Our Team", "require_lock_membership_approval_until": null, "require_personal_info_access_approval": "edit", "require_watch_approval": false, "root_activity_id": "1234", "root_skill_id": null, "managers": [{"id": "23", "login": "jane", "first_name": "Jane", "last_name": "Doe"}, {"id": "21", "login": "john", "first_name": null, "last_name": null}]} |
      | dcef123492 | {"name": "Our Class", "require_lock_membership_approval_until": "2037-01-02T12:30:55Z", "require_personal_info_access_approval": "none", "require_watch_approval": true, "root_activity_id": null, "root_skill_id": "5678", "managers": []}                                                                                                                            |

  Scenario Outline: The code is not valid
    Given I am the user with id "21"
    When I send a GET request to "/groups/is-code-valid?code=<code>"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": false, "reason": "<reason>"}
    """
    Examples:
      | code       | reason                         | description                                             |
      | abcdef     | no_group                       | no such code                                            |
      | 456789abcd | no_group                       | expired                                                 |
      | 5987654abc | conflicting_team_participation | a member of another team participating in same contests |
      | cba9876543 | already_member                 | already a member of the group                           |
      | 87654abcde | frozen_membership              | frozen membership                                       |
      | 3333333333 | no_group                       | the group is a user                                     |

  Scenario: The user is temporary
    Given I am the user with id "22"
    And the application config is:
      """
      domains:
        -
          domains: [127.0.0.1]
          allUsersGroup: 3
      """
    When I send a GET request to "/groups/is-code-valid?code=3456789abc"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "valid": true,
      "group": {
        "name": "Our Team", "require_lock_membership_approval_until": null, "require_personal_info_access_approval": "edit",
        "require_watch_approval": false, "root_activity_id": "1234", "root_skill_id": null,
        "managers": [
          {"id": "23", "login": "jane", "first_name": "Jane", "last_name": "Doe"},
          {"id": "21", "login": "john", "first_name": null, "last_name": null}
        ]
      }
    }
    """

  Scenario: The user is temporary (custom all-users group)
    Given I am the user with id "22"
    And the database has the following table "items":
      | id | default_language_tag | allows_multiple_attempts |
      | 2  | fr                   | false                    |
    And the database table "attempts" also has the following row:
      | participant_id | id | root_item_id |
      | 3              | 1  | 2            |
      | 11             | 1  | 2            |
    And the database has the following table "results":
      | participant_id | attempt_id | item_id | started_at          |
      | 3              | 1          | 2       | 2019-05-30 11:00:00 |
      | 11             | 1          | 2       | 2019-05-30 11:00:00 |
    And the application config is:
      """
      domains:
        -
          domains: [127.0.0.1]
          allUsersGroup: 3
      """
    When I send a GET request to "/groups/is-code-valid?code=3456789abc"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": false, "reason": "team_conditions_not_met"}
    """

  Scenario: Joining would break entry conditions for the team
    Given I am the user with id "21"
    And the database has the following table "items":
      | id | default_language_tag | entry_min_admitted_members_ratio |
      | 2  | fr                   | All                              |
    And the database table "attempts" also has the following row:
      | participant_id | id | root_item_id |
      | 12             | 1  | 2            |
    And the database has the following table "results":
      | participant_id | attempt_id | item_id | started_at          |
      | 12             | 1          | 2       | 2019-05-30 11:00:00 |
    When I send a GET request to "/groups/is-code-valid?code=abc3456789"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {"valid": false, "reason": "team_conditions_not_met"}
    """
