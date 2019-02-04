@wip
Feature: Get item for tree navigation

  Scenario: Get tree structure
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | iVersion |
      | 1  | jdoe   | 0        | 11          | 12           | 0        |
    And the database has the following table 'groups':
      | ID | sName      | sTextId | iGrade | sType     | iVersion |
      | 11 | jdoe       |         | -2     | UserAdmin | 0        |
      | 12 | jdoe-admin |         | -2     | UserAdmin | 0        |
      | 13 | Group B    |         | -2     | Class     | 0        |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | iVersion |
      | 61 | 13            | 11           | 0        |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 71 | 11              | 11           | 1       | 0        |
      | 72 | 12              | 12           | 1       | 0        |
      | 73 | 13              | 13           | 1       | 0        |
      | 74 | 13              | 11           | 0       | 0        |
    And the database has the following table 'items':
      | ID  | sType    | bTeamsEditable | bNoScore | idItemUnlocked | bTransparentFolder | iVersion |
      | 200 | Category | false          | false    | 1234,2345      | true               | 0        |
      | 210 | Chapter  | false          | false    | 1234,2345      | true               | 0        |
      | 220 | Chapter  | false          | false    | 1234,2345      | true               | 0        |
      | 230 | Chapter  | false          | false    | 1234,2345      | true               | 0        |
      | 211 | Task     | false          | false    | 1234,2345      | true               | 0        |
      | 231 | Task     | false          | false    | 1234,2345      | true               | 0        |
      | 232 | Task     | false          | false    | 1234,2345      | true               | 0        |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
      | 43 | 13      | 200    | null            | true              | true                 | true                | 0             | 0        |
      | 44 | 13      | 210    | null            | true              | true                 | true                | 0             | 0        |
      | 45 | 13      | 220    | null            | true              | true                 | true                | 0             | 0        |
      | 46 | 13      | 230    | null            | true              | true                 | true                | 0             | 0        |
      | 47 | 13      | 211    | null            | true              | true                 | true                | 0             | 0        |
      | 48 | 13      | 231    | null            | true              | true                 | true                | 0             | 0        |
      | 49 | 13      | 232    | null            | true              | true                 | true                | 0             | 0        |
    And the database has the following table 'items_items':
      | ID | idItemParent | idItemChild | iChildOrder | bAccessRestricted | iDifficulty | iVersion |
      | 54 | 200          | 210         | 1           | true              | 0           | 0        |
      | 55 | 200          | 220         | 2           | true              | 0           | 0        |
      | 56 | 200          | 230         | 3           | true              | 0           | 0        |
      | 57 | 210          | 211         | 1           | true              | 0           | 0        |
      | 58 | 230          | 231         | 1           | true              | 0           | 0        |
      | 59 | 230          | 232         | 2           | true              | 0           | 0        |
    And the database has the following table 'items_strings':
      | ID | idItem | idLanguage | sTitle     | iVersion |
      | 53 | 200    | 1          | Category 1 | 0        |
      | 54 | 210    | 1          | Chapter A  | 0        |
      | 55 | 220    | 1          | Chapter B  | 0        |
      | 56 | 230    | 1          | Chapter C  | 0        |
      | 57 | 211    | 1          | Task 1     | 0        |
      | 58 | 231    | 1          | Task 2     | 0        |
      | 59 | 232    | 1          | Task 3     | 0        |
    And the database has the following table 'users_items':
      | ID | idUser | idItem | iScore | nbSubmissionsAttempts | bValidated  | bFinished | bKeyObtained | sStartDate           | sFinishDate          | sValidationDate      | iVersion |
      | 1  | 1      | 200    | 12345  | 10                    | true        | true      | true         | 2019-01-30T09:26:41Z | 2019-02-01T09:26:41Z | 2019-01-31T09:26:41Z | 0        |
      | 2  | 1      | 210    | 12345  | 10                    | true        | true      | true         | 2019-01-30T09:26:41Z | 2019-02-01T09:26:41Z | 2019-01-31T09:26:41Z | 0        |
      | 3  | 1      | 220    | 12345  | 10                    | true        | true      | true         | 2019-01-30T09:26:41Z | 2019-02-01T09:26:41Z | 2019-01-31T09:26:41Z | 0        |
      | 4  | 1      | 230    | 12345  | 10                    | true        | true      | true         | 2019-01-30T09:26:41Z | 2019-02-01T09:26:41Z | 2019-01-31T09:26:41Z | 0        |
      | 5  | 1      | 211    | 12345  | 10                    | true        | true      | true         | 2019-01-30T09:26:41Z | 2019-02-01T09:26:41Z | 2019-01-31T09:26:41Z | 0        |
      | 6  | 1      | 231    | 12345  | 10                    | true        | true      | true         | 2019-01-30T09:26:41Z | 2019-02-01T09:26:41Z | 2019-01-31T09:26:41Z | 0        |
      | 7  | 1      | 232    | 12345  | 10                    | true        | true      | true         | 2019-01-30T09:26:41Z | 2019-02-01T09:26:41Z | 2019-01-31T09:26:41Z | 0        |
    And I am the user with ID "1"
    When I send a GET request to "/items/nav-tree/200"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "item_id": 200,
        "type": "Category",
        "transparent_folder": true,
        "has_unlocked_items": true,
        "title": "Category 1",
        "user_score": 12345,
        "user_validated": true,
        "user_finished": true,
        "key_obtained": true,
        "submissions_attempts": 10,
        "start_date": "2019-01-30T09:26:41Z",
        "validation_date": "2019-01-31T09:26:41Z",
        "finish_date": "2019-02-01T09:26:41Z",
        "partial_access": true,
        "full_access": true,
        "gray_access": true,
        "children": [
          {
            "item_id": 210,
            "order": 1,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "title": "Chapter A",
            "user_score": 12345,
            "user_validated": true,
            "user_finished": true,
            "key_obtained": true,
            "submissions_attempts": 10,
            "start_date": "2019-01-30T09:26:41Z",
            "validation_date": "2019-01-31T09:26:41Z",
            "finish_date": "2019-02-01T09:26:41Z",
            "partial_access": true,
            "full_access": true,
            "gray_access": true,
            "children": [
              {
                "item_id": 211,
                "order": 1,
                "access_restricted": true,
                "type": "Task",
                "transparent_folder": true,
                "has_unlocked_items": true,
                "title": "Task 1",
                "user_score": 12345,
                "user_validated": true,
                "user_finished": true,
                "key_obtained": true,
                "submissions_attempts": 10,
                "start_date": "2019-01-30T09:26:41Z",
                "validation_date": "2019-01-31T09:26:41Z",
                "finish_date": "2019-02-01T09:26:41Z",
                "partial_access": true,
                "full_access": true,
                "gray_access": true
              }
            ]
          },
          {
            "item_id": 220,
            "order": 2,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "title": "Chapter B",
            "user_score": 12345,
            "user_validated": true,
            "user_finished": true,
            "key_obtained": true,
            "submissions_attempts": 10,
            "start_date": "2019-01-30T09:26:41Z",
            "validation_date": "2019-01-31T09:26:41Z",
            "finish_date": "2019-02-01T09:26:41Z",
            "partial_access": true,
            "full_access": true,
            "gray_access": true
          },
          {
            "item_id": 230,
            "order": 3,
            "access_restricted": true,
            "type": "Chapter",
            "transparent_folder": true,
            "has_unlocked_items": true,
            "title": "Chapter C",
            "user_score": 12345,
            "user_validated": true,
            "user_finished": true,
            "key_obtained": true,
            "submissions_attempts": 10,
            "start_date": "2019-01-30T09:26:41Z",
            "validation_date": "2019-01-31T09:26:41Z",
            "finish_date": "2019-02-01T09:26:41Z",
            "partial_access": true,
            "full_access": true,
            "gray_access": true,
            "children": [
              {
                "item_id": 231,
                "order": 1,
                "access_restricted": true,
                "type": "Task",
                "transparent_folder": true,
                "has_unlocked_items": true,
                "title": "Task 2",
                "user_score": 12345,
                "user_validated": true,
                "user_finished": true,
                "key_obtained": true,
                "submissions_attempts": 10,
                "start_date": "2019-01-30T09:26:41Z",
                "validation_date": "2019-01-31T09:26:41Z",
                "finish_date": "2019-02-01T09:26:41Z",
                "partial_access": true,
                "full_access": true,
                "gray_access": true
              },
              {
                "item_id": 232,
                "order": 2,
                "access_restricted": true,
                "type": "Task",
                "transparent_folder": true,
                "has_unlocked_items": true,
                "title": "Task 3",
                "user_score": 12345,
                "user_validated": true,
                "user_finished": true,
                "key_obtained": true,
                "submissions_attempts": 10,
                "start_date": "2019-01-30T09:26:41Z",
                "validation_date": "2019-01-31T09:26:41Z",
                "finish_date": "2019-02-01T09:26:41Z",
                "partial_access": true,
                "full_access": true,
                "gray_access": true
              }
            ]
          }]
        }
      """
# TODO: test different languages
