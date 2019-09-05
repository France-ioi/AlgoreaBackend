Feature: Add item

  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned |
      | 1  | jdoe   | 0        | 11          | 12           |
    And the database has the following table 'groups':
      | ID | sName      | sType     |
      | 11 | jdoe       | UserSelf  |
      | 12 | jdoe-admin | UserAdmin |
    And the database has the following table 'items':
      | ID | bTeamsEditable | bNoScore |
      | 21 | false          | false    |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | bCachedManagerAccess | idUserCreated |
      | 41 | 11      | 21     | true                 | 0             |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf |
      | 71 | 11              | 11           | 1       |
      | 72 | 12              | 12           | 1       |
    And the database has the following table 'languages':
      | ID |
      | 3  |

  Scenario: Valid
    Given I am the user with ID "1"
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "language_id": "3",
        "title": "my title",
        "image_url":"http://bit.ly/1234",
        "subtitle": "hard task",
        "description": "the goal of this task is ...",
        "parent_item_id": "21",
        "order": 100
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": { "id": "5577006791947779410" }
      }
      """
    And the table "items" at ID "5577006791947779410" should be:
      | ID                  | sType  | sUrl | idDefaultLanguage | bTeamsEditable | bNoScore | sTextID | bTitleBarVisible | bCustomChapter | bDisplayDetailsInParent | bUsesAPI | bReadOnly | sFullScreen | bShowDifficulty | bShowSource | bHintsAllowed | bFixedRanks | sValidationType | iValidationMin | idItemUnlocked | iScoreMinUnlock | sTeamMode | bTeamsEditable | idTeamInGroup | iTeamMaxMembers | bHasAttempts | sAccessOpenDate      | sDuration | sEndContestDate      | bShowUserInfos | sContestPhase | iLevel | bNoScore | groupCodeEnter |
      | 5577006791947779410 | Course | null | 3                 | 0              | 0        | null    | 1                | 0              | 0                       | 1        | 0         | default     | 0               | 0           | 0             | 0           | All             | null           | null           | 100             | null      | 0              | null          | 0               | 0            | null                 | null      | null                 | 0              | Running       | null   | 0        | 0              |
    And the table "items_strings" should be:
      | ID                  | idItem              | idLanguage | sTitle   | sImageUrl          | sSubtitle | sDescription                 |
      | 6129484611666145821 | 5577006791947779410 | 3          | my title | http://bit.ly/1234 | hard task | the goal of this task is ... |
    And the table "items_items" should be:
      | ID                  | idItemParent | idItemChild         | iChildOrder |
      | 4037200794235010051 | 21           | 5577006791947779410 | 100         |
    And the table "items_ancestors" should be:
      | idItemAncestor | idItemChild         |
      | 21             | 5577006791947779410 |
    And the table "groups_items" at ID "8674665223082153551" should be:
      | ID                  | idGroup | idItem              | idUserCreated | ABS(TIMESTAMPDIFF(SECOND, sFullAccessDate, NOW())) < 3 | bOwnerAccess | bCachedManagerAccess | ABS(TIMESTAMPDIFF(SECOND, sCachedFullAccessDate, NOW())) < 3 | bCachedFullAccess |
      | 8674665223082153551 | 11      | 5577006791947779410 | 1             | 1                                                      | 1            | 1                    | 1                                                            | 1                 |

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
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | bCachedManagerAccess | bOwnerAccess |
      | 42 | 11      | 12     | true                 | false        |
      | 43 | 11      | 34     | false                | true         |
    When I send a POST request to "/items" with the following body:
      """
      {
        "type": "Course",
        "url": "http://myurl.com/",
        "text_id": "Task number 1",
        "title_bar_visible": true,
        "custom_chapter": true,
        "display_details_in_parent": true,
        "uses_api": true,
        "read_only": true,
        "full_screen": "forceYes",
        "show_difficulty": true,
        "show_source": true,
        "hints_allowed": true,
        "fixed_ranks": true,
        "validation_type": "AllButOne",
        "validation_min": 1234,
        "unlocked_item_ids": "12,34",
        "score_min_unlock": 34,
        "team_mode": "All",
        "teams_editable": true,
        "team_in_group_id": "12345",
        "team_max_members": 2345,
        "has_attempts": true,
        "access_open_date": "2018-01-02T03:04:05Z",
        "duration": "01:02:03",
        "end_contest_date": "2019-02-03T04:05:06Z",
        "show_user_infos": true,
        "contest_phase": "Analysis",
        "level": 345,
        "no_score": true,
        "group_code_enter": true,
        "language_id": "3",
        "title": "my title",
        "image_url":"http://bit.ly/1234",
        "subtitle": "hard task",
        "description": "the goal of this task is ...",
        "parent_item_id": "21",
        "order": 100,
        "children": [
          {"item_id": "12", "order": 0},
          {"item_id": "34", "order": 1}
        ]
      }
      """
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": { "id": "5577006791947779410" }
      }
      """
    And the table "items" at ID "5577006791947779410" should be:
      | ID                  | sType  | sUrl              | idDefaultLanguage | bTeamsEditable | bNoScore | sTextID       | bTitleBarVisible | bCustomChapter | bDisplayDetailsInParent | bUsesAPI | bReadOnly | sFullScreen | bShowDifficulty | bShowSource | bHintsAllowed | bFixedRanks | sValidationType | iValidationMin | idItemUnlocked | iScoreMinUnlock | sTeamMode | bTeamsEditable | idTeamInGroup | iTeamMaxMembers | bHasAttempts | sAccessOpenDate      | sDuration | sEndContestDate      | bShowUserInfos | sContestPhase | iLevel | bNoScore | groupCodeEnter |
      | 5577006791947779410 | Course | http://myurl.com/ | 3                 | 1              | 1        | Task number 1 | 1                | 1              | 1                       | 1        | 1         | forceYes    | 1               | 1           | 1             | 1           | AllButOne       | 1234           | 12,34          | 34              | All       | 1              | 12345         | 2345            | 1            | 2018-01-02T03:04:05Z | 01:02:03  | 2019-02-03T04:05:06Z | 1              | Analysis      | 345    | 1        | 1              |
    And the table "items_strings" should be:
      | ID                  | idItem              | idLanguage | sTitle   | sImageUrl          | sSubtitle | sDescription                 |
      | 6129484611666145821 | 5577006791947779410 | 3          | my title | http://bit.ly/1234 | hard task | the goal of this task is ... |
    And the table "items_items" should be:
      | ID                  | idItemParent        | idItemChild         | iChildOrder |
      | 3916589616287113937 | 5577006791947779410 | 12                  | 0           |
      | 4037200794235010051 | 21                  | 5577006791947779410 | 100         |
      | 6334824724549167320 | 5577006791947779410 | 34                  | 1           |
    And the table "items_ancestors" should be:
      | idItemAncestor      | idItemChild         |
      | 21                  | 12                  |
      | 21                  | 34                  |
      | 21                  | 5577006791947779410 |
      | 5577006791947779410 | 12                  |
      | 5577006791947779410 | 34                  |
    And the table "groups_items" at ID "8674665223082153551" should be:
      | ID                  | idGroup | idItem              | idUserCreated | ABS(TIMESTAMPDIFF(SECOND, sFullAccessDate, NOW())) < 3 | bOwnerAccess | bCachedManagerAccess | ABS(TIMESTAMPDIFF(SECOND, sCachedFullAccessDate, NOW())) < 3 | bCachedFullAccess |
      | 8674665223082153551 | 11      | 5577006791947779410 | 1             | 1                                                      | 1            | 1                    | 1                                                            | 1                 |

  Scenario: Valid with empty full_screen
    Given I am the user with ID "1"
    When I send a POST request to "/items" with the following body:
    """
    {
      "type": "Course",
      "full_screen": "",
      "language_id": "3",
      "title": "my title",
      "parent_item_id": "21",
      "order": 100
    }
    """
    Then the response code should be 201
    And the response body should be, in JSON:
    """
    {
      "success": true,
      "message": "created",
      "data": { "id": "5577006791947779410" }
    }
    """
    And the table "items" at ID "5577006791947779410" should be:
      | ID                  | sType  | sUrl | idDefaultLanguage | bTeamsEditable | bNoScore | sTextID | bTitleBarVisible | bCustomChapter | bDisplayDetailsInParent | bUsesAPI | bReadOnly | sFullScreen | bShowDifficulty | bShowSource | bHintsAllowed | bFixedRanks | sValidationType | iValidationMin | idItemUnlocked | iScoreMinUnlock | sTeamMode | bTeamsEditable | idTeamInGroup | iTeamMaxMembers | bHasAttempts | sAccessOpenDate      | sDuration | sEndContestDate      | bShowUserInfos | sContestPhase | iLevel | bNoScore | groupCodeEnter |
      | 5577006791947779410 | Course | null |                 3 |              0 |        0 | null    | 1                | 0              | 0                       | 1        | 0         |             | 0               | 0           | 0             | 0           | All             | null           | null           | 100             | null      | 0              | null          | 0               | 0            | null                 | null      | null                 | 0              | Running       | null   | 0        | 0              |
    And the table "items_strings" should be:
      | ID                  | idItem              | idLanguage | sTitle   | sImageUrl | sSubtitle | sDescription |
      | 6129484611666145821 | 5577006791947779410 | 3          | my title | null      | null      | null         |
    And the table "items_items" should be:
      | ID                  | idItemParent | idItemChild         | iChildOrder |
      | 4037200794235010051 | 21           | 5577006791947779410 | 100         |
    And the table "items_ancestors" should be:
      | idItemAncestor | idItemChild         |
      | 21             | 5577006791947779410 |
    And the table "groups_items" at ID "8674665223082153551" should be:
      | ID                  | idGroup | idItem              | idUserCreated | ABS(TIMESTAMPDIFF(SECOND, sFullAccessDate, NOW())) < 3 | bOwnerAccess | bCachedManagerAccess | ABS(TIMESTAMPDIFF(SECOND, sCachedFullAccessDate, NOW())) < 3 | bCachedFullAccess |
      | 8674665223082153551 | 11      | 5577006791947779410 | 1             | 1                                                      | 1            | 1                    | 1                                                            | 1                 |
