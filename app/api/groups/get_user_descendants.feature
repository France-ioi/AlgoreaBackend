Feature: List user descendants of the group (groupUserDescendantView)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 1  | owner  | 21          | 22           | Jean-Michel | Blanquer  | 10     |
    And the database has the following table 'groups':
      | ID | sType     | sName          | iGrade |
      | 1  | Root      | Root 1         | -2     |
      | 3  | Root      | Root 2         | -2     |
      | 11 | Class     | Our Class      | -2     |
      | 12 | Class     | Other Class    | -2     |
      | 13 | Class     | Special Class  | -2     |
      | 14 | Team      | Super Team     | -2     |
      | 15 | Team      | Our Team       | -1     |
      | 16 | Team      | First Team     | 0      |
      | 17 | Other     | A custom group | -2     |
      | 18 | Club      | Our Club       | -2     |
      | 20 | Friends   | My Friends     | -2     |
      | 21 | UserSelf  | owner          | -2     |
      | 22 | UserAdmin | owner-admin    | -2     |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType  |
      | 1             | 11           | direct |
      | 3             | 13           | direct |
      | 3             | 15           | direct |
      | 11            | 14           | direct |
      | 11            | 16           | direct |
      | 11            | 17           | direct |
      | 11            | 18           | direct |
      | 13            | 14           | direct |
      | 13            | 15           | direct |
      | 22            | 1            | direct |

  Scenario: One group with 3 grand children (different parents; one connected as "direct", one as "invitationAccepted", one as "requestAccepted")
    Given the database table 'users' has also the following rows:
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 11 | johna  | 51          | 52           | null        | Adams     | 1      |
      | 12 | johnb  | 53          | 54           | John        | Baker     | null   |
      | 13 | johnc  | 55          | 56           | John        | null      | 3      |
    And the database table 'groups' has also the following rows:
      | ID | sType     | sName          | iGrade |
      | 51 | UserSelf  | johna          | -2     |
      | 52 | UserAdmin | johna-admin    | -2     |
      | 53 | UserSelf  | johnb          | -2     |
      | 54 | UserAdmin | johnb-admin    | -2     |
      | 55 | UserSelf  | johnc          | -2     |
      | 56 | UserAdmin | johnc-admin    | -2     |
    And the database table 'groups_groups' has also the following rows:
      | idGroupParent | idGroupChild | sType              |
      | 11            | 51           | invitationAccepted |
      | 17            | 53           | requestAccepted    |
      | 16            | 55           | direct             |
    And I am the user with ID "1"
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
      }
    ]
    """

  Scenario: Non-descendant parents should not appear (one group with 1 grand child, having also a parent which is not descendant)
    Given the database table 'users' has also the following rows:
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 11 | johna  | 51          | 52           | null        | Adams     | 1      |
    And the database table 'groups' has also the following rows:
      | ID | sType     | sName          | iGrade |
      | 51 | UserSelf  | johna          | -2     |
      | 52 | UserAdmin | johna-admin    | -2     |
    And the database table 'groups_groups' has also the following rows:
      | idGroupParent | idGroupChild | sType              |
      | 11            | 51           | invitationAccepted |
      | 13            | 51           | invitationAccepted |
    And I am the user with ID "1"
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
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 11 | johna  | 51          | 52           | John        | Adams     | 1      |
      | 12 | johnb  | 53          | 54           | John        | Baker     | 2      |
      | 13 | johnc  | 55          | 56           | John        | null      | 3      |
      | 14 | johnd  | 57          | 58           | null        | Davis     | -1     |
      | 15 | johne  | 59          | 60           | John        | Edwards   | null   |
      | 16 | janea  | 61          | 62           | Jane        | Adams     | 3      |
      | 17 | janeb  | 63          | 64           | Jane        | Baker     | null   |
    And the database table 'groups' has also the following rows:
      | ID | sType     | sName          | iGrade |
      | 51 | UserSelf  | johna          | -2     |
      | 52 | UserAdmin | johna-admin    | -2     |
      | 53 | UserSelf  | johnb          | -2     |
      | 54 | UserAdmin | johnb-admin    | -2     |
      | 55 | UserSelf  | johnc          | -2     |
      | 56 | UserAdmin | johnc-admin    | -2     |
      | 57 | UserSelf  | johnd          | -2     |
      | 58 | UserAdmin | johnd-admin    | -2     |
      | 59 | UserSelf  | johne          | -2     |
      | 60 | UserAdmin | johne-admin    | -2     |
      | 61 | UserSelf  | janea          | -2     |
      | 62 | UserAdmin | janea-admin    | -2     |
    And the database table 'groups_groups' has also the following rows:
      | idGroupParent | idGroupChild | sType              |
      | 11            | 51           | invitationSent     |
      | 11            | 53           | requestSent        |
      | 11            | 55           | invitationRefused  |
      | 11            | 57           | requestRefused     |
      | 11            | 59           | removed            |
      | 11            | 61           | left               |
    And I am the user with ID "1"
    When I send a GET request to "/groups/1/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: No duplication (one group with 1 grand children connected through 2 different parents)
    Given the database table 'users' has also the following rows:
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 11 | johna  | 51          | 52           | null        | Adams     | 1      |
    And the database table 'groups' has also the following rows:
      | ID | sType     | sName          | iGrade |
      | 51 | UserSelf  | johna          | -2     |
      | 52 | UserAdmin | johna-admin    | -2     |
    And the database table 'groups_groups' has also the following rows:
      | idGroupParent | idGroupChild | sType              |
      | 11            | 51           | invitationAccepted |
      | 14            | 51           | requestAccepted    |
    And I am the user with ID "1"
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
    Given I am the user with ID "1"
    When I send a GET request to "/groups/18/user-descendants"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """
