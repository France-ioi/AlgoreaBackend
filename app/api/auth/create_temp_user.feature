Feature: Create a temporary user

  Background:
    Given the application config is:
      """
      domains:
        -
          domains: [127.0.0.1]
          rootGroup: 1
          rootSelfGroup: 2
          rootAdminGroup: 3
          rootTempGroup: 4
      """
    And the database has the following table 'groups':
      | ID | sName      | sType     | sTextId   |
      | 1  | Root       | Base      | Root      |
      | 2  | RootSelf   | Base      | RootSelf  |
      | 3  | RootAdmin  | Base      | RootAdmin |
      | 4  | RootTemp   | UserSelf  | RootTemp  |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | iChildOrder |
      | 1  | 1             | 2            | 1           |
      | 2  | 1             | 3            | 2           |
      | 3  | 2             | 4            | 1           |

  Scenario: Create a new temporary user
    Given the generated auth key is "ny93zqri9a2adn4v1ut6izd76xb3pccw"
    When I send a POST request to "/auth/temp-user"
    Then the response code should be 201
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "created",
        "data": {"access_token": "ny93zqri9a2adn4v1ut6izd76xb3pccw", "expires_in": 7200}
      }
      """
    And logs should contain:
      """
      Generated a session token "ny93zqri9a2adn4v1ut6izd76xb3pccw" expiring in 7200 seconds for a temporary user 5577006791947779410
      """
    And the table "users" at ID "5577006791947779410" should be:
      | ID                  | loginID | sLogin       | tempUser | ABS(TIMESTAMPDIFF(SECOND, sRegistrationDate, NOW())) < 3 | idGroupSelf         | idGroupOwned | sLastIp   |
      | 5577006791947779410 | 0       | tmp-49727887 | true     | true                                                     | 6129484611666145821 | null         | 127.0.0.1 |
    And the table "groups" should stay unchanged but the row with ID "6129484611666145821"
    And the table "groups" at ID "6129484611666145821" should be:
      | ID                  | sName        | sType    | sDescription | ABS(TIMESTAMPDIFF(SECOND, sDateCreated, NOW())) < 3 | bOpened | bSendEmails |
      | 6129484611666145821 | tmp-49727887 | UserSelf | tmp-49727887 | true                                                | false   | false       |
    And the table "groups_groups" should stay unchanged but the row with ID "4037200794235010051"
    And the table "groups_groups" at ID "4037200794235010051" should be:
      | ID                  | idGroupParent | idGroupChild        | iChildOrder |
      | 4037200794235010051 | 4             | 6129484611666145821 | 1           |
    And the table "groups_ancestors" should be:
      | idGroupAncestor     | idGroupChild        | bIsSelf |
      | 1                   | 1                   | true    |
      | 1                   | 2                   | false   |
      | 1                   | 3                   | false   |
      | 1                   | 4                   | false   |
      | 1                   | 6129484611666145821 | false   |
      | 2                   | 2                   | true    |
      | 2                   | 4                   | false   |
      | 2                   | 6129484611666145821 | false   |
      | 3                   | 3                   | true    |
      | 4                   | 4                   | true    |
      | 4                   | 6129484611666145821 | false   |
      | 6129484611666145821 | 6129484611666145821 | true    |
    And the table "sessions" should be:
      | sAccessToken                     | ABS(TIMESTAMPDIFF(SECOND, NOW(), sExpirationDate) - 7200) < 3 | idUser              | ABS(TIMESTAMPDIFF(SECOND, NOW(), sIssuedAtDate)) < 3 | sIssuer |
      | ny93zqri9a2adn4v1ut6izd76xb3pccw | true                                                          | 5577006791947779410 | true                                                 | backend |
