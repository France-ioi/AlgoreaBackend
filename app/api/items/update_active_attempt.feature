Feature: Update active attempt for an item
  Background:
    Given the database has the following table 'users':
      | ID  | sLogin | idGroupSelf |
      | 10  | john   | 101         |
      | 11  | jane   | 111         |
      | 12  | jack   | 121         |
    And the database has the following table 'groups':
      | ID  | idTeamItem | sType    |
      | 101 | null       | UserSelf |
      | 102 | 10         | Team     |
      | 111 | null       | UserSelf |
    And the database has the following table 'groups_groups':
      | idGroupParent | idGroupChild | sType              |
      | 102           | 101          | invitationAccepted |
      | 102           | 121          | joinedByCode       |
      | 103           | 101          | requestAccepted    |
      | 104           | 101          | direct             |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 101             | 101          | 1       |
      | 102             | 101          | 0       |
      | 102             | 102          | 1       |
      | 102             | 121          | 0       |
      | 111             | 111          | 1       |
      | 121             | 121          | 1       |
    And the database has the following table 'items':
      | ID | sUrl                                                                    | sType   | bHasAttempts |
      | 10 | null                                                                    | Chapter | 0            |
      | 50 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Course  | 0            |
      | 60 | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | Task    | 1            |
    And the database has the following table 'items_ancestors':
      | idItemAncestor | idItemChild |
      | 10             | 60          |
    And the database has the following table 'items_items':
      | idItemParent | idItemChild | iChildOrder |
      | 10           | 60          | 0           |
    And the database has the following table 'groups_items':
      | idGroup | idItem | sCachedPartialAccessDate | sCachedFullAccessDate | idUserCreated |
      | 101     | 50     | 2017-05-29 06:38:38      | null                  | 10            |
      | 101     | 60     | 2017-05-29 06:38:38      | null                  | 10            |
      | 111     | 50     | null                     | 2017-05-29 06:38:38   | 10            |
      | 121     | 50     | null                     | 2017-05-29 06:38:38   | 10            |

  Scenario: User is able to update an active attempt (full access)
    Given I am the user with ID "11"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | sLastActivityDate   |
      | 11     | 50     | null            | 2017-05-29 06:38:38 |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sLastActivityDate   | iOrder |
      | 100 | 111     | 50     | 2017-05-29 06:38:38 | 0      |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | idAttemptActive | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 11     | 50     | 100             | done                       | 1                                                        |
    And the table "groups_attempts" should be:
      | ID  | idGroup | idItem | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 100 | 111     | 50     | done                       | 1                                                        |

  Scenario: User is able to fetch an active attempt (partial access)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | sLastActivityDate   |
      | 10     | 50     | null            | 2017-05-29 06:38:38 |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sLastActivityDate   | iOrder |
      | 100 | 101     | 50     | 2017-05-29 06:38:38 | 0      |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | idAttemptActive | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 10     | 50     | 100             | done                       | 1                                                        |
    And the table "groups_attempts" should be:
      | ID  | idGroup | idItem | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 100 | 101     | 50     | done                       | 1                                                        |

  Scenario: User is able to update an active attempt (full access, groups_groups.sType=joinedByCode)
    Given I am the user with ID "11"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | sLastActivityDate   |
      | 11     | 50     | null            | 2017-05-29 06:38:38 |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sLastActivityDate   | iOrder |
      | 100 | 111     | 50     | 2017-05-29 06:38:38 | 0      |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | idAttemptActive | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 11     | 50     | 100             | done                       | 1                                                        |
    And the table "groups_attempts" should be:
      | ID  | idGroup | idItem | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 100 | 111     | 50     | done                       | 1                                                        |

  Scenario: User is able to update an active attempt (bHasAttempts=1, groups_groups.sType=invitationAccepted)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | sLastActivityDate   |
      | 10     | 10     | null            | 2018-05-29 06:38:38 |
      | 10     | 60     | null            | 2017-05-29 06:38:38 |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sLastActivityDate   | iOrder |
      | 200 | 102     | 60     | 2017-05-29 06:38:38 | 0      |
    When I send a PUT request to "/attempts/200/active"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | idAttemptActive | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 10     | 10     | null            | done                       | 1                                                        |
      | 10     | 60     | 200             | done                       | 1                                                        |
    And the table "groups_attempts" should be:
      | ID  | idGroup | idItem | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 200 | 102     | 60     | done                       | 1                                                        |

  Scenario: User is able to update an active attempt (bHasAttempts=1, groups_groups.sType=requestAccepted)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | sLastActivityDate   |
      | 10     | 10     | null            | 2018-05-29 06:38:38 |
      | 10     | 60     | null            | 2017-05-29 06:38:38 |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sLastActivityDate   | iOrder |
      | 200 | 103     | 60     | 2017-05-29 06:38:38 | 0      |
    When I send a PUT request to "/attempts/200/active"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | idAttemptActive | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 10     | 10     | null            | done                       | 1                                                        |
      | 10     | 60     | 200             | done                       | 1                                                        |
    And the table "groups_attempts" should be:
      | ID  | idGroup | idItem | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 200 | 103     | 60     | done                       | 1                                                        |

  Scenario: User is able to update an active attempt (bHasAttempts=1, groups_groups.sType=direct)
    Given I am the user with ID "10"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | sLastActivityDate   |
      | 10     | 10     | null            | 2018-05-29 06:38:38 |
      | 10     | 60     | null            | 2017-05-29 06:38:38 |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sLastActivityDate   | iOrder |
      | 200 | 104     | 60     | 2017-05-29 06:38:38 | 0      |
    When I send a PUT request to "/attempts/200/active"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | idAttemptActive | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 10     | 10     | null            | done                       | 1                                                        |
      | 10     | 60     | 200             | done                       | 1                                                        |
    And the table "groups_attempts" should be:
      | ID  | idGroup | idItem | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 200 | 104     | 60     | done                       | 1                                                        |

  Scenario: User is able to update an active attempt when this attempt is already active
    Given I am the user with ID "11"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | sLastActivityDate   |
      | 11     | 50     | 100             | 2017-05-29 06:38:38 |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sLastActivityDate   | iOrder |
      | 100 | 111     | 50     | 2017-05-29 06:38:38 | 0      |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | idAttemptActive | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 11     | 50     | 100             | done                       | 1                                                        |
    And the table "groups_attempts" should be:
      | ID  | idGroup | idItem | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 100 | 111     | 50     | done                       | 1                                                        |


  Scenario: User is able to update an active attempt when another attempt is active
    Given I am the user with ID "11"
    And the database has the following table 'users_items':
      | idUser | idItem | idAttemptActive | sLastActivityDate   |
      | 11     | 50     | 101             | 2017-05-29 06:38:38 |
    And the database has the following table 'groups_attempts':
      | ID  | idGroup | idItem | sLastActivityDate   | iOrder |
      | 100 | 111     | 50     | 2017-05-29 06:38:38 | 0      |
      | 101 | 111     | 50     | 2018-05-29 06:38:38 | 1      |
    When I send a PUT request to "/attempts/100/active"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "message": "updated",
        "success": true
      }
      """
    And the table "users_items" should be:
      | idUser | idItem | idAttemptActive | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 11     | 50     | 100             | done                       | 1                                                        |
    And the table "groups_attempts" should be:
      | ID  | idGroup | idItem | sAncestorsComputationState | ABS(TIMESTAMPDIFF(SECOND, sLastActivityDate, NOW())) < 3 |
      | 100 | 111     | 50     | done                       | 1                                                        |
      | 101 | 111     | 50     | done                       | 0                                                        |
