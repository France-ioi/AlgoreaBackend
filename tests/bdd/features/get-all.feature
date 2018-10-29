
Feature: Get all groups

Scenario: No filtering
Given the database has the following table 'groups':
  | ID |     sName | sTextId | sType | iVersion |
  |  1 | testGroup |       0 | Other |        0 |
When I make a GET /groups/
Then it should be a JSON array with 1 entry