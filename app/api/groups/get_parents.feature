Feature: Get group parents (groupParentsView)
  Background:
    Given the database has the following table "groups":
      | id  | name                                       | type                | is_public |
      | 21  | owner                                      | User                | false     |
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
      | 400 | jack                                       | User                | false     |
      | 500 | Owner's group                              | Club                | false     |
    And the database has the following table "users":
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | jack  | 400      | Jack        | Smith     |
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

  Scenario: User is a manager of one of the group's ancestors, rows are sorted by name by default (also checks that ContestParticipants are skipped)
    Given I am the user with id "21"
    When I send a GET request to "/groups/400/parents"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "162", "name": "Another managed group", "type": "Club", "current_user_can_grant_group_access": true},
      {"id": "163", "name": "Child of a managed group", "type": "Class", "current_user_can_grant_group_access": true},
      {"id": "164", "name": "Grand-child of a managed group", "type": "Class", "current_user_can_grant_group_access": true},
      {"id": "151", "name": "Grand-parent of a joined group", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "156", "name": "Grand-parent of a managed non-user group", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "153", "name": "Joined group", "type": "Class", "current_user_can_grant_group_access": false},
      {"id": "155", "name": "Joined team", "type": "Team", "current_user_can_grant_group_access": false},
      {"id": "158", "name": "Managed non-user group", "type": "Class", "current_user_can_grant_group_access": false},
      {"id": "152", "name": "Parent of a joined group", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "154", "name": "Parent of a joined team", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "157", "name": "Parent of a managed non-user group", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "165", "name": "Public group", "type": "Club", "current_user_can_grant_group_access": false}
    ]
    """

  Scenario: User is a manager of the group
    Given I am the user with id "21"
    When I send a GET request to "/groups/158/parents"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "157", "name": "Parent of a managed non-user group", "type": "Club", "current_user_can_grant_group_access": false}
    ]
    """

  Scenario: User is a manager of one of the group's descendants
    Given I am the user with id "21"
    When I send a GET request to "/groups/163/parents"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "162", "name": "Another managed group", "type": "Club", "current_user_can_grant_group_access": true}
    ]
    """

  Scenario: User is a member of the group
    Given I am the user with id "21"
    When I send a GET request to "/groups/153/parents"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "152", "name": "Parent of a joined group", "type": "Club", "current_user_can_grant_group_access": false}
    ]
    """

  Scenario: User is a member of one of the group's descendants
    Given I am the user with id "21"
    When I send a GET request to "/groups/152/parents"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "151", "name": "Grand-parent of a joined group", "type": "Club", "current_user_can_grant_group_access": false}
    ]
    """

  Scenario: User is a member of the team
    Given I am the user with id "21"
    When I send a GET request to "/groups/155/parents"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "154", "name": "Parent of a joined team", "type": "Club", "current_user_can_grant_group_access": false}
    ]
    """

  Scenario: User is a member of one of the group's descendant teams
    Given I am the user with id "21"
    When I send a GET request to "/groups/154/parents"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: The group is public
    Given I am the user with id "21"
    When I send a GET request to "/groups/165/parents"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: User is a manager of one of the group's ancestors, rows are sorted by name in descending order, get only two rows
    Given I am the user with id "21"
    When I send a GET request to "/groups/400/parents?sort=-name&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "165", "name": "Public group", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "157", "name": "Parent of a managed non-user group", "type": "Club", "current_user_can_grant_group_access": false}
    ]
    """

  Scenario: User is a manager of one of the group's ancestors, rows are sorted by id
    Given I am the user with id "21"
    When I send a GET request to "/groups/400/parents?sort=id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "151", "name": "Grand-parent of a joined group", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "152", "name": "Parent of a joined group", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "153", "name": "Joined group", "type": "Class", "current_user_can_grant_group_access": false},
      {"id": "154", "name": "Parent of a joined team", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "155", "name": "Joined team", "type": "Team", "current_user_can_grant_group_access": false},
      {"id": "156", "name": "Grand-parent of a managed non-user group", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "157", "name": "Parent of a managed non-user group", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "158", "name": "Managed non-user group", "type": "Class", "current_user_can_grant_group_access": false},
      {"id": "162", "name": "Another managed group", "type": "Club", "current_user_can_grant_group_access": true},
      {"id": "163", "name": "Child of a managed group", "type": "Class", "current_user_can_grant_group_access": true},
      {"id": "164", "name": "Grand-child of a managed group", "type": "Class", "current_user_can_grant_group_access": true},
      {"id": "165", "name": "Public group", "type": "Club", "current_user_can_grant_group_access": false}
    ]
    """

  Scenario: User is a manager of one of the group's ancestors, rows are sorted by id, only the second and the third rows
    Given I am the user with id "21"
    When I send a GET request to "/groups/400/parents?sort=id&from.id=151&limit=2"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "152", "name": "Parent of a joined group", "type": "Club", "current_user_can_grant_group_access": false},
      {"id": "153", "name": "Joined group", "type": "Class", "current_user_can_grant_group_access": false}
    ]
    """
