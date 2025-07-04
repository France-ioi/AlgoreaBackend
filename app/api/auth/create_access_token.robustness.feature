Feature: Login callback - robustness
  Scenario: Should be an error when create_temp_user_if_not_authorized is not a boolean
    When I send a POST request to "/auth/token?create_temp_user_if_not_authorized=invalid"
    Then the response code should be 400
    And the response error message should contain "Wrong value for create_temp_user_if_not_authorized (should have a boolean value (0 or 1))"

  Scenario: Both code and Authorization header are present
    Given the "Authorization" request header is "Bearer 1234567890"
    When I send a POST request to "/auth/token?code=somecode"
    Then the response code should be 400
    And the response error message should contain "Only one of the 'code' parameter and the 'Authorization' header can be given"
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario: Should be an error when nor code given, nor auth token given, and we don't want to create a temporary user
    When I send a POST request to "/auth/token"
    Then the response code should be 401
    And the response error message should contain "No access token provided"
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario: Should be an error when no code given, and auth token is invalid (could have expired), and we don't want to create a temporary user
    When I send a POST request to "/auth/token"
    And the "Authorization" request header is "invalid"
    Then the response code should be 401
    And the response error message should contain "No access token provided"
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario: Invalid JSON data
    Given the "Content-Type" request header is "application/json"
    When I send a POST request to "/auth/token" with the following body:
    """
    code=1234
    """
    Then the response code should be 400
    And the response error message should contain "Invalid character 'c' looking for beginning of value"
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario: Invalid form data
    Given the "Content-Type" request header is "application/x-www-form-urlencoded"
    When I send a POST request to "/auth/token" with the following body:
    """
    %%%%
    """
    Then the response code should be 400
    And the response error message should contain "Invalid URL escape "%%%""
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario: Invalid request content type
    Given the "Content-Type" request header is "application/xml"
    When I send a POST request to "/auth/token" with the following body:
    """
    <code>1234</code>
    """
    Then the response code should be 415
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario: OAuth error
    Given the DB time now is "2019-07-16 22:02:28"
    And the login module "token" endpoint for code "somecode" returns 500 with body:
      """
      Unknown error
      """
    When I send a POST request to "/auth/token?code=somecode"
    Then the response code should be 500
    And the response error message should contain "Unknown error"
    And logs should contain:
      """
      oauth2: cannot fetch token: 500\nResponse: Unknown error
      """
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario: User API error
    Given the DB time now is "2019-07-16 22:02:28"
    And the login module "token" endpoint for code "somecode" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":31622420,
        "access_token":"accesstoken",
        "refresh_token":"refreshtoken"
      }
      """
    And the login module "account" endpoint for token "accesstoken" returns 500 with body:
      """
      Unknown error
      """
    When I send a POST request to "/auth/token?code=somecode"
    Then the response code should be 500
    And the response error message should contain "Unknown error"
    And logs should contain:
      """
      {{ quote(`Can't retrieve user's profile (status code = 500, response = "Unknown error")`) }}
      """
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario: User profile can't be parsed
    Given the DB time now is "2019-07-16 22:02:28"
    And the login module "token" endpoint for code "somecode" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":31622420,
        "access_token":"accesstoken",
        "refresh_token":"refreshtoken"
      }
      """
    And the login module "account" endpoint for token "accesstoken" returns 200 with body:
      """
      Not a JSON
      """
    When I send a POST request to "/auth/token?code=somecode"
    Then the response code should be 500
    And the response error message should contain "Unknown error"
    And logs should contain:
      """
      {{ quote(`Can't parse user's profile (response = "Not a JSON", error = "invalid character 'N' looking for beginning of value")`)}}
      """
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged

  Scenario Outline: User profile is invalid
    Given the DB time now is "2019-07-16 22:02:28"
    And the login module "token" endpoint for code "somecode" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":31622420,
        "access_token":"accesstoken",
        "refresh_token":"refreshtoken"
      }
      """
    And the login module "account" endpoint for token "accesstoken" returns 200 with body:
      """
      <profile_body>
      """
    When I send a POST request to "/auth/token?code=somecode"
    Then the response code should be 500
    And the response error message should contain "Unknown error"
    And logs should contain:
      """
      {{ quote(`User's profile is invalid (response = ` + quote(`<profile_body>`) + `, error = "<error_text>")`) }}
      """
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged
  Examples:
    | profile_body      | error_text                 |
    | {"login":"login"} | no id in user's profile    |
    | {"id":12345}      | no login in user's profile |

  Scenario Outline: Invalid cookie attributes
    Given I send a POST request to "/auth/token<query>"
    Then the response code should be 400
    And the response error message should contain "<expected_error>"
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged
  Examples:
    | query                                            | expected_error                                                                 |
    | ?use_cookie=1                                    | One of cookie_secure and cookie_same_site must be true when use_cookie is true |
    | ?use_cookie=1&cookie_same_site=0&cookie_secure=0 | One of cookie_secure and cookie_same_site must be true when use_cookie is true |
    | ?use_cookie=abc                                  | Wrong value for use_cookie (should have a boolean value (0 or 1))              |
    | ?cookie_same_site=abc                            | Wrong value for cookie_same_site (should have a boolean value (0 or 1))        |
    | ?cookie_secure=abc                               | Wrong value for cookie_secure (should have a boolean value (0 or 1))           |

  Scenario: Should be an error when the login module returns an expired access token
    Given the time now is "2019-07-16T22:02:28Z"
    And the template constant "code_from_oauth" is "somecode"
    And the template constant "access_token_from_oauth" is "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImp0aSI6Ijc3M2IyMjY0ZDU0MDUzNWQ5OTFlMjNlODY0MzljNzJmYjI0MWI5ZWY1ZTI5NjMyYjc3OWQwNjdlNmJmZWRiYmUyZDM4NmQ4YmQ2OTBlNGI3In0.eyJhdWQiOiIxIiwianRpIjoiNzczYjIyNjRkNTQwNTM1ZDk5MWUyM2U4NjQzOWM3MmZiMjQxYjllZjVlMjk2MzJiNzc5ZDA2N2U2YmZlZGJiZTJkMzg2ZDhiZDY5MGU0YjciLCJpYXQiOjE1NjM1Mjk4MjUsIm5iZiI6MTU2MzUyOTgyNSwiZXhwIjoxNTk1MTUyMjI0LCJzdWIiOiIxMDAwMDAwMDEiLCJzY29wZXMiOlsiYWNjb3VudCJdfQ.hcMLfoK8ocb0dpJg-R6EViMePCE4uw_Zzid_CIzFMFT6khY7m1kLorzKgYLWbDBxyxG-RBWTjJIbE-0J96VvLegYoZo5JObHzZP_FQyOUQ-qVe98mjI3Mc0a-dmr5bQyPTS2OC2COlFnletMHhBe4D_DSh2Zi8TfN79kTjsYErN59Vc4Bz0sPPmnLRqdKbg8r6jVX-s6cidN8mgDjujAljiaPkjCCiumdMj9kSfTKLNxMu1e9-4GfN41xc72ikstcBXjvakTyeq2-M9Wcby4XA5fys313kKlKQy3WJAVW3D6qMEwRH566vesEIx-RWUIlkPyV4QvIaE3k4mKdiO6c21LSFFSlIfr6jkVaGDvi8Rc9g77CWgUXaZOsETliW0Yea0tL9fG1negRr9uQGKyOZCM1dxSlBJAKlD3kyLi4ykEw6uTp0tM-AdwRB7mUpu9bw3evpr7f0mN65Nhd-byAuys0PXyegZeSKxZB3i1mAzE6s7vUbADJcBOx0kRmfkpT3kfUkJ4c9QohVCpkIMl80sbxcv9RTck0P9W1J-LGUULTtcPeaLNz85q7DKKbdiTAcbqzQkxZn0hO2wrF-3L0p_ms-yQg8ebu-ZJIzUG5LQq6Szu-QpXyQPP3NdKqHEvMhKoFY-9BZwA9SCEfiB8kMwCm9TAfztZBiCRcS2I4LE"
    And the template constant "refresh_token_from_oauth" is "def502008be6565fe7888139650994031dcf475fd4ec863d9d088562aeff095c4fb5026d189b05385b5d6e834bb26ed98d67b19f21c8e4f70e035083b8aba36027c748eb0a8fc987b900a96734eb3952733d8d87368cbf5194195dfee364ebe774117dc8e51075ea7afe356d985021a38be505ea7328137d0f3552dcf4ed1b7187affee3399964b81d396a597fb9ef78c1651c5203529cd016a9c9584fc024e597e47327c36431981000741c8e6e24066718b3b46d6278a0f13b0d1bd87e2811269a2464b832b765f45d40a878ce4d3bc9da03aad32dc6f17caa52f67befffd89bae734ac0b424d9a32bd2e47c47dfee43e534d36d6cc180759b3d220ddea18ba70d8490501934e960a9ad99012184fcd67f471a16c65db5185f24ace83857efefdd935280cc0a9653150d89f9ca531283ec9e566592de626d0c350ddd682f59ede69f29acfb0bc3104d826afabd0f1e1a246375154c78a9ad27a2c47bde5159686a4264bd91f16ffa185554d09858402a68"
    And the login module "token" endpoint for code "{{code_from_oauth}}" and code_verifier "123456" and redirect_uri "http://my.url" returns 200 with body:
      """
      {
        "token_type":"Bearer",
        "expires_in":9,
        "access_token":"{{access_token_from_oauth}}",
        "refresh_token":"{{refresh_token_from_oauth}}"
      }
      """
    When I send a POST request to "/auth/token?code={{code_from_oauth}}&code_verifier=123456&redirect_uri=http%3A%2F%2Fmy.url"
    Then the response code should be 401
    And the response error message should contain "Got an invalid OAuth2 token"
    And the response header "Set-Cookie" should not be set
    And the table "users" should remain unchanged
    And the table "groups" should remain unchanged
    And the table "groups_groups" should remain unchanged
    And the table "groups_ancestors" should remain unchanged
    And the table "attempts" should remain unchanged
    And the table "group_membership_changes" should remain unchanged
    And the table "sessions" should remain unchanged
    And the table "access_tokens" should remain unchanged
    And the table "group_managers" should remain unchanged
