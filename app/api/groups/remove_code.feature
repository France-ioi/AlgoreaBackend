Feature: Remove the code of the given group

  Background:
    Given the database has the following table 'groups':
      | id | name    | grade | description     | created_at          | type  | code       | code_lifetime | code_expires_at     |
      | 11 | Group A | -3    | Group A is here | 2019-02-06 09:26:40 | Class | ybqybxnlyo | 3600          | 2017-10-13 05:39:48 |
      | 13 | Group B | -2    | Group B is here | 2019-03-06 09:26:40 | Class | 3456789abc | 3600          | 2017-10-14 05:39:48 |
      | 21 | owner   | -4    | owner           | 2019-04-06 09:26:40 | User  | null       | null          | null                |
    And the database has the following table 'users':
      | login | temp_user | group_id | first_name  | last_name | default_language |
      | owner | 0         | 21       | Jean-Michel | Blanquer  | fr               |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_manage  |
      | 13       | 21         | memberships |
    And the groups ancestors are computed

  Scenario: User is a manager of the group
    Given I am the user with id "21"
    When I send a DELETE request to "/groups/13/code"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "deleted"
    }
    """
    And the table "groups" should stay unchanged but the row with id "13"
    And the table "groups" at id "13" should be:
      | id | name    | grade | description     | created_at          | type  | code | code_lifetime | code_expires_at     |
      | 13 | Group B | -2    | Group B is here | 2019-03-06 09:26:40 | Class | null | 3600          | 2017-10-14 05:39:48 |
