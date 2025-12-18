Feature: Explain permissions - pagination
  Background:
    Given the database has the following table "groups":
      | id | name          | type  |
      | 25 | some class    | Class |
      | 27 | some club     | Club  |
      | 28 | other         | Other |
    And the database has the following users:
      | group_id | login | first_name  | last_name |
      | 21       | owner | Jean-Michel | Blanquer  |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 25              | 21             |
      | 27              | 21             |
      | 28              | 27             |
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
    And the database has the following table "permissions_granted":
      | group_id | item_id | source_group_id | origin         | can_view | can_grant_view | can_watch         | can_edit | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 27       | 100     | 25              | other          | none     | enter          | none              | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 27       | 100     | 25              | item_unlocking | content  | none           | none              | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 28       | 100     | 28              | self           | none     | none           | answer_with_grant | none     | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the generated permissions are computed

  Scenario: All rows
    Given I am the user with id "21"
    When I send a GET request to "/groups/27/permissions/101/explain"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {"id": "27", "name": "some club"},
        "item": {
          "id": "100", "language_tag": "en", "requires_explicit_entry": false, "title": "Some Item", "type": "Chapter"
        },
        "source_group": {"id": "25", "name": "some class"},
        "origin": "item_unlocking",
        "granted_permissions": {
          "can_view": "content", "can_grant_view": "none", "can_watch": "none", "can_edit": "none", "is_owner": false
        },
        "propagated_permissions": {
          "can_view_generated": "info", "can_grant_view_generated": "none", "can_watch_generated": "none",
          "can_edit_generated": "none", "is_owner_generated": false
        },
        "user_can_update_permission": false,
        "from": "f4da9716baa9497d807ab4b95a4579b6c3c2f4b5b5c91ab1d2f64eb19ef0992c"
      },
      {
        "group": {"id": "27", "name": "some club"},
        "item": {
          "id": "100", "language_tag": "en", "requires_explicit_entry": false, "title": "Some Item", "type": "Chapter"
        },
        "source_group": {"id": "25", "name": "some class"},
        "origin": "other",
        "granted_permissions": {
          "can_view": "none", "can_grant_view": "enter", "can_watch": "none", "can_edit": "none", "is_owner": false
        },
        "propagated_permissions": {
          "can_view_generated": "none", "can_grant_view_generated": "enter", "can_watch_generated": "none",
          "can_edit_generated": "none", "is_owner_generated": false
        },
        "user_can_update_permission": false,
        "from": "37c73f0ee917bbd508383028e595c5c11908808ed2bfac8ea35b98955b3756ab"
      },
      {
        "group": {"id": "28", "name": "other"},
        "item": {
          "id": "100", "language_tag": "en", "requires_explicit_entry": false, "title": "Some Item", "type": "Chapter"
        },
        "source_group": {"id": "28", "name": "other"},
        "origin": "self",
        "granted_permissions": {
          "can_view": "none", "can_grant_view": "none", "can_watch": "answer_with_grant", "can_edit": "none", "is_owner": false
        },
        "propagated_permissions": {
          "can_view_generated": "none", "can_grant_view_generated": "none", "can_watch_generated": "answer",
          "can_edit_generated": "none", "is_owner_generated": false
        },
        "user_can_update_permission": false,
        "from": "084bb649a83e7565606236777d310437fb6ff19bcd603506c5582272061a59c3"
      }
    ]
    """

  Scenario: Start from the second row
    Given I am the user with id "21"
    When I send a GET request to "/groups/27/permissions/101/explain?from=f4da9716baa9497d807ab4b95a4579b6c3c2f4b5b5c91ab1d2f64eb19ef0992c"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {"id": "27", "name": "some club"},
        "item": {
          "id": "100", "language_tag": "en", "requires_explicit_entry": false, "title": "Some Item", "type": "Chapter"
        },
        "source_group": {"id": "25", "name": "some class"},
        "origin": "other",
        "granted_permissions": {
          "can_view": "none", "can_grant_view": "enter", "can_watch": "none", "can_edit": "none", "is_owner": false
        },
        "propagated_permissions": {
          "can_view_generated": "none", "can_grant_view_generated": "enter", "can_watch_generated": "none",
          "can_edit_generated": "none", "is_owner_generated": false
        },
        "user_can_update_permission": false,
        "from": "37c73f0ee917bbd508383028e595c5c11908808ed2bfac8ea35b98955b3756ab"
      },
      {
        "group": {"id": "28", "name": "other"},
        "item": {
          "id": "100", "language_tag": "en", "requires_explicit_entry": false, "title": "Some Item", "type": "Chapter"
        },
        "source_group": {"id": "28", "name": "other"},
        "origin": "self",
        "granted_permissions": {
          "can_view": "none", "can_grant_view": "none", "can_watch": "answer_with_grant", "can_edit": "none", "is_owner": false
        },
        "propagated_permissions": {
          "can_view_generated": "none", "can_grant_view_generated": "none", "can_watch_generated": "answer",
          "can_edit_generated": "none", "is_owner_generated": false
        },
        "user_can_update_permission": false,
        "from": "084bb649a83e7565606236777d310437fb6ff19bcd603506c5582272061a59c3"
      }
    ]
    """

  Scenario: Start from the second row, take only the first row
    Given I am the user with id "21"
    When I send a GET request to "/groups/27/permissions/101/explain?from=f4da9716baa9497d807ab4b95a4579b6c3c2f4b5b5c91ab1d2f64eb19ef0992c&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group": {"id": "27", "name": "some club"},
        "item": {
          "id": "100", "language_tag": "en", "requires_explicit_entry": false, "title": "Some Item", "type": "Chapter"
        },
        "source_group": {"id": "25", "name": "some class"},
        "origin": "other",
        "granted_permissions": {
          "can_view": "none", "can_grant_view": "enter", "can_watch": "none", "can_edit": "none", "is_owner": false
        },
        "propagated_permissions": {
          "can_view_generated": "none", "can_grant_view_generated": "enter", "can_watch_generated": "none",
          "can_edit_generated": "none", "is_owner_generated": false
        },
        "user_can_update_permission": false,
        "from": "37c73f0ee917bbd508383028e595c5c11908808ed2bfac8ea35b98955b3756ab"
      }
    ]
    """
