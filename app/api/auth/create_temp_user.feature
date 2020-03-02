Feature: Create a temporary user

  Background:
    Given the application config is:
      """
      domains:
        -
          domains: [127.0.0.1]
          rootGroup: 1
          rootSelfGroup: 2
          rootTempGroup: 4
      """
    And the database has the following table 'groups':
      | id | name     | type | text_id  |
      | 1  | Root     | Base | Root     |
      | 2  | RootSelf | Base | RootSelf |
      | 4  | RootTemp | User | RootTemp |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 1               | 2              |
      | 2               | 4              |

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
      Generated a session token expiring in 7200 seconds for a temporary user with group_id = 5577006791947779410
      """
    And the table "users" at group_id "5577006791947779410" should be:
      | group_id            | login_id | login        | temp_user | ABS(TIMESTAMPDIFF(SECOND, registered_at, NOW())) < 3 | last_ip   |
      | 5577006791947779410 | 0        | tmp-49727887 | true      | true                                                 | 127.0.0.1 |
    And the table "groups" should stay unchanged but the row with id "5577006791947779410"
    And the table "groups" at id "5577006791947779410" should be:
      | id                  | name         | type | description  | ABS(TIMESTAMPDIFF(SECOND, created_at, NOW())) < 3 | is_open | send_emails |
      | 5577006791947779410 | tmp-49727887 | User | tmp-49727887 | true                                              | false   | false       |
    And the table "groups_groups" should stay unchanged but the row with child_group_id "5577006791947779410"
    And the table "groups_groups" at child_group_id "5577006791947779410" should be:
      | parent_group_id | child_group_id      |
      | 4               | 5577006791947779410 |
    And the table "groups_ancestors" should be:
      | ancestor_group_id   | child_group_id      | is_self |
      | 1                   | 1                   | true    |
      | 1                   | 2                   | false   |
      | 1                   | 4                   | false   |
      | 1                   | 5577006791947779410 | false   |
      | 2                   | 2                   | true    |
      | 2                   | 4                   | false   |
      | 2                   | 5577006791947779410 | false   |
      | 4                   | 4                   | true    |
      | 4                   | 5577006791947779410 | false   |
      | 5577006791947779410 | 5577006791947779410 | true    |
    And the table "sessions" should be:
      | access_token                     | ABS(TIMESTAMPDIFF(SECOND, NOW(), expires_at) - 7200) < 3 | user_id             | ABS(TIMESTAMPDIFF(SECOND, NOW(), issued_at)) < 3 | issuer  |
      | ny93zqri9a2adn4v1ut6izd76xb3pccw | true                                                     | 5577006791947779410 | true                                             | backend |
