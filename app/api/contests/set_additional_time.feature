Feature: Set additional time in the contest for the group (contestSetAdditionalTime)
  Background:
    Given the database has the following table 'groups':
      | id | name        | type      |
      | 10 | Parent      | Club      |
      | 11 | Group A     | Class     |
      | 13 | Group B     | Other     |
      | 14 | Group B     | Friends   |
      | 21 | owner       | UserSelf  |
      | 22 | owner-admin | UserAdmin |
      | 31 | john        | UserSelf  |
      | 32 | john-admin  | UserAdmin |
      | 33 | item10      | Other     |
      | 34 | item50      | Other     |
      | 35 | item60      | Other     |
      | 36 | item70      | Other     |
    And the database has the following table 'users':
      | login | group_id | owned_group_id |
      | owner | 21       | 22             |
      | john  | 31       | 32             |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | expires_at          |
      | 10              | 11             | 9999-12-31 23:59:59 |
      | 11              | 13             | 9999-12-31 23:59:59 |
      | 13              | 14             | 9999-12-31 23:59:59 |
      | 22              | 13             | 9999-12-31 23:59:59 |
      | 22              | 14             | 9999-12-31 23:59:59 |
      | 22              | 31             | 9999-12-31 23:59:59 |
      | 34              | 13             | 9999-12-31 23:59:59 |
      | 34              | 14             | 9999-12-31 23:59:59 |
      | 36              | 13             | 9999-12-31 23:59:59 |
      | 36              | 14             | 9999-12-31 23:59:59 |
      | 36              | 31             | 9999-12-31 23:59:59 |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self | expires_at          |
      | 10                | 10             | 1       | 9999-12-31 23:59:59 |
      | 10                | 11             | 0       | 9999-12-31 23:59:59 |
      | 10                | 13             | 0       | 9999-12-31 23:59:59 |
      | 10                | 14             | 0       | 9999-12-31 23:59:59 |
      | 11                | 11             | 1       | 9999-12-31 23:59:59 |
      | 11                | 13             | 0       | 9999-12-31 23:59:59 |
      | 11                | 14             | 0       | 9999-12-31 23:59:59 |
      | 13                | 13             | 1       | 9999-12-31 23:59:59 |
      | 13                | 14             | 0       | 9999-12-31 23:59:59 |
      | 14                | 14             | 1       | 9999-12-31 23:59:59 |
      | 21                | 21             | 1       | 9999-12-31 23:59:59 |
      | 22                | 13             | 0       | 9999-12-31 23:59:59 |
      | 22                | 14             | 0       | 9999-12-31 23:59:59 |
      | 22                | 22             | 1       | 9999-12-31 23:59:59 |
      | 22                | 31             | 0       | 9999-12-31 23:59:59 |
      | 31                | 31             | 1       | 9999-12-31 23:59:59 |
      | 32                | 32             | 1       | 9999-12-31 23:59:59 |
      | 33                | 33             | 1       | 9999-12-31 23:59:59 |
      | 34                | 13             | 0       | 9999-12-31 23:59:59 |
      | 34                | 14             | 0       | 9999-12-31 23:59:59 |
      | 34                | 34             | 1       | 9999-12-31 23:59:59 |
      | 35                | 35             | 1       | 9999-12-31 23:59:59 |
      | 36                | 13             | 0       | 9999-12-31 23:59:59 |
      | 36                | 14             | 0       | 9999-12-31 23:59:59 |
      | 36                | 31             | 0       | 9999-12-31 23:59:59 |
      | 36                | 36             | 1       | 9999-12-31 23:59:59 |
    And the database has the following table 'items':
      | id | duration | contest_participants_group_id |
      | 10 | 00:00:02 | 33                            |
      | 50 | 00:00:00 | 34                            |
      | 60 | 00:00:01 | 35                            |
      | 70 | 00:00:03 | 36                            |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       |
      | 10       | 50      | none                     |
      | 11       | 50      | none                     |
      | 11       | 60      | none                     |
      | 11       | 70      | content_with_descendants |
      | 13       | 50      | content                  |
      | 13       | 60      | info                     |
      | 21       | 50      | solution                 |
      | 21       | 60      | content_with_descendants |
      | 21       | 70      | content_with_descendants |
    And the database has the following table 'groups_contest_items':
      | group_id | item_id | additional_time |
      | 10       | 50      | 01:00:00        |
      | 11       | 50      | 00:01:00        |
      | 13       | 50      | 00:00:01        |
      | 13       | 60      | 00:00:30        |
      | 21       | 50      | 00:01:00        |
      | 21       | 60      | 00:01:00        |
      | 21       | 70      | 00:01:00        |
    And the database has the following table 'groups_attempts':
      | group_id | item_id | entered_at          | order |
      | 13       | 50      | 2018-05-29 06:38:38 | 1     |
      | 13       | 70      | 2018-05-29 06:38:38 | 1     |
      | 14       | 50      | 2019-05-29 06:38:38 | 1     |
      | 31       | 70      | 2017-05-29 06:38:38 | 1     |

  Scenario: Updates an existing row
    Given I am the user with group_id "21"
    When I send a PUT request to "/contests/50/groups/13/additional-times?seconds=3020399"
    Then the response code should be 200
    And the response should be "updated"
    And the table "permissions_generated" should stay unchanged
    And the table "groups_contest_items" should be:
      | group_id | item_id | additional_time |
      | 10       | 50      | 01:00:00        |
      | 11       | 50      | 00:01:00        |
      | 13       | 50      | 838:59:59       |
      | 13       | 60      | 00:00:30        |
      | 21       | 50      | 00:01:00        |
      | 21       | 60      | 00:01:00        |
      | 21       | 70      | 00:01:00        |
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | expires_at          |
      | 10              | 11             | 9999-12-31 23:59:59 |
      | 11              | 13             | 9999-12-31 23:59:59 |
      | 13              | 14             | 9999-12-31 23:59:59 |
      | 22              | 13             | 9999-12-31 23:59:59 |
      | 22              | 14             | 9999-12-31 23:59:59 |
      | 22              | 31             | 9999-12-31 23:59:59 |
      | 34              | 13             | 2018-07-03 06:39:37 |
      | 34              | 14             | 2019-07-03 06:39:37 |
      | 36              | 13             | 9999-12-31 23:59:59 |
      | 36              | 14             | 9999-12-31 23:59:59 |
      | 36              | 31             | 9999-12-31 23:59:59 |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self | expires_at          |
      | 10                | 10             | 1       | 9999-12-31 23:59:59 |
      | 10                | 11             | 0       | 9999-12-31 23:59:59 |
      | 10                | 13             | 0       | 9999-12-31 23:59:59 |
      | 10                | 14             | 0       | 9999-12-31 23:59:59 |
      | 11                | 11             | 1       | 9999-12-31 23:59:59 |
      | 11                | 13             | 0       | 9999-12-31 23:59:59 |
      | 11                | 14             | 0       | 9999-12-31 23:59:59 |
      | 13                | 13             | 1       | 9999-12-31 23:59:59 |
      | 13                | 14             | 0       | 9999-12-31 23:59:59 |
      | 14                | 14             | 1       | 9999-12-31 23:59:59 |
      | 21                | 21             | 1       | 9999-12-31 23:59:59 |
      | 22                | 13             | 0       | 9999-12-31 23:59:59 |
      | 22                | 14             | 0       | 9999-12-31 23:59:59 |
      | 22                | 22             | 1       | 9999-12-31 23:59:59 |
      | 22                | 31             | 0       | 9999-12-31 23:59:59 |
      | 31                | 31             | 1       | 9999-12-31 23:59:59 |
      | 32                | 32             | 1       | 9999-12-31 23:59:59 |
      | 33                | 33             | 1       | 9999-12-31 23:59:59 |
      | 34                | 34             | 1       | 9999-12-31 23:59:59 |
      | 35                | 35             | 1       | 9999-12-31 23:59:59 |
      | 36                | 13             | 0       | 9999-12-31 23:59:59 |
      | 36                | 14             | 0       | 9999-12-31 23:59:59 |
      | 36                | 31             | 0       | 9999-12-31 23:59:59 |
      | 36                | 36             | 1       | 9999-12-31 23:59:59 |

  Scenario: Creates a new row
    Given I am the user with group_id "21"
    When I send a PUT request to "/contests/70/groups/13/additional-times?seconds=-3020399"
    Then the response code should be 200
    And the response should be "updated"
    And the table "permissions_generated" should stay unchanged
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
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | expires_at          |
      | 10              | 11             | 9999-12-31 23:59:59 |
      | 11              | 13             | 9999-12-31 23:59:59 |
      | 13              | 14             | 9999-12-31 23:59:59 |
      | 22              | 13             | 9999-12-31 23:59:59 |
      | 22              | 14             | 9999-12-31 23:59:59 |
      | 22              | 31             | 9999-12-31 23:59:59 |
      | 34              | 13             | 9999-12-31 23:59:59 |
      | 34              | 14             | 9999-12-31 23:59:59 |
      | 36              | 13             | 2018-04-24 07:38:42 |
      | 36              | 14             | 9999-12-31 23:59:59 |
      | 36              | 31             | 9999-12-31 23:59:59 |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self | expires_at          |
      | 10                | 10             | 1       | 9999-12-31 23:59:59 |
      | 10                | 11             | 0       | 9999-12-31 23:59:59 |
      | 10                | 13             | 0       | 9999-12-31 23:59:59 |
      | 10                | 14             | 0       | 9999-12-31 23:59:59 |
      | 11                | 11             | 1       | 9999-12-31 23:59:59 |
      | 11                | 13             | 0       | 9999-12-31 23:59:59 |
      | 11                | 14             | 0       | 9999-12-31 23:59:59 |
      | 13                | 13             | 1       | 9999-12-31 23:59:59 |
      | 13                | 14             | 0       | 9999-12-31 23:59:59 |
      | 14                | 14             | 1       | 9999-12-31 23:59:59 |
      | 21                | 21             | 1       | 9999-12-31 23:59:59 |
      | 22                | 13             | 0       | 9999-12-31 23:59:59 |
      | 22                | 14             | 0       | 9999-12-31 23:59:59 |
      | 22                | 22             | 1       | 9999-12-31 23:59:59 |
      | 22                | 31             | 0       | 9999-12-31 23:59:59 |
      | 31                | 31             | 1       | 9999-12-31 23:59:59 |
      | 32                | 32             | 1       | 9999-12-31 23:59:59 |
      | 33                | 33             | 1       | 9999-12-31 23:59:59 |
      | 34                | 13             | 0       | 9999-12-31 23:59:59 |
      | 34                | 14             | 0       | 9999-12-31 23:59:59 |
      | 34                | 34             | 1       | 9999-12-31 23:59:59 |
      | 35                | 35             | 1       | 9999-12-31 23:59:59 |
      | 36                | 14             | 0       | 9999-12-31 23:59:59 |
      | 36                | 31             | 0       | 9999-12-31 23:59:59 |
      | 36                | 36             | 1       | 9999-12-31 23:59:59 |

  Scenario: Doesn't create a new row when seconds=0
    Given I am the user with group_id "21"
    When I send a PUT request to "/contests/70/groups/13/additional-times?seconds=0"
    Then the response code should be 200
    And the response should be "updated"
    And the table "groups_contest_items" should stay unchanged
    And the table "groups_groups" should stay unchanged
    And the table "groups_ancestors" should stay unchanged

  Scenario: Creates a new row for a user group
    Given I am the user with group_id "21"
    When I send a PUT request to "/contests/70/groups/31/additional-times?seconds=-3020399"
    Then the response code should be 200
    And the response should be "updated"
    And the table "permissions_generated" should stay unchanged
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
    And the table "groups_groups" should be:
      | parent_group_id | child_group_id | expires_at          |
      | 10              | 11             | 9999-12-31 23:59:59 |
      | 11              | 13             | 9999-12-31 23:59:59 |
      | 13              | 14             | 9999-12-31 23:59:59 |
      | 22              | 13             | 9999-12-31 23:59:59 |
      | 22              | 14             | 9999-12-31 23:59:59 |
      | 22              | 31             | 9999-12-31 23:59:59 |
      | 34              | 13             | 9999-12-31 23:59:59 |
      | 34              | 14             | 9999-12-31 23:59:59 |
      | 36              | 13             | 9999-12-31 23:59:59 |
      | 36              | 14             | 9999-12-31 23:59:59 |
      | 36              | 31             | 2017-04-24 07:38:42 |
    And the table "groups_ancestors" should be:
      | ancestor_group_id | child_group_id | is_self | expires_at          |
      | 10                | 10             | 1       | 9999-12-31 23:59:59 |
      | 10                | 11             | 0       | 9999-12-31 23:59:59 |
      | 10                | 13             | 0       | 9999-12-31 23:59:59 |
      | 10                | 14             | 0       | 9999-12-31 23:59:59 |
      | 11                | 11             | 1       | 9999-12-31 23:59:59 |
      | 11                | 13             | 0       | 9999-12-31 23:59:59 |
      | 11                | 14             | 0       | 9999-12-31 23:59:59 |
      | 13                | 13             | 1       | 9999-12-31 23:59:59 |
      | 13                | 14             | 0       | 9999-12-31 23:59:59 |
      | 14                | 14             | 1       | 9999-12-31 23:59:59 |
      | 21                | 21             | 1       | 9999-12-31 23:59:59 |
      | 22                | 13             | 0       | 9999-12-31 23:59:59 |
      | 22                | 14             | 0       | 9999-12-31 23:59:59 |
      | 22                | 22             | 1       | 9999-12-31 23:59:59 |
      | 22                | 31             | 0       | 9999-12-31 23:59:59 |
      | 31                | 31             | 1       | 9999-12-31 23:59:59 |
      | 32                | 32             | 1       | 9999-12-31 23:59:59 |
      | 33                | 33             | 1       | 9999-12-31 23:59:59 |
      | 34                | 13             | 0       | 9999-12-31 23:59:59 |
      | 34                | 14             | 0       | 9999-12-31 23:59:59 |
      | 34                | 34             | 1       | 9999-12-31 23:59:59 |
      | 35                | 35             | 1       | 9999-12-31 23:59:59 |
      | 36                | 13             | 0       | 9999-12-31 23:59:59 |
      | 36                | 14             | 0       | 9999-12-31 23:59:59 |
      | 36                | 36             | 1       | 9999-12-31 23:59:59 |
