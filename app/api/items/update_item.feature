Feature: Update item

Background:
  Given the database has the following table 'users':
    | ID | sLogin | tempUser | idGroupSelf | idGroupOwned |
    | 1  | jdoe   | 0        | 11          | 12           |
  And the database has the following table 'groups':
    | ID | sName      | sType     |
    | 11 | jdoe       | UserSelf  |
    | 12 | jdoe-admin | UserAdmin |
  And the database has the following table 'items':
    | ID | sType   | sUrl                 | idDefaultLanguage | bNoScore | sTextID | bTitleBarVisible | bCustomChapter | bDisplayDetailsInParent | bUsesAPI | bReadOnly | sFullScreen | bShowDifficulty | bShowSource | bHintsAllowed | bFixedRanks | sValidationType | iValidationMin | idItemUnlocked | iScoreMinUnlock | sTeamMode | bTeamsEditable | idTeamInGroup | iTeamMaxMembers | bHasAttempts | sAccessOpenDate      | sDuration | sEndContestDate      | bShowUserInfos | sContestPhase | iLevel | groupCodeEnter |
    | 21 | Chapter | http://someurl1.com/ | 2                 | 1        | Task 1  | 0                | 1              | 1                       | 0        | 1         | forceNo     | 1               | 1           | 1             | 1           | One             | 12             | 1              | 99              | Half      | 1              | 2             | 10              | 1            | 2016-01-02T03:04:05Z | 01:20:30  | 2017-01-02T03:04:05Z | 1              | Closed        | 3      | 1              |
    | 50 | Chapter | http://someurl2.com/ | 2                 | 1        | Task 2  | 0                | 1              | 1                       | 0        | 1         | forceNo     | 1               | 1           | 1             | 1           | One             | 12             | 1              | 99              | Half      | 1              | 2             | 10              | 1            | 2016-01-02T03:04:05Z | 01:20:30  | 2017-01-02T03:04:05Z | 1              | Closed        | 3      | 1              |
    | 60 | Chapter | http://someurl2.com/ | 2                 | 1        | Task 3  | 0                | 1              | 1                       | 0        | 1         | forceNo     | 1               | 1           | 1             | 1           | One             | 12             | 1              | 99              | Half      | 1              | 2             | 10              | 1            | 2016-01-02T03:04:05Z | 01:20:30  | 2017-01-02T03:04:05Z | 1              | Closed        | 3      | 1              |
  And the database has the following table 'items_items':
    | idItemParent | idItemChild |
    | 21           | 60          |
    | 50           | 21          |
  And the database has the following table 'items_ancestors':
    | idItemAncestor | idItemChild |
    | 21             | 60          |
    | 50             | 21          |
    | 50             | 60          |
  And the database has the following table 'groups_items':
    | ID | idGroup | idItem | bManagerAccess | bOwnerAccess |
    | 40 | 11      | 50     | false          | true         |
    | 41 | 11      | 21     | true           | false        |
    | 42 | 11      | 60     | false          | true         |
  And the database has the following table 'groups_ancestors':
    | ID | idGroupAncestor | idGroupChild | bIsSelf |
    | 71 | 11              | 11           | 1       |
    | 72 | 12              | 12           | 1       |
  And the database has the following table 'languages':
    | ID |
    | 2  |
    | 3  |

