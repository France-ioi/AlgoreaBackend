Feature: Add item

Background:
  Given the database has the following table 'users':
    | ID | sLogin | tempUser | idGroupSelf | iVersion |
    | 1  | jdoe   | 0        | 11          | 0        |
  And the database has the following table 'groups':
    | ID | sName      | sTextId | iGrade | sType     | iVersion |
    | 11 | jdoe       |         | -2     | UserAdmin | 0        |
  And the database has the following table 'items':
    | ID | bTeamsEditable | bNoScore | iVersion |
    | 21 | false          | false    | 0        |
  And the database has the following table 'groups_items':
    | ID | idGroup | idItem | bManagerAccess | idUserCreated | iVersion |
    | 41 | 11      | 21     | true           | 0             | 0        |
  And the database has the following table 'groups_ancestors':
    | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
    | 71 | 11              | 11           | 1       | 0        |

Scenario: Valid, id is given
  Given I am the user with ID "1"
  And the time now is "2018-01-01T00:00:00Z"
  When I send a POST request to "/items/" with the following body:
    """
    {
      "id": 2,
      "type": "Course",
      "strings": [
        { "language_id": 3, "title": "my title", "image_url":"http://bit.ly/1234", "subtitle": "hard task", "description": "the goal of this task is ..." }
      ],
      "parents": [
        { "id": 21, "order": 100 }
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
  And the table "items" at ID "2" should be:
    | ID | sType  | sUrl | idDefaultLanguage | bTeamsEditable | bNoScore |
    |  2 | Course | null |                 3 |              0 |        0 |
  And the table "items_strings" should be:
    |                  ID | idItem  | idLanguage |   sTitle |          sImageUrl | sSubtitle |                 sDescription |
    | 8674665223082153551 |      2  |          3 | my title | http://bit.ly/1234 | hard task | the goal of this task is ... |
  And the table "items_items" should be:
    |                  ID | idItemParent | idItemChild | iChildOrder |
    | 6129484611666145821 |           21 |           2 |         100 |
  And the table "groups_items" at ID "5577006791947779410" should be:
    |                  ID | idGroup | idItem | idUserCreated |     sFullAccessDate  |  bOwnerAccess | bManagerAccess | sCachedFullAccessDate | bCachedFullAccess |
    | 5577006791947779410 |      11 |      2 |             1 | 2018-01-01T00:00:00Z |             1 |              1 |  2018-01-01T00:00:00Z |                 1 |

Scenario: Id not given
  Given I am the user with ID "1"
  When I send a POST request to "/items/" with the following body:
    """
    {
      "type": "Course",
      "strings": [
        { "language_id": 3, "title": "my title" }
      ],
      "parents": [
        { "id": 21, "order": 100 }
      ]
    }
    """
  Then the response code should be 201
  And the table "items" at ID "5577006791947779410" should be:
    | sType  | sUrl |
    | Course | null |
