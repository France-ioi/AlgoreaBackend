Feature: Add item

Scenario:
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
  |  2 | Course | NULL |
And the table "items_strings" should be:
  | ID | idItem  | idLanguage |   sTitle |
  |  1 |      2  |          3 | my title |
And the table "items_items" should be:
  | ID | idItemParent | idItemChild | iChildOrder |
  |  1 |            4 |           2 |         100 |
And the table "groups_items" should be:
  | ID | idGroup | idItem |     sFullAccessDate | bCachedFullAccess | bOwnerAccess | idUserCreated |
  |  1 |       6 |      2 | 2018-01-01 00:00:00 |                 1 |            1 |             9 |
