Feature: Get pending requests for managed groups - robustness
  Background:
    Given the database has the following users:
      | login | temp_user | group_id | first_name  | last_name | grade |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | 3     |
      | user  | 0         | 11       | John        | Doe       | 1     |
      | jane  | 0         | 31       | Jane        | Doe       | 2     |
    And the database has the following table "groups":
      | id |
      | 13 |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage  |
      | 13       | 21         | memberships |
      | 13       | 31         | none        |
      | 21       | 31         | memberships |
    And the groups ancestors are computed

  Scenario: invalid group_id
    Given I am the user with id "11"
    When I send a GET request to "/groups/user-requests?group_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: invalid include_descendant_groups
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=13&include_descendant_groups=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for include_descendant_groups (should have a boolean value (0 or 1))"

  Scenario: include_descendant_groups is given while group_id is not given
    Given I am the user with id "11"
    When I send a GET request to "/groups/user-requests?include_descendant_groups=1"
    Then the response code should be 400
    And the response error message should contain "'include_descendant_groups' should not be given when 'group_id' is not given"

  Scenario: invalid types
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=13&types=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value in 'types': "abc""

  Scenario: User is not a manager of the group_id
    Given I am the user with id "11"
    When I send a GET request to "/groups/user-requests?group_id=13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User is a manager of the group_id, but doesn't have enough permissions on it
    Given I am the user with id "31"
    When I send a GET request to "/groups/user-requests?group_id=13"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User has enough permissions on the group_id, but the group_id is a user
    Given I am the user with id "31"
    When I send a GET request to "/groups/user-requests?group_id=21"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/groups/user-requests?group_id=13"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: sort is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/user-requests?group_id=13&sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""
