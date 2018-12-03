Feature: Not found fallback

Scenario: A request to a not found url is redirected to the fallback server
Given a server is running as fallback
When I send a GET request to "/non-existing-path"
Then the response code should be 200
And the response header "X-Got-Query" should be "/non-existing-path"
