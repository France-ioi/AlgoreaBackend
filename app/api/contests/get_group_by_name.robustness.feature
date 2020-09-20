Feature: Get group by name (contestGetGroupByName) - robustness
  Background:
    Given the database has the following table 'groups':
      | id | name    | type |
      | 12 | Group A | Team |
      | 13 | Group B | Team |
      | 14 | Group A | Team |
      | 15 | Group A | Team |
      | 21 | owner   | User |
      | 31 | john    | User |
    And the database has the following table 'users':
      | login | group_id |
      | owner | 21       |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_grant_group_access | can_watch_members |
      | 13       | 21         | true                   | true              |
      | 14       | 21         | true                   | false             |
      | 15       | 21         | false                  | true              |
      | 31       | 21         | true                   | true              |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id | duration | default_language_tag | entry_participant_type |
      | 10 | 00:00:02 | fr                   | Team                   |
      | 11 | 00:00:02 | fr                   | Team                   |
      | 12 | 00:00:02 | fr                   | Team                   |
      | 50 | 00:00:00 | fr                   | User                   |
      | 60 | null     | fr                   | User                   |
      | 70 | 00:00:03 | fr                   | Team                   |
      | 80 | 00:00:03 | fr                   | Team                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_watch_generated |
      | 13       | 10      | info                     | none                     | none                |
      | 13       | 11      | info                     | none                     | none                |
      | 13       | 12      | info                     | none                     | none                |
      | 13       | 50      | content                  | none                     | none                |
      | 13       | 60      | info                     | none                     | none                |
      | 13       | 70      | content_with_descendants | none                     | none                |
      | 13       | 80      | none                     | enter                    | result              |
      | 15       | 70      | info                     | none                     | none                |
      | 21       | 10      | content                  | none                     | result              |
      | 21       | 11      | content                  | enter                    | none                |
      | 21       | 12      | none                     | enter                    | result              |
      | 21       | 50      | none                     | none                     | none                |
      | 21       | 60      | content_with_descendants | enter                    | result              |
      | 21       | 70      | content_with_descendants | enter                    | result              |
      | 21       | 80      | none                     | enter                    | result              |
      | 31       | 60      | info                     | none                     | none                |

  Scenario: Wrong item_id
    Given I am the user with id "21"
    When I send a GET request to "/contests/abc/groups/by-name?name=Group%20B"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: name is missing
    Given I am the user with id "21"
    When I send a GET request to "/contests/50/groups/by-name"
    Then the response code should be 400
    And the response error message should contain "Missing name"

  Scenario: No such item
    Given I am the user with id "21"
    When I send a GET request to "/contests/90/groups/by-name?name=Group%20B"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Cannot grant view permission on the item
    Given I am the user with id "21"
    When I send a GET request to "/contests/10/groups/by-name?name=Group%20B"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Cannot watch results on the item
    Given I am the user with id "21"
    When I send a GET request to "/contests/11/groups/by-name?name=Group%20B"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Cannot view content of the item
    Given I am the user with id "21"
    When I send a GET request to "/contests/12/groups/by-name?name=Group%20B"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The item is not a timed contest
    Given I am the user with id "21"
    When I send a GET request to "/contests/60/groups/by-name?name=john"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The user is not a manager of the group
    Given I am the user with id "21"
    When I send a GET request to "/contests/70/groups/by-name?name=Group%20A"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: The group cannot view/enter the item
    Given I am the user with id "21"
    When I send a GET request to "/contests/80/groups/by-name?name=Group%20B"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No such group (space)
    Given I am the user with id "21"
    When I send a GET request to "/contests/70/groups/by-name?name=Group%20B%20"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No such group (wildcards should not work)
    Given I am the user with id "21"
    When I send a GET request to "/contests/70/groups/by-name?name=%25"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
