Feature: Get item view information

  Background:
    Given the database has the following table 'users':
      | ID | sLogin | tempUser | idGroupSelf | idGroupOwned | iVersion |
      | 1  | jdoe   | 0        | 11          | 12           | 0        |
    And the database has the following table 'groups':
      | ID | sName      | sTextId | iGrade | sType     | iVersion |
      | 11 | jdoe       |         | -2     | UserAdmin | 0        |
      | 12 | jdoe-admin |         | -2     | UserAdmin | 0        |
      | 13 | Group B    |         | -2     | Class     | 0        |
    And the database has the following table 'items':
      | ID  | sType    | bNoScore | idItemUnlocked | sAccessOpenDate      | bDisplayDetailsInParent | sValidationType | iScoreMinUnlock | sTeamMode | bTeamsEditable | iTeamMaxMembers | bHasAttempts | sDuration | sEndContestDate      | groupCodeEnter | bTitleBarVisible | bReadOnly | sFullScreen | bShowSource | iValidationMin | bShowUserInfos | sContestPhase | sUrl           | bUsesAPI | bHintsAllowed | iVersion |
      | 200 | Category | true     | 1234,2345      | 2019-02-06T09:26:40Z | true                    | All             | 100             | All       | true           | 10              | true         | 10:20:30  | 2019-03-06T09:26:40Z | true           | true             | true      | forceYes    | true        | 100            | true           | Running       | http://someurl | true     | true          | 0        |
      | 210 | Chapter  | true     | 1234,2345      | 2019-02-06T09:26:41Z | true                    | All             | 100             | All       | true           | 10              | true         | 10:20:31  | 2019-03-06T09:26:41Z | true           | true             | true      | forceYes    | true        | 100            | true           | Running       | http://someurl | true     | true          | 0        |
      | 220 | Chapter  | true     | 1234,2345      | 2019-02-06T09:26:42Z | true                    | All             | 100             | All       | true           | 10              | true         | 10:20:32  | 2019-03-06T09:26:42Z | true           | true             | true      | forceYes    | true        | 100            | true           | Running       | http://someurl | true     | true          |          |
    And the database has the following table 'items_strings':
      | ID | idItem | idLanguage | sTitle      | sImageUrl                  | sSubtitle    | sDescription  | sEduComment  | iVersion |
      | 53 | 200    | 1          | Category 1  | http://example.com/my0.jpg | Subtitle 0   | Description 0 | Some comment | 0        |
      | 54 | 210    | 1          | Chapter A   | http://example.com/my1.jpg | Subtitle 1   | Description 1 | Some comment | 0        |
      | 55 | 220    | 1          | Chapter B   | http://example.com/my2.jpg | Subtitle 2   | Description 2 | Some comment | 0        |
      | 63 | 200    | 2          | Catégorie 1 | http://example.com/mf0.jpg | Sous-titre 0 | texte 0       | Some comment | 0        |
      | 64 | 210    | 2          | Chapitre A  | http://example.com/mf1.jpg | Sous-titre 1 | texte 1       | Some comment | 0        |
      | 66 | 220    | 2          | Chapitre b  | http://example.com/mf2.jpg | Sous-titre 2 | texte 2       | Some comment | 0        |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | iVersion |
      | 61 | 13            | 11           | 0        |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 71 | 11              | 11           | 1       | 0        |
      | 72 | 12              | 12           | 1       | 0        |
      | 73 | 13              | 13           | 1       | 0        |
      | 74 | 13              | 11           | 0       | 0        |
    And the database has the following table 'items_items':
      | ID | idItemParent | idItemChild | iChildOrder | sCategory | bAlwaysVisible | bAccessRestricted | iVersion |
      | 54 | 200          | 210         | 1           | Discovery | true           | true              | 0        |
      | 55 | 200          | 220         | 2           | Discovery | true           | true              | 0        |
    And the database has the following table 'users_items':
      | ID | idUser | idItem | idAttemptActive | iScore | nbSubmissionsAttempts | bValidated  | bFinished | bKeyObtained | nbHintsCached | sStartDate           | sFinishDate          | sValidationDate      | sContestStartDate    | sState     | sAnswer     | iVersion |
      | 1  | 1      | 200    | 100             | 12341  | 11                    | true        | true      | true         | 11            | 2019-01-30T09:26:41Z | 2019-02-01T09:26:41Z | 2019-01-31T09:26:41Z | 2019-02-01T06:26:41Z | Some state | Some answer | 0        |
      | 2  | 1      | 210    | 100             | 12342  | 12                    | true        | true      | true         | 11            | 2019-01-30T09:26:42Z | 2019-02-01T09:26:42Z | 2019-01-31T09:26:42Z | 2019-02-01T06:26:42Z | Some state | Some answer | 0        |
      | 3  | 1      | 220    | 100             | 12344  | 14                    | true        | true      | true         | 11            | 2019-01-30T09:26:44Z | 2019-02-01T09:26:44Z | 2019-01-31T09:26:44Z | 2019-02-01T06:26:44Z | Some state | Some answer | 0        |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | bCachedAccessSolutions | idUserCreated | iVersion |
      | 43 | 13      | 200    | null            | true              | true                 | true                | true                   | 0             | 0        |
      | 44 | 13      | 210    | null            | true              | true                 | true                | true                   | 0             | 0        |
      | 45 | 13      | 220    | null            | true              | true                 | true                | true                   | 0             | 0        |

  Scenario: Full access on all items
    Given I am the user with ID "1"
    When I send a GET request to "/items/200"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": 200,
      "type": "Category",
      "display_details_in_parent": true,
      "validation_type": "All",
      "has_unlocked_items": true,
      "score_min_unlock": 100,
      "team_mode": "All",
      "teams_editable": true,
      "team_max_members": 10,
      "has_attempts": true,
      "access_open_date": "2019-02-06T09:26:40Z",
      "duration": "10:20:30",
      "end_contest_date": "2019-03-06T09:26:40Z",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_source": true,
      "validation_min": 100,
      "show_user_infos": true,
      "contest_phase": "Running",
      "url": "http://someurl",
      "uses_API": true,
      "hints_allowed": true,

      "string": {
        "language_id": 1,
        "title": "Category 1",
        "image_url": "http://example.com/my0.jpg",
        "subtitle": "Subtitle 0",
        "description": "Description 0",
        "edu_comment": "Some comment"
      },

      "user": {
        "active_attempt_id": 100,
        "score": 12341,
        "submissions_attempts": 11,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "start_date": "2019-01-30T09:26:41Z",
        "validation_date": "2019-01-31T09:26:41Z",
        "finish_date": "2019-02-01T09:26:41Z",
        "contest_start_date": "2019-02-01T06:26:41Z",

        "state": "Some state",
        "answer": "Some answer"
      },

      "children": [
        {
          "id": 210,

          "order": 1,
          "category": "Discovery",
          "always_visible": true,
          "access_restricted": true,

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "team_mode": "All",
          "teams_editable": true,
          "team_max_members": 10,
          "has_attempts": true,
          "access_open_date": "2019-02-06T09:26:41Z",
          "duration": "10:20:31",
          "end_contest_date": "2019-03-06T09:26:41Z",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": 1,
            "title": "Chapter A",
            "image_url": "http://example.com/my1.jpg",
            "subtitle": "Subtitle 1",
            "description": "Description 1"
          },

          "user": {
            "active_attempt_id": 100,
            "score": 12342,
            "submissions_attempts": 12,
            "validated": true,
            "finished": true,
            "key_obtained": true,
            "hints_cached": 11,
            "start_date": "2019-01-30T09:26:42Z",
            "validation_date": "2019-01-31T09:26:42Z",
            "finish_date": "2019-02-01T09:26:42Z",
            "contest_start_date": "2019-02-01T06:26:42Z"
          }
        },
        {
          "id": 220,
          "order": 2,
          "category": "Discovery",
          "always_visible": true,
          "access_restricted": true,

          "type": "Chapter",
          "display_details_in_parent": true,
          "validation_type": "All",
          "has_unlocked_items": true,
          "score_min_unlock": 100,
          "team_mode": "All",
          "teams_editable": true,
          "team_max_members": 10,
          "has_attempts": true,
          "access_open_date": "2019-02-06T09:26:42Z",
          "duration": "10:20:32",
          "end_contest_date": "2019-03-06T09:26:42Z",
          "no_score": true,
          "group_code_enter": true,

          "string": {
            "language_id": 1,
            "title": "Chapter B",
            "image_url": "http://example.com/my2.jpg",
            "subtitle": "Subtitle 2",
            "description": "Description 2"
          },

          "user": {
            "active_attempt_id": 100,
            "score": 12344,
            "submissions_attempts": 14,
            "validated": true,
            "finished": true,
            "key_obtained": true,
            "hints_cached": 11,
            "start_date": "2019-01-30T09:26:44Z",
            "validation_date": "2019-01-31T09:26:44Z",
            "finish_date": "2019-02-01T09:26:44Z",
            "contest_start_date": "2019-02-01T06:26:44Z"
          }
        }
      ]
    }
    """

  Scenario: Chapter as a root node (full access)
    Given I am the user with ID "1"
    When I send a GET request to "/items/210"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "id": 210,
      "type": "Chapter",
      "display_details_in_parent": true,
      "validation_type": "All",
      "has_unlocked_items": true,
      "score_min_unlock": 100,
      "team_mode": "All",
      "teams_editable": true,
      "team_max_members": 10,
      "has_attempts": true,
      "access_open_date": "2019-02-06T09:26:41Z",
      "duration": "10:20:31",
      "end_contest_date": "2019-03-06T09:26:41Z",
      "no_score": true,
      "group_code_enter": true,

      "title_bar_visible": true,
      "read_only": true,
      "full_screen": "forceYes",
      "show_source": true,
      "validation_min": 100,
      "show_user_infos": true,
      "contest_phase": "Running",

      "string": {
        "language_id": 1,
        "title": "Chapter A",
        "image_url": "http://example.com/my1.jpg",
        "subtitle": "Subtitle 1",
        "description": "Description 1",
        "edu_comment": "Some comment"
      },

      "user": {
        "active_attempt_id": 100,
        "score": 12342,
        "submissions_attempts": 12,
        "validated": true,
        "finished": true,
        "key_obtained": true,
        "hints_cached": 11,
        "start_date": "2019-01-30T09:26:42Z",
        "validation_date": "2019-01-31T09:26:42Z",
        "finish_date": "2019-02-01T09:26:42Z",
        "contest_start_date": "2019-02-01T06:26:42Z"
      }
    }
    """


# TODO:
#  Chapter as a root node
#  Root node without AccessSolutions
#  Grayed Root
#  Grayed Child
#  Root node without solution access
#  Custom language
