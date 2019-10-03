Feature: Set additional time in the contest for the group (contestSetAdditionalTime)
  Background:
    Given the database has the following table 'users':
      | id | login | self_group_id | owned_group_id |
      | 1  | owner | 21            | 22             |
      | 2  | john  | 31            | 32             |
    And the database has the following table 'groups':
      | id | name        | type      |
      | 10 | Parent      | Club      |
      | 11 | Group A     | Class     |
      | 13 | Group B     | Other     |
      | 14 | Group B     | Friends   |
      | 21 | owner       | UserSelf  |
      | 22 | owner-admin | UserAdmin |
      | 31 | john        | UserSelf  |
      | 32 | john-admin  | UserAdmin |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 10                | 10             | 1       |
      | 10                | 11             | 0       |
      | 10                | 13             | 0       |
      | 11                | 11             | 1       |
      | 11                | 13             | 0       |
      | 13                | 13             | 1       |
      | 14                | 14             | 1       |
      | 21                | 21             | 1       |
      | 22                | 13             | 0       |
      | 22                | 14             | 0       |
      | 22                | 31             | 0       |
      | 22                | 22             | 1       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
    And the database has the following table 'items':
      | id | duration |
      | 50 | 00:00:00 |
      | 60 | 00:00:01 |
      | 10 | 00:00:02 |
      | 70 | 00:00:03 |
    And the database has the following table 'groups_items':
      | id | group_id | item_id | cached_partial_access_since | cached_grayed_access_since | cached_full_access_since | cached_solutions_access_since | creator_user_id |
      | 1  | 10       | 50      | null                        | null                       | null                     | null                          | 3               |
      | 2  | 11       | 50      | null                        | null                       | null                     | null                          | 3               |
      | 3  | 13       | 50      | 2017-05-29 06:38:38         | null                       | null                     | null                          | 3               |
      | 4  | 11       | 60      | null                        | null                       | null                     | null                          | 3               |
      | 5  | 13       | 60      | null                        | 2017-05-29 06:38:38        | null                     | null                          | 3               |
      | 6  | 11       | 70      | null                        | null                       | 2017-05-29 06:38:38      | null                          | 3               |
      | 7  | 21       | 50      | null                        | null                       | null                     | 2018-05-29 06:38:38           | 3               |
      | 8  | 21       | 60      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 3               |
      | 9  | 21       | 70      | null                        | null                       | 2018-05-29 06:38:38      | null                          | 3               |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | additional_time |
      | 10       | 50      | 01:00:00        |
      | 11       | 50      | 00:01:00        |
      | 13       | 50      | 00:00:01        |
      | 13       | 60      | 00:00:30        |
      | 21       | 50      | 00:01:00        |
      | 21       | 60      | 00:01:00        |
      | 21       | 70      | 00:01:00        |

  Scenario: Updates an existing row
    Given I am the user with id "1"
    When I send a PUT request to "/contests/50/groups/13/additional-times?seconds=3020399"
    Then the response code should be 200
    And the response should be "updated"
    And the table "groups_items" should stay unchanged
    And the table "groups_contest_items" should be:
      | group_id | item_id | additional_time |
      | 10       | 50      | 01:00:00        |
      | 11       | 50      | 00:01:00        |
      | 13       | 50      | 838:59:59       |
      | 13       | 60      | 00:00:30        |
      | 21       | 50      | 00:01:00        |
      | 21       | 60      | 00:01:00        |
      | 21       | 70      | 00:01:00        |

  Scenario: Creates a new row
    Given I am the user with id "1"
    When I send a PUT request to "/contests/70/groups/13/additional-times?seconds=-3020399"
    Then the response code should be 200
    And the response should be "updated"
    And the table "groups_items" should stay unchanged
    And the table "groups_contest_items" should be:
      | group_id | item_id | additional_time |
      | 10       | 50      | 01:00:00        |
      | 11       | 50      | 00:01:00        |
      | 13       | 50      | 00:00:01        |
      | 13       | 60      | 00:00:30        |
      | 13       | 70      | -838:59:59      |
      | 21       | 50      | 00:01:00        |
      | 21       | 60      | 00:01:00        |
      | 21       | 70      | 00:01:00        |

  Scenario: Doesn't create a new row when seconds=0
    Given I am the user with id "1"
    When I send a PUT request to "/contests/70/groups/13/additional-times?seconds=0"
    Then the response code should be 200
    And the response should be "updated"
    And the table "groups_contest_items" should stay unchanged

  Scenario: Creates a new row for a user group
    Given I am the user with id "1"
    When I send a PUT request to "/contests/70/groups/31/additional-times?seconds=-3020399"
    Then the response code should be 200
    And the response should be "updated"
    And the table "groups_items" should stay unchanged
    And the table "groups_contest_items" should be:
      | group_id | item_id | additional_time |
      | 10       | 50      | 01:00:00        |
      | 11       | 50      | 00:01:00        |
      | 13       | 50      | 00:00:01        |
      | 13       | 60      | 00:00:30        |
      | 21       | 50      | 00:01:00        |
      | 21       | 60      | 00:01:00        |
      | 21       | 70      | 00:01:00        |
      | 31       | 70      | -838:59:59      |
