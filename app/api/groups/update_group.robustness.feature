Feature: Update a group (groupEdit) - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName |
      | 1  | owner  | 0        | 21          | 22           | Jean-Michel | Blanquer  |
      | 2  | user   | 0        | 31          | 32           | John        | Doe       |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 75 | 22              | 13           | 0       | 0        |
      | 76 | 13              | 11           | 0       | 0        |
      | 77 | 32              | 15           | 0       | 0        |
    And the database has the following table 'groups':
      | ID | sName      | iGrade | sDescription    | sDateCreated         | sType     | sRedirectPath                          | bOpened | bFreeAccess | sPassword  | sPasswordTimer | sPasswordEnd        | bOpenContest |
      | 11 | Group A    | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00       | 2017-10-13 05:39:48 | true         |
      | 13 | Group B    | -2     | Group B is here | 2019-03-06T09:26:40Z | Class     | 182529188317717610/1672978871462145461 | true    | true        | ybabbxnlyo | 01:00:00       | 2017-10-14 05:39:48 | true         |
      | 14 | Group C    | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | true        | null       | null           | null                | false        |

  Scenario: Should fail if the user is not an owner of the group
    Given I am the user with ID "2"
    When I send a POST request to "/groups/13" with the following body:
    """
    {}
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should fail if the user is an owner of the group, but the group itself doesn't exist
    Given I am the user with ID "2"
    When I send a POST request to "/groups/15" with the following body:
    """
    {"type":"Club"}
    """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: User is an owner of the group, but required fields are not filled in correctly
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13" with the following body:
    """
    {
      "type": "Root",
      "free_access": 15,
      "name": 123,
      "grade": "grade",
      "description": 14.5,
      "opened": "true",
      "password_timer": 1234,
      "password_end": "the end",
      "open_contest": 12,
      "redirect_path": "some path"
    }
    """
    Then the response code should be 400
    And the response body should be, in JSON:
		"""
		{
    	"error_text": "Invalid input data",
 	    "errors": {
				"description": ["expected type 'string', got unconvertible type 'float64'"],
				"free_access": ["expected type 'bool', got unconvertible type 'float64'"],
				"grade": ["expected type 'int64', got unconvertible type 'string'"],
				"name": ["expected type 'string', got unconvertible type 'float64'"],
				"open_contest": ["expected type 'bool', got unconvertible type 'float64'"],
				"opened": ["expected type 'bool', got unconvertible type 'string'"],
				"password_end": ["decoding error: parsing time \"the end\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"the end\" as \"2006\""],
				"password_timer": ["expected type 'string', got unconvertible type 'float64'"],
				"redirect_path": ["some path does not validate as matches(^(\\d+(/\\d+)*)*$)"],
				"type": ["Root does not validate as in(Class|Team|Club|Friends|Other)"]
 	    },
 	    "message": "Bad Request",
      "success": false
		}
    """
    And the table "groups" should be:
      | ID | sName      | iGrade | sDescription    | sDateCreated         | sType     | sRedirectPath                          | bOpened | bFreeAccess | sPassword  | sPasswordTimer | sPasswordEnd         | bOpenContest |
      | 11 | Group A    | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00       | 2017-10-13T05:39:48Z | true         |
      | 13 | Group B    | -2     | Group B is here | 2019-03-06T09:26:40Z | Class     | 182529188317717610/1672978871462145461 | true    | true        | ybabbxnlyo | 01:00:00       | 2017-10-14T05:39:48Z | true         |
      | 14 | Group C    | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | true        | null       | null           | null                 | false        |

  Scenario: User is an owner of the group, but no fields provided
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13" with the following body:
    """
    {
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
      | ID | sName      | iGrade | sDescription    | sDateCreated         | sType     | sRedirectPath                          | bOpened | bFreeAccess | sPassword  | sPasswordTimer | sPasswordEnd         | bOpenContest |
      | 11 | Group A    | -3     | Group A is here | 2019-02-06T09:26:40Z | Class     | 182529188317717510/1672978871462145361 | true    | true        | ybqybxnlyo | 01:00:00       | 2017-10-13T05:39:48Z | true         |
      | 13 | Group B    | -2     | Group B is here | 2019-03-06T09:26:40Z | Class     | 182529188317717610/1672978871462145461 | true    | true        | ybabbxnlyo | 01:00:00       | 2017-10-14T05:39:48Z | true         |
      | 14 | Group C    | -4     | Admin Group     | 2019-04-06T09:26:40Z | UserAdmin | null                                   | true    | true        | null       | null           | null                 | false        |
