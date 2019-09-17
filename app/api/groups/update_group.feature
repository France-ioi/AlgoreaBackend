Feature: Update a group (groupEdit)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | sDefaultLanguage |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | fr               |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf |
      | 75 | 22              | 13           | 0       |
      | 76 | 13              | 11           | 0       |
      | 77 | 22              | 14           | 0       |
    And the database has the following table 'groups':
      | ID | sName      | iGrade | sDescription    | sDateCreated        | sType     | sRedirectPath                          | bOpened | bFreeAccess | sCode      | sCodeTimer | sCodeEnd            | bOpenContest |
      | 11 | Group A    | -3     | Group A is here | 2019-02-06 09:26:40 | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00   | 2017-10-13 05:39:48 | true         |
      | 13 | Group B    | -2     | Group B is here | 2019-03-06 09:26:40 | Class     | 182529188317717610/1672978871462145461 | true    | true        | ybabbxnlyo | 01:00:00   | 2017-10-14 05:39:48 | true         |
      | 14 | Group C    | -4     | Admin Group     | 2019-04-06 09:26:40 | UserAdmin | null                                   | true    | false       | null       | null       | null                | false        |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              |
      | 75 | 13            | 21           | invitationSent     |
      | 76 | 13            | 22           | requestSent        |
      | 77 | 13            | 23           | invitationAccepted |
      | 78 | 13            | 24           | requestSent        |
      | 79 | 14            | 22           | requestSent        |

  Scenario: User is an owner of the group, all fields are not nulls, updates groups_groups
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "free_access": false,
      "name": "Team B",
      "grade": 10,
      "description": "Team B is here",
      "opened": false,
      "code_timer": "99:59:59",
      "code_end": "2019-12-31T23:59:59Z",
      "open_contest": false,
      "redirect_path": "1234/5678"
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with ID "13"
    And the table "groups" at ID "13" should be:
      | ID | sName   | iGrade | sDescription   | sDateCreated        | sType     | sRedirectPath | bOpened | bFreeAccess | sCode      | sCodeTimer | sCodeEnd            | bOpenContest |
      | 13 | Team B  | 10     | Team B is here | 2019-03-06 09:26:40 | Class     | 1234/5678     | false   | false       | ybabbxnlyo | 99:59:59   | 2019-12-31 23:59:59 | false        |
    And the table "groups_groups" should be:
      | ID | idGroupParent | idGroupChild | sType              |
      | 75 | 13            | 21           | invitationSent     |
      | 76 | 13            | 22           | requestRefused     |
      | 77 | 13            | 23           | invitationAccepted |
      | 78 | 13            | 24           | requestRefused     |
      | 79 | 14            | 22           | requestSent        |

  Scenario: User is an owner of the group, nullable fields are nulls
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "free_access": false,
      "name": "Club B",
      "description": null,
      "opened": false,
      "open_contest": false,
      "redirect_path": null,
      "grade": 0,
      "code_end": null,
      "code_timer": null
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with ID "13"
    And the table "groups" at ID "13" should be:
      | ID | sName   | iGrade | sDescription | sDateCreated        | sType | sRedirectPath | bOpened | bFreeAccess | sCode      | sCodeTimer | sCodeEnd | bOpenContest |
      | 13 | Club B  | 0      | null         | 2019-03-06 09:26:40 | Class | null          | false   | false       | ybabbxnlyo | null       | null     | false        |

  Scenario: User is an owner of the group, does not update groups_groups (free_access is still true)
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "free_access": true,
      "name": "Club B",
      "description": null,
      "opened": false,
      "open_contest": false,
      "redirect_path": null,
      "grade": 0,
      "code_end": null,
      "code_timer": null
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with ID "13"
    And the table "groups" at ID "13" should be:
      | ID | sName   | iGrade | sDescription    | sDateCreated        | sType | sRedirectPath | bOpened | bFreeAccess | sCode      | sCodeTimer | sCodeEnd | bOpenContest |
      | 13 | Club B  | 0      | null            | 2019-03-06 09:26:40 | Class | null          | false   | true        | ybabbxnlyo | null       | null     | false        |
    And the table "groups_groups" should stay unchanged

  Scenario: User is an owner of the group, does not update groups_groups (free_access is not changed)
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "name": "Club B",
      "description": null,
      "opened": false,
      "open_contest": false,
      "redirect_path": null,
      "grade": 0,
      "code_end": null,
      "code_timer": null
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with ID "13"
    And the table "groups" at ID "13" should be:
      | ID | sName   | iGrade | sDescription    | sDateCreated        | sType | sRedirectPath | bOpened | bFreeAccess | sCode      | sCodeTimer | sCodeEnd | bOpenContest |
      | 13 | Club B  | 0      | null            | 2019-03-06 09:26:40 | Class | null          | false   | true        | ybabbxnlyo | null       | null     | false        |
    And the table "groups_groups" should stay unchanged

  Scenario: User is an owner of the group, does not update groups_groups (free_access changes from false to true)
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/14" with the following body:
    """
    {
      "free_access": true
    }
    """
    Then the response should be "updated"
    And the table "groups" should stay unchanged but the row with ID "14"
    And the table "groups" at ID "14" should be:
      | ID | sName   | iGrade | sDescription    | sDateCreated        | sType     | sRedirectPath | bOpened | bFreeAccess | sCode  | sCodeTimer | sCodeEnd | bOpenContest |
      | 14 | Group C | -4     | Admin Group     | 2019-04-06 09:26:40 | UserAdmin | null          | true    | true        | null   | null       | null     | false        |
    And the table "groups_groups" should stay unchanged
