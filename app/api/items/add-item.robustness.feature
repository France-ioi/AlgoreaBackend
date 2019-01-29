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
And the response error message should contain "Wrong value for 'type': must be given"

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
And the response error message should contain "Exactly one string per item is supported"

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
And the response error message should contain "Exactly one string per item is supported"

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
And the response error message should contain "Exactly one parent item is supported"

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
And the response error message should contain "Exactly one parent item is supported"

Scenario: Not existing parent
Given the database has the following table 'users':
  | ID | sLogin | tempUser | idGroupSelf | iVersion |
  | 1  | jdoe   | 0        | 11          | 0        |
And I am the user with ID "1"
When I send a POST request to "/items/" with the following body:
  """
  {
    "id": 2,
    "type": "Course",
    "strings": [
      { "language_id": 3, "title": "my title" }
    ],
    "parents": [
      { "id": 21, "order": 100 }
    ]
  }
  """
Then the response code should be 403
And the response error message should contain "Cannot find the parent item"

Scenario: Not enough perm on parent
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
  | 41 | 11      | 21     | false           | 0             | 0        |
And the database has the following table 'groups_ancestors':
  | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
  | 71 | 11              | 11           | 1       | 0        |
And I am the user with ID "1"
When I send a POST request to "/items/" with the following body:
  """
  {
    "id": 2,
    "type": "Course",
    "strings": [
      { "language_id": 3, "title": "my title" }
    ],
    "parents": [
      { "id": 21, "order": 100 }
    ]
  }
  """
Then the response code should be 403
And the response error message should contain "Insufficient access on the parent item"
