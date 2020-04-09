Feature: Get item for tree navigation
  Background:
    Given the database has the following table 'groups':
      | id | name      | text_id | grade | type |
      | 11 | jdoe      |         | -2    | User |
      | 13 | Group B   |         | -2    | Team |
      | 14 | info_root |         | -2    | User |
      | 16 | info_mid  |         | -2    | User |
      | 18 | french    |         | -2    | User |
      | 19 | Group C   |         | -2    | Team |
    And the database has the following table 'languages':
      | tag |
      | fr  |
    And the database has the following table 'users':
      | login     | temp_user | group_id | default_language |
      | jdoe      | 0         | 11       |                  |
      | info_root | 0         | 14       |                  |
      | info_mid  | 0         | 16       |                  |
      | fr_user   | 0         | 18       | fr               |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 11             |
      | 19              | 11             |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id |
      | 11                | 11             |
      | 13                | 13             |
      | 13                | 11             |
      | 14                | 14             |
      | 16                | 16             |
      | 18                | 18             |
      | 19                | 11             |
      | 19                | 19             |
    And the database has the following table 'items':
      | id  | type    | default_language_tag | teams_editable | no_score |
      | 200 | Course  | en                   | false          | false    |
      | 210 | Chapter | en                   | false          | false    |
      | 220 | Chapter | en                   | false          | false    |
      | 230 | Chapter | en                   | false          | false    |
      | 211 | Task    | en                   | false          | false    |
      | 231 | Task    | en                   | false          | false    |
      | 232 | Task    | en                   | false          | false    |
      | 250 | Task    | en                   | false          | false    |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 11       | 200     | solution                 |
      | 11       | 210     | content                  |
      | 11       | 230     | info                     |
      | 13       | 200     | content_with_descendants |
      | 13       | 210     | content_with_descendants |
      | 13       | 220     | content_with_descendants |
      | 13       | 230     | content_with_descendants |
      | 13       | 211     | content_with_descendants |
      | 13       | 231     | content_with_descendants |
      | 13       | 232     | content_with_descendants |
      | 13       | 250     | content_with_descendants |
      | 14       | 200     | info                     |
      | 14       | 210     | content_with_descendants |
      | 14       | 220     | content_with_descendants |
      | 14       | 230     | content_with_descendants |
      | 14       | 211     | content_with_descendants |
      | 14       | 231     | content_with_descendants |
      | 14       | 232     | content_with_descendants |
      | 16       | 200     | content_with_descendants |
      | 16       | 210     | info                     |
      | 16       | 220     | content_with_descendants |
      | 16       | 230     | info                     |
      | 16       | 211     | content_with_descendants |
      | 16       | 231     | content_with_descendants |
      | 16       | 232     | content_with_descendants |
      | 18       | 200     | content                  |
      | 18       | 210     | content                  |
      | 18       | 220     | content                  |
      | 18       | 230     | content                  |
      | 18       | 211     | content                  |
      | 18       | 231     | content                  |
      | 18       | 232     | content                  |
      | 19       | 200     | content                  |
      | 19       | 210     | content                  |
      | 19       | 211     | content                  |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order | content_view_propagation |
      | 200            | 210           | 3           | none                     |
      | 200            | 220           | 2           | as_info                  |
      | 200            | 230           | 1           | as_content               |
      | 210            | 211           | 1           | none                     |
      | 230            | 231           | 2           | none                     |
      | 230            | 232           | 1           | none                     |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title       |
      | 200     | en           | Category 1  |
      | 210     | en           | Chapter A   |
      | 220     | en           | Chapter B   |
      | 230     | en           | Chapter C   |
      | 211     | en           | Task 1      |
      | 231     | en           | Task 2      |
      | 232     | en           | Task 3      |
      | 200     | fr           | Catégorie 1 |
      | 210     | fr           | Chapitre A  |
      | 230     | fr           | Chapitre C  |
      | 211     | fr           | Tâche 1     |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          |
      | 0  | 11             | 2019-01-30 08:26:41 |
      | 1  | 11             | 2019-01-30 08:26:41 |
      | 0  | 13             | 2018-01-30 09:26:42 |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | score_computed | submissions | started_at          | validated_at        |
      | 0          | 11             | 200     | 91             | 11          | 2019-01-30 09:26:41 | null                |
      | 0          | 11             | 210     | 92             | 12          | 2019-01-30 09:26:42 | 2019-01-31 09:26:42 |
      | 0          | 11             | 211     | 93             | 13          | 2019-01-30 09:26:43 | null                |
      | 0          | 11             | 220     | 94             | 14          | 2019-01-30 09:26:44 | 2019-01-31 09:26:44 |
      | 0          | 11             | 230     | 95             | 15          | 2019-01-30 09:26:45 | 2019-01-31 09:26:45 |
      | 0          | 11             | 231     | 96             | 16          | 2019-01-30 09:26:46 | 2019-01-31 09:26:46 |
      | 0          | 11             | 232     | 97             | 17          | 2019-01-30 09:26:47 | 2019-01-31 09:26:47 |
      | 1          | 11             | 200     | 90             | 11          | 2019-01-30 09:26:41 | 2019-01-31 09:26:41 |
      | 0          | 13             | 200     | 45             | 2           | 2018-01-30 09:26:42 | null                |
      | 0          | 13             | 210     | 56             | 5           | 2018-01-30 09:26:42 | 2018-01-31 09:26:42 |
      | 0          | 13             | 230     | 78             | 4           | 2018-01-30 09:26:42 | 2018-01-31 09:26:42 |

  Scenario: Get tree structure
    Given I am the user with id "11"
    When I send a GET request to "/items/200/nav-tree"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Course",
        "string": {
          "title": "Category 1"
        },
        "best_score": 91,
        "validated": true,
        "access_rights": {
          "can_view": "solution"
        },
        "children": [
          {
            "id": "230",
            "order": 1,
            "content_view_propagation": "as_content",
            "type": "Chapter",
            "string": {
              "title": "Chapter C"
            },
            "best_score": 95,
            "validated": true,
            "access_rights": {
              "can_view": "content_with_descendants"
            },
            "children": [
              {
                "id": "232",
                "order": 1,
                "content_view_propagation": "none",
                "type": "Task",
                "string": {
                  "title": "Task 3"
                },
                "best_score": 97,
                "validated": true,
                "access_rights": {
                  "can_view": "content_with_descendants"
                },
                "children": null
              },
              {
                "id": "231",
                "order": 2,
                "content_view_propagation": "none",
                "type": "Task",
                "string": {
                  "title": "Task 2"
                },
                "best_score": 96,
                "validated": true,
                "access_rights": {
                  "can_view": "content_with_descendants"
                },
                "children": null
              }
            ]
          },
          {
            "id": "220",
            "order": 2,
            "content_view_propagation": "as_info",
            "type": "Chapter",
            "string": {
              "title": "Chapter B"
            },
            "best_score": 94,
            "validated": true,
            "access_rights": {
              "can_view": "content_with_descendants"
            },
            "children": []
          },
          {
            "id": "210",
            "order": 3,
            "content_view_propagation": "none",
            "type": "Chapter",
            "string": {
              "title": "Chapter A"
            },
            "best_score": 92,
            "validated": true,
            "access_rights": {
              "can_view": "content_with_descendants"
            },
            "children": [
              {
                "id": "211",
                "order": 1,
                "content_view_propagation": "none",
                "type": "Task",
                "string": {
                  "title": "Task 1"
                },
                "best_score": 93,
                "validated": false,
                "access_rights": {
                  "can_view": "content_with_descendants"
                },
                "children": null
              }
            ]
          }
        ]
      }
      """

  Scenario: Should return only one node if the root item doesn't have children
    Given I am the user with id "11"
    When I send a GET request to "/items/232/nav-tree"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "232",
        "type": "Task",
        "string": {
          "title": "Task 3"
        },
        "best_score": 97,
        "validated": true,
        "access_rights": {
          "can_view": "content_with_descendants"
        },
        "children": []
      }
      """

  Scenario: Should return a subtree having two levels if the root item doesn't have grandchildren
    Given I am the user with id "11"
    When I send a GET request to "/items/230/nav-tree"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "230",
        "type": "Chapter",
        "string": {
          "title": "Chapter C"
        },
        "best_score": 95,
        "validated": true,
        "access_rights": {
          "can_view": "content_with_descendants"
        },
        "children": [
          {
            "id": "232",
            "order": 1,
            "content_view_propagation": "none",
            "type": "Task",
            "string": {
              "title": "Task 3"
            },
            "best_score": 97,
            "validated": true,
            "access_rights": {
              "can_view": "content_with_descendants"
            },
            "children": []
          },
          {
            "id": "231",
            "order": 2,
            "content_view_propagation": "none",
            "type": "Task",
            "string": {
              "title": "Task 2"
            },
            "best_score": 96,
            "validated": true,
            "access_rights": {
              "can_view": "content_with_descendants"
            },
            "children": []
          }
        ]
      }
      """

  Scenario: Should return only one node if the user has only info access to the root item
    Given I am the user with id "14"
    When I send a GET request to "/items/200/nav-tree"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Course",
        "string": {
          "title": "Category 1"
        },
        "best_score": 0,
        "validated": false,
        "access_rights": {
          "can_view": "info"
        },
        "children": []
      }
      """

  Scenario: Should skip children of nodes with 'info' access
    Given I am the user with id "16"
    When I send a GET request to "/items/200/nav-tree"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Course",
        "string": {
          "title": "Category 1"
        },
        "best_score": 0,
        "validated": false,
        "access_rights": {
          "can_view": "content_with_descendants"
        },
        "children": [
          {
            "id": "230",
            "order": 1,
            "content_view_propagation": "as_content",
            "type": "Chapter",
            "string": {
              "title": "Chapter C"
            },
            "best_score": 0,
            "validated": false,
            "access_rights": {
              "can_view": "info"
            },
            "children": []
          },
          {
            "id": "220",
            "order": 2,
            "content_view_propagation": "as_info",
            "type": "Chapter",
            "string": {
              "title": "Chapter B"
            },
            "best_score": 0,
            "validated": false,
            "access_rights": {
              "can_view": "content_with_descendants"
            },
            "children": []
          },
          {
            "id": "210",
            "order": 3,
            "content_view_propagation": "none",
            "type": "Chapter",
            "string": {
              "title": "Chapter A"
            },
            "best_score": 0,
            "validated": false,
            "access_rights": {
              "can_view": "info"
            },
            "children": []
          }
        ]
      }
      """

  Scenario: Should prefer the user's default language for titles
    Given I am the user with id "18"
    When I send a GET request to "/items/200/nav-tree"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Course",
        "string": {
          "title": "Catégorie 1"
        },
        "best_score": 0,
        "validated": false,
        "access_rights": {
          "can_view": "content"
        },
        "children": [
          {
            "id": "230",
            "order": 1,
            "content_view_propagation": "as_content",
            "type": "Chapter",
            "string": {
              "title": "Chapitre C"
            },
            "best_score": 0,
            "validated": false,
            "access_rights": {
              "can_view": "content"
            },
            "children": [
              {
                "id": "232",
                "order": 1,
                "content_view_propagation": "none",
                "type": "Task",
                "string": {
                  "title": "Task 3"
                },
                "best_score": 0,
                "validated": false,
                "access_rights": {
                  "can_view": "content"
                },
                "children": null
              },
              {
                "id": "231",
                "order": 2,
                "content_view_propagation": "none",
                "type": "Task",
                "string": {
                  "title": "Task 2"
                },
                "best_score": 0,
                "validated": false,
                "access_rights": {
                  "can_view": "content"
                },
                "children": null
              }
            ]
          },
          {
            "id": "220",
            "order": 2,
            "content_view_propagation": "as_info",
            "type": "Chapter",
            "string": {
              "title": "Chapter B"
            },
            "best_score": 0,
            "validated": false,
            "access_rights": {
              "can_view": "content"
            },
            "children": []
          },
          {
            "id": "210",
            "order": 3,
            "content_view_propagation": "none",
            "type": "Chapter",
            "string": {
              "title": "Chapitre A"
            },
            "best_score": 0,
            "validated": false,
            "access_rights": {
              "can_view": "content"
            },
            "children": [
              {
                "id": "211",
                "order": 1,
                "content_view_propagation": "none",
                "type": "Task",
                "string": {
                  "title": "Tâche 1"
                },
                "best_score": 0,
                "validated": false,
                "access_rights": {
                  "can_view": "content"
                },
                "children": null
              }
            ]
          }
        ]
      }
      """

  Scenario: Get tree structure (as team)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/nav-tree?as_team_id=13"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Course",
        "string": {
          "title": "Category 1"
        },
        "best_score": 45,
        "validated": false,
        "access_rights": {
          "can_view": "content_with_descendants"
        },
        "children": [
          {
            "id": "230",
            "order": 1,
            "content_view_propagation": "as_content",
            "type": "Chapter",
            "string": {
              "title": "Chapter C"
            },
            "best_score": 78,
            "validated": true,
            "access_rights": {
              "can_view": "content_with_descendants"
            },
            "children": [
              {
                "id": "232",
                "order": 1,
                "content_view_propagation": "none",
                "type": "Task",
                "string": {
                  "title": "Task 3"
                },
                "best_score": 0,
                "validated": false,
                "access_rights": {
                  "can_view": "content_with_descendants"
                },
                "children": null
              },
              {
                "id": "231",
                "order": 2,
                "content_view_propagation": "none",
                "type": "Task",
                "string": {
                  "title": "Task 2"
                },
                "best_score": 0,
                "validated": false,
                "access_rights": {
                  "can_view": "content_with_descendants"
                },
                "children": null
              }
            ]
          },
          {
            "id": "220",
            "order": 2,
            "content_view_propagation": "as_info",
            "type": "Chapter",
            "string": {
              "title": "Chapter B"
            },
            "best_score": 0,
            "validated": false,
            "access_rights": {
              "can_view": "content_with_descendants"
            },
            "children": []
          },
          {
            "id": "210",
            "order": 3,
            "content_view_propagation": "none",
            "type": "Chapter",
            "string": {
              "title": "Chapter A"
            },
            "best_score": 56,
            "validated": true,
            "access_rights": {
              "can_view": "content_with_descendants"
            },
            "children": [
              {
                "id": "211",
                "order": 1,
                "content_view_propagation": "none",
                "type": "Task",
                "string": {
                  "title": "Task 1"
                },
                "best_score": 0,
                "validated": false,
                "access_rights": {
                  "can_view": "content_with_descendants"
                },
                "children": null
              }
            ]
          }
        ]
      }
      """

  Scenario: Get tree structure (as another team)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/nav-tree?as_team_id=19"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": "200",
        "type": "Course",
        "string": {
          "title": "Category 1"
        },
        "best_score": 0,
        "validated": false,
        "access_rights": {
          "can_view": "content"
        },
        "children": [
          {
            "id": "210",
            "order": 3,
            "content_view_propagation": "none",
            "type": "Chapter",
            "string": {
              "title": "Chapter A"
            },
            "best_score": 0,
            "validated": false,
            "access_rights": {
              "can_view": "content"
            },
            "children": [
              {
                "id": "211",
                "order": 1,
                "content_view_propagation": "none",
                "type": "Task",
                "string": {
                  "title": "Task 1"
                },
                "best_score": 0,
                "validated": false,
                "access_rights": {
                  "can_view": "content"
                },
                "children": null
              }
            ]
          }
        ]
      }
      """
