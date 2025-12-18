Feature: Explain permissions - pagination
  Background:
    Given the database has the following table "groups":
      | id | name               | type  |
      | 25 | some class         | Class |
      | 26 | some team          | Team  |
      | 27 | some club          | Club  |
      | 28 | other              | Other |
      | 29 | team's parent      | Club  |
    And the database has the following users:
      | group_id | login | first_name  | last_name |
      | 21       | owner | Jean-Michel | Blanquer  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 25              | 21             |
      | 26              | 21             |
      | 27              | 21             |
      | 28              | 27             |
      | 29              | 26             |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id  | default_language_tag |
      | 100 | en                   |
      | 101 | en                   |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order | content_view_propagation | grant_view_propagation | watch_propagation |
      | 100            | 101           | 1           | as_info                  | true                   | true              |
    And the items ancestors are computed
    And the database has the following table "items_strings":
      | item_id | language_tag | title      |
      | 100     | en           | Some Item  |
      | 101     | en           | Child Item |

  Scenario: The item is visible to the user
    Given I am the user with id "21"
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 27       | 100     | 25              | item_unlocking | content  | none           | none      | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/27/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].item.id in JSON should be:
    """
    ["100"]
    """

  Scenario: The item is visible to some team of the user
    Given I am the user with id "21"
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 29       | 100     | 29              | item_unlocking | content  | none           | none      | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/26/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].item.id in JSON should be:
    """
    ["100"]
    """

  Scenario: The item is invisible
    Given I am the user with id "21"
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 29       | 100     | 29              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/26/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].item in JSON should be:
    """
    []
    """

  Scenario: The group is public
    Given I am the user with id "21"
    And the database table "groups" also has the following row:
      | id | name       | is_public | type |
      | 30 | some group | true      | User |
    And the database table "group_managers" also has the following row:
      | group_id | manager_id | can_manage  |
      | 30       | 21         | memberships |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 30       | 100     | 29              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/30/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].group.id in JSON should be:
    """
    ["30"]
    """

  Scenario: The group is an ancestor of the current user
    Given I am the user with id "21"
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 28       | 100     | 29              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/28/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].group.id in JSON should be:
    """
    ["28"]
    """

  Scenario: The group is an ancestor of a team the current user is a member of
    Given I am the user with id "21"
    And the database table "groups" also has the following row:
      | id | name               | type |
      | 40 | team's grandparent | Club |
    And the database table "groups_groups" also has the following row:
      | parent_group_id | child_group_id |
      | 40              | 29             |
    And the groups ancestors are computed
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 40       | 100     | 29              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/26/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].group.id in JSON should be:
    """
    ["40"]
    """

  Scenario Outline: The user is a manager of a non-user descendant of the group
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 30 | a group            | Club  |
      | 31 | a child group      | Club  |
      | 32 | a grandchild group | Other |
      | 33 | a descendant group | Club  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage   | can_watch_members   | can_grant_group_access   |
      | 33       | 28         | <can_manage> | <can_watch_members> | <can_grant_group_access> |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 101     | 21              | self           | none     | enter          | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 30       | 100     | 28              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/33/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].group.id in JSON should be:
    """
    ["30"]
    """
    Examples:
      | can_manage            | can_watch_members | can_grant_group_access |
      | memberships           | false             | false                  |
      | memberships_and_group | false             | false                  |
      | none                  | true              | false                  |
      | none                  | false             | true                   |

  Scenario: The user is a manager of a non-user descendant of the group without any managerial permissions (the group should be invisible)
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 30 | a group            | Club  |
      | 31 | a child group      | Club  |
      | 32 | a grandchild group | Other |
      | 33 | a descendant group | Club  |
      | 34 | a descendant user  | User  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 33             |
      | 33              | 34             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage  | can_watch_members | can_grant_group_access |
      | 34       | 21         | memberships | false             | false                  |
      | 33       | 28         | none        | false             | false                  |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 101     | 21              | self           | none     | enter          | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 31       | 100     | 28              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/34/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].group in JSON should be:
    """
    []
    """

  Scenario: The user is a manager of a descendant user of the group (the group should be invisible)
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 30 | a group            | Club  |
      | 31 | a child group      | Club  |
      | 32 | a grandchild group | Other |
      | 33 | a descendant user  | User  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage            | can_watch_members | can_grant_group_access |
      | 33       | 28         | memberships_and_group | true              | true                   |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 101     | 21              | self           | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 30       | 100     | 28              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/33/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].group in JSON should be:
    """
    []
    """

  Scenario Outline: The group is a user implicitly managed by the current user
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 30 | a group            | Club  |
      | 31 | a child group      | Club  |
      | 32 | a grandchild group | Other |
      | 33 | a descendant user  | User  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage   | can_watch_members   | can_grant_group_access   |
      | 30       | 28         | <can_manage> | <can_watch_members> | <can_grant_group_access> |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 101     | 21              | self           | none     | enter          | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 33       | 100     | 28              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/33/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].group.id in JSON should be:
    """
    ["33"]
    """
    Examples:
      | can_manage            | can_watch_members | can_grant_group_access |
      | memberships           | false             | false                  |
      | memberships_and_group | false             | false                  |
      | none                  | true              | false                  |
      | none                  | false             | true                   |

  Scenario: The group is a user implicitly managed by the current user without any managerial permissions (the group should be invisible)
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 30 | a group            | Club  |
      | 31 | a child group      | Club  |
      | 32 | a grandchild group | Other |
      | 33 | a descendant user  | User  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage  | can_watch_members | can_grant_group_access |
      | 30       | 28         | none        | false             | false                  |
      | 33       | 28         | memberships | false             | false                  |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 101     | 21              | self           | none     | enter          | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 33       | 100     | 28              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/33/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].group in JSON should be:
    """
    []
    """

  Scenario: The group is a user explicitly managed by the current user (the group should be invisible)
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 33 | a descendant user  | User  |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage            | can_watch_members   | can_grant_group_access |
      | 33       | 28         | memberships_and_group | true                | true                   |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 101     | 21              | self           | none     | enter          | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 33       | 100     | 28              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/33/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].group in JSON should be:
    """
    []
    """

  Scenario Outline: The group is a user in a team managed by the current user
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name              | type |
      | 30 | a group           | Club |
      | 31 | a child group     | Club |
      | 32 | a grandchild team | Team |
      | 33 | a descendant user | User |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 21             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage   | can_watch_members   | can_grant_group_access   |
      | 30       | 28         | <can_manage> | <can_watch_members> | <can_grant_group_access> |
      | 33       | 21         | memberships  | false               | false                    |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 101     | 21              | self           | none     | enter          | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 33       | 100     | 28              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/33/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].group.id in JSON should be:
    """
    ["33"]
    """
    Examples:
      | can_manage            | can_watch_members | can_grant_group_access |
      | memberships           | false             | false                  |
      | memberships_and_group | false             | false                  |
      | none                  | true              | false                  |
      | none                  | false             | true                   |

  Scenario: The group is a user in a team managed by the current user, but the current user doesn't have any managerial permission (the group should be invisible)
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name              | type |
      | 30 | a group           | Club |
      | 31 | a child group     | Club |
      | 32 | a grandchild team | Team |
      | 33 | a descendant user | User |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 21             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage   | can_watch_members   | can_grant_group_access |
      | 30       | 28         | none         | false               | false                  |
      | 33       | 21         | memberships  | false               | false                  |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 101     | 21              | self           | none     | enter          | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 33       | 100     | 28              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/33/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].group in JSON should be:
    """
    []
    """

  Scenario: The source group is public
    Given I am the user with id "21"
    And the database table "groups" also has the following row:
      | id | name       | is_public | type |
      | 30 | some group | true      | User |
    And the database table "group_managers" also has the following row:
      | group_id | manager_id | can_manage  |
      | 30       | 21         | memberships |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 30              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].source_group.id in JSON should be:
    """
    ["30"]
    """

  Scenario: The source group is an ancestor of the current user
    Given I am the user with id "21"
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 28              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].source_group.id in JSON should be:
    """
    ["28"]
    """

  Scenario: The source group is an ancestor of a team the current user is a member of
    Given I am the user with id "21"
    And the database table "groups" also has the following row:
      | id | name               | type |
      | 40 | team's grandparent | Club |
    And the database table "groups_groups" also has the following row:
      | parent_group_id | child_group_id |
      | 40              | 29             |
    And the groups ancestors are computed
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 40              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].source_group.id in JSON should be:
    """
    ["40"]
    """

  Scenario Outline: The user is a manager of a non-user descendant of the source group
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 30 | a group            | Club  |
      | 31 | a child group      | Club  |
      | 32 | a grandchild group | Other |
      | 33 | a descendant group | Club  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage   | can_watch_members   | can_grant_group_access   |
      | 33       | 28         | <can_manage> | <can_watch_members> | <can_grant_group_access> |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 30              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].source_group.id in JSON should be:
    """
    ["30"]
    """
    Examples:
      | can_manage            | can_watch_members | can_grant_group_access |
      | memberships           | false             | false                  |
      | memberships_and_group | false             | false                  |
      | none                  | true              | false                  |
      | none                  | false             | true                   |

  Scenario: The user is a manager of a non-user descendant of the source group without any managerial permissions (the source group should be invisible)
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 30 | a group            | Club  |
      | 31 | a child group      | Club  |
      | 32 | a grandchild group | Other |
      | 33 | a descendant group | Club  |
      | 34 | a descendant user  | User  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 33             |
      | 33              | 34             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage  | can_watch_members | can_grant_group_access |
      | 33       | 28         | none        | false             | false                  |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 31              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].source_group in JSON should be:
    """
    []
    """

  Scenario: The user is a manager of a descendant user of the source group (the source group should be invisible)
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 30 | a group            | Club  |
      | 31 | a child group      | Club  |
      | 32 | a grandchild group | Other |
      | 33 | a descendant user  | User  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage            | can_watch_members | can_grant_group_access |
      | 33       | 28         | memberships_and_group | true              | true                   |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 30              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].source_group in JSON should be:
    """
    []
    """

  Scenario Outline: The source group is a user implicitly managed by the current user
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 30 | a group            | Club  |
      | 31 | a child group      | Club  |
      | 32 | a grandchild group | Other |
      | 33 | a descendant user  | User  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage   | can_watch_members   | can_grant_group_access   |
      | 30       | 28         | <can_manage> | <can_watch_members> | <can_grant_group_access> |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 33              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].source_group.id in JSON should be:
    """
    ["33"]
    """
    Examples:
      | can_manage            | can_watch_members | can_grant_group_access |
      | memberships           | false             | false                  |
      | memberships_and_group | false             | false                  |
      | none                  | true              | false                  |
      | none                  | false             | true                   |

  Scenario: The source group is a user implicitly managed by the current user without any managerial permissions (the source group should be invisible)
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 30 | a group            | Club  |
      | 31 | a child group      | Club  |
      | 32 | a grandchild group | Other |
      | 33 | a descendant user  | User  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage  | can_watch_members | can_grant_group_access |
      | 30       | 28         | none        | false             | false                  |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 33              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].source_group in JSON should be:
    """
    []
    """

  Scenario: The source_group is a user explicitly managed by the current user (the source group should be invisible)
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name               | type  |
      | 33 | a descendant user  | User  |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage            | can_watch_members   | can_grant_group_access |
      | 33       | 28         | memberships_and_group | true                | true                   |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 33              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].source_group in JSON should be:
    """
    []
    """

  Scenario Outline: The source group is a user in a team managed by the current user
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name              | type |
      | 30 | a group           | Club |
      | 31 | a child group     | Club |
      | 32 | a grandchild team | Team |
      | 33 | a descendant user | User |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 21             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage   | can_watch_members   | can_grant_group_access   |
      | 30       | 28         | <can_manage> | <can_watch_members> | <can_grant_group_access> |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 33              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].source_group.id in JSON should be:
    """
    ["33"]
    """
    Examples:
      | can_manage            | can_watch_members | can_grant_group_access |
      | memberships           | false             | false                  |
      | memberships_and_group | false             | false                  |
      | none                  | true              | false                  |
      | none                  | false             | true                   |

  Scenario: The source group is a user in a team managed by the current user, but the current user doesn't have any managerial permission (the source group should be invisible)
    Given I am the user with id "21"
    And the database table "groups" also has the following rows:
      | id | name              | type |
      | 30 | a group           | Club |
      | 31 | a child group     | Club |
      | 32 | a grandchild team | Team |
      | 33 | a descendant user | User |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 30              | 31             |
      | 31              | 32             |
      | 32              | 21             |
      | 32              | 33             |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | group_id | manager_id | can_manage   | can_watch_members   | can_grant_group_access |
      | 30       | 28         | none         | false               | false                  |
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 21       | 100     | 33              | item_unlocking | none     | none           | result    | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed
    When I send a GET request to "/groups/21/permissions/101/explain"
    Then the response code should be 200
    And the response should be a JSON array with 1 entry
    And the response at $[*].source_group in JSON should be:
    """
    []
    """
