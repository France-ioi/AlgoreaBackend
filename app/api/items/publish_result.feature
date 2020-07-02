Feature: Publish a result to LTI
  Background:
    Given the database has the following table 'groups':
      | id  | type  |
      | 21  | User  |
      | 31  | User  |
      | 99  | Class |
      | 100 | Team  |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 99              | 21             |
      | 100             | 31             |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id  | default_language_tag |
      | 123 | fr                   |
      | 124 | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 21       | 123     | content                  |
      | 99       | 124     | content_with_descendants |
    And the database has the following table 'attempts':
      | id  | participant_id |
      | 0   | 21             |
      | 1   | 21             |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | score_computed |
      | 0          | 21             | 123     | 12.3           |
      | 1          | 21             | 123     | 15.6           |
      | 1          | 21             | 124     | 20.1           |
      | 1          | 31             | 123     | 9.5            |
    And the database has the following table 'users':
      | temp_user | login | group_id | login_id |
      | 0         | john  | 21       | 1234567  |
      | 1         | jane  | 31       | null     |
    And the application config is:
      """
      auth:
        loginModuleURL: "https://login.algorea.org"
        clientID: "1"
        clientSecret: "tzxsLyFtJiGnmD6sjZMqSEidVpVsL3hEoSxIXCpI"
      """

  Scenario: The user has a result on the item
    Given I am the user with id "21"
    And the login module "lti_result/send" endpoint for user id "1234567", content id "123", score "15.6" returns 200 with encoded body:
      """
      {"success":true}
      """
    When I send a POST request to "/items/123/attempts/1/publish"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "published"
      }
      """

  Scenario: The user has no results on the item inside the attempt
    Given I am the user with id "21"
    And the login module "lti_result/send" endpoint for user id "1234567", content id "124", score "0" returns 200 with encoded body:
      """
      {"success":true}
      """
    When I send a POST request to "/items/124/attempts/0/publish"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": true,
        "message": "published"
      }
      """

  Scenario: Login module fails
    Given I am the user with id "21"
    And the login module "lti_result/send" endpoint for user id "1234567", content id "123", score "15.6" returns 200 with encoded body:
      """
      {"success":false}
      """
    When I send a POST request to "/items/123/attempts/1/publish"
    Then the response code should be 200
    And the response body should be, in JSON:
      """
      {
        "success": false,
        "message": "failed"
      }
      """
