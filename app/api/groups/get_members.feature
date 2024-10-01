Feature: Get members of group_id
  Background:
    Given the database has the following users:
      | login | group_id | first_name  | last_name  | grade |
      | owner | 21       | Jean-Michel | Blanquer   | 3     |
      | user  | 11       | John        | Doe        | 1     |
      | jane  | 31       | Jane        | Doe        | 2     |
      | john  | 41       | John        | Connor     | -1    |
      | billg | 51       | Bill        | Gates      | 5     |
      | zuck  | 61       | Mark        | Zuckerberg | 9     |
      | jeff  | 71       | Jeff        | Bezos      | 7     |
      | larry | 81       | Larry       | Ellison    | 8     |
      | lp    | 91       | Larry       | Page       | 6     |
    And the database has the following table "groups":
      | id |
      | 13 |
      | 14 |
      | 22 |
    And the database has the following table "group_managers":
      | group_id | manager_id |
      | 13       | 21         |
      | 13       | 91         |
      | 22       | 21         |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | expires_at          | personal_info_view_approved_at |
      | 13              | 51             | 9999-12-31 23:59:59 | null                           | # still shows personal info because of 22-51
      | 13              | 61             | 9999-12-31 23:59:59 | 2019-05-30 11:00:00            |
      | 13              | 91             | 9999-12-31 23:59:59 | null                           |
      | 14              | 11             | 9999-12-31 23:59:59 | 2019-05-30 11:00:00            |
      | 14              | 31             | 9999-12-31 23:59:59 | 2019-05-30 11:00:00            |
      | 14              | 21             | 2019-05-30 11:00:00 | 2019-05-30 11:00:00            |
      | 14              | 41             | 9999-12-31 23:59:59 | 2019-05-30 11:00:00            |
      | 22              | 13             | 9999-12-31 23:59:59 | null                           |
      | 22              | 51             | 9999-12-31 23:59:59 | 2019-05-30 11:00:00            |
    And the groups ancestors are computed
    And the database has the following table "group_membership_changes":
      | group_id | member_id | action                | at                  | initiator_id |
      | 13       | 11        | invitation_refused    | 2017-11-29 06:38:38 | 31           |
      | 13       | 21        | invitation_created    | 2017-10-29 06:38:38 | 11           |
      | 13       | 31        | join_request_created  | 2017-09-29 06:38:38 | 21           |
      | 13       | 41        | join_request_refused  | 2017-08-29 06:38:38 | 11           |
      | 13       | 51        | invitation_accepted   | 2017-07-29 06:38:38 | 11           |
      | 13       | 61        | join_request_accepted | 2017-06-29 06:38:38 | 31           |
      | 13       | 71        | removed               | 2017-05-29 06:38:38 | 21           |
      | 13       | 81        | left                  | 2017-04-29 06:38:38 | 11           |
      | 13       | 91        | added_directly        | 2017-03-29 06:38:38 | null         |
      | 14       | 11        | invitation_accepted   | 2017-02-28 06:38:38 | 11           |
      | 14       | 31        | join_request_accepted | 2017-01-29 06:38:38 | 31           |
      | 14       | 21        | added_directly        | 2016-12-29 06:38:38 | null         |
      | 14       | 22        | join_request_refused  | 2016-11-29 06:38:38 | 11           |
      | 22       | 13        | added_directly        | 2016-10-29 06:38:38 | null         |

  Scenario: User is a manager of the group (default sort)
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/members"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "user": {
          "first_name": "Bill",
          "group_id": "51",
          "last_name": "Gates",
          "login": "billg",
          "grade": 5
        },
        "member_since": "2017-07-29T06:38:38Z",
        "action": "invitation_accepted"
      },
      {
        "id": "61",
        "user": {
          "first_name": "Mark",
          "group_id": "61",
          "last_name": "Zuckerberg",
          "login": "zuck",
          "grade": 9
        },
        "member_since": "2017-06-29T06:38:38Z",
        "action": "join_request_accepted"
      },
      {
        "id": "91",
        "user": {
          "group_id": "91",
          "login": "lp",
          "grade": 6
        },
        "member_since": "2017-03-29T06:38:38Z",
        "action": "added_directly"
      }
    ]
    """

  Scenario: User is a manager of the group (default sort, different approvals)
    Given I am the user with id "91"
    When I send a GET request to "/groups/13/members"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "user": {
          "group_id": "51",
          "login": "billg",
          "grade": 5
        },
        "member_since": "2017-07-29T06:38:38Z",
        "action": "invitation_accepted"
      },
      {
        "id": "61",
        "user": {
          "first_name": "Mark",
          "group_id": "61",
          "last_name": "Zuckerberg",
          "login": "zuck",
          "grade": 9
        },
        "member_since": "2017-06-29T06:38:38Z",
        "action": "join_request_accepted"
      },
      {
        "id": "91",
        "user": {
          "first_name": "Larry",
          "group_id": "91",
          "last_name": "Page",
          "login": "lp",
          "grade": 6
        },
        "member_since": "2017-03-29T06:38:38Z",
        "action": "added_directly"
      }
    ]
    """

  Scenario: User is a manager of the group (sort by user's grade)
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/members?sort=user.grade,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "user": {
          "first_name": "Bill",
          "group_id": "51",
          "last_name": "Gates",
          "login": "billg",
          "grade": 5
        },
        "member_since": "2017-07-29T06:38:38Z",
        "action": "invitation_accepted"
      },
      {
        "id": "91",
        "user": {
          "group_id": "91",
          "login": "lp",
          "grade": 6
        },
        "member_since": "2017-03-29T06:38:38Z",
        "action": "added_directly"
      },
      {
        "id": "61",
        "user": {
          "first_name": "Mark",
          "group_id": "61",
          "last_name": "Zuckerberg",
          "login": "zuck",
          "grade": 9
        },
        "member_since": "2017-06-29T06:38:38Z",
        "action": "join_request_accepted"
      }
    ]
    """

  Scenario: User is a manager of the group (sort by user's login in descending order)
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/members?sort=-user.login,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "61",
        "user": {
          "first_name": "Mark",
          "group_id": "61",
          "last_name": "Zuckerberg",
          "login": "zuck",
          "grade": 9
        },
        "member_since": "2017-06-29T06:38:38Z",
        "action": "join_request_accepted"
      },
      {
        "id": "91",
        "user": {
          "group_id": "91",
          "login": "lp",
          "grade": 6
        },
        "member_since": "2017-03-29T06:38:38Z",
        "action": "added_directly"
      },
      {
        "id": "51",
        "user": {
          "first_name": "Bill",
          "group_id": "51",
          "last_name": "Gates",
          "login": "billg",
          "grade": 5
        },
        "member_since": "2017-07-29T06:38:38Z",
        "action": "invitation_accepted"
      }
    ]
    """

  Scenario: User is a manager of the group; request the first row
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/members?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "user": {
          "first_name": "Bill",
          "group_id": "51",
          "last_name": "Gates",
          "login": "billg",
          "grade": 5
        },
        "member_since": "2017-07-29T06:38:38Z",
        "action": "invitation_accepted"
      }
    ]
    """

  Scenario: User is a manager of the group; request the second row
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/members?limit=1&from.id=51"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "61",
        "user": {
          "first_name": "Mark",
          "group_id": "61",
          "last_name": "Zuckerberg",
          "login": "zuck",
          "grade": 9
        },
        "member_since": "2017-06-29T06:38:38Z",
        "action": "join_request_accepted"
      }
    ]
    """

  Scenario: Non-user members are not listed
    Given I am the user with id "21"
    When I send a GET request to "/groups/22/members"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "user": {
          "first_name": "Bill",
          "grade": 5,
          "group_id": "51",
          "last_name": "Gates",
          "login": "billg"
        }
      }
    ]
    """

