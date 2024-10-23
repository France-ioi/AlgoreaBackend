Feature: Get permissions granted to a group
  Background:
    Given the database has the following table "groups":
      | id | name        | type  |
      | 8  | Club        | Club  |
      | 9  | Class       | Class |
      | 10 | Other       | Other |
      | 25 | some class  | Class |
      | 26 | other class | Class |
      | 27 | third class | Class |
    And the database has the following users:
      | group_id | login | first_name  | last_name | default_language |
      | 21       | owner | Jean-Michel | Blanquer  | fr               |
      | 23       | user  | John        | Doe       | en               |
      | 31       | jane  | Jane        | Doe       | en               |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_grant_group_access |
      | 25       | 8          | 1                      |
      | 26       | 8          | 1                      |
      | 10       | 21         | 0                      |
      | 31       | 21         | 0                      |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 8               | 21             |
      | 9               | 10             |
      | 10              | 25             |
      | 25              | 23             |
      | 25              | 31             |
      | 25              | 27             |
      | 26              | 25             |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id  | default_language_tag | requires_explicit_entry | type    |
      | 100 | fr                   | true                    | Chapter |
      | 101 | en                   | false                   | Task    |
      | 102 | fr                   | true                    | Chapter |
      | 103 | fr                   | false                   | Chapter |
      | 104 | fr                   | false                   | Chapter |
    And the database has the following table "items_strings":
      | item_id  | language_tag | title      |
      | 101      | en           | Task A     |
      | 102      | en           | Chapter B  |
      | 102      | fr           | Chapitre B |
      | 104      | fr           | Chapitre C |
    And the database table "permissions_granted" also has the following rows:
      | group_id | item_id | source_group_id | origin           | can_view | can_grant_view | can_watch | can_edit | is_owner | can_request_help_to | can_make_session_official | can_enter_from      | can_enter_until     |
      | 31       | 102     | 10              | group_membership | info     | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 101     | 25              | group_membership | none     | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 101     | 25              | item_unlocking   | info     | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 101     | 25              | self             | info     | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 101     | 25              | other            | info     | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 25              | group_membership | info     | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 101     | 26              | group_membership | none     | enter          | none      | none     | false    | 25                  | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 104     | 26              | group_membership | none     | none           | result    | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 101     | 25              | group_membership | none     | none           | none      | children | false    | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 25              | group_membership | none     | none           | none      | none     | true     | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 101     | 26              | group_membership | none     | none           | none      | none     | false    | null                | true                      | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 26              | group_membership | none     | none           | none      | none     | false    | null                | false                     | 2999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 9        | 101     | 25              | group_membership | none     | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 3999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content  | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 3999-12-31 23:59:59 |
      | 23       | 102     | 25              | group_membership | content  | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 3999-12-31 23:59:59 |
      | 25       | 101     | 10              | group_membership | none     | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 3999-12-31 23:59:59 |
      | 25       | 103     | 25              | group_membership | content  | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 31       | 101     | 27              | group_membership | content  | none           | none      | none     | false    | null                | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table "permissions_generated" also has the following rows:
      | group_id | item_id | can_grant_view_generated | can_watch_generated | can_edit_generated |
      | 8        | 101     | enter                    | none                | none               |
      | 21       | 102     | none                     | answer_with_grant   | none               |
      | 21       | 103     | none                     | answer              | all                |
      | 21       | 104     | none                     | none                | all_with_grant     |

  Scenario: can_grant_group_access=1 for the group
    Given I am the user with id "21"
    When I send a GET request to "/groups/25/granted_permissions"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {"id": "10", "name": "Other"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "2999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "26", "name": "other class"}
      },
      {
        "group": {"id": "10", "name": "Other"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": true,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      },
      {
        "group": {"id": "25", "name": "some class"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "info", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      },
      {
        "group": {"id": "25", "name": "some class"},
        "item": {"id": "104", "language_tag": "fr", "requires_explicit_entry": false, "title": "Chapitre C", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "result",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "26", "name": "other class"}
      },
      {
        "group": {"id": "10", "name": "Other"},
        "item": {"id": "101", "language_tag": "en", "requires_explicit_entry": false, "title": "Task A", "type": "Task"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": true, "can_view": "none", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "26", "name": "other class"}
      },
      {
        "group": {"id": "25", "name": "some class"},
        "item": {"id": "101", "language_tag": "en", "requires_explicit_entry": false, "title": "Task A", "type": "Task"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "enter", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": 25
        },
        "source_group": {"id": "26", "name": "other class"}
      },
      {
        "group": {"id": "9", "name": "Class"},
        "item": {"id": "101", "language_tag": "en", "requires_explicit_entry": false, "title": "Task A", "type": "Task"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "3999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      },
      {
        "group": {"id": "10", "name": "Other"},
        "item": {"id": "101", "language_tag": "en", "requires_explicit_entry": false, "title": "Task A", "type": "Task"},
        "permissions": {
          "can_edit": "children", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      }
    ]
    """

  Scenario: can_grant_group_access=1 for a ancestor group
    Given I am the user with id "21"
    And the database table "group_managers" also has the following rows:
      | group_id | manager_id | can_grant_group_access |
      | 9        | 8          | 1                      |
    When I send a GET request to "/groups/25/granted_permissions"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {"id": "10", "name": "Other"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "2999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "26", "name": "other class"}
      },
      {
        "group": {"id": "10", "name": "Other"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": true,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      },
      {
        "group": {"id": "25", "name": "some class"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "info", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      },
      {
        "group": {"id": "25", "name": "some class"},
        "item": {"id": "104", "language_tag": "fr", "requires_explicit_entry": false, "title": "Chapitre C", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "result",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "26", "name": "other class"}
      },
      {
        "group": {"id": "25", "name": "some class"},
        "item": {"id": "101", "language_tag": "en", "requires_explicit_entry": false, "title": "Task A", "type": "Task"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "3999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "10", "name": "Other"}
      },
      {
        "group": {"id": "10", "name": "Other"},
        "item": {"id": "101", "language_tag": "en", "requires_explicit_entry": false, "title": "Task A", "type": "Task"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": true, "can_view": "none", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "26", "name": "other class"}
      },
      {
        "group": {"id": "25", "name": "some class"},
        "item": {"id": "101", "language_tag": "en", "requires_explicit_entry": false, "title": "Task A", "type": "Task"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "enter", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": 25
        },
        "source_group": {"id": "26", "name": "other class"}
      },
      {
        "group": {"id": "9", "name": "Class"},
        "item": {"id": "101", "language_tag": "en", "requires_explicit_entry": false, "title": "Task A", "type": "Task"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "3999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      },
      {
        "group": {"id": "10", "name": "Other"},
        "item": {"id": "101", "language_tag": "en", "requires_explicit_entry": false, "title": "Task A", "type": "Task"},
        "permissions": {
          "can_edit": "children", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      }
    ]
    """

  Scenario: Get only the second row
    Given I am the user with id "21"
    When I send a GET request to "/groups/25/granted_permissions?from.source_group.id=26&from.group.id=10&from.item.id=102&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {"id": "10", "name": "Other"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "none", "can_watch": "none",
          "is_owner": true,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      }
    ]
    """

  Scenario: Get only the third row
    Given I am the user with id "21"
    When I send a GET request to "/groups/25/granted_permissions?from.source_group.id=25&from.group.id=10&from.item.id=102&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {"id": "25", "name": "some class"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "info", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      }
    ]
    """

  Scenario: can_grant_group_access=1 for the group's ancestor, descendants=1
    Given I am the user with id "21"
    And the database table "group_managers" also has the following rows:
      | group_id | manager_id | can_grant_group_access |
      | 9        | 8          | 1                      |
    When I send a GET request to "/groups/25/granted_permissions?descendants=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {"id": "31", "name": "jane"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "info", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "10", "name": "Other"}
      },
      {
        "group": {"id": "23", "name": "user"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "3999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "content", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      },
      {
        "group": {"id": "31", "name": "jane"},
        "item": {"id": "101", "language_tag": "en", "requires_explicit_entry": false, "title": "Task A", "type": "Task"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "content", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "27", "name": "third class"}
      }
    ]
    """

  Scenario: can_grant_group_access=1 for the group's ancestor, descendants=1, order by group.name, source_group.name, item.title
    Given I am the user with id "21"
    And the database table "group_managers" also has the following rows:
      | group_id | manager_id | can_grant_group_access |
      | 9        | 8          | 1                      |
    When I send a GET request to "/groups/25/granted_permissions?descendants=1&sort=group.name,source_group.name,item.title"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {"id": "31", "name": "jane"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "info", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "10", "name": "Other"}
      },
      {
        "group": {"id": "31", "name": "jane"},
        "item": {"id": "101", "language_tag": "en", "requires_explicit_entry": false, "title": "Task A", "type": "Task"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "content", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "27", "name": "third class"}
      },
      {
        "group": {"id": "23", "name": "user"},
        "item": {"id": "102", "language_tag": "fr", "requires_explicit_entry": true, "title": "Chapitre B", "type": "Chapter"},
        "permissions": {
          "can_edit": "none", "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "3999-12-31T23:59:59Z",
          "can_grant_view": "none", "can_make_session_official": false, "can_view": "content", "can_watch": "none",
          "is_owner": false,
          "can_request_help_to": null
        },
        "source_group": {"id": "25", "name": "some class"}
      }
    ]
    """
