Feature: List official sessions for item_id
  Background:
    Given the database has the following table 'groups':
      | id | name          | description | type    | is_official_session | is_public | root_activity_id | require_lock_membership_approval_until | require_personal_info_access_approval | require_watch_approval | require_members_to_join_parent | open_activity_when_joining | expected_start      | organizer              | address_city | address_country | address_line1               | address_line2        | address_postcode |
      | 11 | jdoe          | null        | User    | false               | false     | null             | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 13 | Group B       | null        | Class   | false               | false     | null             | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 21 | other         | null        | User    | false               | false     | null             | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 23 | Group C       | null        | Team    | false               | true      | null             | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 50 | Session 200   | null        | Session | true                | true      | 200              | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 51 | Session 200.1 | null        | Session | false               | true      | null             | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 52 | Session 200.2 | null        | Session | true                | false     | null             | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 53 | Session 200.3 | Official    | Session | true                | true      | 200              | 2019-05-30 11:00:00                    | view                                  | true                   | true                           | true                       | 2019-05-30 12:00:00 | Association France-ioi | Paris        | France          | Chez Jacques-Henri Jourdan, | 42, rue de Cronstadt | 75015            |
      | 60 | Session 210   | null        | Session | true                | true      | 210              | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 70 | Parent3       | null        | Other   | false               | true      | null             | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 71 | Parent        | null        | Club    | false               | false     | null             | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 72 | Parent        | null        | Club    | false               | false     | null             | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 73 | Parent2       | null        | Friends | false               | false     | null             | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
      | 80 | Group         | null        | Friends | false               | false     | null             | null                                   | none                                  | false                  | false                          | false                      | null                | null                   | null         | null            | null                        | null                 | null             |
    And the database has the following table 'users':
      | login | group_id | first_name | last_name |
      | jdoe  | 11       | John       | Doe       |
      | other | 21       | George     | Bush      |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id |
      | 13              | 11             |
      | 13              | 21             |
      | 23              | 21             |
      | 23              | 31             |
      | 53              | 11             |
      | 70              | 50             |
      | 71              | 50             |
      | 72              | 50             |
      | 72              | 11             |
      | 73              | 50             |
      | 80              | 71             |
    And the groups ancestors are computed
    And the database has the following table 'group_managers':
      | manager_id | group_id |
      | 13         | 80       |
    And the database has the following table 'items':
      | id  | allows_multiple_attempts | default_language_tag |
      | 200 | 0                        | fr                   |
      | 210 | 1                        | fr                   |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 13       | 200     | info                     |
      | 13       | 210     | content                  |
      | 23       | 210     | content_with_descendants |

  Scenario Outline: User has access to the item
    Given I am the user with id "11"
    When I send a GET request to "/items/200/official-sessions<sort>"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "53",
        "name": "Session 200.3",
        "address_city": "Paris",
        "address_country": "France",
        "address_line1": "Chez Jacques-Henri Jourdan,",
        "address_line2": "42, rue de Cronstadt",
        "address_postcode": "75015",
        "current_user_is_manager": false,
        "current_user_is_member": true,
        "description": "Official",
        "expected_start": "2019-05-30T12:00:00Z",
        "is_public": true,
        "open_activity_when_joining": true,
        "organizer": "Association France-ioi",
        "parents": [],
        "require_lock_membership_approval_until": "2019-05-30T11:00:00Z",
        "require_members_to_join_parent": true,
        "require_personal_info_access_approval": "view",
        "require_watch_approval": true
      },
      {
        "group_id": "50",
        "name": "Session 200",
        "address_city": null,
        "address_country": null,
        "address_line1": null,
        "address_line2": null,
        "address_postcode": null,
        "current_user_is_manager": true,
        "current_user_is_member": false,
        "description": null,
        "expected_start": null,
        "is_public": true,
        "open_activity_when_joining": false,
        "organizer": null,
        "parents": [
          {"current_user_is_manager": true, "current_user_is_member": false, "id": "71", "is_public": false, "name": "Parent"},
          {"current_user_is_manager": false, "current_user_is_member": true, "id": "72", "is_public": false, "name": "Parent"},
          {"current_user_is_manager": false, "current_user_is_member": false, "id": "70", "is_public": true, "name": "Parent3"}
        ],
        "require_lock_membership_approval_until": null,
        "require_members_to_join_parent": false,
        "require_personal_info_access_approval": "none",
        "require_watch_approval": false
      }
    ]
    """
  Examples:
    | sort           |
    |                |
    | ?sort=-group_id |
    | ?sort=-name     |

  Scenario: User has access to the item (sort by expected start, null first)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/official-sessions?sort=expected_start"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "50",
        "name": "Session 200",
        "address_city": null,
        "address_country": null,
        "address_line1": null,
        "address_line2": null,
        "address_postcode": null,
        "current_user_is_manager": true,
        "current_user_is_member": false,
        "description": null,
        "expected_start": null,
        "is_public": true,
        "open_activity_when_joining": false,
        "organizer": null,
        "parents": [
          {"current_user_is_manager": true, "current_user_is_member": false, "id": "71", "is_public": false, "name": "Parent"},
          {"current_user_is_manager": false, "current_user_is_member": true, "id": "72", "is_public": false, "name": "Parent"},
          {"current_user_is_manager": false, "current_user_is_member": false, "id": "70", "is_public": true, "name": "Parent3"}
        ],
        "require_lock_membership_approval_until": null,
        "require_members_to_join_parent": false,
        "require_personal_info_access_approval": "none",
        "require_watch_approval": false
      },
      {
        "group_id": "53",
        "name": "Session 200.3",
        "address_city": "Paris",
        "address_country": "France",
        "address_line1": "Chez Jacques-Henri Jourdan,",
        "address_line2": "42, rue de Cronstadt",
        "address_postcode": "75015",
        "current_user_is_manager": false,
        "current_user_is_member": true,
        "description": "Official",
        "expected_start": "2019-05-30T12:00:00Z",
        "is_public": true,
        "open_activity_when_joining": true,
        "organizer": "Association France-ioi",
        "parents": [],
        "require_lock_membership_approval_until": "2019-05-30T11:00:00Z",
        "require_members_to_join_parent": true,
        "require_personal_info_access_approval": "view",
        "require_watch_approval": true
      }
    ]
    """

  Scenario: User has access to the item (sort by group_id, limit=1)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/official-sessions?sort=group_id&limit=1"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "50",
        "name": "Session 200",
        "address_city": null,
        "address_country": null,
        "address_line1": null,
        "address_line2": null,
        "address_postcode": null,
        "current_user_is_manager": true,
        "current_user_is_member": false,
        "description": null,
        "expected_start": null,
        "is_public": true,
        "open_activity_when_joining": false,
        "organizer": null,
        "parents": [
          {"current_user_is_manager": true, "current_user_is_member": false, "id": "71", "is_public": false, "name": "Parent"},
          {"current_user_is_manager": false, "current_user_is_member": true, "id": "72", "is_public": false, "name": "Parent"},
          {"current_user_is_manager": false, "current_user_is_member": false, "id": "70", "is_public": true, "name": "Parent3"}
        ],
        "require_lock_membership_approval_until": null,
        "require_members_to_join_parent": false,
        "require_personal_info_access_approval": "none",
        "require_watch_approval": false
      }
    ]
    """

  Scenario: User has access to the item (default sort, start from the second row)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/official-sessions?from.group_id=53"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "50",
        "name": "Session 200",
        "address_city": null,
        "address_country": null,
        "address_line1": null,
        "address_line2": null,
        "address_postcode": null,
        "current_user_is_manager": true,
        "current_user_is_member": false,
        "description": null,
        "expected_start": null,
        "is_public": true,
        "open_activity_when_joining": false,
        "organizer": null,
        "parents": [
          {"current_user_is_manager": true, "current_user_is_member": false, "id": "71", "is_public": false, "name": "Parent"},
          {"current_user_is_manager": false, "current_user_is_member": true, "id": "72", "is_public": false, "name": "Parent"},
          {"current_user_is_manager": false, "current_user_is_member": false, "id": "70", "is_public": true, "name": "Parent3"}
        ],
        "require_lock_membership_approval_until": null,
        "require_members_to_join_parent": false,
        "require_personal_info_access_approval": "none",
        "require_watch_approval": false
      }
    ]
    """

  Scenario: User has access to the item (sort=expected_start, start from the second row)
    Given I am the user with id "11"
    When I send a GET request to "/items/200/official-sessions?sort=expected_start&from.group_id=50"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    [
      {
        "group_id": "53",
        "name": "Session 200.3",
        "address_city": "Paris",
        "address_country": "France",
        "address_line1": "Chez Jacques-Henri Jourdan,",
        "address_line2": "42, rue de Cronstadt",
        "address_postcode": "75015",
        "current_user_is_manager": false,
        "current_user_is_member": true,
        "description": "Official",
        "expected_start": "2019-05-30T12:00:00Z",
        "is_public": true,
        "open_activity_when_joining": true,
        "organizer": "Association France-ioi",
        "parents": [],
        "require_lock_membership_approval_until": "2019-05-30T11:00:00Z",
        "require_members_to_join_parent": true,
        "require_personal_info_access_approval": "view",
        "require_watch_approval": true
      }
    ]
    """
