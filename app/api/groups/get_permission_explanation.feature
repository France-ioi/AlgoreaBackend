Feature: Explain permissions
  Background:
    Given the database has the following users:
      | group_id | login | first_name  | last_name | default_language |
      | 21       | owner | Jean-Michel | Blanquer  | fr               |
    And the database has the following table "items":
      | id  | default_language_tag |
      | 100 | en                   |
      | 101 | en                   |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order | content_view_propagation | grant_view_propagation | watch_propagation |
      | 100            | 101           | 1           | none                     | false                  | false             |
    And the items ancestors are computed
    And the database has the following table "items_strings":
      | item_id | language_tag | title      |
      | 100     | en           | Some Item  |
      | 101     | en           | Child Item |

  Scenario: Excludes not affecting permissions
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name          | type  |
      | 25 | some class    | Class |
      | 27 | some club     | Club  |
      | 28 | other         | Other |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 25              | 21             |
      | 27              | 21             |
    And the groups ancestors are computed
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin | can_view                 | can_grant_view      | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 21              | self   | content_with_descendants | solution_with_grant | answer_with_grant | all_with_grant | true     | true                      | 1000-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 21       | 101     | 21              | self   | none                     | none                | none              | none           | false    | true                      | 1000-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 100     | 25              | self   | content_with_descendants | solution_with_grant | answer_with_grant | all_with_grant | true     | true                      | 1000-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 101     | 25              | self   | none                     | none                | none              | none           | false    | true                      | 1000-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 27       | 100     | 27              | self   | content_with_descendants | solution_with_grant | answer_with_grant | all_with_grant | true     | true                      | 1000-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 27       | 101     | 27              | self   | none                     | none                | none              | none           | false    | true                      | 1000-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 28       | 101     | 28              | self   | content_with_descendants | solution_with_grant | answer_with_grant | all_with_grant | true     | true                      | 1000-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: Displays item titles in the user's default language if possible
    Given I am the user with id "21"
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 21              | self   | content  | none           | none      | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    And the database table "items_strings" also has the following row:
      | item_id | language_tag | title             |
      | 100     | fr           | Certains Articles |
    When I send a GET request to "/groups/21/permissions/100/explain"
    Then the response code should be 200
    And the response at $[*].item in JSON should be:
    """
    [{
      "id": "100", "language_tag": "fr", "requires_explicit_entry": false,
      "title": "Certains Articles", "type": "Chapter"
    }]
    """

  Scenario: Displays item titles in the item's default language when the title in the user's default language is null
    Given I am the user with id "21"
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 21              | self   | content  | none           | none      | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    And the database table "items_strings" also has the following row:
      | item_id | language_tag | title |
      | 100     | fr           | null  |
    When I send a GET request to "/groups/21/permissions/100/explain"
    Then the response code should be 200
    And the response at $[*].item in JSON should be:
    """
    [{
      "id": "100", "language_tag": "en", "requires_explicit_entry": false,
      "title": "Some Item", "type": "Chapter"
    }]
    """

  Scenario Outline: user_can_update_permission is true
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name          | type  |
      | 25 | some class    | Class |
      | 27 | some club     | Club  |
      | 28 | other         | Other |
      | 29 | club's parent | Other |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 25              | 21             |
      | 27              | 21             |
      | 29              | 27             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access |
      | 29       | 21         | true                   |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin           | can_view | can_grant_view   | can_watch   | can_edit   | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 25       | 100     | 27              | group_membership | content  | <can_grant_view> | <can_watch> | <can_edit> | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 200
    And the response at $[*].user_can_update_permission in JSON should be:
    """
    [true]
    """
  Examples:
    | can_grant_view           | can_watch         | can_edit       |
    | enter                    | none              | none           |
    | content                  | none              | none           |
    | content_with_descendants | none              | none           |
    | solution                 | none              | none           |
    | solution_with_grant      | none              | none           |
    | none                     | answer_with_grant | none           |
    | none                     | none              | all_with_grant |

  Scenario Outline: user_can_update_permission is false because of insufficient permissions
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name          | type  |
      | 25 | some class    | Class |
      | 27 | some club     | Club  |
      | 28 | other         | Other |
      | 29 | club's parent | Other |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 25              | 21             |
      | 27              | 21             |
      | 29              | 27             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access |
      | 29       | 21         | true                   |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin           | can_view | can_grant_view | can_watch   | can_edit   | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 25       | 100     | 27              | group_membership | content  | none           | <can_watch> | <can_edit> | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 200
    And the response at $[*].user_can_update_permission in JSON should be:
    """
    [false]
    """
    Examples:
      | can_watch | can_edit |
      | none      | none     |
      | result    | none     |
      | answer    | none     |
      | answer    | children |
      | answer    | all      |

  Scenario Outline: user_can_update_permission is false because origin != group_membership
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name          | type  |
      | 25 | some class    | Class |
      | 27 | some club     | Club  |
      | 28 | other         | Other |
      | 29 | club's parent | Other |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 25              | 21             |
      | 27              | 21             |
      | 29              | 27             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access |
      | 29       | 21         | true                   |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin   | can_view | can_grant_view      | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 25       | 100     | 27              | <origin> | content  | solution_with_grant | answer_with_grant | all_with_grant | true     | true                      | 1000-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 200
    And the response at $[*].user_can_update_permission in JSON should be:
    """
    [false]
    """
    Examples:
      | origin         |
      | item_unlocking |
      | self           |
      | other          |

  Scenario: user_can_update_permission is false because the user can't grant group access
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name          | type  |
      | 25 | some class    | Class |
      | 27 | some club     | Club  |
      | 28 | other         | Other |
      | 29 | club's parent | Other |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 25              | 21             |
      | 27              | 21             |
      | 29              | 27             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access |
      | 29       | 21         | false                  |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin           | can_view | can_grant_view      | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 25       | 100     | 27              | group_membership | content  | solution_with_grant | answer_with_grant | all_with_grant | true     | true                      | 1000-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/25/permissions/100/explain"
    Then the response code should be 200
    And the response at $[*].user_can_update_permission in JSON should be:
    """
    [false]
    """
