
Feature: Get all groups

Scenario: No filtering
Given the database has the following table 'groups':
  | ID |     sName | sTextId | sType | iVersion |
  |  1 | testGroup |       0 | Other |        0 |
When I send a GET request to "/groups/"
Then the response code should be 200
And the response should match json:
  """
  [
    { "id": 1, "name": "testGroup" }
  ]
  """
