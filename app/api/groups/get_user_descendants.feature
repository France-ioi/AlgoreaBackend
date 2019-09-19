Feature: List user descendants of the group (groupUserDescendantView)
  Background:
    Given the database has the following table 'users':
      | id | login | group_self_id | group_owned_id | first_name  | last_name | grade |
      | 1  | owner | 21            | 22             | Jean-Michel | Blanquer  | 10    |
    And the database has the following table 'groups':
      | id | type      | name           | grade |
      | 1  | Base      | Root 1         | -2    |
      | 3  | Base      | Root 2         | -2    |
      | 11 | Class     | Our Class      | -2    |
      | 12 | Class     | Other Class    | -2    |
      | 13 | Class     | Special Class  | -2    |
      | 14 | Team      | Super Team     | -2    |
      | 15 | Team      | Our Team       | -1    |
      | 16 | Team      | First Team     | 0     |
      | 17 | Other     | A custom group | -2    |
      | 18 | Club      | Our Club       | -2    |
      | 20 | Friends   | My Friends     | -2    |
      | 21 | UserSelf  | owner          | -2    |
      | 22 | UserAdmin | owner-admin    | -2    |
    And the database has the following table 'groups_groups':
      | group_parent_id | group_child_id | type   |
      | 1               | 11             | direct |
      | 3               | 13             | direct |
      | 3               | 15             | direct |
      | 11              | 14             | direct |
      | 11              | 16             | direct |
      | 11              | 17             | direct |
      | 11              | 18             | direct |
      | 13              | 14             | direct |
      | 13              | 15             | direct |
    And the database has the following table 'groups_ancestors':
      | group_ancestor_id | group_child_id | is_self |
      | 1                 | 1              | 1       |
      | 1                 | 11             | 0       |
      | 1                 | 12             | 0       |
      | 1                 | 14             | 0       |
      | 1                 | 16             | 0       |
      | 1                 | 17             | 0       |
      | 1                 | 18             | 0       |
      | 3                 | 3              | 1       |
      | 3                 | 13             | 0       |
      | 3                 | 15             | 0       |
      | 11                | 11             | 1       |
      | 11                | 14             | 0       |
      | 11                | 16             | 0       |
      | 11                | 17             | 0       |
      | 11                | 18             | 0       |
      | 12                | 12             | 1       |
      | 13                | 13             | 1       |
      | 13                | 14             | 0       |
      | 13                | 15             | 0       |
      | 14                | 14             | 1       |
      | 15                | 15             | 1       |
      | 16                | 16             | 1       |
      | 20                | 20             | 1       |
      | 20                | 21             | 0       |
      | 21                | 21             | 1       |
      | 22                | 1              | 0       |
      | 22                | 11             | 0       |
      | 22                | 12             | 0       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 15             | 0       |
      | 22                | 16             | 0       |
      | 22                | 17             | 0       |
      | 22                | 18             | 0       |
      | 22                | 22             | 1       |

  Scenario: One group with 4 grand children (different parents; one connected as "direct", one as "invitationAccepted", one as "requestAccepted", one as "joinedByCode")
    Given the database table 'users' has also the following rows:
      | id | login | group_self_id | group_owned_id | first_name | last_name | grade |
      | 11 | johna | 51            | 52             | null       | Adams     | 1     |
      | 12 | johnb | 53            | 54             | John       | Baker     | null  |
      | 13 | johnc | 55            | 56             | John       | null      | 3     |
      | 14 | johnd | 57            | 58             | John       | Doe       | 3     |
    And the database table 'groups' has also the following rows:
      | id | type      | name        | grade |
      | 51 | UserSelf  | johna       | -2    |
      | 52 | UserAdmin | johna-admin | -2    |
      | 53 | UserSelf  | johnb       | -2    |
      | 54 | UserAdmin | johnb-admin | -2    |
      | 55 | UserSelf  | johnc       | -2    |
      | 56 | UserAdmin | johnc-admin | -2    |
      | 57 | UserSelf  | johnd       | -2    |
      | 58 | UserAdmin | johnd-admin | -2    |
    And the database table 'groups_groups' has also the following rows:
      | group_parent_id | group_child_id | type               |
      | 11              | 51             | invitationAccepted |
      | 17              | 53             | requestAccepted    |
      | 16              | 55             | direct             |
      | 18              | 57             | joinedByCode       |
    And the database table 'groups_ancestors' has also the following rows:
      | group_ancestor_id | group_child_id | is_self |
      | 1                 | 51             | 0       |
      | 1                 | 53             | 0       |
      | 1                 | 55             | 0       |
      | 1                 | 57             | 0       |
      | 3                 | 53             | 0       |
      | 11                | 51             | 0       |
      | 11                | 53             | 0       |
      | 11                | 55             | 0       |
      | 11                | 57             | 0       |
      | 16                | 55             | 0       |
      | 17                | 53             | 0       |
      | 18                | 57             | 0       |
      | 22                | 51             | 0       |
      | 22                | 53             | 0       |
      | 22                | 55             | 0       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
      | 53                | 53             | 1       |
      | 54                | 54             | 1       |
      | 55                | 55             | 1       |
      | 56                | 56             | 1       |
    And I am the user with id "1"
    When I send a GET request to "/groups/1/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "name": "johna",
        "parents": [{"id": "11", "name": "Our Class"}],
        "user": {"first_name": null, "grade": 1, "id": "11", "last_name": "Adams", "login": "johna"}
      },
      {
        "id": "53",
        "name": "johnb",
        "parents": [{"id": "17", "name": "A custom group"}],
        "user": {"first_name": "John", "grade": null, "id": "12", "last_name": "Baker", "login": "johnb"}
      },
      {
        "id": "55",
        "name": "johnc",
        "parents": [{"id": "16", "name": "First Team"}],
        "user": {"first_name": "John", "grade": 3, "id": "13", "last_name": null, "login": "johnc"}
      },
      {
        "id": "57",
        "name": "johnd",
        "parents": [{"id": "18", "name": "Our Club"}],
        "user": {"first_name": "John", "grade": 3, "id": "14", "last_name": "Doe", "login": "johnd"}
      }
    ]
    """
    When I send a GET request to "/groups/1/user-descendants?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "name": "johna",
        "parents": [{"id": "11", "name": "Our Class"}],
        "user": {"first_name": null, "grade": 1, "id": "11", "last_name": "Adams", "login": "johna"}
      }
    ]
    """
    When I send a GET request to "/groups/1/user-descendants?from.name=johna&from.id=51"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "53",
        "name": "johnb",
        "parents": [{"id": "17", "name": "A custom group"}],
        "user": {"first_name": "John", "grade": null, "id": "12", "last_name": "Baker", "login": "johnb"}
      },
      {
        "id": "55",
        "name": "johnc",
        "parents": [{"id": "16", "name": "First Team"}],
        "user": {"first_name": "John", "grade": 3, "id": "13", "last_name": null, "login": "johnc"}
      },
      {
        "id": "57",
        "name": "johnd",
        "parents": [{"id": "18", "name": "Our Club"}],
        "user": {"first_name": "John", "grade": 3, "id": "14", "last_name": "Doe", "login": "johnd"}
      }
    ]
    """

  Scenario: Non-descendant parents should not appear (one group with 1 grand child, having also a parent which is not descendant)
    Given the database table 'users' has also the following rows:
      | id | login | group_self_id | group_owned_id | first_name | last_name | grade |
      | 11 | johna | 51            | 52             | null       | Adams     | 1     |
    And the database table 'groups' has also the following rows:
      | id | type      | name        | grade |
      | 51 | UserSelf  | johna       | -2    |
      | 52 | UserAdmin | johna-admin | -2    |
    And the database table 'groups_groups' has also the following rows:
      | group_parent_id | group_child_id | type               |
      | 11              | 51             | invitationAccepted |
      | 13              | 51             | invitationAccepted |
    And the database table 'groups_ancestors' has also the following rows:
      | group_ancestor_id | group_child_id | is_self |
      | 1                 | 51             | 0       |
      | 3                 | 51             | 0       |
      | 11                | 51             | 0       |
      | 13                | 51             | 0       |
      | 22                | 51             | 0       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
    And I am the user with id "1"
    When I send a GET request to "/groups/1/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "name": "johna",
        "parents": [{"id": "11", "name": "Our Class"}],
        "user": {"first_name": null, "grade": 1, "id": "11", "last_name": "Adams", "login": "johna"}
      }
    ]
    """

  Scenario: Only actual memberships count
    Given the database table 'users' has also the following rows:
      | id | login | group_self_id | group_owned_id | first_name | last_name | grade |
      | 11 | johna | 51            | 52             | John       | Adams     | 1     |
      | 12 | johnb | 53            | 54             | John       | Baker     | 2     |
      | 13 | johnc | 55            | 56             | John       | null      | 3     |
      | 14 | johnd | 57            | 58             | null       | Davis     | -1    |
      | 15 | johne | 59            | 60             | John       | Edwards   | null  |
      | 16 | janea | 61            | 62             | Jane       | Adams     | 3     |
      | 17 | janeb | 63            | 64             | Jane       | Baker     | null  |
    And the database table 'groups' has also the following rows:
      | id | type      | name        | grade |
      | 51 | UserSelf  | johna       | -2    |
      | 52 | UserAdmin | johna-admin | -2    |
      | 53 | UserSelf  | johnb       | -2    |
      | 54 | UserAdmin | johnb-admin | -2    |
      | 55 | UserSelf  | johnc       | -2    |
      | 56 | UserAdmin | johnc-admin | -2    |
      | 57 | UserSelf  | johnd       | -2    |
      | 58 | UserAdmin | johnd-admin | -2    |
      | 59 | UserSelf  | johne       | -2    |
      | 60 | UserAdmin | johne-admin | -2    |
      | 61 | UserSelf  | janea       | -2    |
      | 62 | UserAdmin | janea-admin | -2    |
    And the database table 'groups_groups' has also the following rows:
      | group_parent_id | group_child_id | type              |
      | 11              | 51             | invitationSent    |
      | 11              | 53             | requestSent       |
      | 11              | 55             | invitationRefused |
      | 11              | 57             | requestRefused    |
      | 11              | 59             | removed           |
      | 11              | 61             | left              |
    And the database table 'groups_ancestors' has also the following rows:
      | group_ancestor_id | group_child_id | is_self |
      | 22                | 51             | 0       |
      | 22                | 53             | 0       |
      | 22                | 55             | 0       |
      | 22                | 57             | 0       |
      | 22                | 59             | 0       |
      | 22                | 61             | 0       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
      | 53                | 53             | 1       |
      | 54                | 54             | 1       |
      | 55                | 55             | 1       |
      | 56                | 56             | 1       |
      | 57                | 57             | 1       |
      | 58                | 58             | 1       |
      | 59                | 59             | 1       |
      | 60                | 60             | 1       |
      | 61                | 61             | 1       |
      | 62                | 62             | 1       |
    And I am the user with id "1"
    When I send a GET request to "/groups/1/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: No duplication (one group with 1 grand children connected through 2 different parents)
    Given the database table 'users' has also the following rows:
      | id | login | group_self_id | group_owned_id | first_name | last_name | grade |
      | 11 | johna | 51            | 52             | null       | Adams     | 1     |
    And the database table 'groups' has also the following rows:
      | id | type      | name        | grade |
      | 51 | UserSelf  | johna       | -2    |
      | 52 | UserAdmin | johna-admin | -2    |
    And the database table 'groups_groups' has also the following rows:
      | group_parent_id | group_child_id | type               |
      | 11              | 51             | invitationAccepted |
      | 14              | 51             | requestAccepted    |
    And the database table 'groups_ancestors' has also the following rows:
      | group_ancestor_id | group_child_id | is_self |
      | 1                 | 51             | 0       |
      | 11                | 51             | 0       |
      | 14                | 51             | 0       |
      | 22                | 51             | 0       |
      | 51                | 51             | 1       |
      | 52                | 52             | 1       |
    And I am the user with id "1"
    When I send a GET request to "/groups/1/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "id": "51",
        "name": "johna",
        "parents": [{"id": "11", "name": "Our Class"}, {"id": "14", "name": "Super Team"}],
        "user": {"first_name": null, "grade": 1, "id": "11", "last_name": "Adams", "login": "johna"}
      }
    ]
    """

  Scenario: No users
    Given I am the user with id "1"
    When I send a GET request to "/groups/18/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
