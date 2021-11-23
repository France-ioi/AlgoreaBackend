Feature: Get permissions for a group
  Background:
    Given the database has the following table 'groups':
      | id | name       | type  |
      | 10 | Other      | Other |
      | 21 | owner      | User  |
      | 23 | user       | User  |
      | 25 | some class | Class |
      | 31 | jane       | User  |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | user  | 23       | John        | Doe       |
      | jane  | 31       | Jane        | Doe       |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_grant_group_access |
      | 25       | 21         | 1                      |
      | 31       | 21         | 0                      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 10              | 25             |
      | 25              | 23             |
      | 25              | 31             |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id  | default_language_tag |
      | 100 | fr                   |
      | 101 | fr                   |
      | 102 | fr                   |
      | 103 | fr                   |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | content_view_propagation | grant_view_propagation | watch_propagation | edit_propagation | child_order |
      | 100            | 101           | as_info                  | false                  | false             | false            | 0           |
      | 101            | 102           | as_content               | false                  | false             | false            | 0           |
      | 102            | 103           | as_content               | true                   | true              | true             | 0           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 100              | 101           |
      | 100              | 102           |
      | 100              | 103           |
      | 101              | 102           |
      | 101              | 103           |
      | 102              | 103           |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated        | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 23       | 100     | content_with_descendants  | none                     | none                | none               | false              |
      | 23       | 101     | info                      | none                     | none                | none               | false              |
      | 23       | 103     | solution                  | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view      | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 100     | 23              | other            | content                  | none                | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 103     | 23              | group_membership | solution                 | solution_with_grant | answer_with_grant | all_with_grant | true     | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 103     | 25              | group_membership | solution                 | solution_with_grant | answer_with_grant | all_with_grant | true     | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 103     | 23              | group_membership | solution                 | solution_with_grant | answer_with_grant | all_with_grant | true     | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 103     | 23              | item_unlocking   | solution                 | solution_with_grant | answer_with_grant | all_with_grant | true     | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 103     | 23              | item_unlocking   | solution                 | solution_with_grant | answer_with_grant | all_with_grant | true     | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 103     | 25              | self             | solution                 | solution_with_grant | answer_with_grant | all_with_grant | true     | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 103     | 23              | self             | solution                 | solution_with_grant | answer_with_grant | all_with_grant | true     | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 103     | 25              | other            | solution                 | solution_with_grant | answer_with_grant | all_with_grant | true     | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 103     | 23              | other            | solution                 | solution_with_grant | answer_with_grant | all_with_grant | true     | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |

  Scenario: No permissions
    Given I am the user with id "21"
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 21       | 103     | solution           | solution                 | answer              | all                | true               |
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | can_view | can_grant_view      | can_watch         | can_edit       | source_group_id | latest_update_at    |
      | 21       | 102     | solution | solution_with_grant | answer_with_grant | all_with_grant | 23              | 2019-05-30 11:00:00 |
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "none", "can_grant_view": "none", "can_edit": "none", "can_watch": "none",
          "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "none", "can_grant_view": "none", "can_edit": "none", "can_watch": "none",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_group_membership": {
          "can_view": "none", "can_grant_view": "none", "can_edit": "none", "can_watch": "none",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "none", "can_grant_view": "none", "can_edit": "none", "can_watch": "none",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "none", "can_grant_view": "none", "can_edit": "none", "can_watch": "none",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "none", "can_grant_view": "none", "can_edit": "none", "can_watch": "none",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        }
      }
    """

  Scenario: Maximum permissions given directly
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2017-12-31 23:59:59 | 9998-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | content                  | content                  | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | info                     | enter                    | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | content                  | content_with_descendants | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2017-12-31T23:59:59Z", "can_enter_until": "9998-12-31T23:59:59Z",
          "can_make_session_official": true, "is_owner": true
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        }
      }
    """

  Scenario: Maximum permissions given via group membership
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2017-12-31 23:59:59 | 9998-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | content                  | content                  | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | info                     | enter                    | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | content                  | content_with_descendants | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_item_unlocking": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        }
      }
    """

  Scenario: Maximum permissions given via ancestor's group membership
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 25              | group_membership | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2017-12-31 23:59:59 | 9998-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | content                  | content                  | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | info                     | enter                    | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | content                  | content_with_descendants | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_item_unlocking": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        }
      }
    """

  Scenario: Maximum permissions given via item unlocking
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content                  | content                  | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2017-12-31 23:59:59 | 9998-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | info                     | enter                    | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | content                  | content_with_descendants | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_self": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        }
      }
    """

  Scenario: Maximum permissions given via ancestor's item unlocking
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content                  | content                  | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | item_unlocking   | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2017-12-31 23:59:59 | 9998-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | info                     | enter                    | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | content                  | content_with_descendants | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_self": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        }
      }
    """

  Scenario: Maximum permissions given via self
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content                  | content                  | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | info                     | enter                    | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2017-12-31 23:59:59 | 9998-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | content                  | content_with_descendants | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_other": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        }
      }
    """

  Scenario: Maximum permissions given via ancestor's self
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content                  | content                  | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | info                     | enter                    | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2017-12-31 23:59:59 | 9998-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | content                  | content_with_descendants | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_other": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        }
      }
    """

  Scenario: Maximum permissions given via other
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content                  | content                  | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | info                     | enter                    | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | content                  | content_with_descendants | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2017-12-31 23:59:59 | 9998-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        }
      }
    """

  Scenario: Maximum permissions given via ancestor's other
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content                  | content                  | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | info                     | enter                    | answer            | all            | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | content                  | content_with_descendants | result            | children       | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | none                     | none                     | none              | none           | false    | false                     | 9999-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2017-12-31 23:59:59 | 9998-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        }
      }
    """

  Scenario: can_enter_from aggregation
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 2021-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content                  | content                  | result            | children       | false    | false                     | 2022-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 2023-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | info                     | enter                    | answer            | all            | false    | false                     | 2024-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 2025-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | content                  | content_with_descendants | result            | children       | false    | false                     | 2026-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 2027-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | none                     | none                     | none              | none           | false    | false                     | 2028-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2029-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "2021-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2021-12-31T23:59:59Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "2022-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "2024-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "2026-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2028-12-31T23:59:59Z",
          "can_make_session_official": true, "is_owner": true
        }
      }
    """

  Scenario: can_enter_from aggregation (dates in the reverse order)
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 2029-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content                  | content                  | result            | children       | false    | false                     | 2028-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 2027-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | info                     | enter                    | answer            | all            | false    | false                     | 2026-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 2025-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | content                  | content_with_descendants | result            | children       | false    | false                     | 2024-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 2023-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | none                     | none                     | none              | none           | false    | false                     | 2022-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2021-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "2029-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2021-12-31T23:59:59Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "2027-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "2025-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "2023-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2021-12-31T23:59:59Z",
          "can_make_session_official": true, "is_owner": true
        }
      }
    """

  Scenario: can_enter_from aggregation (ignores rows with can_enter_until in the past)
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 2029-12-31 23:59:59 | 2020-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content                  | content                  | result            | children       | false    | false                     | 2028-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 2027-12-31 23:59:59 | 2020-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | info                     | enter                    | answer            | all            | false    | false                     | 2026-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 2025-12-31 23:59:59 | 2020-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | content                  | content_with_descendants | result            | children       | false    | false                     | 2024-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 2023-12-31 23:59:59 | 2020-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | none                     | none                     | none              | none           | false    | false                     | 2022-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2021-12-31 23:59:59 | 2020-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2021-01-10 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "2029-12-31T23:59:59Z", "can_enter_until": "2020-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2022-12-31T23:59:59Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "2028-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "2026-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "2024-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2022-12-31T23:59:59Z",
          "can_make_session_official": true, "is_owner": true
        }
      }
    """

  Scenario: can_enter_from aggregation (can_enter_from in the past)
    Given I am the user with id "21"
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | source_group_id | origin           | can_view                 | can_grant_view           | can_watch         | can_edit       | is_owner | can_make_session_official | can_enter_from      | can_enter_until     |
      | 23       | 102     | 25              | group_membership | content_with_descendants | solution                 | answer            | all            | false    | false                     | 2011-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 23              | group_membership | content                  | content                  | result            | children       | false    | false                     | 2012-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | group_membership | none                     | none                     | none              | none           | false    | false                     | 2013-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | item_unlocking   | info                     | enter                    | answer            | all            | false    | false                     | 2014-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 25       | 102     | 1               | item_unlocking   | none                     | none                     | none              | none           | false    | false                     | 2015-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | self             | content                  | content_with_descendants | result            | children       | false    | false                     | 2016-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | self             | none                     | none                     | none              | none           | false    | false                     | 2017-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 23       | 102     | 2               | other            | none                     | none                     | none              | none           | false    | false                     | 2018-12-31 23:59:59 | 9999-12-31 23:59:59 |
      | 10       | 102     | 1               | other            | solution                 | solution_with_grant      | answer_with_grant | all_with_grant | true     | true                      | 2019-12-31 23:59:59 | 9999-12-31 23:59:59 |
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 23       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 10       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
      | 25       | 102     | solution           | solution_with_grant      | answer_with_grant   | all_with_grant     | true               |
    Given the DB time now is "2019-07-16 22:02:28"
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
      {
        "granted": {
          "can_view": "content_with_descendants", "can_grant_view": "solution", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "2011-12-31T23:59:59Z", "can_enter_until": "9999-12-31T23:59:59Z",
          "can_make_session_official": false, "is_owner": false
        },
        "computed": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        },
        "granted_via_group_membership": {
          "can_view": "content", "can_grant_view": "content", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_item_unlocking": {
          "can_view": "info", "can_grant_view": "enter", "can_edit": "all", "can_watch": "answer",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_self": {
          "can_view": "content", "can_grant_view": "content_with_descendants", "can_edit": "children", "can_watch": "result",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": false, "is_owner": false
        },
        "granted_via_other": {
          "can_view": "solution", "can_grant_view": "solution_with_grant", "can_edit": "all_with_grant", "can_watch": "answer_with_grant",
          "can_enter_from": "2019-07-16T22:02:28Z",
          "can_make_session_official": true, "is_owner": true
        }
      }
    """

  Scenario Outline: Access rights for the current user
    Given I am the user with id "21"
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated   | can_watch_generated   | can_edit_generated   | is_owner_generated |
      | 21       | 102     | none               | <can_grant_view_generated> | <can_watch_generated> | <can_edit_generated> | false              |
    When I send a GET request to "/groups/25/permissions/23/102"
    Then the response code should be 200
  Examples:
    | can_grant_view_generated | can_watch_generated | can_edit_generated |
    | enter                    | none                | none               |
    | none                     | answer_with_grant   | none               |
    | none                     | none                | all_with_grant     |
