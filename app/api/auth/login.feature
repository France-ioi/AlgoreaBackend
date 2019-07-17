Feature: Generate a login state, set the cookie, and redirect to the auth url
  Scenario: Successful redirect
    Given the generated auth random strings are "o5yuy6wmpe607bknrmvrrduy5xe60zd7","ny93zqri9a2adn4v1ut6izd76xb3pccw"
    And the time now is "2019-07-16T22:02:29Z"
    And the DB time now is "2019-07-16T22:02:29Z"
    And the database has the following table 'login_states':
      | sCookie                          | sState                           | sExpirationDate     |
      | asdfasdflaw923rkaskdfa92jffjflak | 9023raslkdk92jf293favbasdf23892a | 2019-07-16T22:02:29 |
      | zfjpi09i2j9fj29fg28ghf298fnksds9 | 29fn29fnj2jnf92nkkfn92nf92nf9u22 | 2019-07-16T22:02:30 |
    When I send a POST request to "/auth/login"
    Then the response code should be 302
    And the response header "Location" should be "http://127.0.0.1:8000/oauth/authorize?approval_prompt=auto&client_id=1&redirect_uri=http%3A%2F%2F127.0.0.1%3A8080%2Fauth%2Flogin-callback&response_type=code&scope=account&state=o5yuy6wmpe607bknrmvrrduy5xe60zd7"
    And the response header "Set-Cookie" should be "login_csrf=ny93zqri9a2adn4v1ut6izd76xb3pccw; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:29 GMT; Max-Age=7200; HttpOnly"
    # old cookies are deleted
    And the table "login_states" should be:
      | sCookie                          | sState                           | ABS(TIMESTAMPDIFF(SECOND, NOW(), sExpirationDate) - 7200) < 3 |
      | ny93zqri9a2adn4v1ut6izd76xb3pccw | o5yuy6wmpe607bknrmvrrduy5xe60zd7 | true                                                          |
      | zfjpi09i2j9fj29fg28ghf298fnksds9 | 29fn29fnj2jnf92nkkfn92nf92nf9u22 | false                                                         |

    When the generated auth random strings are "23ri9n9ii92nflsdkn9ii92923dafadf","o2fjooij0ilasdjf02ijsadlkfjj0ksc"
    And the time now is "2019-07-16T22:02:30Z"
    And the DB time now is "2019-07-16T22:02:30Z"
    When I send a POST request to "/auth/login"
    Then the response code should be 302
    And the response header "Location" should be "http://127.0.0.1:8000/oauth/authorize?approval_prompt=auto&client_id=1&redirect_uri=http%3A%2F%2F127.0.0.1%3A8080%2Fauth%2Flogin-callback&response_type=code&scope=account&state=23ri9n9ii92nflsdkn9ii92923dafadf"
    And the response header "Set-Cookie" should be "login_csrf=o2fjooij0ilasdjf02ijsadlkfjj0ksc; Path=/; Domain=127.0.0.1; Expires=Wed, 17 Jul 2019 00:02:30 GMT; Max-Age=7200; HttpOnly"
    # doesn't delete old cookies on each request
    But the table "login_states" should be:
      | sCookie                          | sState                           | ABS(TIMESTAMPDIFF(SECOND, NOW(), sExpirationDate) - 7200) < 3 |
      | ny93zqri9a2adn4v1ut6izd76xb3pccw | o5yuy6wmpe607bknrmvrrduy5xe60zd7 | true                                                          |
      | o2fjooij0ilasdjf02ijsadlkfjj0ksc | 23ri9n9ii92nflsdkn9ii92923dafadf | true                                                          |
      | zfjpi09i2j9fj29fg28ghf298fnksds9 | 29fn29fnj2jnf92nkkfn92nf92nf9u22 | false                                                         |
