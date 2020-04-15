Feature: Get group by groupID (groupView)
  Background:
    Given the database has the following table 'groups':
      | id | name    | grade | description     | created_at          | type  | activity_id         | is_open | is_public | code       | code_lifetime | code_expires_at     | open_contest |
      | 11 | Group A | -3    | Group A is here | 2019-02-06 09:26:40 | Class | 1672978871462145361 | true    | false     | ybqybxnlyo | 01:00:00      | 2017-10-13 05:39:48 | true         |
      | 13 | Group B | -2    | Group B is here | 2019-03-06 09:26:40 | Class | 1672978871462145461 | true    | false     | ybabbxnlyo | 01:00:00      | 2017-10-14 05:39:48 | true         |
      | 15 | Group D | -4    | Other Group     | 2019-04-06 09:26:40 | Other | null                | false   | true      | abcdefghij | null          | null                | false        |
      | 21 | owner   | 0     | null            | 2019-01-06 09:26:40 | User  | null                | false   | false     | null       | null          | null                | false        |
      | 31 | john    | 0     | null            | 2019-01-06 09:26:40 | User  | null                | false   | false     | null       | null          | null                | false        |
      | 41 | jane    | 0     | null            | 2019-01-06 09:26:40 | User  | null                | false   | false     | null       | null          | null                | false        |
      | 51 | rick    | 0     | null            | 2019-01-06 09:26:40 | User  | null                | false   | false     | null       | null          | null                | false        |
      | 61 | ian     | 0     | null            | 2019-01-06 09:26:40 | User  | null                | false   | false     | null       | null          | null                | false        |
      | 71 | dirk    | 0     | null            | 2019-01-06 09:26:40 | User  | null                | false   | false     | null       | null          | null                | false        |
      | 81 | chuck   | 0     | null            | 2019-01-06 09:26:40 | User  | null                | false   | false     | null       | null          | null                | false        |
    And the database has the following table 'users':
      | login | group_id |
      | owner | 21       |
      | john  | 31       |
      | jane  | 41       |
      | rick  | 51       |
      | ian   | 61       |
      | dirk  | 71       |
      | chuck | 81       |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 11              | 31             |
      | 13              | 11             |
      | 13              | 51             |
      | 13              | 61             |
      | 13              | 71             |
      | 13              | 81             |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 13       | 21         |

  Scenario: The user is a manager of the group
    Given I am the user with id "21"
    When I send a GET request to "/groups/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "13",
      "name": "Group B",
      "grade": -2,
      "description": "Group B is here",
      "created_at": "2019-03-06T09:26:40Z",
      "type": "Class",
      "activity_id": "1672978871462145461",
      "is_open": true,
      "is_public": false,
      "code": "ybabbxnlyo",
      "code_lifetime": "01:00:00",
      "code_expires_at": "2017-10-14T05:39:48Z",
      "open_contest": true,
      "current_user_is_manager": true,
      "current_user_is_member": false
    }
    """

  Scenario: The user is a manager of the group's ancestor
    Given I am the user with id "21"
    When I send a GET request to "/groups/11"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "11",
      "name": "Group A",
      "grade": -3,
      "description": "Group A is here",
      "created_at": "2019-02-06T09:26:40Z",
      "type": "Class",
      "activity_id": "1672978871462145361",
      "is_open": true,
      "is_public": false,
      "code": "ybqybxnlyo",
      "code_lifetime": "01:00:00",
      "code_expires_at": "2017-10-13T05:39:48Z",
      "open_contest": true,
      "current_user_is_manager": true,
      "current_user_is_member": false
    }
    """

  Scenario: The user is a descendant of the group
    Given I am the user with id "31"
    When I send a GET request to "/groups/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "13",
      "name": "Group B",
      "grade": -2,
      "description": "Group B is here",
      "created_at": "2019-03-06T09:26:40Z",
      "type": "Class",
      "activity_id": "1672978871462145461",
      "is_open": true,
      "is_public": false,
      "open_contest": true,
      "current_user_is_manager": false,
      "current_user_is_member": false
    }
    """

  Scenario Outline: The user is a member of the group
    Given I am the user with id "<user_id>"
    When I send a GET request to "/groups/13"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "13",
      "name": "Group B",
      "grade": -2,
      "description": "Group B is here",
      "created_at": "2019-03-06T09:26:40Z",
      "type": "Class",
      "activity_id": "1672978871462145461",
      "is_open": true,
      "is_public": false,
      "open_contest": true,
      "current_user_is_manager": false,
      "current_user_is_member": true
    }
    """
  Examples:
    | user_id |
    | 51      |
    | 61      |
    | 71      |
    | 81      |

  Scenario: The group has is_public = 1
    Given I am the user with id "41"
    When I send a GET request to "/groups/15"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": "15",
      "name": "Group D",
      "grade": -4,
      "description": "Other Group",
      "created_at": "2019-04-06T09:26:40Z",
      "type": "Other",
      "activity_id": null,
      "is_open": false,
      "is_public": true,
      "open_contest": false,
      "current_user_is_manager": false,
      "current_user_is_member": false
    }
    """
