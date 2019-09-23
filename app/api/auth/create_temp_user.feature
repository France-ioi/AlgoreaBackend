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
      | id | name       | type      | text_id   |
      | 1  | Root       | Base      | Root      |
      | 2  | RootSelf   | Base      | RootSelf  |
      | 3  | RootAdmin  | Base      | RootAdmin |
      | 4  | RootTemp   | UserSelf  | RootTemp  |
    And the database has the following table 'groups_groups':
      | id | parent_group_id | child_group_id | child_order |
      | 1  | 1               | 2              | 1           |
      | 2  | 1               | 3              | 2           |
      | 3  | 2               | 4              | 1           |

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
      Generated a session token expiring in 7200 seconds for a temporary user 5577006791947779410
      """
    And the table "users" at id "5577006791947779410" should be:
      | id                  | login_id | login        | temp_user | ABS(TIMESTAMPDIFF(SECOND, registration_date, NOW())) < 3 | self_group_id       | owned_group_id | last_ip   |
      | 5577006791947779410 | 0        | tmp-49727887 | true      | true                                                     | 6129484611666145821 | null           | 127.0.0.1 |
    And the table "groups" should stay unchanged but the row with id "6129484611666145821"
    And the table "groups" at id "6129484611666145821" should be:
      | id                  | name         | type     | description  | ABS(TIMESTAMPDIFF(SECOND, date_created, NOW())) < 3 | opened | send_emails |
      | 6129484611666145821 | tmp-49727887 | UserSelf | tmp-49727887 | true                                                | false  | false       |
    And the table "groups_groups" should stay unchanged but the row with id "4037200794235010051"
    And the table "groups_groups" at id "4037200794235010051" should be:
      | id                  | parent_group_id | child_group_id      | child_order |
      | 4037200794235010051 | 4               | 6129484611666145821 | 1           |
    And the table "groups_ancestors" should be:
      | ancestor_group_id   | child_group_id      | is_self |
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
      | access_token                     | ABS(TIMESTAMPDIFF(SECOND, NOW(), expiration_date) - 7200) < 3 | user_id             | ABS(TIMESTAMPDIFF(SECOND, NOW(), issued_at_date)) < 3 | issuer  |
      | ny93zqri9a2adn4v1ut6izd76xb3pccw | true                                                          | 5577006791947779410 | true                                                  | backend |
