Feature: Find all breadcrumbs to an item - robustness
  Background:
    Given the database has the following table "groups":
      | id  | type  | root_activity_id | root_skill_id |
      | 90  | Class | 10               | null          |
      | 91  | Other | 50               | null          |
      | 102 | Team  | 60               | null          |
    And the database has the following users:
      | group_id | login | default_language |
      | 101      | john  | en               |
      | 111      | jane  | fr               |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 90              | 102            |
      | 91              | 111            |
      | 102             | 101            |
    And the groups ancestors are computed
    And the database has the following table "group_managers":
      | manager_id | group_id | can_watch_members |
      | 91         | 90       | true              |
      | 111        | 111      | false             |
    And the database has the following table "items":
      | id | url                    | type    | default_language_tag | requires_explicit_entry | text_id |
      | 10 | null                   | Chapter | en                   | false                   | id10    |
      | 60 | http://taskplatform/60 | Task    | en                   | false                   | id60    |
      | 70 | http://taskplatform/70 | Task    | fr                   | false                   | id70    |
    And the database has the following table "items_strings":
      | item_id | language_tag | title            |
      | 10      | fr           | Graphe: Methodes |
      | 10      | en           | Graph: Methods   |
      | 60      | en           | Reduce Graph     |
      | 70      | fr           | null             |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order |
      | 10             | 60            | 1           |
      | 60             | 70            | 2           |
    And the database has the following table "items_ancestors":
      | ancestor_item_id | child_item_id |
      | 10               | 60            |
      | 10               | 70            |
      | 60               | 70            |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated |
      | 102      | 60      | info               |
      | 111      | 10      | info               |
      | 111      | 60      | none               |
      | 111      | 70      | none               |
    And the database has the following table "attempts":
      | id | participant_id | root_item_id | parent_attempt_id |
      | 0  | 101            | null         | null              |
      | 0  | 102            | null         | null              |
      | 0  | 111            | null         | null              |
      | 1  | 102            | 10           | null              |
      | 2  | 102            | 10           | null              |
      | 3  | 102            | 10           | null              |
      | 4  | 102            | 10           | null              |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | started_at          |
      | 1          | 102            | 10      | 2020-01-01 00:00:00 |
      | 2          | 102            | 10      | 2020-01-01 00:00:00 |
      | 2          | 102            | 60      | 2020-01-01 00:00:00 |
      | 3          | 102            | 10      | 2020-01-01 00:00:00 |
      | 3          | 102            | 60      | 2020-01-01 00:00:00 |
      | 3          | 102            | 70      | 2020-01-01 00:00:00 |
      | 0          | 111            | 10      | 2020-01-01 00:00:00 |

  Scenario: Invalid item_id
    And I am the user with id "111"
    When I send a GET request to "/items/100000000000000000000000/breadcrumbs-from-roots"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: text_id not found
    And I am the user with id "111"
    When I send a GET request to "/items/by-text-id/abc/breadcrumbs-from-roots"
    Then the response code should be 400
    And the response error message should contain "No item found with text_id"

  Scenario Outline: Invalid participant_id
    And I am the user with id "111"
    When I send a GET request to "<service_url>?participant_id=abc"
    Then the response code should be 400
    And the response error message should contain "Wrong value for participant_id (should be int64)"
    Examples:
      | service_url                                   |
      | /items/10/breadcrumbs-from-roots              |
      | /items/by-text-id/id10/breadcrumbs-from-roots |

  Scenario Outline: No access to participant_id
    Given I am the user with id "111"
    When I send a GET request to "<service_url>?participant_id=<participant_id>"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    Examples:
      | service_url                                   | participant_id |
      | /items/10/breadcrumbs-from-roots              | 404            |
      | /items/by-text-id/id10/breadcrumbs-from-roots | 404            |
      | /items/10/breadcrumbs-from-roots              | 111            |
      | /items/by-text-id/id10/breadcrumbs-from-roots | 111            |

  Scenario Outline: No paths
    Given I am the user with id "111"
    When I send a GET request to "<service_url>?participant_id=102"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
  Examples:
    | service_url                                   |
    | /items/70/breadcrumbs-from-roots              |
    | /items/by-text-id/id70/breadcrumbs-from-roots |
    | /items/60/breadcrumbs-from-roots              |
    | /items/by-text-id/id60/breadcrumbs-from-roots |
