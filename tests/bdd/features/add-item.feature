Feature: Add item

Scenario: Basic
When I send a POST request to "/items/" with the following body:
  """
  {
    "id": 2,
    "type": "Course",
    "strings": [
      { "language_id": 3, "title": "my title" }
    ],
    "parents": [
      { "id": 4, "order": 100 }
    ]
  }
  """
Then the response code should be 200
And the table "items" should be:
  | ID | sType  | sUrl |
  | 2  | Course | NULL |
And the table "items_strings" should be:
  | ID | idItem  | idLanguage | sTitle   |
  | 1  | 2       |          3 | my title |
