Feature: Path Not found

Scenario: A request to a path without service returns a 404 error
When I send a GET request to "/non-existing-path"
Then the response code should be 404
And the response body should be, in JSON:
"""
{
  "success": false,
  "message": "Not Found"
}
"""
