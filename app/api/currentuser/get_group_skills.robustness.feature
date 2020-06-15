Feature: Get root skills for a participant group - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name      | text_id | grade | type  | root_activity_id | created_at          |
      | 11 | jdoe      |         | -2    | User  | null             | 2019-01-30 08:26:48 |
      | 13 | Group B   |         | -2    | Team  | 230              | 2019-01-30 08:26:46 |
    And the database has the following table 'languages':
      | tag |
      | fr  |
    And the database has the following table 'users':
      | login     | temp_user | group_id |
      | jdoe      | 0         | 11       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 1               | 11             |
    And the groups ancestors are computed

  Scenario: Wrong value for as_team_id
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships/skills?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: The current user is not a member of the as_team_id team
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships/skills?as_team_id=13"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: as_team_id is not a team
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships/skills?as_team_id=1"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"
