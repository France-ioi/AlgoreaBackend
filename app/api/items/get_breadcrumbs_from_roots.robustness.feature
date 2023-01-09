Feature: Find all breadcrumbs to an item - robustness
  Background:
    Given the database has the following table 'groups':
      | id  | type  | root_activity_id | root_skill_id |
      | 90  | Class | 10               | 20            |
      | 91  | Other | 50               | null          |
      | 101 | User  | null             | null          |
      | 102 | Team  | 60               | 30            |
      | 111 | User  | null             | null          |
    And the database has the following table 'users':
      | login | group_id | default_language |
      | john  | 101      | en               |
      | jane  | 111      | fr               |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 90              | 102            |
      | 91              | 111            |
      | 102             | 101            |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | manager_id | group_id | can_watch_members |
      | 91         | 90       | true              |
      | 111        | 111      | false             |
    And the database has the following table 'items':
      | id | url                                                                     | type    | default_language_tag | requires_explicit_entry |
      | 10 | null                                                                    | Chapter | en                   | false                   |
      | 20 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | en                   | true                    |
      | 30 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | en                   | false                   |
      | 40 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | en                   | false                   |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | en                   | false                   |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | en                   | false                   |
      | 70 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | fr                   | false                   |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title            |
      | 10      | fr           | Graphe: Methodes |
      | 10      | en           | Graph: Methods   |
      | 20      | en           | BFS              |
      | 50      | en           | DFS              |
      | 60      | en           | Reduce Graph     |
      | 70      | fr           | null             |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 10             | 60            | 1           |
      | 10             | 20            | 1           |
      | 30             | 40            | 1           |
      | 60             | 70            | 2           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
      | 10               | 20            |
      | 10               | 70            |
      | 30               | 40            |
      | 60               | 70            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 102      | 60      | info                     |
      | 111      | 10      | info                     |
      | 111      | 20      | info                     |
      | 111      | 60      | none                     |
      | 111      | 70      | none                     |
      | 111      | 30      | content_with_descendants |
      | 111      | 40      | content_with_descendants |
      | 111      | 50      | content_with_descendants |
    And the database has the following table 'attempts':
      | id | participant_id | root_item_id | parent_attempt_id |
      | 0  | 101            | null         | null              |
      | 0  | 102            | null         | null              |
      | 0  | 111            | null         | null              |
      | 1  | 102            | 10           | null              |
      | 2  | 102            | 10           | null              |
      | 3  | 102            | 10           | null              |
      | 4  | 102            | 10           | null              |
      | 5  | 102            | 30           | null              |
      | 6  | 102            | 40           | 4                 |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          |
      | 1          | 102            | 10      | 2019-05-30 11:00:00 |
      | 2          | 102            | 10      | 2019-05-30 11:00:00 |
      | 2          | 102            | 60      | 2019-05-30 11:00:00 |
      | 3          | 102            | 10      | 2019-05-30 11:00:00 |
      | 3          | 102            | 60      | 2019-05-30 11:00:00 |
      | 3          | 102            | 70      | 2019-05-30 11:00:00 |
      | 3          | 102            | 20      | null                |
      | 4          | 102            | 20      | 2019-05-30 11:00:00 |
      | 5          | 102            | 30      | 2019-05-30 11:00:00 |
      | 6          | 102            | 40      | 2019-05-30 11:00:00 |
      | 0          | 111            | 10      | 2019-05-30 11:00:00 |
      | 0          | 111            | 50      | 2019-05-30 11:00:00 |

  Scenario: Invalid item_id
    And I am the user with id "111"
    When I send a GET request to "/items/100000000000000000000000/breadcrumbs-from-roots"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Invalid participant_id
    And I am the user with id "111"
    When I send a GET request to "/items/10/breadcrumbs-from-roots?participant_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for participant_id (should be int64)"

  Scenario Outline: No access to participant_id
    Given I am the user with id "111"
    When I send a GET request to "/items/10/breadcrumbs-from-roots?participant_id=<participant_id>"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    Examples:
      | participant_id |
      | 404            |
      | 111            |

  Scenario Outline: No paths
    Given I am the user with id "111"
    When I send a GET request to "/items/<item_id>/breadcrumbs-from-roots?participant_id=102"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
  Examples:
    | item_id |
    | 70      |
    | 60      |
    | 20      |
    | 40      |
