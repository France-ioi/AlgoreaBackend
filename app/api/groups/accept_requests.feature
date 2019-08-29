Feature: Accept group requests
  Background:
    Given the database has the following table 'users':
      | ID | sLogin | idGroupSelf | idGroupOwned | sFirstName  | sLastName | iGrade |
      | 1  | owner  | 21          | 22           | Jean-Michel | Blanquer  | 3      |
    And the database has the following table 'groups':
      | ID  |
      | 11  |
      | 13  |
      | 14  |
      | 21  |
      | 22  |
      | 31  |
      | 111 |
      | 121 |
      | 122 |
      | 123 |
      | 131 |
      | 141 |
      | 151 |
    And the database has the following table 'groups_ancestors':
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 13           | 1       |
      | 13              | 111          | 0       |
      | 13              | 121          | 0       |
      | 13              | 123          | 0       |
      | 13              | 151          | 0       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 13           | 0       |
      | 22              | 22           | 1       |
      | 31              | 31           | 1       |
      | 111             | 111          | 1       |
      | 121             | 121          | 1       |
      | 122             | 122          | 1       |
      | 123             | 123          | 1       |
      | 151             | 151          | 1       |
    And the database has the following table 'groups_groups':
      | ID | idGroupParent | idGroupChild | sType              | sStatusDate               |
      | 1  | 13            | 21           | invitationSent     | {{relativeTime("-170h")}} |
      | 2  | 13            | 11           | invitationRefused  | {{relativeTime("-169h")}} |
      | 3  | 13            | 31           | requestSent        | {{relativeTime("-168h")}} |
      | 5  | 14            | 11           | invitationSent     | null                      |
      | 6  | 14            | 31           | invitationRefused  | null                      |
      | 7  | 14            | 21           | requestSent        | null                      |
      | 8  | 14            | 22           | requestRefused     | null                      |
      | 9  | 13            | 121          | invitationAccepted | 2017-05-29T06:38:38Z      |
      | 10 | 13            | 111          | requestAccepted    | null                      |
      | 11 | 13            | 131          | removed            | null                      |
      | 12 | 13            | 122          | left               | null                      |
      | 13 | 13            | 123          | direct             | null                      |
      | 14 | 13            | 141          | requestSent        | null                      |
      | 15 | 22            | 13           | direct             | null                      |
      | 16 | 13            | 151          | joinedByCode       | null                      |

  Scenario: Accept requests
    Given I am the user with ID "1"
    When I send a POST request to "/groups/13/requests/accept?group_ids=31,141,21,11,13,22,151"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    {
      "data": {
        "141": "success",
        "31": "success",
        "11": "invalid",
        "13": "invalid",
        "21": "invalid",
        "22": "invalid",
        "151": "invalid"
      },
      "message": "updated",
      "success": true
    }
    """
    And the table "groups_groups" should stay unchanged but the row with ID "3,14"
    And the table "groups_groups" at ID "3,14" should be:
      | ID | idGroupParent | idGroupChild | sType              | (sStatusDate IS NOT NULL) AND (ABS(TIMESTAMPDIFF(SECOND, sStatusDate, NOW())) < 3) |
      | 3  | 13            | 31           | requestAccepted    | 1                                                                                  |
      | 14 | 13            | 141          | requestAccepted    | 1                                                                                  |
    And the table "groups_ancestors" should be:
      | idGroupAncestor | idGroupChild | bIsSelf |
      | 11              | 11           | 1       |
      | 13              | 13           | 1       |
      | 13              | 31           | 0       |
      | 13              | 111          | 0       |
      | 13              | 121          | 0       |
      | 13              | 123          | 0       |
      | 13              | 141          | 0       |
      | 13              | 151          | 0       |
      | 14              | 14           | 1       |
      | 21              | 21           | 1       |
      | 22              | 13           | 0       |
      | 22              | 22           | 1       |
      | 22              | 31           | 0       |
      | 22              | 111          | 0       |
      | 22              | 121          | 0       |
      | 22              | 123          | 0       |
      | 22              | 141          | 0       |
      | 22              | 151          | 0       |
      | 31              | 31           | 1       |
      | 111             | 111          | 1       |
      | 121             | 121          | 1       |
      | 122             | 122          | 1       |
      | 123             | 123          | 1       |
      | 131             | 131          | 1       |
      | 141             | 141          | 1       |
      | 151             | 151          | 1       |
