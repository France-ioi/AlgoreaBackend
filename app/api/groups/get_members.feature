Feature: Get members of group_id
  Background:
    Given the database has the following users:
      | login | group_id | owned_group_id | first_name  | last_name  | grade |
      | owner | 21       | 22             | Jean-Michel | Blanquer   | 3     |
      | user  | 11       | 12             | John        | Doe        | 1     |
      | jane  | 31       | 32             | Jane        | Doe        | 2     |
      | john  | 41       | 42             | John        | Connor     | -1    |
      | billg | 51       | 52             | Bill        | Gates      | 5     |
      | zuck  | 61       | 62             | Mark        | Zuckerberg | 9     |
      | jeff  | 71       | 72             | Jeff        | Bezos      | 7     |
      | larry | 81       | 82             | Larry       | Ellison    | 8     |
      | lp    | 91       | 92             | Larry       | Page       | 6     |
    And the database has the following table 'groups':
      | id |
      | 13 |
      | 14 |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 11                | 11             | 1       |
      | 12                | 12             | 1       |
      | 13                | 13             | 1       |
      | 13                | 51             | 0       |
      | 13                | 61             | 0       |
      | 13                | 91             | 0       |
      | 14                | 11             | 0       |
      | 14                | 14             | 1       |
      | 14                | 21             | 0       |
      | 14                | 31             | 0       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 22             | 1       |
      | 22                | 51             | 0       |
      | 22                | 61             | 0       |
      | 22                | 91             | 0       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
      | 41                | 41             | 1       |
      | 42                | 42             | 1       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
      | 61                | 61             | 1       |
      | 62                | 62             | 1       |
      | 71                | 71             | 1       |
      | 72                | 72             | 1       |
      | 81                | 81             | 1       |
      | 82                | 82             | 1       |
      | 91                | 91             | 1       |
      | 92                | 92             | 1       |
      | 22                | 11             | 0       |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | expires_at          |
      | 9  | 13              | 51             | 9999-12-31 23:59:59 |
      | 10 | 13              | 61             | 9999-12-31 23:59:59 |
      | 13 | 13              | 91             | 9999-12-31 23:59:59 |
      | 5  | 14              | 11             | 9999-12-31 23:59:59 |
      | 6  | 14              | 31             | 9999-12-31 23:59:59 |
      | 7  | 14              | 21             | 2019-05-30 11:00:00 |
      | 8  | 14              | 41             | 9999-12-31 23:59:59 |
      | 15 | 22              | 13             | 9999-12-31 23:59:59 |
    And the database has the following table 'group_membership_changes':
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

  Scenario: User is an admin of the group (default sort)
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/members"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "9",
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
        "id": "10",
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
        "id": "13",
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

  Scenario: User is an admin of the group (sort by user's grade)
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/members?sort=user.grade,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "9",
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
        "id": "13",
        "user": {
          "first_name": "Larry",
          "group_id": "91",
          "last_name": "Page",
          "login": "lp",
          "grade": 6
        },
        "member_since": "2017-03-29T06:38:38Z",
        "action": "added_directly"
      },
      {
        "id": "10",
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

  Scenario: User is an admin of the group (sort by user's login in descending order)
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/members?sort=-user.login,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "10",
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
        "id": "13",
        "user": {
          "first_name": "Larry",
          "group_id": "91",
          "last_name": "Page",
          "login": "lp",
          "grade": 6
        },
        "member_since": "2017-03-29T06:38:38Z",
        "action": "added_directly"
      },
      {
        "id": "9",
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

  Scenario: User is an admin of the group; request the first row
    Given I am the user with id "21"
    When I send a GET request to "/groups/13/members?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "9",
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

  Scenario: The member is not a user
    Given I am the user with id "21"
    When I send a GET request to "/groups/22/members?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "15",
        "user": null,
        "member_since": "2016-10-29T06:38:38Z",
        "action": "added_directly"
      }
    ]
    """

