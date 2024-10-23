Feature: Get root activities for a participant group - robustness
  Background:
    Given the database has the following table "groups":
      | id | name      | type  | root_activity_id | created_at          |
      | 11 | jdoe      | User  | null             | 2019-01-30 08:26:48 |
      | 13 | Group B   | Team  | 230              | 2019-01-30 08:26:46 |
      | 14 | Group C   | Team  | 230              | 2019-01-30 08:26:46 |
    And the database has the following table "languages":
      | tag |
      | fr  |
    And the database has the following user:
      | group_id | login |
      | 11       | jdoe  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 1               | 11             |
      | 14              | 11             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | manager_id | group_id | can_watch_members |
      | 11         | 13       | true              |

  Scenario: Wrong value for as_team_id
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships/activities?as_team_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for as_team_id (should be int64)"

  Scenario: The current user is not a member of the as_team_id team
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships/activities?as_team_id=13"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: as_team_id is not a team
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships/activities?as_team_id=1"
    Then the response code should be 403
    And the response error message should contain "Can't use given as_team_id as a user's team"

  Scenario: watched_group_id is invalid
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships/activities?watched_group_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for watched_group_id (should be int64)"

  Scenario: Both watched_group_id and as_team_id are given
    Given I am the user with id "11"
    When I send a GET request to "/current-user/group-memberships/activities?watched_group_id=13&as_team_id=14"
    Then the response code should be 400
    And the response error message should contain "Only one of as_team_id and watched_group_id can be given"
