Feature: Add item

Scenario: Valid, id is given
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
Then the response code should be 201
And the response body should be, in JSON:
"""
{
  "success": true,
  "message": "success",
  "data": { "ID": 2 }
}
"""
And the table "items" should be:
  | ID | sType  | sUrl |
  |  2 | Course | NULL |
And the table "items_strings" should be:
  |                  ID | idItem  | idLanguage |   sTitle |
  | 8674665223082153551 |      2  |          3 | my title |
And the table "items_items" should be:
  |                  ID | idItemParent | idItemChild | iChildOrder |
  | 6129484611666145821 |            4 |           2 |         100 |
And the table "groups_items" should be:
  |                  ID | idGroup | idItem |     sFullAccessDate | bCachedFullAccess | bOwnerAccess | idUserCreated |
  | 5577006791947779410 |       6 |      2 | 2018-01-01 00:00:00 |                 0 |            0 |             0 |

Scenario: Id not given
When I send a POST request to "/items/" with the following body:
  """
  {
    "type": "Course",
    "strings": [
      { "language_id": 3, "title": "my title" }
    ],
    "parents": [
      { "id": 4, "order": 100 }
    ]
  }
  """
Then the response code should be 201
And the table "items" should be:
  |                  ID | sType  | sUrl |
  | 5577006791947779410 | Course | NULL |
