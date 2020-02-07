Feature: Update a group (groupEdit)
  Background:
    Given the database has the following table 'groups':
      | id | name    | grade | description     | created_at          | type    | activity_id         | is_open | is_public | code       | code_lifetime | code_expires_at     | open_contest | is_official_session | require_members_to_join_parent | organizer | address_line1 | address_line2 | address_postcode | address_city | address_country | expected_start |
      | 11 | Group A | -3    | Group A is here | 2019-02-06 09:26:40 | Class   | 1672978871462145361 | true    | true      | ybqybxnlyo | 01:00:00      | 2017-10-13 05:39:48 | true         | false               | false                          | null      | null          | null          | null             | null         | null            | null           |
      | 13 | Group B | -2    | Group B is here | 2019-03-06 09:26:40 | Class   | 1672978871462145461 | true    | true      | ybabbxnlyo | 01:00:00      | 2017-10-14 05:39:48 | true         | false               | false                          | null      | null          | null          | null             | null         | null            | null           |
      | 14 | Group C | -4    | Group C         | 2019-04-06 09:26:40 | Club    | null                | true    | false     | null       | null          | null                | false        | false               | false                          | null      | null          | null          | null             | null         | null            | null           |
      | 15 | Group D | -4    | Group D         | 2019-04-06 09:26:40 | Session | null                | true    | false     | null       | null          | null                | false        | true                | false                          | null      | null          | null          | null             | null         | null            | null           |
      | 21 | owner   | -4    | owner           | 2019-04-06 09:26:40 | User    | null                | false   | false     | null       | null          | null                | false        | false               | false                          | null      | null          | null          | null             | null         | null            | null           |
      | 24 | other   | -4    | other           | 2019-04-06 09:26:40 | User    | null                | false   | false     | null       | null          | null                | false        | false               | false                          | null      | null          | null          | null             | null         | null            | null           |
      | 31 | user    | -4    | owner           | 2019-04-06 09:26:40 | User    | null                | false   | false     | null       | null          | null                | false        | false               | false                          | null      | null          | null          | null             | null         | null            | null           |
      | 50 | Admins  | -4    | Admins          | 2019-04-06 09:26:40 | Club    | null                | false   | false     | null       | null          | null                | false        | false               | false                          | null      | null          | null          | null             | null         | null            | null           |
    And the database has the following table 'users':
      | login | temp_user | group_id | first_name  | last_name | default_language |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | fr               |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
      | 14       | 21         |
      | 15       | 50         |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 13                | 11             |
      | 13                | 13             |
      | 14                | 14             |
      | 15                | 15             |
      | 21                | 21             |
      | 50                | 21             |
      | 50                | 50             |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 23             |
      | 50              | 21             |
    And the database has the following table 'group_pending_requests':
      | group_id | member_id | type         |
      | 13       | 21        | invitation   |
      | 13       | 24        | join_request |
      | 13       | 31        | join_request |
      | 14       | 31        | join_request |
    And the database has the following table 'items':
      | id   | default_language_tag |
      | 123  | fr                   |
      | 5678 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated |
      | 21       | 5678    | info               |
      | 50       | 123     | info               |
    And the database has the following table 'permissions_granted':
      | group_id | item_id             | can_make_session_official | source_group_id |
      | 21       | 1672978871462145461 | true                      | 13              |
      | 50       | 123                 | true                      | 13              |

  Scenario: User is a manager of the group, all fields are not nulls, updates group_pending_requests
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "is_public": false,
      "name": "Team B",
      "grade": 10,
      "description": "Team B is here",
      "is_open": false,
      "code_lifetime": "99:59:59",
      "code_expires_at": "2019-12-31T23:59:59Z",
      "open_contest": false,
      "activity_id": "5678",

      "require_members_to_join_parent": true,
      "organizer": "Association France-ioi",
      "address_line1": "Chez Jacques-Henri Jourdan,",
      "address_line2": "42, rue de Cronstadt",
      "address_postcode": "75015",
      "address_city": "Paris",
      "address_country": "France",
      "expected_start": "2019-05-03T12:00:00+01:00"
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "13"
    And the table "groups" at id "13" should be:
      | id | name   | grade | description    | created_at          | type  | activity_id | is_open | is_public | code       | code_lifetime | code_expires_at     | open_contest | require_members_to_join_parent | organizer              | address_line1               | address_line2        | address_postcode | address_city | address_country | expected_start      |
      | 13 | Team B | 10    | Team B is here | 2019-03-06 09:26:40 | Class | 5678        | false   | false     | ybabbxnlyo | 99:59:59      | 2019-12-31 23:59:59 | false        | true                           | Association France-ioi | Chez Jacques-Henri Jourdan, | 42, rue de Cronstadt | 75015            | Paris        | France          | 2019-05-03 11:00:00 |
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
      "is_public": false,
      "name": "Club B",
      "description": null,
      "is_open": false,
      "open_contest": false,
      "activity_id": null,
      "grade": 0,
      "code_expires_at": null,
      "code_lifetime": null
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "13"
    And the table "groups" at id "13" should be:
      | id | name   | grade | description | created_at          | type  | activity_id | is_open | is_public | code       | code_lifetime | code_expires_at | open_contest |
      | 13 | Club B | 0     | null        | 2019-03-06 09:26:40 | Class | null        | false   | false     | ybabbxnlyo | null          | null            | false        |

  Scenario: User is a manager of the group, does not update group_pending_requests (is_public is still true)
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "is_public": true,
      "name": "Club B",
      "description": null,
      "is_open": false,
      "open_contest": false,
      "activity_id": null,
      "grade": 0,
      "code_expires_at": null,
      "code_lifetime": null
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "13"
    And the table "groups" at id "13" should be:
      | id | name   | grade | description | created_at          | type  | activity_id | is_open | is_public | code       | code_lifetime | code_expires_at | open_contest |
      | 13 | Club B | 0     | null        | 2019-03-06 09:26:40 | Class | null        | false   | true      | ybabbxnlyo | null          | null            | false        |
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should stay unchanged

  Scenario: User is a manager of the group, does not update group_pending_requests (is_public is not changed)
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "name": "Club B",
      "description": null,
      "is_open": false,
      "open_contest": false,
      "activity_id": null,
      "grade": 0,
      "code_expires_at": null,
      "code_lifetime": null
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "13"
    And the table "groups" at id "13" should be:
      | id | name   | grade | description | created_at          | type  | activity_id | is_open | is_public | code       | code_lifetime | code_expires_at | open_contest |
      | 13 | Club B | 0     | null        | 2019-03-06 09:26:40 | Class | null        | false   | true      | ybabbxnlyo | null          | null            | false        |
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should stay unchanged

  Scenario: User is a manager of the group, does not update group_pending_requests (is_public changes from false to true)
    Given I am the user with id "21"
    When I send a PUT request to "/groups/14" with the following body:
    """
    {
      "is_public": true
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "14"
    And the table "groups" at id "14" should be:
      | id | name    | grade | description | created_at          | type | activity_id | is_official_session | is_open | is_public | code | code_lifetime | code_expires_at | open_contest |
      | 14 | Group C | -4    | Group C     | 2019-04-06 09:26:40 | Club | null        | false               | true    | true      | null | null          | null            | false        |
    And the table "groups_groups" should stay unchanged
    And the table "group_pending_requests" should stay unchanged
    And the table "group_membership_changes" should stay unchanged

  Scenario: User is a manager of the group, but no fields provided
    Given I am the user with id "21"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged
    And the table "groups_groups" should stay unchanged

  Scenario: User attaches the group to the activity as an official session
    Given I am the user with id "21"
    When I send a PUT request to "/groups/14" with the following body:
    """
    {
      "activity_id": "123",
      "is_official_session": true
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "14"
    And the table "groups" at id "14" should be:
      | id | name    | grade | description | created_at          | type | activity_id | is_official_session | is_open | is_public | code | code_lifetime | code_expires_at | open_contest |
      | 14 | Group C | -4    | Group C     | 2019-04-06 09:26:40 | Club | 123         | true                | true    | false     | null | null          | null            | false        |
    And the table "groups_groups" should stay unchanged

  Scenario: User replaces the activity of the official session
    Given I am the user with id "21"
    When I send a PUT request to "/groups/15" with the following body:
    """
    {
      "activity_id": "123"
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with id "15"
    And the table "groups" at id "15" should be:
      | id | name    | grade | description | created_at          | type    | activity_id | is_official_session | is_open | is_public | code | code_lifetime | code_expires_at | open_contest |
      | 15 | Group D | -4    | Group D     | 2019-04-06 09:26:40 | Session | 123         | true                | true    | false     | null | null          | null            | false        |
    And the table "groups_groups" should stay unchanged
