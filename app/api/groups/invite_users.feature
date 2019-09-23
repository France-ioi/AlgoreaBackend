Feature: Invite users
  Background:
    Given the database has the following table 'users':
      | id | login | self_group_id | owned_group_id | first_name  | last_name |
      | 1  | owner | 21            | 22             | Jean-Michel | Blanquer  |
      | 10 | john  | 101           | 111            | John        | Doe       |
      | 11 | jane  | 102           | 112            | Jane        | Doe       |
      | 12 | Jane  | 103           | 113            | Jane        | Smith     |
    And the database has the following table 'groups':
      | id  | type      | team_item_id |
      | 13  | Team      | 1234         |
      | 21  | UserSelf  | null         |
      | 22  | UserAdmin | null         |
      | 101 | UserSelf  | null         |
      | 102 | UserSelf  | null         |
      | 103 | UserSelf  | null         |
      | 111 | UserAdmin | null         |
      | 112 | UserAdmin | null         |
      | 113 | UserAdmin | null         |
      | 444 | Team      | 1234         |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 13                | 13             | 1       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |
      | 101               | 101            | 1       |
      | 102               | 102            | 1       |
      | 103               | 103            | 1       |
      | 111               | 111            | 1       |
      | 112               | 112            | 1       |
      | 113               | 113            | 1       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | type   | status_date |
      | 22              | 13             | direct | null        |

  Scenario: Successfully invite users
    Given I am the user with id "1"
    And the database table 'groups_ancestors' has also the following rows:
      | ancestor_group_id | child_group_id | is_self |
      | 444               | 444            | 1       |
    When I send a POST request to "/groups/13/invitations" with the following body:
      """
      {
        "logins": ["john", "jane", "owner", "barack"]
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "data": {
          "john": "success",
          "jane": "success",
          "owner": "success",
          "barack": "not_found"
        },
        "message": "created",
        "success": true
      }
      """
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | type           | role   | inviting_user_id | child_order = 0 | (status_date IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, status_date, NOW())) < 3) |
      | 13              | 21             | invitationSent | member | 1                | 0               | 1                                                                                  |
      | 13              | 101            | invitationSent | member | 1                | 0               | 1                                                                                  |
      | 13              | 102            | invitationSent | member | 1                | 0               | 1                                                                                  |
      | 22              | 13             | direct         | member | null             | 1               | 0                                                                                  |
    And the table "groups_groups" should be:
      | child_order |
      | 0           |
      | 1           |
      | 2           |
      | 3           |
    And the table "groups_ancestors" should stay unchanged

  Scenario: Successfully invite users into a team skipping those who are members of other teams with the same team_item_id
    Given I am the user with id "1"
    And the database table 'groups_groups' has also the following rows:
      | parent_group_id | child_group_id | type               | status_date |
      | 444             | 21             | joinedByCode       | null        |
      | 444             | 101            | invitationAccepted | null        |
      | 444             | 102            | requestAccepted    | null        |
    And the database table 'groups_ancestors' has also the following rows:
      | ancestor_group_id | child_group_id | is_self |
      | 444               | 21             | 0       |
      | 444               | 101            | 0       |
      | 444               | 102            | 0       |
      | 444               | 444            | 1       |
    When I send a POST request to "/groups/13/invitations" with the following body:
      """
      {
        "logins": ["john", "jane", "owner", "barack"]
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "data": {
          "john": "in_another_team",
          "jane": "in_another_team",
          "owner": "in_another_team",
          "barack": "not_found"
        },
        "message": "created",
        "success": true
      }
      """
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged
