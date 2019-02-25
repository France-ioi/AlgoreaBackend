Feature: Get recent activity for group_id and item_id - robustness
  Background:
    Given the database has the following table 'users':
      | ID | sLogin  | tempUser | idGroupSelf | idGroupOwned | sFirstName  | sLastName |
      | 1  | someone | 0        | 21          | 22           | Bill        | Clinton   |
      | 2  | user    | 0        | 11          | 12           | John        | Doe       |
      | 3  | owner   | 0        | 23          | 24           | Jean-Michel | Blanquer  |
    And the database has the following table 'groups_ancestors':
      | ID | idGroupAncestor | idGroupChild | bIsSelf | iVersion |
      | 75 | 24              | 13           | 0       | 0        |
      | 76 | 13              | 11           | 0       | 0        |
      | 77 | 22              | 11           | 0       | 0        |
      | 78 | 21              | 21           | 1       | 0        |
      | 79 | 23              | 23           | 1       | 0        |
    And the database has the following table 'users_answers':
      | ID | idUser | idItem | idAttempt | sName            | sType      | sState  | sLangProg | sSubmissionDate     | iScore | bValidated |
      | 1  | 2      | 200    | 100       | My answer        | Submission | Current | python    | 2017-05-29 06:38:38 | 100    | true       |
      | 2  | 2      | 200    | 101       | My second anwser | Submission | Current | python    | 2017-05-29 06:38:38 | 100    | true       |
    And the database has the following table 'items':
      | ID  | sType    | bTeamsEditable | bNoScore | idItemUnlocked | bTransparentFolder | iVersion |
      | 200 | Category | false          | false    | 1234,2345      | true               | 0        |
    And the database has the following table 'groups_items':
      | ID | idGroup | idItem | sFullAccessDate | bCachedFullAccess | bCachedPartialAccess | bCachedGrayedAccess | idUserCreated | iVersion |
      | 43 | 21      | 200    | null            | true              | true                 | true                | 0             | 0        |
      | 44 | 23      | 200    | null            | false             | false                | false               | 0             | 0        |
    And the database has the following table 'items_ancestors':
      | ID | idItemAncestor | idItemChild | iVersion |
      | 1  | 200            | 200         | 0        |

  Scenario: Should fail when user is not an admin of the group
    Given I am the user with ID "1"
    When I send a GET request to "/groups/13/recent_activity?item_id=200"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: Should return empty array when user is an admin of the group, but has no access rights to the item
    Given I am the user with ID "3"
    When I send a GET request to "/groups/13/recent_activity?item_id=200"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: Should fail when from.id is given, but from.submission_date is not
    Given I am the user with ID "3"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&from.id=1"
    Then the response code should be 400
    And the response error message should contain "Both from.id and from.submission_date or none of them must be present"

  Scenario: Should fail when from.submission_date is given, but from.id is not
    Given I am the user with ID "3"
    When I send a GET request to "/groups/13/recent_activity?item_id=200&from.submission_date=2017-05-30T06:38:38Z"
    Then the response code should be 400
    And the response error message should contain "Both from.id and from.submission_date or none of them must be present"
