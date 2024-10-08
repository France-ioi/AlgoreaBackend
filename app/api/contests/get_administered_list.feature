Feature: Get the contests that the user has administration rights on (contestAdminList)
  Background:
    Given the database has the following users:
      | login          | group_id | default_language |
      | possesseur     | 21       | fr               |
      | owner          | 31       | en               |
      | administrateur | 41       | fr               |
      | admin          | 51       | en               |
      | guest          | 61       | en               |
      | panas          | 71       | uk               |
    And the database has the following table "groups":
      | id | type  | name       |
      | 80 | Other | Some group |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 80              | 21             |
    And the groups ancestors are computed
    And the database has the following table "languages":
      | tag |
      | en  |
      | fr  |
      | sl  |
    And the database has the following table "items":
      | id | duration | default_language_tag | allows_multiple_attempts | entry_participant_type | requires_explicit_entry |
      | 10 | 00:00:02 | en                   | 0                        | Team                   | true                    |
      | 40 | 00:00:00 | fr                   | 0                        | Team                   | false                   |
      | 50 | 00:00:00 | fr                   | 0                        | Team                   | true                    |
      | 60 | 00:00:01 | en                   | 1                        | User                   | true                    |
      | 70 | 00:00:03 | fr                   | 0                        | User                   | true                    |
      | 80 | 00:00:03 | sl                   | 0                        | Team                   | true                    |
      | 90 | 00:00:03 | sl                   | 0                        | User                   | true                    |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order |
      | 10             | 60            | 0           |
      | 10             | 70            | 1           |
      | 60             | 70            | 0           |
      | 90             | 80            | 0           |
    And the database has the following table "items_strings":
      | item_id | language_tag | title      |
      | 10      | en           | Chapter    |
      | 10      | fr           | Chapitre   |
      | 50      | fr           | null       |
      | 60      | en           | Contest    |
      | 70      | en           | Contest 2  |
      | 70      | fr           | Concours 2 |
      | 80      | sl           | null       |
      | 90      | sl           | null       |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated       | can_grant_view_generated | can_watch_generated |
      | 21       | 40      | solution                 | enter                    | result              |
      | 21       | 50      | solution                 | enter                    | result              |
      | 21       | 60      | content                  | none                     | none                |
      | 21       | 70      | content_with_descendants | content                  | answer              |
      | 21       | 80      | content                  | none                     | answer              |
      | 21       | 90      | info                     | enter                    | answer              |
      | 31       | 50      | solution                 | enter                    | result              |
      | 31       | 60      | content_with_descendants | content                  | answer              |
      | 31       | 70      | content                  | enter                    | result              |
      | 31       | 80      | content_with_descendants | enter                    | none                |
      | 41       | 10      | content                  | none                     | none                |
      | 41       | 50      | solution                 | enter                    | result              |
      | 41       | 60      | solution                 | enter                    | result              |
      | 41       | 70      | content_with_descendants | enter                    | result              |
      | 51       | 10      | info                     | none                     | none                |
      | 51       | 50      | solution                 | enter                    | result              |
      | 51       | 60      | solution                 | enter                    | result              |
      | 51       | 70      | content_with_descendants | enter                    | result              |
      | 71       | 80      | content_with_descendants | enter                    | result              |
      | 71       | 90      | content_with_descendants | enter                    | result              |
      | 80       | 60      | none                     | enter                    | result              |

  Scenario: User's default language is French (most parents are invisible)
    Given I am the user with id "21"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "allows_multiple_attempts": false, "entry_participant_type": "Team", "parents": [], "title": null, "language_tag": "fr"},
      {"id": "70", "allows_multiple_attempts": false, "entry_participant_type": "User", "parents": [{"title": "Contest", "language_tag": "en"}],
       "title": "Concours 2", "language_tag": "fr"},
      {"id": "60", "allows_multiple_attempts": true, "entry_participant_type": "User", "parents": [], "title": "Contest", "language_tag": "en"}
    ]
    """

  Scenario: User's default language is English  (most parents are invisible)
    Given I am the user with id "31"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "allows_multiple_attempts": false, "entry_participant_type": "Team", "parents": [], "title": null, "language_tag": "fr"},
      {"id": "60", "allows_multiple_attempts": true, "entry_participant_type": "User", "parents": [], "title": "Contest", "language_tag": "en"},
      {"id": "70", "allows_multiple_attempts": false, "entry_participant_type": "User", "parents": [{"title": "Contest", "language_tag": "en"}],
       "title": "Contest 2", "language_tag": "en"}
    ]
    """

  Scenario: User's default language is French (parents are visible)
    Given I am the user with id "41"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "allows_multiple_attempts": false, "entry_participant_type": "Team", "parents": [], "title": null, "language_tag": "fr"},
      {"id": "70", "allows_multiple_attempts": false, "entry_participant_type": "User",
       "parents": [{"title": "Chapitre", "language_tag": "fr"}, {"title": "Contest", "language_tag": "en"}],
       "title": "Concours 2", "language_tag": "fr"},
      {"id": "60", "allows_multiple_attempts": true, "entry_participant_type": "User", "parents": [{"title": "Chapitre", "language_tag": "fr"}],
       "title": "Contest", "language_tag": "en"}
    ]
    """

  Scenario: User's default language is English  (parents are visible)
    Given I am the user with id "51"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "allows_multiple_attempts": false, "entry_participant_type": "Team", "parents": [], "title": null, "language_tag": "fr"},
      {"id": "60", "allows_multiple_attempts": true, "entry_participant_type": "User", "parents": [{"title": "Chapter", "language_tag": "en"}],
       "title": "Contest", "language_tag": "en"},
      {"id": "70", "allows_multiple_attempts": false, "entry_participant_type": "User",
       "parents": [{"title": "Chapter", "language_tag": "en"}, {"title": "Contest", "language_tag": "en"}],
       "title": "Contest 2", "language_tag": "en"}
    ]
    """

  Scenario: Empty result
    Given I am the user with id "61"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
    ]
    """

  Scenario: User's default language is English  (parents are visible), limit=1
    Given I am the user with id "51"
    When I send a GET request to "/contests/administered?limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "50", "allows_multiple_attempts": false, "entry_participant_type": "Team", "parents": [], "title": null, "language_tag": "fr"}
    ]
    """

  Scenario: User's default language is English  (parents are visible), start from the second row, limit=1
    Given I am the user with id "51"
    When I send a GET request to "/contests/administered?from.id=50&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "60", "allows_multiple_attempts": true, "entry_participant_type": "User", "parents": [{"title": "Chapter", "language_tag": "en"}],
       "title": "Contest", "language_tag": "en"}
    ]
    """

  Scenario: User's default language is English  (parents are visible), inverse order
    Given I am the user with id "51"
    When I send a GET request to "/contests/administered?sort=-title,id"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "70", "allows_multiple_attempts": false, "entry_participant_type": "User",
       "parents": [{"title": "Chapter", "language_tag": "en"}, {"title": "Contest", "language_tag": "en"}],
       "title": "Contest 2", "language_tag": "en"},
      {"id": "60", "allows_multiple_attempts": true, "entry_participant_type": "User", "parents": [{"title": "Chapter", "language_tag": "en"}],
       "title": "Contest", "language_tag": "en"},
      {"id": "50", "allows_multiple_attempts": false, "entry_participant_type": "Team", "parents": [], "title": null, "language_tag": "fr"}
    ]
    """

  Scenario: Keeps parents with nil titles
    Given I am the user with id "71"
    When I send a GET request to "/contests/administered"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {"id": "80", "allows_multiple_attempts": false, "entry_participant_type": "Team", "parents": [{"language_tag": "sl", "title": null}],
       "title": null, "language_tag": "sl"},
      {"id": "90", "allows_multiple_attempts": false, "entry_participant_type": "User", "parents": [], "title": null, "language_tag": "sl"}
    ]
    """
