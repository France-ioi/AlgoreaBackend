Feature: Fetch active attempt for an item
  Background:
    Given the database has the following table 'users':
      | ID  | sLogin | idGroupSelf |
      | 10  | john   | 101         |
      | 11  | jane   | 111         |
    And the database has the following table 'groups':
      | ID  | idTeamItem | sType    |
      | 101 | null       | UserSelf |
      | 102 | 10         | Team     |
      | 111 | null       | UserSelf |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild |
      | 102           | 101          |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 101             | 101          | 1       |
      | 102             | 101          | 0       |
      | 102             | 102          | 1       |
      | 111             | 111          | 1       |
    And the database has the following table 'items':
      | ID | sUrl                                                                    | sType   | bHasAttempts | bHintsAllowed | sTextId | sSupportedLangProg |
      | 10 | null                                                                    | Chapter | 0            | 0             | null    | null               |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | 0            | 1             | task1   | null               |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course  | 1            | 0             | null    | c,python           |
    And the database has the following table 'items_ancestors':
      | idItemAncestor | idItemChild |
      | 10             | 60          |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate | sCachedFullAccessDate | sCachedAccessSolutionsDate |
      | 101     | 50     | 2017-05-29T06:38:38Z     | null                  | null                       |
      | 101     | 60     | 2017-05-29T06:38:38Z     | null                  | 2017-05-29T06:38:38Z       |
      | 111     | 50     | null                     | 2017-05-29T06:38:38Z  | null                       |
    And time is frozen

  Scenario: User is able to fetch an active attempt (no active attempt set)
    Given I am the user with ID "11"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | iScore | sBestAnswerDate | sValidationDate | sStartDate | sHintsRequested | nbHintsCached |
      | 11     | 50     | null            | 0      | null            | null            | null       | 1,2,3           | 3             |
    When I send a PUT request to "/items/50/active-attempt"
    Then the response code should be 200
    And the response body decoded as "FetchActiveAttemptResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "bAccessSolutions": false,
            "bHintsAllowed": true,
            "bIsAdmin": false,
            "bReadAnswers": true,
            "bSubmissionPossible": true,
            "idAttempt": "8674665223082153551",
            "idUser": "11",
            "idItemLocal": "50",
            "idItem": "task1",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "nbHintsGiven": "0",
            "randomSeed": "8674665223082153551",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          }
        },
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | ABS(sStartDate - NOW()) < 3 |
      | 11     | 50     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 1                           |
    And the table "groups_attempts" should be:
      | ID                  | idGroup | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | ABS(sStartDate - NOW()) < 3 |
      | 8674665223082153551 | 111     | 50     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 1                           |

  Scenario: User is able to fetch an active attempt (no active attempt set, only full access)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | iScore | sBestAnswerDate | sValidationDate | sStartDate |
      | 10     | 50     | null            | 0      | null            | null            | null       |
    When I send a PUT request to "/items/50/active-attempt"
    Then the response code should be 200
    And the response body decoded as "FetchActiveAttemptResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "bAccessSolutions": false,
            "bHintsAllowed": true,
            "bIsAdmin": false,
            "bReadAnswers": true,
            "bSubmissionPossible": true,
            "idAttempt": "8674665223082153551",
            "idUser": "10",
            "idItem": "task1",
            "idItemLocal": "50",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "nbHintsGiven": "0",
            "randomSeed": "8674665223082153551",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          }
        },
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | ABS(sStartDate - NOW()) < 3 |
      | 10     | 50     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 1                           |
    And the table "groups_attempts" should be:
      | ID                  | idGroup | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | ABS(sStartDate - NOW()) < 3 |
      | 8674665223082153551 | 101     | 50     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 1                           |

  Scenario: User is able to fetch an active attempt (no active attempt and item.bHasAttempts=1)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | iScore | sBestAnswerDate | sValidationDate | sStartDate |
      | 10     | 60     | null            | 0      | null            | null            | null       |
    When I send a PUT request to "/items/60/active-attempt"
    Then the response code should be 200
    And the response body decoded as "FetchActiveAttemptResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "bAccessSolutions": true,
            "bHintsAllowed": false,
            "bIsAdmin": false,
            "bReadAnswers": true,
            "bSubmissionPossible": true,
            "idAttempt": "8674665223082153551",
            "idUser": "10",
            "idItemLocal": "60",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "nbHintsGiven": "0",
            "sSupportedLangProg": "c,python",
            "randomSeed": "8674665223082153551",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          }
        },
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | ABS(sStartDate - NOW()) < 3 |
      | 10     | 60     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 1                           |
    And the table "groups_attempts" should be:
      | ID                  | idGroup | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | ABS(sStartDate - NOW()) < 3 |
      | 8674665223082153551 | 102     | 60     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 1                           |

  Scenario: User is able to fetch an active attempt (with active attempt set)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | iScore | sBestAnswerDate | sValidationDate | sStartDate |
      | 10     | 50     | 100             | 0      | null            | null            | null       |
    When I send a PUT request to "/items/50/active-attempt"
    Then the response code should be 200
    And the response body decoded as "FetchActiveAttemptResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "bAccessSolutions": false,
            "bHintsAllowed": true,
            "bIsAdmin": false,
            "bReadAnswers": true,
            "bSubmissionPossible": true,
            "idAttempt": "100",
            "idUser": "10",
            "idItem": "task1",
            "idItemLocal": "50",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "nbHintsGiven": "0",
            "randomSeed": "100",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          }
        },
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | ABS(sStartDate - NOW()) < 3 |
      | 10     | 50     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 1                           |
    And the table "groups_attempts" should stay unchanged

  Scenario: User is able to fetch an active attempt (no active attempt set, but there are some in the DB)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | iScore | sBestAnswerDate | sValidationDate | sStartDate |
      | 10     | 50     | null            | 0      | null            | null            | null       |
    And the database has the following table 'groups_attempts':
      | ID | idGroup | idItem | sLastActivityDate    | sStartDate | iScore | sBestAnswerDate | sValidationDate | sHintsRequested | nbHintsCached |
      | 1  | 101     | 50     | 2017-05-29T06:38:38Z | null       | 0      | null            | null            | null            | 0             |
      | 2  | 101     | 50     | 2018-05-29T06:38:38Z | null       | 0      | null            | null            | [1,2,3,4]       | 4             |
      | 3  | 102     | 50     | 2019-05-29T06:38:38Z | null       | 0      | null            | null            | null            | 0             |
      | 4  | 101     | 51     | 2019-04-29T06:38:38Z | null       | 0      | null            | null            | null            | 0             |
    When I send a PUT request to "/items/50/active-attempt"
    Then the response code should be 200
    And the response body decoded as "FetchActiveAttemptResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "bAccessSolutions": false,
            "bHintsAllowed": true,
            "bIsAdmin": false,
            "bReadAnswers": true,
            "bSubmissionPossible": true,
            "idAttempt": "2",
            "idUser": "10",
            "idItemLocal": "50",
            "idItem": "task1",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "nbHintsGiven": "4",
            "sHintsRequested": "[1,2,3,4]",
            "randomSeed": "2",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          }
        },
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | ABS(sStartDate - NOW()) < 3 |
      | 10     | 50     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 1                           |
    And the table "groups_attempts" should stay unchanged but the row with ID "2"
    And the table "groups_attempts" at ID "2" should be:
      | ID | idGroup | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | ABS(sStartDate - NOW()) < 3 |
      | 2  | 101     | 50     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 1                           |

  Scenario: User is able to fetch an active attempt (no active attempt set, but there are some in the DB and items.bHasAttempts=1)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | iScore | sBestAnswerDate | sValidationDate | sStartDate |
      | 10     | 60     | null            | 0      | null            | null            | null       |
    And the database has the following table 'groups_attempts':
      | ID | idGroup | idItem | sLastActivityDate    | sStartDate | iScore | sBestAnswerDate | sValidationDate | sHintsRequested | nbHintsCached |
      | 1  | 102     | 60     | 2017-05-29T06:38:38Z | null       | 0      | null            | null            | null            | 0             |
      | 2  | 102     | 60     | 2018-05-29T06:38:38Z | null       | 0      | null            | null            | [1,2,3,4]       | 4             |
      | 3  | 101     | 60     | 2019-05-29T06:38:38Z | null       | 0      | null            | null            | null            | 0             |
      | 4  | 102     | 61     | 2019-04-29T06:38:38Z | null       | 0      | null            | null            | null            | 0             |
    When I send a PUT request to "/items/60/active-attempt"
    Then the response code should be 200
    And the response body decoded as "FetchActiveAttemptResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "bAccessSolutions": true,
            "bHintsAllowed": false,
            "bIsAdmin": false,
            "bReadAnswers": true,
            "bSubmissionPossible": true,
            "idAttempt": "2",
            "idUser": "10",
            "idItemLocal": "60",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "nbHintsGiven": "4",
            "sHintsRequested": "[1,2,3,4]",
            "sSupportedLangProg": "c,python",
            "randomSeed": "2",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          }
        },
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | ABS(sStartDate - NOW()) < 3 |
      | 10     | 60     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 1                           |
    And the table "groups_attempts" should stay unchanged but the row with ID "2"
    And the table "groups_attempts" at ID "2" should be:
      | ID | idGroup | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | ABS(sStartDate - NOW()) < 3 |
      | 2  | 102     | 60     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 1                           |

  Scenario: Keeps previous sStartDate values
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | iScore | sBestAnswerDate | sValidationDate | sStartDate           |
      | 10     | 50     | null            | 0      | null            | null            | 2017-05-29T06:38:38Z |
    And the database has the following table 'groups_attempts':
      | ID | idGroup | idItem | sLastActivityDate    | sStartDate           | iScore | sBestAnswerDate | sValidationDate |
      | 2  | 101     | 50     | 2018-05-29T06:38:38Z | 2017-05-29T06:38:38Z | 0      | null            | null            |
    When I send a PUT request to "/items/50/active-attempt"
    Then the response code should be 200
    And the response body decoded as "FetchActiveAttemptResponse" should be, in JSON:
      """
      {
        "data": {
          "task_token": {
            "date": "{{currentTimeInFormat("02-01-2006")}}",
            "bAccessSolutions": false,
            "bHintsAllowed": true,
            "bIsAdmin": false,
            "bReadAnswers": true,
            "bSubmissionPossible": true,
            "idAttempt": "2",
            "idUser": "10",
            "idItemLocal": "50",
            "idItem": "task1",
            "itemUrl": "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
            "nbHintsGiven": "0",
            "randomSeed": "2",
            "platformName": "{{app().TokenConfig.PlatformName}}"
          }
        },
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | sStartDate           |
      | 10     | 50     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 2017-05-29T06:38:38Z |
    And the table "groups_attempts" should stay unchanged but the row with ID "2"
    And the table "groups_attempts" at ID "2" should be:
      | ID | idGroup | idItem | iScore | nbTasksTried | bValidated | bKeyObtained | sAncestorsComputationState | ABS(sLastActivityDate - NOW()) < 3 | ABS(sLastAnswerDate - NOW()) < 3 | ABS(sBestAnswerDate - NOW()) < 3 | ABS(sValidationDate - NOW()) < 3 | sStartDate           |
      | 2  | 101     | 50     | 0      | 0            | 0          | 0            | done                       | 1                                  | null                             | null                             | null                             | 2017-05-29T06:38:38Z |

