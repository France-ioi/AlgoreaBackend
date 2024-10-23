Feature: Find all breadcrumbs to an item
  Background:
    Given the database has the following table "groups":
      | id  | type  | root_activity_id | root_skill_id |
      | 90  | Class | 10               | null          |
      | 91  | Other | 50               | null          |
      | 92  | Club  | 80               | null          |
      | 93  | Class | null             | 90            |
      | 94  | Club  | null             | null          |
      | 102 | Team  | 100              | 60            |
    And the database has the following users:
      | group_id | login | default_language |
      | 101      | john  | en               |
      | 111      | jane  | fr               |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 90              | 111            |
      | 90              | 102            |
      | 91              | 111            |
      | 94              | 93             |
      | 102             | 101            |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | manager_id | group_id | can_watch_members |
      | 91         | 90       | true              |
      | 91         | 94       | false             |
      | 111        | 92       | false             |
    And the database has the following table "items":
      | id  | url                    | type    | default_language_tag | text_id               | requires_explicit_entry |
      | 10  | null                   | Chapter | en                   | id10                  | false                   |
      | 50  | http://taskplatform/50 | Task    | en                   | -_ '#&?:=/\.,+%¤€aéàd | false                   |
      | 60  | http://taskplatform/60 | Task    | en                   | id60                  | false                   |
      | 70  | http://taskplatform/70 | Task    | fr                   | id70                  | false                   |
      | 80  | null                   | Chapter | en                   | id80                  | false                   |
      | 90  | null                   | Chapter | en                   | id90                  | false                   |
      | 100 | null                   | Chapter | en                   | id100                 | false                   |
      | 101 | null                   | Task    | en                   | id101                 | true                    |
    And the database has the following table "items_strings":
      | item_id | language_tag | title                                         |
      | 10      | fr           | Graphe: Methodes                              |
      | 10      | en           | Graph: Methods                                |
      | 50      | en           | DFS                                           |
      | 60      | en           | Reduce Graph                                  |
      | 70      | fr           | null                                          |
      | 80      | en           | Trees                                         |
      | 90      | en           | Queues                                        |
      | 100     | en           | Chapter Containing Explicit Entry Not Started |
      | 101     | en           | Explicit Entry Not Started                    |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order |
      | 10             | 60            | 1           |
      | 60             | 70            | 1           |
      | 80             | 90            | 1           |
      | 100            | 101           | 1           |
    And the database has the following table "items_ancestors":
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
      | 10               | 70            |
      | 60               | 70            |
      | 80               | 90            |
      | 100              | 101           |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated       |
      | 102      | 60      | none                     |
      | 111      | 10      | content_with_descendants |
      | 111      | 60      | content                  |
      | 111      | 70      | info                     |
      | 111      | 50      | content_with_descendants |
      | 111      | 80      | content                  |
      | 111      | 90      | info                     |
      | 111      | 100     | content                  |
      | 111      | 101     | content                  |
    And the database has the following table "attempts":
      | id | participant_id | root_item_id | parent_attempt_id |
      | 0  | 101            | null         | null              |
      | 0  | 102            | null         | null              |
      | 0  | 111            | null         | null              |
      | 1  | 111            | 80           | 0                 |
      | 1  | 102            | 10           | null              |
      | 2  | 102            | 10           | null              |
      | 3  | 102            | 60           | 1                 |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | started_at          |
      | 1          | 102            | 10      | 2020-01-01 00:00:00 |
      | 2          | 102            | 60      | 2020-01-01 00:00:00 |
      | 3          | 102            | 60      | 2020-01-01 00:00:00 |
      | 3          | 102            | 70      | 2020-01-01 00:00:00 |
      | 0          | 111            | 10      | 2020-01-01 00:00:00 |
      | 0          | 111            | 50      | 2020-01-01 00:00:00 |
      | 1          | 111            | 80      | 2020-01-01 00:00:00 |
      | 1          | 111            | 90      | 2020-01-01 00:00:00 |
      | 0          | 111            | 100     | 2020-01-01 00:00:00 |

  Scenario Outline: Find breadcrumbs for the current user
    Given I am the user with id "111"
    When I send a GET request to "<service_url>"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      <expected_output>
      """
  Examples:
    | service_url                                                                                                   | expected_output                                                                                                                                                                                                                                                                                                       |
    | /items/50/breadcrumbs-from-roots                                                                              | [{"started_by_participant": true, "path": [{"id": "50", "title": "DFS", "language_tag": "en", "type": "Task"}]}]                                                                                                                                                                                                      |
    | /items/by-text-id/-_%20%27%23%26%3F%3A%3D%2F%5C.%2C%2B%25%C2%A4%E2%82%ACa%C3%A9%C3%A0d/breadcrumbs-from-roots | [{"started_by_participant": true, "path": [{"id": "50", "title": "DFS", "language_tag": "en", "type": "Task"}]}]                                                                                                                                                                                                      |
    | /items/10/breadcrumbs-from-roots                                                                              | [{"started_by_participant": true, "path": [{"id": "10", "title": "Graphe: Methodes", "language_tag": "fr", "type": "Chapter"}]}]                                                                                                                                                                                      |
    | /items/by-text-id/id10/breadcrumbs-from-roots                                                                 | [{"started_by_participant": true, "path": [{"id": "10", "title": "Graphe: Methodes", "language_tag": "fr", "type": "Chapter"}]}]                                                                                                                                                                                      |
    | /items/90/breadcrumbs-from-roots                                                                              | [{"started_by_participant": true, "path": [{"id": "80", "title": "Trees", "language_tag": "en", "type": "Chapter"}, {"id": "90", "title": "Queues", "language_tag": "en", "type": "Chapter"}]}, {"started_by_participant": true, "path": [{"id": "90", "title": "Queues", "language_tag": "en", "type": "Chapter"}]}] |
    | /items/by-text-id/id90/breadcrumbs-from-roots                                                                 | [{"started_by_participant": true, "path": [{"id": "80", "title": "Trees", "language_tag": "en", "type": "Chapter"}, {"id": "90", "title": "Queues", "language_tag": "en", "type": "Chapter"}]}, {"started_by_participant": true, "path": [{"id": "90", "title": "Queues", "language_tag": "en", "type": "Chapter"}]}] |

  Scenario: Should return a breadcrumb when there are missing results, like path-from-root
    Given the database has the following user:
      | group_id | login | default_language |
      | 1000     | user  | en               |
    And the database has the following table "groups":
      | id   | type  | root_activity_id |
      | 1001 | Class | 1010             |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 1001            | 1000           |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id   | url                      | type    | default_language_tag | text_id |
      | 1010 | null                     | Chapter | en                   | id1010  |
      | 1011 | null                     | Chapter | en                   | id1011  |
      | 1012 | null                     | Chapter | en                   | id1012  |
      | 1020 | http://taskplatform/1020 | Task    | en                   | id1020  |
    And the database has the following table "items_strings":
      | item_id | language_tag | title     |
      | 1010    | en           | Chapter 1 |
      | 1011    | en           | Chapter 2 |
      | 1012    | en           | Chapter 3 |
      | 1020    | en           | Item      |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order |
      | 1010           | 1011          | 1           |
      | 1011           | 1012          | 1           |
      | 1012           | 1020          | 1           |
    And the database has the following table "items_ancestors":
      | ancestor_item_id | child_item_id |
      | 1010             | 1011          |
      | 1010             | 1012          |
      | 1010             | 1020          |
      | 1011             | 1012          |
      | 1011             | 1020          |
      | 1012             | 1020          |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated       |
      | 1000     | 1010    | content_with_descendants |
      | 1000     | 1011    | content_with_descendants |
      | 1000     | 1012    | content_with_descendants |
      | 1000     | 1020    | content                  |
      And the database has the following table "attempts":
      | id | participant_id | root_item_id | parent_attempt_id |
      | 0  | 1000           | null         | null              |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | started_at          |
      | 0          | 1000           | 1010    | 2020-01-01 00:00:00 |
    And I am the user with id "1000"
    When I send a GET request to "/items/1020/breadcrumbs-from-roots"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      [
        {
          "started_by_participant": false,
          "path": [
            {"id": "1010", "title": "Chapter 1", "language_tag": "en", "type": "Chapter"},
            {"id": "1011", "title": "Chapter 2", "language_tag": "en", "type": "Chapter"},
            {"id": "1012", "title": "Chapter 3", "language_tag": "en", "type": "Chapter"},
            {"id": "1020", "title": "Item", "language_tag": "en", "type": "Task"}
          ]
        }
      ]
      """

  Scenario Outline: Find breadcrumbs for a team
    Given I am the user with id "111"
    When I send a GET request to "<service_url>?participant_id=102"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      [
        {
          "started_by_participant": true,
          "path": [
            {"id": "60", "title": "Reduce Graph", "language_tag": "en", "type": "Task"},
            {"id": "70","title": null, "language_tag": "fr", "type": "Task"}
          ]
        },
        {
          "started_by_participant": true,
          "path": [
            {"id": "10", "title": "Graphe: Methodes", "language_tag": "fr", "type": "Chapter"},
            {"id": "60", "title": "Reduce Graph", "language_tag": "en", "type": "Task"},
            {"id": "70", "title": null, "language_tag": "fr", "type": "Task"}
          ]
        }
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
        {
          "started_by_participant": true,
          "path": [
            {"id": "60", "title": "Reduce Graph", "language_tag": "en", "type": "Task"}
          ]
        },
        {
          "started_by_participant": true,
          "path": [
            {"id": "10", "title": "Graphe: Methodes", "language_tag": "fr", "type": "Chapter"},
            {"id": "60", "title": "Reduce Graph", "language_tag": "en", "type": "Task"}
          ]
        }
      ]
      """
    Examples:
      | service_url                                   |
      | /items/60/breadcrumbs-from-roots              |
      | /items/by-text-id/id60/breadcrumbs-from-roots |

  Scenario Outline: Should return not started paths
    Given I am the user with id "111"
    When I send a GET request to "<service_url>?participant_id=102"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      [
        {
          "started_by_participant": false,
          "path": [
            {"id": "100", "title": "Chapter Containing Explicit Entry Not Started", "language_tag": "en", "type": "Chapter"},
            {"id": "101", "title": "Explicit Entry Not Started", "language_tag": "en", "type": "Task"}
          ]
        }
      ]
      """
    Examples:
      | service_url                                    |
      | /items/101/breadcrumbs-from-roots              |
      | /items/by-text-id/id101/breadcrumbs-from-roots |
