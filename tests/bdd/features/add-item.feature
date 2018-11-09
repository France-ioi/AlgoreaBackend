Feature: Add item

Scenario: Basic
When I send a POST request to "/items/" with the following body:
  """
  {
    "id": 1,
    "type": "Course",
    "strings": [
      { "LanguageID": 1, "Title": "my title" }
    ],
    "parents": [
      { "ID": 3, "Order": 100 }
    ]
  }
  """
Then the response code should be 200
And the table "items" should be:
  | ID | sType  | sUrl |
  | 1  | Course | NULL |
