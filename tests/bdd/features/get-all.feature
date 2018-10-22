
Feature: Get all groups

Scenario: No filtering
Given the database has the following table 'groups':
  | id |
  |  1 |
When I make a GET /groups/
Then it should be a JSON array with 1 entry