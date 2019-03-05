Feature: Update a group (groupEdit)
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName | sDefaultLanguage |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  | fr               |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 75 | 22              | 13           | 0       | 0        |
      | 76 | 13              | 11           | 0       | 0        |
      | 77 | 22              | 14           | 0       | 0        |
    And the database has the following table 'groups':
      | ID | sName      | iGrade | sDescription    | sDateCreated         | sType     | sRedirectPath                          | bOpened | bFreeAccess | sPassword  | sPasswordTimer | sPasswordEnd         | bOpenContest |
      | 11 | Group A    | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00       | 2017-10-13T05:39:48Z | true         |
      | 13 | Group B    | -2     | Group B is here | 2019-03-06T09:26:40Z | Class     | 182529188317717610/1672978871462145461 | true    | true        | ybabbxnlyo | 01:00:00       | 2017-10-14T05:39:48Z | true         |
      | 14 | Group C    | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | false       | null       | null           | null                 | false        |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | iVersion |
      | 75 | 13            | 21           | invitationSent     | 0        |
      | 76 | 13            | 22           | requestSent        | 0        |
      | 77 | 13            | 23           | invitationAccepted | 0        |
      | 78 | 13            | 24           | requestSent        | 0        |
      | 79 | 14            | 22           | requestSent        | 0        |

  Scenario: User is an owner of the group, all fields are not nulls, updates groups_groups
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "type": "Team",
      "free_access": false,
      "name": "Team B",
      "grade": 10,
      "description": "Team B is here",
      "opened": false,
      "password_timer": "99:59:59",
      "password_end": "9999-12-31T23:59:59Z",
      "open_contest": false,
      "redirect_path": "1234/5678"
    }
    """
    Then the response code should be 200
    And the response body should be, in JSON:
		"""
		{
      "message": "success",
      "success": true
		}
    """
    And the table "groups" should be:
      | ID | sName   | iGrade | sDescription   | sDateCreated          | sType     | sRedirectPath                          | bOpened | bFreeAccess | sPassword  | sPasswordTimer | sPasswordEnd         | bOpenContest |
      | 11 | Group A | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00       | 2017-10-13T05:39:48Z | true         |
      | 13 | Team B  | 10     | Team B is here  | 2019-03-06T09:26:40Z | Team      | 1234/5678                              | false   | false       | ybabbxnlyo | 99:59:59       | 9999-12-31T23:59:59Z | false        |
      | 14 | Group C | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | false       | null       | null           | null                 | false        |
    And the table "groups_groups" should be:
      | ID | idGroupParent | idGroupChild | sType              | iVersion |
      | 75 | 13            | 21           | invitationSent     | 0        |
      | 76 | 13            | 22           | requestRefused     | 0        |
      | 77 | 13            | 23           | invitationAccepted | 0        |
      | 78 | 13            | 24           | requestRefused     | 0        |
      | 79 | 14            | 22           | requestSent        | 0        |

  Scenario: User is an owner of the group, nullable fields are nulls
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "type": "Club",
      "free_access": false,
      "name": "Club B",
      "description": null,
      "opened": false,
      "open_contest": false,
      "redirect_path": null,
      "grade": 0,
      "password_end": null,
      "password_timer": null
    }
    """
    Then the response code should be 200
    And the response body should be, in JSON:
		"""
		{
      "message": "success",
      "success": true
		}
    """
    And the table "groups" should be:
      | ID | sName   | iGrade | sDescription    | sDateCreated         | sType     | sRedirectPath                          | bOpened | bFreeAccess | sPassword  | sPasswordTimer | sPasswordEnd         | bOpenContest |
      | 11 | Group A | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00       | 2017-10-13T05:39:48Z | true         |
      | 13 | Club B  | 0      | null            | 2019-03-06T09:26:40Z | Club      | null                                   | false   | false       | ybabbxnlyo | null           | null                 | false        |
      | 14 | Group C | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | false       | null       | null           | null                 | false        |

  Scenario: User is an owner of the group, does not update groups_groups (free_access is still true)
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "type": "Club",
      "free_access": true,
      "name": "Club B",
      "description": null,
      "opened": false,
      "open_contest": false,
      "redirect_path": null,
      "grade": 0,
      "password_end": null,
      "password_timer": null
    }
    """
    Then the response code should be 200
    And the response body should be, in JSON:
		"""
		{
      "message": "success",
      "success": true
		}
    """
    And the table "groups" should be:
      | ID | sName   | iGrade | sDescription    | sDateCreated         | sType     | sRedirectPath                          | bOpened | bFreeAccess | sPassword  | sPasswordTimer | sPasswordEnd         | bOpenContest |
      | 11 | Group A | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00       | 2017-10-13T05:39:48Z | true         |
      | 13 | Club B  | 0      | null            | 2019-03-06T09:26:40Z | Club      | null                                   | false   | true        | ybabbxnlyo | null           | null                 | false        |
      | 14 | Group C | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | false       | null       | null           | null                 | false        |
    And the table "groups_groups" should be:
      | ID | idGroupParent | idGroupChild | sType              | iVersion |
      | 75 | 13            | 21           | invitationSent     | 0        |
      | 76 | 13            | 22           | requestSent        | 0        |
      | 77 | 13            | 23           | invitationAccepted | 0        |
      | 78 | 13            | 24           | requestSent        | 0        |
      | 79 | 14            | 22           | requestSent        | 0        |

  Scenario: User is an owner of the group, does not update groups_groups (free_access is not changed)
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/13" with the following body:
    """
    {
      "type": "Club",
      "name": "Club B",
      "description": null,
      "opened": false,
      "open_contest": false,
      "redirect_path": null,
      "grade": 0,
      "password_end": null,
      "password_timer": null
    }
    """
    Then the response code should be 200
    And the response body should be, in JSON:
		"""
		{
      "message": "success",
      "success": true
		}
    """
    And the table "groups" should be:
      | ID | sName   | iGrade | sDescription    | sDateCreated         | sType     | sRedirectPath                          | bOpened | bFreeAccess | sPassword  | sPasswordTimer | sPasswordEnd         | bOpenContest |
      | 11 | Group A | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00       | 2017-10-13T05:39:48Z | true         |
      | 13 | Club B  | 0      | null            | 2019-03-06T09:26:40Z | Club      | null                                   | false   | true        | ybabbxnlyo | null           | null                 | false        |
      | 14 | Group C | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | false       | null       | null           | null                 | false        |
    And the table "groups_groups" should be:
      | ID | idGroupParent | idGroupChild | sType              | iVersion |
      | 75 | 13            | 21           | invitationSent     | 0        |
      | 76 | 13            | 22           | requestSent        | 0        |
      | 77 | 13            | 23           | invitationAccepted | 0        |
      | 78 | 13            | 24           | requestSent        | 0        |
      | 79 | 14            | 22           | requestSent        | 0        |

  Scenario: User is an owner of the group, does not update groups_groups (free_access changes from false to true)
    Given I am the user with ID "1"
    When I send a PUT request to "/groups/14" with the following body:
    """
    {
      "free_access": true
    }
    """
    Then the response code should be 200
    And the response body should be, in JSON:
		"""
		{
      "message": "success",
      "success": true
		}
    """
    And the table "groups" should be:
      | ID | sName   | iGrade | sDescription    | sDateCreated         | sType     | sRedirectPath                          | bOpened | bFreeAccess | sPassword  | sPasswordTimer | sPasswordEnd         | bOpenContest |
      | 11 | Group A | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00       | 2017-10-13T05:39:48Z | true         |
      | 13 | Group B | -2     | Group B is here | 2019-03-06T09:26:40Z | Class     | 182529188317717610/1672978871462145461 | true    | true        | ybabbxnlyo | 01:00:00       | 2017-10-14T05:39:48Z | true         |
      | 14 | Group C | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | true        | null       | null           | null                 | false        |
    And the table "groups_groups" should be:
      | ID | idGroupParent | idGroupChild | sType              | iVersion |
      | 75 | 13            | 21           | invitationSent     | 0        |
      | 76 | 13            | 22           | requestSent        | 0        |
      | 77 | 13            | 23           | invitationAccepted | 0        |
      | 78 | 13            | 24           | requestSent        | 0        |
      | 79 | 14            | 22           | requestSent        | 0        |

