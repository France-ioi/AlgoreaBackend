Feature: List owner groups for an item
  Background:
    Given the database has the following table "groups":
      | id | name        | type  |
      | 9  | Class       | Class |
      | 10 | Team Alpha  | Team  |
      | 11 | Team Beta   | Team  |
      | 25 | some class  | Class |
    And the database has the following users:
      | group_id | login | first_name  | last_name | default_language |
      | 21       | owner | Jean-Michel | Blanquer  | fr               |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 9               | 10             |
      | 9               | 21             |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id  | default_language_tag | requires_explicit_entry | type    |
      | 102 | fr                   | true                    | Chapter |
    And the database has the following table "items_strings":
      | item_id | language_tag | title      |
      | 102     | fr           | Chapitre B |
    And the database table "permissions_granted" also has the following rows:
      | group_id | item_id | source_group_id | origin           | can_view | can_grant_view | can_watch | can_edit | is_owner | can_request_help_to | can_make_session_official | can_enter_from      | can_enter_until     |
      | 9        | 102     | 25              | group_membership | none     | none           | none      | none     | true     | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 25              | group_membership | none     | none           | none      | none     | true     | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 9               | group_membership | none     | none           | none      | none     | true     | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 11       | 102     | 25              | group_membership | none     | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 25              | group_membership | info     | none           | none      | none     | true     | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 9               | group_membership | none     | none           | none      | none     | true     | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table "permissions_generated" also has the following rows:
      | group_id | item_id | can_edit_generated |
      | 21       | 102     | all                |
    And the generated permissions are computed

  Scenario: List owner groups
    Given I am the user with id "21"
    When I send a GET request to "/items/102/owners"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "9", "name": "Class", "type": "Class"},
      {"id": "25", "name": "some class", "type": "Class"},
      {"id": "10", "name": "Team Alpha", "type": "Team"}
    ]
    """

  Scenario: Empty list when no owner groups
    Given the database has the following table "items":
      | id  | default_language_tag | requires_explicit_entry | type    |
      | 103 | fr                   | false                   | Chapter |
    And the database table "permissions_generated" also has the following rows:
      | group_id | item_id | can_edit_generated |
      | 21       | 103     | all                |
    And the generated permissions are computed
    And I am the user with id "21"
    When I send a GET request to "/items/103/owners"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: Sort by name descending
    Given I am the user with id "21"
    When I send a GET request to "/items/102/owners?sort=-name"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "10", "name": "Team Alpha", "type": "Team"},
      {"id": "25", "name": "some class", "type": "Class"},
      {"id": "9", "name": "Class", "type": "Class"}
    ]
    """

  Scenario: Group with is_owner_generated is listed
    Given the database has the following table "groups":
      | id | name             | type  |
      | 12 | Generated owner  | Class |
    And the database table "permissions_generated" also has the following rows:
      | group_id | item_id | is_owner_generated |
      | 12       | 102     | true               |
    And the generated permissions are computed
    And I am the user with id "21"
    When I send a GET request to "/items/102/owners"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "9", "name": "Class", "type": "Class"},
      {"id": "12", "name": "Generated owner", "type": "Class"},
      {"id": "25", "name": "some class", "type": "Class"},
      {"id": "10", "name": "Team Alpha", "type": "Team"}
    ]
    """
