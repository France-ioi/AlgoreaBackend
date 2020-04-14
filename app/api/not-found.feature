Feature: Not found fallback

Scenario: A request to a not found url is redirected to the fallback server
When I send a GET request to "/non-existing-path"
Then the response code should be 404
And the response body should be, in JSON:
"""
{
  "success": false,
  "message": "Not Found"
}
"""
