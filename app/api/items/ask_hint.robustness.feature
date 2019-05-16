Feature: Ask for a hint - robustness
  Background:
    Given the database has the following table 'users':
      | ID  | sLogin | idGroupSelf |
      | 10  | john   | 101         |
    And the database has the following table 'groups':
      | ID  |
      | 101 |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 101             | 101          | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate |
      | 15 | 22            | 13           | direct             | null        |
    And the database has the following table 'platforms':
      | ID | bUsesTokens | sRegexp                                           | sPublicKey                |
      | 10 | 1           | http://taskplatform.mblockelet.info/task.html\?.* | {{taskPlatformPublicKey}} |
    And the database has the following table 'items':
      | ID | idPlatform | sUrl                                                                    | bReadOnly |
      | 50 | 10         | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | 1         |
      | 10 | 10         | http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936 | 0         |
    And the database has the following table 'items_items':
      | idItemParent | idItemChild |
      | 10           | 50          |
    And the database has the following table 'items_ancestors':
      | idItemAncestor | idItemChild |
      | 10             | 50          |
    And the database has the following table 'users_items':
      | idUser | idItem | sHintsRequested                 | nbHintsCached | nbSubmissionsAttempts | idAttemptActive |
      | 10     | 10     | null                            | 0             | 0                     | null            |
      | 10     | 50     | [{"rotorIndex":0,"cellRank":0}] | 12            | 2                     | 100             |
    And time is frozen

  Scenario: Wrong JSON in request
    Given I am the user with ID "10"
    When I send a POST request to "/items/ask_hint" with the following body:
      """
      []
      """
    Then the response code should be 400
    And the response error message should contain "Json: cannot unmarshal array into Go value of type items.askHintRequestWrapper"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: User not found
    Given I am the user with ID "404"
    When I send a POST request to "/items/ask_hint" with the following body:
      """
      {
        "task_token": {{generateToken(map(
          "idUser", "404",
          "idItemLocal", "50",
          "platformName", app().TokenConfig.PlatformName,
        ), app().TokenConfig.PrivateKey)}},
        "hint_requested": {{generateToken(map(
          "idUser", "404",
	        "itemURL", "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
	        "askedHint", `{"rotorIndex":1,"cellRank":1}`,
        ), taskPlatformPrivateKey)}}
      }
      """
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: idUser in task_token doesn't match the user's ID
    Given I am the user with ID "10"
    When I send a POST request to "/items/ask_hint" with the following body:
      """
      {
        "task_token": {{generateToken(map(
          "idUser", "20",
          "idItemLocal", "50",
          "platformName", app().TokenConfig.PlatformName,
        ), app().TokenConfig.PrivateKey)}},
        "hint_requested": {{generateToken(map(
          "idUser", "10",
	        "itemURL", "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
	        "askedHint", `{"rotorIndex":1,"cellRank":1}`,
        ), taskPlatformPrivateKey)}}
      }
      """
    Then the response code should be 400
    And the response error message should contain "Token in task_token doesn't correspond to user session: got idUser=20, expected 10"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: idUser in hint_requested doesn't match the user's ID
    Given I am the user with ID "10"
    When I send a POST request to "/items/ask_hint" with the following body:
      """
      {
        "task_token": {{generateToken(map(
          "idUser", "10",
          "idItemLocal", "50",
          "platformName", app().TokenConfig.PlatformName,
        ), app().TokenConfig.PrivateKey)}},
        "hint_requested": {{generateToken(map(
          "idUser", "20",
	        "itemURL", "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
	        "askedHint", `{"rotorIndex":1,"cellRank":1}`,
        ), taskPlatformPrivateKey)}}
      }
      """
    Then the response code should be 400
    And the response error message should contain "Token in hint_requested doesn't correspond to user session: got idUser=20, expected 10"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: No submission rights
    Given I am the user with ID "10"
    When I send a POST request to "/items/ask_hint" with the following body:
      """
      {
        "task_token": {{generateToken(map(
          "idUser", "10",
          "idItemLocal", "50",
          "platformName", app().TokenConfig.PlatformName,
        ), app().TokenConfig.PrivateKey)}},
        "hint_requested": {{generateToken(map(
          "idUser", "10",
	        "itemURL", "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
	        "askedHint", `{"rotorIndex":1,"cellRank":1}`,
        ), taskPlatformPrivateKey)}}
      }
      """
    Then the response code should be 403
    And the response error message should contain "Item is read-only"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged

  Scenario: idAttempt not found
    Given I am the user with ID "10"
    When I send a POST request to "/items/ask_hint" with the following body:
      """
      {
        "task_token": {{generateToken(map(
          "idUser", "10",
          "idItemLocal", "10",
          "idAttempt", "100",
          "platformName", app().TokenConfig.PlatformName,
        ), app().TokenConfig.PrivateKey)}},
        "hint_requested": {{generateToken(map(
          "idUser", "10",
	        "itemURL", "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
	        "askedHint", `{"rotorIndex":1,"cellRank":1}`,
        ), taskPlatformPrivateKey)}}
      }
      """
    Then the response code should be 404
    And the response error message should contain "Can't find previously requested hints info"
    And the table "users_items" should stay unchanged
    And the table "groups_attempts" should stay unchanged
