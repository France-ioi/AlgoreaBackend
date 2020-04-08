Feature: Get group children (groupChildrenView) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name    | grade | type  | is_open | is_public | code       |
      | 11 | Group A | -3    | Class | true    | true      | ybqybxnlyo |
      | 13 | Group B | -2    | Class | true    | true      | ybabbxnlyo |
    And the database has the following users:
      | login | temp_user | group_id | first_name  | last_name | default_language |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | fr               |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |
    And the groups ancestors are computed

  Scenario: User is not a manager of the parent group
    Given I am the user with id "21"
    When I send a GET request to "/groups/11/children"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/groups/11/children"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Invalid group_id given
    Given I am the user with id "21"
    When I send a GET request to "/groups/1_1/children"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Invalid sorting rules given
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?sort=code"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "code""

  Scenario: Invalid type in types_include
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?types_include=Teacher"
    Then the response code should be 400
    And the response error message should contain "Wrong value in 'types_include': "Teacher""

  Scenario: Invalid type in types_exclude
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/children?types_exclude=Manager"
    Then the response code should be 400
    And the response error message should contain "Wrong value in 'types_exclude': "Manager""
