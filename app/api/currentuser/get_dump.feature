Feature: Export the current user's data
  Background:
    Given the DB time now is "2019-07-16T22:02:28Z"
    And the time now is "2019-07-16T22:02:28Z"
    And the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 2  | user   | 11          | 12           | John        | Doe       | 1      |
      | 4  | jane   | 31          | 32           | Jane        | Doe       | 2      |
    And the database has the following table 'refresh_tokens':
      | idUser | sRefreshToken    |
      | 1      | refreshTokenFor1 |
      | 2      | refreshTokenFor2 |
    And the database has the following table 'groups':
      | ID | sType     | sName              | sDescription           |
      | 1  | Class     | Our Class          | Our class group        |
      | 2  | Team      | Our Team           | Our team group         |
      | 3  | Club      | Our Club           | Our club group         |
      | 4  | Friends   | Our Friends        | Group for our friends  |
      | 5  | Other     | Other people       | Group for other people |
      | 6  | Class     | Another Class      | Another class group    |
      | 7  | Team      | Another Team       | Another team group     |
      | 8  | Club      | Another Club       | Another club group     |
      | 9  | Friends   | Some other friends | Another friends group  |
      | 11 | UserSelf  | user self          |                        |
      | 12 | UserAdmin | user admin         |                        |
      | 31 | UserSelf  | jane               |                        |
      | 32 | UserAdmin | jane-admin         |                        |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate               | idUserInviting |
      | 2  | 1             | 11           | invitationSent     | {{relativeTime("-169h")}} | null           |
      | 3  | 2             | 11           | invitationAccepted | {{relativeTime("-168h")}} | 1              |
      | 4  | 3             | 11           | requestSent        | {{relativeTime("-167h")}} | 1              |
      | 5  | 4             | 11           | requestRefused     | {{relativeTime("-166h")}} | 2              |
      | 6  | 5             | 11           | invitationAccepted | {{relativeTime("-165h")}} | 2              |
      | 7  | 6             | 11           | requestAccepted    | {{relativeTime("-164h")}} | 2              |
      | 8  | 7             | 11           | removed            | {{relativeTime("-163h")}} | 1              |
      | 9  | 8             | 11           | left               | {{relativeTime("-162h")}} | 1              |
      | 10 | 9             | 11           | direct             | {{relativeTime("-161h")}} | 2              |
      | 11 | 1             | 12           | invitationSent     | {{relativeTime("-170h")}} | 2              |
      | 12 | 12            | 1            | direct             | null                      | null           |
      | 13 | 12            | 2            | direct             | null                      | null           |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 1               | 1            | true    |
      | 2               | 2            | true    |
      | 2               | 11           | false   |
      | 3               | 3            | true    |
      | 4               | 4            | true    |
      | 5               | 5            | true    |
      | 5               | 11           | false   |
      | 6               | 6            | true    |
      | 6               | 11           | false   |
      | 7               | 7            | true    |
      | 8               | 8            | true    |
      | 9               | 9            | true    |
      | 9               | 11           | false   |
      | 11              | 11           | true    |
      | 12              | 1            | false   |
      | 12              | 2            | false   |
      | 12              | 12           | true    |
    And the database has the following table 'users_answers':
      | ID | idUser |
      | 1  | 2      |
      | 2  | 3      |
    And the database has the following table 'users_items':
      | ID | idUser |
      | 11 | 2      |
      | 12 | 3      |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup |
      | 111 | 11      |
      | 112 | 2       |
      | 113 | 1       |

  Scenario: Full data
    Given I am the user with ID "2"
    When I send a GET request to "/current-user/dump"
    Then the response code should be 200
    And the response header "Content-Type" should be "application/json; charset=utf-8"
    And the response header "Content-Disposition" should be "attachment; filename=user_data.json"
    And the response body should be, in JSON:
    """
    {
    	"current_user": {
    		"ID": "2", "allowSubgroups": null, "bBasicEditorMode": 1, "bEmailVerified": 0, "bIsAdmin": 0,
    		"bNoRanking": 0, "bNotifyNews": 0, "bPhotoAutoload": 0, "bPublicFirstName": 0, "bPublicLastName": 0,
    		"creatorID": null, "iGrade": 1, "iGraduationYear": 0, "iMemberState": 0, "iStepLevelInSite": 0,
    		"idGroupAccess": null, "idGroupOwned": "12", "idGroupSelf": "11", "idUserGodfather": null, "loginID": null,
    		"loginModulePrefix": null, "nbHelpGiven": 0, "nbSpacesForTab": 3, "sAddress": null, "sBirthDate": null,
    		"sCellPhoneNumber": null, "sCity": null, "sCountryCode": "", "sDefaultLanguage": "fr", "sEmail": null,
    		"sFirstName": "John", "sFreeText": null, "sLandLineNumber": null, "sLangProg": "Python",
    		"sLastActivityDate": null, "sLastIP": null, "sLastLoginDate": null, "sLastName": "Doe", "sLogin": "user",
    		"sNotificationReadDate": null, "sNotify": "Answers", "sOpenIdIdentity": null, "sPasswordMd5": null,
    		"sRecover": null, "sRegistrationDate": null, "sSalt": null, "sSex": null, "sStudentId": null, "sTimeZone": null,
    		"sWebSite": null, "sZipcode": null, "tempUser": 0
    	},
    	"groups_attempts": [
    		{
    			"ID": "111", "bFinished": 0, "bKeyObtained": 0, "bRanked": 0, "bValidated": 0, "iAutonomy": 0, "iMinusScore": -0,
    			"iOrder": 0, "iPrecision": 0, "iScore": 0, "iScoreComputed": 0, "iScoreDiffManual": 0, "iScoreReeval": 0,
    			"idGroup": "11", "idItem": "0", "idUserCreator": null, "nbChildrenValidated": 0,
    			"nbCorrectionsRead": 0, "nbHintsCached": 0, "nbSubmissionsAttempts": 0, "nbTasksSolved": 0, "nbTasksTried": 0,
    			"nbTasksWithHelp": 0, "sAdditionalTime": null, "sAllLangProg": null, "sAncestorsComputationState": "done",
    			"sBestAnswerDate": null, "sContestStartDate": null, "sFinishDate": null, "sHintsRequested": null,
    			"sLastActivityDate": null, "sLastAnswerDate": null, "sLastHintDate": null, "sScoreDiffComment": "",
    			"sStartDate": null, "sThreadStartDate": null, "sValidationDate": null
    		},
    		{
    			"ID": "112", "bFinished": 0, "bKeyObtained": 0, "bRanked": 0, "bValidated": 0, "iAutonomy": 0, "iMinusScore": -0,
    			"iOrder": 0, "iPrecision": 0, "iScore": 0, "iScoreComputed": 0, "iScoreDiffManual": 0, "iScoreReeval": 0,
    			"idGroup": "2", "idItem": "0", "idUserCreator": null, "nbChildrenValidated": 0,
    			"nbCorrectionsRead": 0, "nbHintsCached": 0, "nbSubmissionsAttempts": 0, "nbTasksSolved": 0, "nbTasksTried": 0,
    			"nbTasksWithHelp": 0, "sAdditionalTime": null, "sAllLangProg": null, "sAncestorsComputationState": "done",
    			"sBestAnswerDate": null, "sContestStartDate": null, "sFinishDate": null, "sHintsRequested": null,
    			"sLastActivityDate": null, "sLastAnswerDate": null, "sLastHintDate": null, "sScoreDiffComment": "",
    			"sStartDate": null, "sThreadStartDate": null, "sValidationDate": null
    		}
    	],
    	"groups_groups": [
    		{
    		  "ID": "2", "iChildOrder": 0, "idGroupChild": "11", "idGroupParent": "1", "idUserInviting": null,
    			"sName": "Our Class", "sRole": "member", "sStatusDate": "2019-07-09T21:02:28Z", "sType": "invitationSent"
    		},
    		{
    			"ID": "3", "iChildOrder": 0, "idGroupChild": "11", "idGroupParent": "2", "idUserInviting": "1",
    			"sName": "Our Team", "sRole": "member", "sStatusDate": "2019-07-09T22:02:28Z", "sType": "invitationAccepted"
    		},
    		{
    			"ID": "4", "iChildOrder": 0, "idGroupChild": "11", "idGroupParent": "3", "idUserInviting": "1",
    			"sName": "Our Club", "sRole": "member", "sStatusDate": "2019-07-09T23:02:28Z", "sType": "requestSent"
    		},
    		{
    			"ID": "5", "iChildOrder": 0, "idGroupChild": "11", "idGroupParent": "4", "idUserInviting": "2",
    			"sName": "Our Friends", "sRole": "member", "sStatusDate": "2019-07-10T00:02:28Z", "sType": "requestRefused"
    		},
    		{
    			"ID": "6", "iChildOrder": 0, "idGroupChild": "11", "idGroupParent": "5", "idUserInviting": "2",
    			"sName": "Other people", "sRole": "member", "sStatusDate": "2019-07-10T01:02:28Z", "sType": "invitationAccepted"
    		},
    		{
    			"ID": "7", "iChildOrder": 0, "idGroupChild": "11", "idGroupParent": "6", "idUserInviting": "2",
    			"sName": "Another Class", "sRole": "member", "sStatusDate": "2019-07-10T02:02:28Z", "sType": "requestAccepted"
    		},
    		{
    			"ID": "8", "iChildOrder": 0, "idGroupChild": "11", "idGroupParent": "7", "idUserInviting": "1",
    			"sName": "Another Team", "sRole": "member", "sStatusDate": "2019-07-10T03:02:28Z", "sType": "removed"
    		},
    		{
    			"ID": "9", "iChildOrder": 0, "idGroupChild": "11", "idGroupParent": "8", "idUserInviting": "1",
    			"sName": "Another Club", "sRole": "member", "sStatusDate": "2019-07-10T04:02:28Z", "sType": "left"
    		},
    		{
    			"ID": "10", "iChildOrder": 0, "idGroupChild": "11", "idGroupParent": "9", "idUserInviting": "2",
    			"sName": "Some other friends", "sRole": "member", "sStatusDate": "2019-07-10T05:02:28Z", "sType": "direct"
    		}
    	],
    	"joined_groups": [
    		{"ID": "2", "sName": "Our Team"},
    		{"ID": "5", "sName": "Other people"},
    		{"ID": "6", "sName": "Another Class"},
    		{"ID": "9", "sName": "Some other friends"}
    	],
    	"owned_groups": [
    		{"ID": "1", "sName": "Our Class"},
    		{"ID": "2", "sName": "Our Team"}
    	],
    	"refresh_token": {"idUser": "2", "sRefreshToken": "***"},
    	"sessions": [
    		{
    			"idUser": "2", "sAccessToken": "***", "sExpirationDate": "2019-07-17T00:02:28Z",
    			"sIssuedAtDate": "2019-07-16T22:02:28Z", "sIssuer": null
    		}
    	],
    	"users_answers": [
    		{
    			"ID": "1", "bValidated": null, "iScore": null, "iVersion": 0, "idAttempt": null, "idItem": "0",
    			"idUser": "2", "idUserGrader": null, "sAnswer": null, "sGradingDate": null, "sLangProg": null,
    			"sName": null, "sState": null, "sSubmissionDate": "0001-01-01T00:00:00Z", "sType": "Submission"
    		}
    	],
    	"users_items": [
    		{
    			"ID": "11", "bFinished": 0, "bKeyObtained": 0, "bPlatformDataRemoved": 0, "bRanked": 0, "bValidated": 0,
    			"iAutonomy": 0, "iPrecision": 0, "iScore": 0, "iScoreComputed": 0, "iScoreDiffManual": 0, "iScoreReeval": 0,
    			"idAttemptActive": null, "idItem": "0", "idUser": "2", "nbChildrenValidated": 0,
    			"nbCorrectionsRead": 0, "nbHintsCached": 0, "nbSubmissionsAttempts": 0, "nbTasksSolved": 0, "nbTasksTried": 0,
    			"nbTasksWithHelp": 0, "sAdditionalTime": null, "sAllLangProg": null, "sAncestorsComputationState": "todo",
    			"sAnswer": null, "sBestAnswerDate": null, "sContestStartDate": null, "sFinishDate": null,
    			"sHintsRequested": null, "sLastActivityDate": null, "sLastAnswerDate": null, "sLastHintDate": null,
    			"sScoreDiffComment": "", "sStartDate": null, "sState": null, "sThreadStartDate": null, "sValidationDate": null
    		}
    	]
    }
    """

  Scenario: All optional arrays and objects are empty
    Given I am the user with ID "4"
    When I send a GET request to "/current-user/dump"
    Then the response code should be 200
    And the response header "Content-Type" should be "application/json; charset=utf-8"
    And the response header "Content-Disposition" should be "attachment; filename=user_data.json"
    And the response body should be, in JSON:
    """
    {
    	"current_user": {
    		"ID": "4", "allowSubgroups": null, "bBasicEditorMode": 1, "bEmailVerified": 0, "bIsAdmin": 0, "bNoRanking": 0,
    		"bNotifyNews": 0, "bPhotoAutoload": 0, "bPublicFirstName": 0, "bPublicLastName": 0, "creatorID": null,
    		"iGrade": 2, "iGraduationYear": 0, "iMemberState": 0, "iStepLevelInSite": 0, "idGroupAccess": null,
    		"idGroupOwned": "32", "idGroupSelf": "31", "idUserGodfather": null, "loginID": null, "loginModulePrefix": null,
    		"nbHelpGiven": 0, "nbSpacesForTab": 3, "sAddress": null, "sBirthDate": null, "sCellPhoneNumber": null,
    		"sCity": null, "sCountryCode": "", "sDefaultLanguage": "fr", "sEmail": null, "sFirstName": "Jane",
    		"sFreeText": null, "sLandLineNumber": null, "sLangProg": "Python", "sLastActivityDate": null, "sLastIP": null,
    		"sLastLoginDate": null, "sLastName": "Doe", "sLogin": "jane", "sNotificationReadDate": null, "sNotify": "Answers",
    		"sOpenIdIdentity": null, "sPasswordMd5": null, "sRecover": null, "sRegistrationDate": null, "sSalt": null,
    		"sSex": null, "sStudentId": null, "sTimeZone": null, "sWebSite": null, "sZipcode": null, "tempUser": 0
    	},
    	"groups_attempts": [],
    	"groups_groups": [],
    	"joined_groups": [],
    	"owned_groups": [],
    	"refresh_token": null,
    	"sessions": [
    		{
    			"idUser": "4", "sAccessToken": "***", "sExpirationDate": "2019-07-17T00:02:28Z",
    			"sIssuedAtDate": "2019-07-16T22:02:28Z", "sIssuer": null
    		}
    	],
    	"users_answers": [],
    	"users_items": []
    }
    """
