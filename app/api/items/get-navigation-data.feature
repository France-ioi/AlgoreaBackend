Feature: Get item for tree navigation
  Background:
    Given the database has the following table 'users':
      | ID | sLogin    | tempUser | idGroupSelf | idGroupOwned | sDefaultLanguage | iVersion |
      | 1  | jdoe      | 0        | 11          | 12           |                  | 0        |
      | 2  | gray_root | 0        | 14          | 15           |                  | 0        |
      | 3  | gray_mid  | 0        | 16          | 17           |                  | 0        |
      | 4  | fr_user   | 0        | 18          | 19           | fr               | 0        |
    And the database has the following table 'languages':
      | ID | sCode | iVersion |
      | 2  | fr    | 0        |
    And the database has the following table 'groups':
      | ID | sName           | sTextId | iGrade | sType     | iVersion |
      | 11 | jdoe            |         | -2     | UserAdmin | 0        |
      | 12 | jdoe-admin      |         | -2     | UserAdmin | 0        |
      | 13 | Group B         |         | -2     | Class     | 0        |
      | 14 | gray_root       |         | -2     | UserAdmin | 0        |
      | 15 | gray_root-admin |         | -2     | UserAdmin | 0        |
      | 16 | gray_mid        |         | -2     | UserAdmin | 0        |
      | 17 | gray_mid-admin  |         | -2     | UserAdmin | 0        |
      | 18 | french          |         | -2     | UserAdmin | 0        |
      | 19 | french-admin    |         | -2     | UserAdmin | 0        |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | iVersion |
      | 61 | 13            | 11           | 0        |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 71 | 11              | 11           | 1       | 0        |
      | 72 | 12              | 12           | 1       | 0        |
      | 73 | 13              | 13           | 1       | 0        |
      | 74 | 13              | 11           | 0       | 0        |
      | 75 | 14              | 14           | 1       | 0        |
      | 76 | 16              | 16           | 1       | 0        |
      | 77 | 18              | 18           | 1       | 0        |
    And the database has the following table 'items':
      | ID  | sType    | bTeamsEditable | bNoScore | idItemUnlocked | bTransparentFolder | iVersion |
      | 200 | Category | false          | false    | 1234,2345      | true               | 0        |
      | 210 | Chapter  | false          | false    | 1234,2345      | true               | 0        |
      | 220 | Chapter  | false          | false    | 1234,2345      | true               | 0        |
      | 230 | Chapter  | false          | false    | 1234,2345      | true               | 0        |
      | 211 | Task     | false          | false    | 1234,2345      | true               | 0        |
      | 231 | Task     | false          | false    | 1234,2345      | true               | 0        |
      | 232 | Task     | false          | false    | 1234,2345      | true               | 0        |
      | 250 | Task     | false          | false    | 1234,2345      | true               | 0        |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
      | 43 | 13      | 200    | null            | true              | true                 | true                | 0             | 0        |
      | 44 | 13      | 210    | null            | true              | true                 | true                | 0             | 0        |
      | 45 | 13      | 220    | null            | true              | true                 | true                | 0             | 0        |
      | 46 | 13      | 230    | null            | true              | true                 | true                | 0             | 0        |
      | 47 | 13      | 211    | null            | true              | true                 | true                | 0             | 0        |
      | 48 | 13      | 231    | null            | true              | true                 | true                | 0             | 0        |
      | 49 | 13      | 232    | null            | true              | true                 | true                | 0             | 0        |
      | 42 | 13      | 250    | null            | true              | true                 | true                | 0             | 0        |
      | 50 | 14      | 200    | null            | false             | false                | true                | 0             | 0        |
      | 51 | 14      | 210    | null            | true              | true                 | true                | 0             | 0        |
      | 52 | 14      | 220    | null            | true              | true                 | true                | 0             | 0        |
      | 53 | 14      | 230    | null            | true              | true                 | true                | 0             | 0        |
      | 54 | 14      | 211    | null            | true              | true                 | true                | 0             | 0        |
      | 55 | 14      | 231    | null            | true              | true                 | true                | 0             | 0        |
      | 56 | 14      | 232    | null            | true              | true                 | true                | 0             | 0        |
      | 60 | 16      | 200    | null            | true              | true                 | true                | 0             | 0        |
      | 61 | 16      | 210    | null            | false             | false                | true                | 0             | 0        |
      | 62 | 16      | 220    | null            | true              | true                 | true                | 0             | 0        |
      | 63 | 16      | 230    | null            | false             | false                | true                | 0             | 0        |
      | 64 | 16      | 211    | null            | true              | true                 | true                | 0             | 0        |
      | 65 | 16      | 231    | null            | true              | true                 | true                | 0             | 0        |
      | 66 | 16      | 232    | null            | true              | true                 | true                | 0             | 0        |
      | 70 | 18      | 200    | null            | false             | true                 | true                | 0             | 0        |
      | 71 | 18      | 210    | null            | false             | true                 | true                | 0             | 0        |
      | 72 | 18      | 220    | null            | false             | true                 | true                | 0             | 0        |
      | 73 | 18      | 230    | null            | false             | true                 | true                | 0             | 0        |
      | 74 | 18      | 211    | null            | false             | true                 | true                | 0             | 0        |
      | 75 | 18      | 231    | null            | false             | true                 | true                | 0             | 0        |
      | 76 | 18      | 232    | null            | false             | true                 | true                | 0             | 0        |
    And the database has the following table 'items_items':
      | ID | idItemParent | idItemChild | iChildOrder | bAccessRestricted | iDifficulty | iVersion |
      | 54 | 200          | 210         | 3           | true              | 0           | 0        |
      | 55 | 200          | 220         | 2           | true              | 0           | 0        |
      | 56 | 200          | 230         | 1           | true              | 0           | 0        |
      | 57 | 210          | 211         | 1           | true              | 0           | 0        |
      | 58 | 230          | 231         | 2           | true              | 0           | 0        |
      | 59 | 230          | 232         | 1           | true              | 0           | 0        |
    And the database has the following table 'items_strings':
      | ID | idItem | idLanguage | sTitle      | iVersion |
      | 53 | 200    | 1          | Category 1  | 0        |
      | 54 | 210    | 1          | Chapter A   | 0        |
      | 55 | 220    | 1          | Chapter B   | 0        |
      | 56 | 230    | 1          | Chapter C   | 0        |
      | 57 | 211    | 1          | Task 1      | 0        |
      | 58 | 231    | 1          | Task 2      | 0        |
      | 59 | 232    | 1          | Task 3      | 0        |
      | 63 | 200    | 2          | Catégorie 1 | 0        |
      | 64 | 210    | 2          | Chapitre A  | 0        |
      | 66 | 230    | 2          | Chapitre C  | 0        |
      | 67 | 211    | 2          | Tâche 1     | 0        |
    And the database has the following table 'users_items':
      | ID | idUser | idItem | iScore | nbSubmissionsAttempts | bValidated  | bFinished | bKeyObtained | sStartDate           | sFinishDate          | sValidationDate      | iVersion |
      | 1  | 1      | 200    | 12341  | 11                    | true        | true      | true         | 2019-01-30T09:26:41Z | 2019-02-01T09:26:41Z | 2019-01-31T09:26:41Z | 0        |
      | 2  | 1      | 210    | 12342  | 12                    | true        | true      | true         | 2019-01-30T09:26:42Z | 2019-02-01T09:26:42Z | 2019-01-31T09:26:42Z | 0        |
      | 5  | 1      | 211    | 12343  | 13                    | true        | true      | true         | 2019-01-30T09:26:43Z | 2019-02-01T09:26:43Z | 2019-01-31T09:26:43Z | 0        |
      | 3  | 1      | 220    | 12344  | 14                    | true        | true      | true         | 2019-01-30T09:26:44Z | 2019-02-01T09:26:44Z | 2019-01-31T09:26:44Z | 0        |
      | 4  | 1      | 230    | 12345  | 15                    | true        | true      | true         | 2019-01-30T09:26:45Z | 2019-02-01T09:26:45Z | 2019-01-31T09:26:45Z | 0        |
      | 6  | 1      | 231    | 12346  | 16                    | true        | true      | true         | 2019-01-30T09:26:46Z | 2019-02-01T09:26:46Z | 2019-01-31T09:26:46Z | 0        |
      | 7  | 1      | 232    | 12347  | 17                    | true        | true      | true         | 2019-01-30T09:26:47Z | 2019-02-01T09:26:47Z | 2019-01-31T09:26:47Z | 0        |

  Scenario: Get tree structure
    Given I am the user with ID "1"
    When I send a GET request to "/items/nav-tree/200"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": 200,
        "type": "Category",
        "transparent_folder": true,
        "has_unlocked_items": true,
        "string": {
          "title": "Category 1"
        },
        "user": {
          "score": 12341,
          "validated": true,
          "finished": true,
          "key_obtained": true,
          "submissions_attempts": 11,
          "start_date": "2019-01-30T09:26:41Z",
          "validation_date": "2019-01-31T09:26:41Z",
          "finish_date": "2019-02-01T09:26:41Z"
        },
        "access_rights": {
          "partial_access": true,
          "full_access": true,
          "gray_access": true
        },
        "children": [
          {
            "id": 230,
            "order": 1,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "string": {
              "title": "Chapter C"
            },
            "user": {
              "score": 12345,
              "validated": true,
              "finished": true,
              "key_obtained": true,
              "submissions_attempts": 15,
              "start_date": "2019-01-30T09:26:45Z",
              "validation_date": "2019-01-31T09:26:45Z",
              "finish_date": "2019-02-01T09:26:45Z"
            },
            "access_rights": {
              "partial_access": true,
              "full_access": true,
              "gray_access": true
            },
            "children": [
              {
                "id": 232,
                "order": 1,
                "access_restricted": true,
                "type": "Task",
                "transparent_folder": true,
                "has_unlocked_items": true,
                "string": {
                  "title": "Task 3"
                },
                "user": {
                  "score": 12347,
                  "validated": true,
                  "finished": true,
                  "key_obtained": true,
                  "submissions_attempts": 17,
                  "start_date": "2019-01-30T09:26:47Z",
                  "validation_date": "2019-01-31T09:26:47Z",
                  "finish_date": "2019-02-01T09:26:47Z"
                },
                "access_rights": {
                  "partial_access": true,
                  "full_access": true,
                  "gray_access": true
                }
              },
              {
                "id": 231,
                "order": 2,
                "access_restricted": true,
                "type": "Task",
                "transparent_folder": true,
                "has_unlocked_items": true,
                "string": {
                  "title": "Task 2"
                },
                "user": {
                  "score": 12346,
                  "validated": true,
                  "finished": true,
                  "key_obtained": true,
                  "submissions_attempts": 16,
                  "start_date": "2019-01-30T09:26:46Z",
                  "validation_date": "2019-01-31T09:26:46Z",
                  "finish_date": "2019-02-01T09:26:46Z"
                },
                "access_rights": {
                  "partial_access": true,
                  "full_access": true,
                  "gray_access": true
                }
              }
            ]
          },
          {
            "id": 220,
            "order": 2,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "string": {
              "title": "Chapter B"
            },
            "user": {
              "score": 12344,
              "validated": true,
              "finished": true,
              "key_obtained": true,
              "submissions_attempts": 14,
              "start_date": "2019-01-30T09:26:44Z",
              "validation_date": "2019-01-31T09:26:44Z",
              "finish_date": "2019-02-01T09:26:44Z"
            },
            "access_rights": {
              "partial_access": true,
              "full_access": true,
              "gray_access": true
            }
          },
          {
            "id": 210,
            "order": 3,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "string": {
              "title": "Chapter A"
            },
            "user": {
              "score": 12342,
              "validated": true,
              "finished": true,
              "key_obtained": true,
              "submissions_attempts": 12,
              "start_date": "2019-01-30T09:26:42Z",
              "validation_date": "2019-01-31T09:26:42Z",
              "finish_date": "2019-02-01T09:26:42Z"
            },
            "access_rights": {
              "partial_access": true,
              "full_access": true,
              "gray_access": true
            },
            "children": [
              {
                "id": 211,
                "order": 1,
                "access_restricted": true,
                "type": "Task",
                "transparent_folder": true,
                "has_unlocked_items": true,
                "string": {
                  "title": "Task 1"
                },
                "user": {
                  "score": 12343,
                  "validated": true,
                  "finished": true,
                  "key_obtained": true,
                  "submissions_attempts": 13,
                  "start_date": "2019-01-30T09:26:43Z",
                  "validation_date": "2019-01-31T09:26:43Z",
                  "finish_date": "2019-02-01T09:26:43Z"
                },
                "access_rights": {
                  "partial_access": true,
                  "full_access": true,
                  "gray_access": true
                }
              }
            ]
          }
        ]
      }
      """

  Scenario: Should return only one node if the root item doesn't have children
    Given I am the user with ID "1"
    When I send a GET request to "/items/nav-tree/232"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": 232,
        "type": "Task",
        "transparent_folder": true,
        "has_unlocked_items": true,
        "string": {
          "title": "Task 3"
        },
        "user": {
          "score": 12347,
          "validated": true,
          "finished": true,
          "key_obtained": true,
          "submissions_attempts": 17,
          "start_date": "2019-01-30T09:26:47Z",
          "validation_date": "2019-01-31T09:26:47Z",
          "finish_date": "2019-02-01T09:26:47Z"
        },
        "access_rights": {
          "partial_access": true,
          "full_access": true,
          "gray_access": true
        }
      }
      """

  Scenario: Should return a subtree having two levels if the root item doesn't have grandchildren
    Given I am the user with ID "1"
    When I send a GET request to "/items/nav-tree/230"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": 230,
        "type": "Chapter",
        "transparent_folder": true,
        "has_unlocked_items": true,
        "string": {
          "title": "Chapter C"
        },
        "user": {
          "score": 12345,
          "validated": true,
          "finished": true,
          "key_obtained": true,
          "submissions_attempts": 15,
          "start_date": "2019-01-30T09:26:45Z",
          "validation_date": "2019-01-31T09:26:45Z",
          "finish_date": "2019-02-01T09:26:45Z"
        },
        "access_rights": {
          "partial_access": true,
          "full_access": true,
          "gray_access": true
        },
        "children": [
          {
            "id": 232,
            "order": 1,
            "access_restricted": true,
            "type": "Task",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "string": {
              "title": "Task 3"
            },
            "user": {
              "score": 12347,
              "validated": true,
              "finished": true,
              "key_obtained": true,
              "submissions_attempts": 17,
              "start_date": "2019-01-30T09:26:47Z",
              "validation_date": "2019-01-31T09:26:47Z",
              "finish_date": "2019-02-01T09:26:47Z"
            },
            "access_rights": {
              "partial_access": true,
              "full_access": true,
              "gray_access": true
            }
          },
          {
            "id": 231,
            "order": 2,
            "access_restricted": true,
            "type": "Task",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "string": {
              "title": "Task 2"
            },
            "user": {
              "score": 12346,
              "validated": true,
              "finished": true,
              "key_obtained": true,
              "submissions_attempts": 16,
              "start_date": "2019-01-30T09:26:46Z",
              "validation_date": "2019-01-31T09:26:46Z",
              "finish_date": "2019-02-01T09:26:46Z"
            },
            "access_rights": {
              "partial_access": true,
              "full_access": true,
              "gray_access": true
            }
          }
        ]
      }
      """

  Scenario: Should return only one node if the user has only grayed access to the root item
    Given I am the user with ID "2"
    When I send a GET request to "/items/nav-tree/200"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": 200,
        "type": "Category",
        "transparent_folder": true,
        "has_unlocked_items": true,
        "string": {
          "title": "Category 1"
        },
        "user": {
          "finish_date": "",
          "finished": false,
          "key_obtained": false,
          "score": 0,
          "start_date": "",
          "submissions_attempts": 0,
          "validated": false,
          "validation_date": ""
        },
        "access_rights": {
          "partial_access": false,
          "full_access": false,
          "gray_access": true
        }
      }
      """

  Scenario: Should skip children of grayed nodes
    Given I am the user with ID "3"
    When I send a GET request to "/items/nav-tree/200"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": 200,
        "type": "Category",
        "transparent_folder": true,
        "has_unlocked_items": true,
        "string": {
          "title": "Category 1"
        },
        "user": {
          "finish_date": "",
          "finished": false,
          "key_obtained": false,
          "score": 0,
          "start_date": "",
          "submissions_attempts": 0,
          "validated": false,
          "validation_date": ""
        },
        "access_rights": {
          "partial_access": true,
          "full_access": true,
          "gray_access": true
        },
        "children": [
          {
            "id": 230,
            "order": 1,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "string": {
              "title": "Chapter C"
            },
            "user": {
              "finish_date": "",
              "finished": false,
              "key_obtained": false,
              "score": 0,
              "start_date": "",
              "submissions_attempts": 0,
              "validated": false,
              "validation_date": ""
            },
            "access_rights": {
              "partial_access": false,
              "full_access": false,
              "gray_access": true
            }
          },
          {
            "id": 220,
            "order": 2,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "string": {
              "title": "Chapter B"
            },
            "user": {
              "finish_date": "",
              "finished": false,
              "key_obtained": false,
              "score": 0,
              "start_date": "",
              "submissions_attempts": 0,
              "validated": false,
              "validation_date": ""
            },
            "access_rights": {
              "partial_access": true,
              "full_access": true,
              "gray_access": true
            }
          },
          {
            "id": 210,
            "order": 3,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "string": {
              "title": "Chapter A"
            },
            "user": {
              "finish_date": "",
              "finished": false,
              "key_obtained": false,
              "score": 0,
              "start_date": "",
              "submissions_attempts": 0,
              "validated": false,
              "validation_date": ""
            },
            "access_rights": {
              "partial_access": false,
              "full_access": false,
              "gray_access": true
            }
          }
        ]
      }
      """

  Scenario: Should prefer the user's default language for titles
    Given I am the user with ID "4"
    When I send a GET request to "/items/nav-tree/200"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "id": 200,
        "type": "Category",
        "transparent_folder": true,
        "has_unlocked_items": true,
        "string": {
          "title": "Catégorie 1"
        },
        "user": {
          "finish_date": "",
          "finished": false,
          "key_obtained": false,
          "score": 0,
          "start_date": "",
          "submissions_attempts": 0,
          "validated": false,
          "validation_date": ""
        },
        "access_rights": {
          "partial_access": true,
          "full_access": false,
          "gray_access": true
        },
        "children": [
          {
            "id": 230,
            "order": 1,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "string": {
              "title": "Chapitre C"
            },
            "user": {
              "finish_date": "",
              "finished": false,
              "key_obtained": false,
              "score": 0,
              "start_date": "",
              "submissions_attempts": 0,
              "validated": false,
              "validation_date": ""
            },
            "access_rights": {
              "partial_access": true,
              "full_access": false,
              "gray_access": true
            },
            "children": [
              {
                "id": 232,
                "order": 1,
                "access_restricted": true,
                "type": "Task",
                "transparent_folder": true,
                "has_unlocked_items": true,
                "string": {
                  "title": "Task 3"
                },
                "user": {
                  "finish_date": "",
                  "finished": false,
                  "key_obtained": false,
                  "score": 0,
                  "start_date": "",
                  "submissions_attempts": 0,
                  "validated": false,
                  "validation_date": ""
                },
                "access_rights": {
                  "partial_access": true,
                  "full_access": false,
                  "gray_access": true
                }
              },
              {
                "id": 231,
                "order": 2,
                "access_restricted": true,
                "type": "Task",
                "transparent_folder": true,
                "has_unlocked_items": true,
                "string": {
                  "title": "Task 2"
                },
                "user": {
                  "finish_date": "",
                  "finished": false,
                  "key_obtained": false,
                  "score": 0,
                  "start_date": "",
                  "submissions_attempts": 0,
                  "validated": false,
                  "validation_date": ""
                },
                "access_rights": {
                  "partial_access": true,
                  "full_access": false,
                  "gray_access": true
                }
              }
            ]
          },
          {
            "id": 220,
            "order": 2,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "string": {
              "title": "Chapter B"
            },
            "user": {
              "finish_date": "",
              "finished": false,
              "key_obtained": false,
              "score": 0,
              "start_date": "",
              "submissions_attempts": 0,
              "validated": false,
              "validation_date": ""
            },
            "access_rights": {
              "partial_access": true,
              "full_access": false,
              "gray_access": true
            }
          },
          {
            "id": 210,
            "order": 3,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "string": {
              "title": "Chapitre A"
            },
            "user": {
              "finish_date": "",
              "finished": false,
              "key_obtained": false,
              "score": 0,
              "start_date": "",
              "submissions_attempts": 0,
              "validated": false,
              "validation_date": ""
            },
            "access_rights": {
              "partial_access": true,
              "full_access": false,
              "gray_access": true
            },
            "children": [
              {
                "id": 211,
                "order": 1,
                "access_restricted": true,
                "type": "Task",
                "transparent_folder": true,
                "has_unlocked_items": true,
                "string": {
                  "title": "Tâche 1"
                },
                "user": {
                  "finish_date": "",
                  "finished": false,
                  "key_obtained": false,
                  "score": 0,
                  "start_date": "",
                  "submissions_attempts": 0,
                  "validated": false,
                  "validation_date": ""
                },
                "access_rights": {
                  "partial_access": true,
                  "full_access": false,
                  "gray_access": true
                }
              }
            ]
          }
        ]
      }
      """
