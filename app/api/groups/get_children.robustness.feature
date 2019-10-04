Feature: Get group children (groupChildrenView) - robustness
  Background:
    Given the database has the following table 'users':
      | id | login | temp_user | self_group_id | owned_group_id | first_name  | last_name | default_language |
      | 1  | owner | 0         | 21            | 22             | Jean-Michel | Blanquer  | fr               |
    And the database has the following table 'groups_ancestors':
      | id | ancestor_group_id | child_group_id | is_self |
      | 75 | 22                | 13             | 0       |
    And the database has the following table 'groups':
      | id | name    | grade | type  | opened | free_access | code       |
      | 11 | Group A | -3    | Class | true   | true        | ybqybxnlyo |
      | 13 | Group B | -2    | Class | true   | true        | ybabbxnlyo |

  Scenario: User is not an owner of the parent group
    Given I am the user with id "1"
    When I send a GET request to "/groups/11/children"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User doesn't exist
    Given I am the user with id "10"
    When I send a GET request to "/groups/11/children"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: Invalid group_id given
    Given I am the user with id "1"
    When I send a GET request to "/groups/1_1/children"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: Invalid sorting rules given
    Given I am the user with id "1"
    When I send a GET request to "/groups/13/children?sort=code"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "code""

  Scenario: Invalid type in types_include
    Given I am the user with id "1"
    When I send a GET request to "/groups/13/children?types_include=Teacher"
    Then the response code should be 400
    And the response error message should contain "Wrong value in 'types_include': "Teacher""

  Scenario: Invalid type in types_exclude
    Given I am the user with id "1"
    When I send a GET request to "/groups/13/children?types_exclude=Manager"
    Then the response code should be 400
    And the response error message should contain "Wrong value in 'types_exclude': "Manager""
