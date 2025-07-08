Feature: Get group parents (groupParentsView) - robustness
  Background:
    Given the database has the following table "groups":
      | id  | name                                       | type                | is_public |
      | 151 | Grand-parent of a joined group             | Club                | false     |
      | 152 | Parent of a joined group                   | Club                | false     |
      | 153 | Joined group                               | Class               | false     |
      | 154 | Parent of a joined team                    | Club                | false     |
      | 155 | Joined team                                | Team                | false     |
      | 156 | Grand-parent of a managed non-user group   | Club                | false     |
      | 157 | Parent of a managed non-user group         | Club                | false     |
      | 158 | Managed non-user group                     | Class               | false     |
      | 159 | Grand-parent of a managed user group       | Class               | false     |
      | 160 | Parent of a managed user group             | Class               | false     |
      | 161 | Managed user group                         | User                | false     |
      | 162 | Another managed group                      | Club                | false     |
      | 163 | Child of a managed group                   | Class               | false     |
      | 164 | Grand-child of a managed group             | Class               | false     |
      | 165 | Public group                               | Club                | true      |
      | 166 | ContestParticipants group                  | ContestParticipants | true      |
      | 251 | Grand-parent of a joined group 2           | Club                | false     |
      | 252 | Parent of a joined group 2                 | Club                | false     |
      | 253 | Joined group 2                             | Class               | false     |
      | 254 | Parent of a joined team 2                  | Club                | false     |
      | 255 | Joined team 2                              | Team                | false     |
      | 256 | Grand-parent of a managed non-user group 2 | Club                | false     |
      | 257 | Parent of a managed non-user group 2       | Club                | false     |
      | 258 | Managed non-user group 2                   | Class               | false     |
      | 259 | Grand-parent of a managed user group 2     | Class               | false     |
      | 260 | Parent of a managed user group 2           | Class               | false     |
      | 261 | Managed user group 2                       | User                | false     |
      | 262 | Another managed group 2                    | Club                | false     |
      | 263 | Child of a managed group 2                 | Class               | false     |
      | 264 | Grand-child of a managed group 2           | Class               | false     |
      | 265 | Public group 2                             | Club                | true      |
      | 500 | Owner's group                              | Club                | false     |
    And the database has the following users:
      | group_id | login | first_name  | last_name |
      | 21       | owner | Jean-Michel | Blanquer  |
      | 400      | jack  | Jack        | Smith     |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access |
      | 158      | 21         | false                  |
      | 161      | 500        | true                   |
      | 162      | 21         | false                  |
      | 162      | 500        | true                   |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 151             | 152            |
      | 151             | 400            |
      | 152             | 153            |
      | 152             | 400            |
      | 153             | 21             |
      | 153             | 400            |
      | 154             | 155            |
      | 154             | 400            |
      | 155             | 21             |
      | 155             | 400            |
      | 156             | 157            |
      | 156             | 400            |
      | 157             | 158            |
      | 157             | 400            |
      | 158             | 400            |
      | 159             | 160            |
      | 159             | 400            |
      | 160             | 161            |
      | 160             | 400            |
      | 162             | 163            |
      | 162             | 400            |
      | 163             | 164            |
      | 163             | 400            |
      | 164             | 400            |
      | 165             | 400            |
      | 166             | 400            |
      | 251             | 252            |
      | 252             | 253            |
      | 253             | 21             |
      | 254             | 255            |
      | 255             | 21             |
      | 256             | 257            |
      | 257             | 258            |
      | 259             | 260            |
      | 260             | 261            |
      | 262             | 263            |
      | 263             | 264            |
      | 500             | 21             |
    And the groups ancestors are computed

  Scenario: Group id is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/abc/parents"
    Then the response code should be 400
    And the response error message should contain "Wrong value for group_id (should be int64)"

  Scenario: User manages a only a user-group that is a descendant of the group
    Given I am the user with id "21"
    When I send a GET request to "/groups/159/parents"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user doesn't manage any of the group's ancestors/descendants, is not a member of the group's descendants, the group is not public
    Given I am the user with id "21"
    When I send a GET request to "/groups/21/parents"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: sort is incorrect
    Given I am the user with id "21"
    When I send a GET request to "/groups/400/parents?sort=myname"
    Then the response code should be 400
    And the response error message should contain "Unallowed field in sorting parameters: "myname""
