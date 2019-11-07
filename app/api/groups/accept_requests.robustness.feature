Feature: Accept group requests - robustness
  Background:
    Given the database has the following table 'groups':
      | id  |
      | 11  |
      | 13  |
      | 14  |
      | 21  |
      | 22  |
      | 31  |
      | 111 |
      | 121 |
      | 122 |
      | 123 |
      | 131 |
      | 141 |
    And the database has the following table 'users':
      | login | group_id | owned_group_id | first_name  | last_name | grade |
      | owner | 21       | 22             | Jean-Michel | Blanquer  | 3     |
      | user  | 11       | 12             | John        | Doe       | 1     |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 13                | 13             | 1       |
      | 13                | 111            | 0       |
      | 13                | 121            | 0       |
      | 13                | 123            | 0       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |
      | 31                | 31             | 1       |
      | 111               | 111            | 1       |
      | 121               | 121            | 1       |
      | 122               | 122            | 1       |
      | 123               | 123            | 1       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | type               | type_changed_at           |
      | 1  | 13              | 21             | invitationSent     | {{relativeTime("-170h")}} |
      | 2  | 13              | 11             | invitationRefused  | {{relativeTime("-169h")}} |
      | 3  | 13              | 31             | requestSent        | {{relativeTime("-168h")}} |
      | 5  | 14              | 11             | invitationSent     | null                      |
      | 6  | 14              | 31             | invitationRefused  | null                      |
      | 7  | 14              | 21             | requestSent        | null                      |
      | 8  | 14              | 22             | requestRefused     | null                      |
      | 9  | 13              | 121            | invitationAccepted | 2017-05-29 06:38:38       |
      | 10 | 13              | 111            | requestAccepted    | null                      |
      | 11 | 13              | 131            | removed            | null                      |
      | 12 | 13              | 122            | left               | null                      |
      | 13 | 13              | 123            | direct             | null                      |
      | 14 | 13              | 141            | requestSent        | null                      |
      | 15 | 22              | 13             | direct             | null                      |

  Scenario: Fails when the user is not an owner of the parent group
    Given I am the user with id "11"
    When I send a POST request to "/groups/13/requests/accept?group_ids=31,141,21,11,13,22"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the user doesn't exist
    Given I am the user with id "404"
    When I send a POST request to "/groups/13/requests/accept?group_ids=31,141,21,11,13,22"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when the parent group id is wrong
    Given I am the user with id "21"
    When I send a POST request to "/groups/abc/requests/accept?group_ids=31,141,21,11,13,22"
    Then the response code should be 400
    And the response error message should contain "Wrong value for parent_group_id (should be int64)"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Fails when group_ids is wrong
    Given I am the user with id "21"
    When I send a POST request to "/groups/13/requests/accept?group_ids=31,abc,11,13,22"
    Then the response code should be 400
    And the response error message should contain "Unable to parse one of the integers given as query args (value: 'abc', param: 'group_ids')"
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
