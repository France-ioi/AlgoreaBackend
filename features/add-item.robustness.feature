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
And the response error message should contain "must be given"

Scenario: No strings given
When I send a POST request to "/items/" with the following body:
  """
  {
    "id": 2,
    "type": "Course",
    "strings": [
    ],
    "parents": [
      { "id": 4, "order": 100 }
    ]
  }
  """
Then the response code should be 400
And the response error message should contain "one string per item"

Scenario: More than one string given
When I send a POST request to "/items/" with the following body:
  """
  {
    "id": 2,
    "type": "Course",
    "strings": [
      { "language_id": 3, "title": "my title" },
      { "language_id": 1, "title": "mon titre" }
    ],
    "parents": [
      { "id": 4, "order": 100 }
    ]
  }
  """
Then the response code should be 400
And the response error message should contain "one string per item"

Scenario: No parents given
When I send a POST request to "/items/" with the following body:
  """
  {
    "id": 2,
    "type": "Course",
    "strings": [
      { "language_id": 3, "title": "my title" }
    ],
    "parents": [
    ]
  }
  """
Then the response code should be 400
And the response error message should contain "one parent item"

Scenario: More than 1 parent given
When I send a POST request to "/items/" with the following body:
  """
  {
    "id": 2,
    "type": "Course",
    "strings": [
      { "language_id": 3, "title": "my title" }
    ],
    "parents": [
      { "id": 4, "order": 100 },
      { "id": 3, "order": 101 }
    ]
  }
  """
Then the response code should be 400
And the response error message should contain "one parent item"
