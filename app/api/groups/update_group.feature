Feature: Update a group (groupEdit)
  Background:
    Given the database has the following table 'groups':
      | id | name    | grade | description     | created_at          | type  | redirect_path                          | opened | free_access | code       | code_lifetime | code_expires_at     | open_contest |
      | 11 | Group A | -3    | Group A is here | 2019-02-06 09:26:40 | Class | 182529188317717510/1672978871462145361 | true   | true        | ybqybxnlyo | 01:00:00      | 2017-10-13 05:39:48 | true         |
      | 13 | Group B | -2    | Group B is here | 2019-03-06 09:26:40 | Class | 182529188317717610/1672978871462145461 | true   | true        | ybabbxnlyo | 01:00:00      | 2017-10-14 05:39:48 | true         |
      | 14 | Group C | -4    | Group           | 2019-04-06 09:26:40 | Club  | null                                   | true   | false       | null       | null          | null                | false        |
      | 21 | owner   | -4    | owner           | 2019-04-06 09:26:40 | User  | null                                   | false  | false       | null       | null          | null                | false        |
      | 24 | other   | -4    | other           | 2019-04-06 09:26:40 | User  | null                                   | false  | false       | null       | null          | null                | false        |
      | 31 | user    | -4    | owner           | 2019-04-06 09:26:40 | User  | null                                   | false  | false       | null       | null          | null                | false        |
    And the database has the following table 'users':
      | login | temp_user | group_id | first_name  | last_name | default_language |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | fr               |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
      | 14       | 21         |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 13                | 11             |
      | 13                | 13             |
      | 14                | 14             |
      | 21                | 21             |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 23             |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         |
      | 13       | 21        | invitation   |
      | 13       | 24        | join_request |
      | 13       | 31        | join_request |
      | 14       | 31        | join_request |

  Scenario: User is a manager of the group, all fields are not nulls, updates group_pending_requests
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "free_access": false,
      "name": "Team B",
      "grade": 10,
      "description": "Team B is here",
      "opened": false,
      "code_lifetime": "99:59:59",
      "code_expires_at": "2019-12-31T23:59:59Z",
      "open_contest": false,
      "redirect_path": "1234/5678"
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "13"
    And the table "groups" at id "13" should be:
      | id | name   | grade | description    | created_at          | type  | redirect_path | opened | free_access | code       | code_lifetime | code_expires_at     | open_contest |
      | 13 | Team B | 10    | Team B is here | 2019-03-06 09:26:40 | Class | 1234/5678     | false  | false       | ybabbxnlyo | 99:59:59      | 2019-12-31 23:59:59 | false        |
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should be:
      | group_id | member_id | type         |
      | 13       | 21        | invitation   |
      | 14       | 31        | join_request |
    And the table "group_membership_changes" should be:
      | group_id | member_id | action               | ABS(TIMESTAMPDIFF(SECOND, at, NOW())) < 3 |
      | 13       | 24        | join_request_refused | 1                                         |
      | 13       | 31        | join_request_refused | 1                                         |

  Scenario: User is a manager of the group, nullable fields are nulls
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "free_access": false,
      "name": "Club B",
      "description": null,
      "opened": false,
      "open_contest": false,
      "redirect_path": null,
      "grade": 0,
      "code_expires_at": null,
      "code_lifetime": null
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "13"
    And the table "groups" at id "13" should be:
      | id | name   | grade | description | created_at          | type  | redirect_path | opened | free_access | code       | code_lifetime | code_expires_at | open_contest |
      | 13 | Club B | 0     | null        | 2019-03-06 09:26:40 | Class | null          | false  | false       | ybabbxnlyo | null          | null            | false        |

  Scenario: User is a manager of the group, does not update group_pending_requests (free_access is still true)
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "free_access": true,
      "name": "Club B",
      "description": null,
      "opened": false,
      "open_contest": false,
      "redirect_path": null,
      "grade": 0,
      "code_expires_at": null,
      "code_lifetime": null
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "13"
    And the table "groups" at id "13" should be:
      | id | name   | grade | description | created_at          | type  | redirect_path | opened | free_access | code       | code_lifetime | code_expires_at | open_contest |
      | 13 | Club B | 0     | null        | 2019-03-06 09:26:40 | Class | null          | false  | true        | ybabbxnlyo | null          | null            | false        |
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should stay unchanged

  Scenario: User is a manager of the group, does not update group_pending_requests (free_access is not changed)
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "name": "Club B",
      "description": null,
      "opened": false,
      "open_contest": false,
      "redirect_path": null,
      "grade": 0,
      "code_expires_at": null,
      "code_lifetime": null
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "13"
    And the table "groups" at id "13" should be:
      | id | name   | grade | description | created_at          | type  | redirect_path | opened | free_access | code       | code_lifetime | code_expires_at | open_contest |
      | 13 | Club B | 0     | null        | 2019-03-06 09:26:40 | Class | null          | false  | true        | ybabbxnlyo | null          | null            | false        |
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should stay unchanged

  Scenario: User is a manager of the group, does not update group_pending_requests (free_access changes from false to true)
    Given I am the user with id "21"
    When I send a PUT request to "/groups/14" with the following body:
    """
    {
      "free_access": true
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "14"
    And the table "groups" at id "14" should be:
      | id | name    | grade | description | created_at          | type | redirect_path | opened | free_access | code | code_lifetime | code_expires_at | open_contest |
      | 14 | Group C | -4    | Group       | 2019-04-06 09:26:40 | Club | null          | true   | true        | null | null          | null            | false        |
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should stay unchanged
