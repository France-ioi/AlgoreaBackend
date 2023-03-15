Feature: Find all breadcrumbs to an item
  Background:
    Given the database has the following table 'groups':
      | id  | type  | root_activity_id | root_skill_id |
      | 90  | Class | 10               | null          |
      | 91  | Other | 50               | null          |
      | 92  | Club  | 80               | null          |
      | 93  | Class | null             | 90            |
      | 94  | Club  | null             | null          |
      | 101 | User  | null             | null          |
      | 102 | Team  | null             | 60            |
      | 111 | User  | null             | null          |
    And the database has the following table 'users':
      | login | group_id | default_language |
      | john  | 101      | en               |
      | jane  | 111      | fr               |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 90              | 111            |
      | 90              | 102            |
      | 91              | 111            |
      | 94              | 93             |
      | 102             | 101            |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | manager_id | group_id | can_watch_members |
      | 91         | 90       | true              |
      | 91         | 94       | false             |
      | 111        | 92       | false             |
    And the database has the following table 'items':
      | id | url                    | type    | default_language_tag | text_id |
      | 10 | null                   | Chapter | en                   | id10    |
      | 50 | http://taskplatform/50 | Task    | en                   | id50    |
      | 60 | http://taskplatform/60 | Task    | en                   | id60    |
      | 70 | http://taskplatform/70 | Task    | fr                   | id70    |
      | 80 | null                   | Chapter | en                   | id80    |
      | 90 | null                   | Chapter | en                   | id90    |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title            |
      | 10      | fr           | Graphe: Methodes |
      | 10      | en           | Graph: Methods   |
      | 50      | en           | DFS              |
      | 60      | en           | Reduce Graph     |
      | 70      | fr           | null             |
      | 80      | en           | Trees            |
      | 90      | en           | Queues           |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 10             | 60            | 1           |
      | 60             | 70            | 1           |
      | 80             | 90            | 1           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
      | 10               | 70            |
      | 60               | 70            |
      | 80               | 90            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 102      | 60      | none                     |
      | 111      | 10      | content_with_descendants |
      | 111      | 60      | content                  |
      | 111      | 70      | info                     |
      | 111      | 50      | content_with_descendants |
      | 111      | 80      | content                  |
      | 111      | 90      | info                     |
    And the database has the following table 'attempts':
      | id | participant_id | root_item_id | parent_attempt_id |
      | 0  | 101            | null         | null              |
      | 0  | 102            | null         | null              |
      | 0  | 111            | null         | null              |
      | 1  | 111            | 80           | 0                 |
      | 1  | 102            | 10           | null              |
      | 2  | 102            | 10           | null              |
      | 3  | 102            | 60           | 1                 |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          |
      | 1          | 102            | 10      | 2020-01-01 00:00:00 |
      | 2          | 102            | 60      | 2020-01-01 00:00:00 |
      | 3          | 102            | 60      | 2020-01-01 00:00:00 |
      | 3          | 102            | 70      | 2020-01-01 00:00:00 |
      | 0          | 111            | 10      | 2020-01-01 00:00:00 |
      | 0          | 111            | 50      | 2020-01-01 00:00:00 |
      | 1          | 111            | 80      | 2020-01-01 00:00:00 |
      | 1          | 111            | 90      | 2020-01-01 00:00:00                    |

  Scenario Outline: Find breadcrumbs for the current user
    Given I am the user with id "111"
    When I send a GET request to "<service_url>"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      <expected_output>
      """
  Examples:
    | service_url                                   | expected_output                                                                                                                                                                                                                   |
    | /items/50/breadcrumbs-from-roots              | [[{"id": "50", "title": "DFS", "language_tag": "en", "type": "Task"}]]                                                                                                                                                            |
    | /items/by-text-id/id50/breadcrumbs-from-roots | [[{"id": "50", "title": "DFS", "language_tag": "en", "type": "Task"}]]                                                                                                                                                            |
    | /items/10/breadcrumbs-from-roots              | [[{"id": "10", "title": "Graphe: Methodes", "language_tag": "fr", "type": "Chapter"}]]                                                                                                                                            |
    | /items/by-text-id/id10/breadcrumbs-from-roots | [[{"id": "10", "title": "Graphe: Methodes", "language_tag": "fr", "type": "Chapter"}]]                                                                                                                                            |
    | /items/90/breadcrumbs-from-roots              | [[{"id": "80", "title": "Trees", "language_tag": "en", "type": "Chapter"}, {"id": "90", "title": "Queues", "language_tag": "en", "type": "Chapter"}], [{"id": "90", "title": "Queues", "language_tag": "en", "type": "Chapter"}]] |
    | /items/by-text-id/id90/breadcrumbs-from-roots | [[{"id": "80", "title": "Trees", "language_tag": "en", "type": "Chapter"}, {"id": "90", "title": "Queues", "language_tag": "en", "type": "Chapter"}], [{"id": "90", "title": "Queues", "language_tag": "en", "type": "Chapter"}]] |

  Scenario Outline: Find breadcrumbs for a team
    Given I am the user with id "111"
    When I send a GET request to "<service_url>?participant_id=102"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      [
        [
          {"id": "10", "title": "Graphe: Methodes", "language_tag": "fr", "type": "Chapter"},
          {"id": "60", "title": "Reduce Graph", "language_tag": "en", "type": "Task"},
          {"id": "70", "title": null, "language_tag": "fr", "type": "Task"}
        ],
        [
          {"id": "60", "title": "Reduce Graph", "language_tag": "en", "type": "Task"},
          {"id": "70","title": null, "language_tag": "fr", "type": "Task"}
        ]
      ]
      """
    Examples:
      | service_url                                   |
      | /items/70/breadcrumbs-from-roots              |
      | /items/by-text-id/id70/breadcrumbs-from-roots |

  Scenario Outline: Find breadcrumbs for a team for another item
    Given I am the user with id "111"
    When I send a GET request to "<service_url>?participant_id=102"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      [
        [
          {"id": "10", "title": "Graphe: Methodes", "language_tag": "fr", "type": "Chapter"},
          {"id": "60", "title": "Reduce Graph", "language_tag": "en", "type": "Task"}
        ],
        [{"id": "60", "title": "Reduce Graph", "language_tag": "en", "type": "Task"}]
      ]
      """
    Examples:
      | service_url                                   |
      | /items/60/breadcrumbs-from-roots              |
      | /items/by-text-id/id60/breadcrumbs-from-roots |
