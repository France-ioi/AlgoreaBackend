Feature: Add item - robustness

Scenario: Missing required attribute
When I send a POST request to "/items/" with the following body:
  """
  {
    "id": 2,
    "strings": [
      { "language_id": 3, "title": "my title" }
    ],
    "parents": [
      { "id": 4, "order": 100 }
    ]
  }
  """
Then the response code should be 400