Scenario: Valid
  Given I am the user with ID "1"
  When I send a PUT request to "/items/50" with the following body:
    """
    {
      "type": "Course"
    }
    """
  Then the response should be "updated"
  And the table "items" at ID "50" should be:
    | ID | sType  | sUrl                 | idDefaultLanguage | bNoScore | sTextID | bTitleBarVisible | bCustomChapter | bDisplayDetailsInParent | bUsesAPI | bReadOnly | sFullScreen | bShowDifficulty | bShowSource | bHintsAllowed | bFixedRanks | sValidationType | iValidationMin | idItemUnlocked | iScoreMinUnlock | sTeamMode | bTeamsEditable | idTeamInGroup | iTeamMaxMembers | bHasAttempts | sAccessOpenDate      | sDuration | sEndContestDate      | bShowUserInfos | sContestPhase | iLevel | groupCodeEnter |
    | 50 | Course | http://someurl2.com/ | 2                 | 1        | Task 2  | 0                | 1              | 1                       | 0        | 1         | forceNo     | 1               | 1           | 1             | 1           | One             | 12             | 1              | 99              | Half      | 1              | 2             | 10              | 1            | 2016-01-02T03:04:05Z | 01:20:30  | 2017-01-02T03:04:05Z | 1              | Closed        | 3      | 1              |
  And the table "items_strings" should stay unchanged
  And the table "items_items" should stay unchanged
  And the table "items_ancestors" should stay unchanged
  And the table "groups_items" should be:
    | idGroup | idItem | bManagerAccess | bOwnerAccess |
    | 11      | 21     | true           | false        |
    | 11      | 50     | false          | true         |
    | 11      | 60     | false          | true         |

  Scenario: Valid (all the fields are set)
    Given I am the user with ID "1"
    And the database has the following table 'groups':
      | ID    |
      | 12345 |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf |
      | 73 | 12              | 12345        | 0       |
    And the database has the following table 'items':
      | ID |
      | 12 |
      | 34 |
    And the database has the following table 'items_strings':
      | idLanguage | idItem |
      | 3          | 50     |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | bManagerAccess | bOwnerAccess |
      | 43 | 11      | 12     | true           | false        |
      | 44 | 11      | 34     | false          | true         |
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "type": "Course",
        "url": "http://myurl.com/",
        "text_id": "Task number 1",
        "title_bar_visible": true,
        "custom_chapter": false,
        "display_details_in_parent": false,
        "uses_api": true,
        "read_only": false,
        "full_screen": "forceYes",
        "show_difficulty": false,
        "show_source": false,
        "hints_allowed": false,
        "fixed_ranks": false,
        "validation_type": "AllButOne",
        "validation_min": 1234,
        "unlocked_item_ids": "12,34",
        "score_min_unlock": 34,
        "team_mode": "All",
        "teams_editable": false,
        "team_in_group_id": "12345",
        "team_max_members": 2345,
        "has_attempts": false,
        "access_open_date": "2018-01-02T03:04:05Z",
        "duration": "01:02:03",
        "end_contest_date": "2019-02-03T04:05:06Z",
        "show_user_infos": false,
        "contest_phase": "Analysis",
        "level": 345,
        "no_score": false,
        "group_code_enter": false,
        "default_language_id": "3",
        "children": [
          {"item_id": "12", "order": 0},
          {"item_id": "34", "order": 1}
        ]
      }
      """
    Then the response should be "updated"
    And the table "items" at ID "50" should be:
      | ID | sType  | sUrl              | idDefaultLanguage | bTeamsEditable | bNoScore | sTextID       | bTitleBarVisible | bCustomChapter | bDisplayDetailsInParent | bUsesAPI | bReadOnly | sFullScreen | bShowDifficulty | bShowSource | bHintsAllowed | bFixedRanks | sValidationType | iValidationMin | idItemUnlocked | iScoreMinUnlock | sTeamMode | bTeamsEditable | idTeamInGroup | iTeamMaxMembers | bHasAttempts | sAccessOpenDate      | sDuration | sEndContestDate      | bShowUserInfos | sContestPhase | iLevel | groupCodeEnter |
      | 50 | Course | http://myurl.com/ | 3                 | 0              | 0        | Task number 1 | 1                | 0              | 0                       | 1        | 0         | forceYes    | 0               | 0           | 0             | 0           | AllButOne       | 1234           | 12,34          | 34              | All       | 0              | 12345         | 2345            | 0            | 2018-01-02T03:04:05Z | 01:02:03  | 2019-02-03T04:05:06Z | 0              | Analysis      | 345    | 0              |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should be:
      | idItemParent | idItemChild |
      | 21           | 60          |
      | 50           | 12          |
      | 50           | 34          |
    And the table "items_ancestors" should be:
      | idItemAncestor | idItemChild |
      | 21             | 60          |
      | 50             | 12          |
      | 50             | 34          |
    And the table "groups_items" should stay unchanged

  Scenario: Valid with empty full_screen
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50" with the following body:
      """
      {
        "full_screen": ""
      }
      """
    Then the response should be "updated"
    And the table "items" at ID "50" should be:
      | ID | sType   | sUrl                 | idDefaultLanguage | bNoScore | sTextID | bTitleBarVisible | bCustomChapter | bDisplayDetailsInParent | bUsesAPI | bReadOnly | sFullScreen | bShowDifficulty | bShowSource | bHintsAllowed | bFixedRanks | sValidationType | iValidationMin | idItemUnlocked | iScoreMinUnlock | sTeamMode | bTeamsEditable | idTeamInGroup | iTeamMaxMembers | bHasAttempts | sAccessOpenDate      | sDuration | sEndContestDate      | bShowUserInfos | sContestPhase | iLevel | groupCodeEnter |
      | 50 | Chapter | http://someurl2.com/ | 2                 | 1        | Task 2  | 0                | 1              | 1                       | 0        | 1         |             | 1               | 1           | 1             | 1           | One             | 12             | 1              | 99              | Half      | 1              | 2             | 10              | 1            | 2016-01-02T03:04:05Z | 01:20:30  | 2017-01-02T03:04:05Z | 1              | Closed        | 3      | 1              |
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Valid without any fields
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50" with the following body:
    """
    {
    }
    """
    Then the response should be "updated"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should stay unchanged
    And the table "items_ancestors" should stay unchanged
    And the table "groups_items" should stay unchanged

  Scenario: Valid with empty children array
    Given I am the user with ID "1"
    When I send a PUT request to "/items/50" with the following body:
    """
    {
      "children": []
    }
    """
    Then the response should be "updated"
    And the table "items" should stay unchanged
    And the table "items_strings" should stay unchanged
    And the table "items_items" should be:
      | idItemParent | idItemChild |
      | 21           | 60          |
    And the table "items_ancestors" should be:
      | idItemAncestor | idItemChild |
      | 21             | 60          |
    And the table "groups_items" should stay unchanged
