Feature: Change item access rights for a group
  Background:
    Given the database has the following table 'groups':
      | id | name       | type     |
      | 21 | owner      | UserSelf |
      | 23 | user       | UserSelf |
      | 25 | some class | Class    |
      | 31 | jane       | UserSelf |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name |
      | owner | 21       | Jean-Michel | Blanquer  |
      | user  | 23       | John        | Doe       |
      | jane  | 31       | Jane        | Doe       |
    And the database has the following table 'group_managers':
      | group_id | manager_id |
      | 23       | 21         |
      | 31       | 21         |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 21                | 21             | 1       |
      | 23                | 23             | 1       |
      | 25                | 23             | 0       |
      | 25                | 25             | 1       |
      | 31                | 31             | 1       |
    And the database has the following table 'items':
      | id  |
      | 100 |
      | 101 |
      | 102 |
      | 103 |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | content_view_propagation | child_order |
      | 100            | 101           | as_info                  | 0           |
      | 101            | 102           | as_content               | 0           |
      | 102            | 103           | as_content               | 0           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 100              | 101           |
      | 100              | 102           |
      | 100              | 103           |
      | 101              | 102           |
      | 101              | 103           |
      | 102              | 103           |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated        |
      | 25       | 100     | content_with_descendants  |
      | 25       | 101     | info                      |
      | 25       | 102     | info                      |
      | 25       | 103     | info                      |
    And the database has the following table 'permissions_granted':
      | group_id | item_id | can_view | giver_group_id | latest_update_on    |
      | 25       | 100     | content  | 23             | 2019-05-30 11:00:00 |

  Scenario Outline: Create a new permissions_granted row
    Given I am the user with id "21"
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated |
      | 21       | 102     | solution           | solution                 | answer              | all                |
      | 21       | 103     | solution           | solution                 | answer              | all                |
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | can_view | can_grant_view | can_watch | can_edit | giver_group_id | latest_update_on    |
      | 21       | 102     | solution | transfer       | transfer  | transfer | 23             | 2019-05-30 11:00:00 |
    When I send a PUT request to "/groups/23/items/102" with the following body:
      """
      {
        "can_view": "<can_view>"
      }
      """
    Then the response should be "updated"
    And the table "permissions_granted" should be:
      | group_id | item_id | giver_group_id | can_view   | TIMESTAMPDIFF(SECOND, latest_update_on, NOW()) < 3 |
      | 21       | 102     | 23             | solution   | 0                                                  |
      | 23       | 102     | 21             | <can_view> | 1                                                  |
      | 25       | 100     | 23             | content    | 0                                                  |
    And the table "permissions_generated" should be:
      | group_id | item_id | can_view_generated    | can_grant_view_generated | can_watch_generated | can_edit_generated |
      | 21       | 102     | solution              | transfer                 | transfer            | transfer           |
      | 21       | 103     | content               | none                     | none                | none               |
      | 23       | 102     | <can_view>            | none                     | none                | none               |
      | 23       | 103     | <propagated_can_view> | none                     | none                | none               |
      | 25       | 100     | content               | none                     | none                | none               |
      | 25       | 101     | info                  | none                     | none                | none               |
      | 25       | 102     | none                  | none                     | none                | none               |
      | 25       | 103     | none                  | none                     | none                | none               |
  Examples:
    | can_view | propagated_can_view |
    | solution | content             |
    | info     | none                |

  Scenario: Update an existing permissions_granted row
    Given I am the user with id "21"
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated |
      | 21       | 102     | solution           | solution                 | answer              | all                |
      | 23       | 102     | none               | none                     | none                | none               |
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | can_view | can_grant_view | can_watch | can_edit | giver_group_id | latest_update_on    |
      | 21       | 102     | solution | solution       | answer    | all      | 23             | 2019-05-30 11:00:00 |
      | 23       | 102     | none     | none           | none      | none     | 21             | 2019-05-30 11:00:00 |
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response should be "updated"
    And the table "permissions_granted" should be:
      | group_id | item_id | giver_group_id | can_view | can_grant_view | can_watch | can_edit | TIMESTAMPDIFF(SECOND, latest_update_on, NOW()) < 3 |
      | 21       | 102     | 23             | solution | solution       | answer    | all      | 0                                                  |
      | 23       | 102     | 21             | solution | none           | none      | none     | 1                                                  |
      | 25       | 100     | 23             | content  | none           | none      | none     | 0                                                  |
    And the table "permissions_generated" should be:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated |
      | 21       | 102     | solution           | solution                 | answer              | all                |
      | 21       | 103     | content            | none                     | none                | none               |
      | 23       | 102     | solution           | none                     | none                | none               |
      | 23       | 103     | content            | none                     | none                | none               |
      | 25       | 100     | content            | none                     | none                | none               |
      | 25       | 101     | info               | none                     | none                | none               |
      | 25       | 102     | none               | none                     | none                | none               |
      | 25       | 103     | none               | none                     | none                | none               |

  Scenario: Create a new permissions_granted row (the group has only 'content' access on the item's parent)
    Given I am the user with id "21"
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | transfer                 | transfer            | transfer           | 1                  |
      | 21       | 103     | none               | none                     | none                | none               | 0                  |
      | 31       | 101     | content            | none                     | none                | none               | 0                  |
      | 31       | 102     | none               | none                     | none                | none               | 0                  |
      | 31       | 103     | none               | none                     | none                | none               | 0                  |
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | can_view | is_owner | giver_group_id | latest_update_on    |
      | 21       | 102     | none     | 1        | 23             | 2019-05-30 11:00:00 |
      | 31       | 101     | content  | 0        | 23             | 2019-05-30 11:00:00 |
    When I send a PUT request to "/groups/31/items/102" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response should be "updated"
    And the table "permissions_granted" should be:
      | group_id | item_id | can_view | is_owner | giver_group_id | TIMESTAMPDIFF(SECOND, latest_update_on, NOW()) < 3 |
      | 21       | 102     | none     | 1        | 23             | 0                                                  |
      | 25       | 100     | content  | 0        | 23             | 0                                                  |
      | 31       | 101     | content  | 0        | 23             | 0                                                  |
      | 31       | 102     | solution | 0        | 21             | 1                                                  |
    And the table "permissions_generated" should be:
      | group_id | item_id | can_view_generated | can_grant_view_generated | can_watch_generated | can_edit_generated | is_owner_generated |
      | 21       | 102     | solution           | transfer                 | transfer            | transfer           | 1                  |
      | 21       | 103     | content            | none                     | none                | none               | 0                  |
      | 25       | 100     | content            | none                     | none                | none               | 0                  |
      | 25       | 101     | info               | none                     | none                | none               | 0                  |
      | 25       | 102     | none               | none                     | none                | none               | 0                  |
      | 25       | 103     | none               | none                     | none                | none               | 0                  |
      | 31       | 101     | content            | none                     | none                | none               | 0                  |
      | 31       | 102     | solution           | none                     | none                | none               | 0                  |
      | 31       | 103     | content            | none                     | none                | none               | 0                  |

  Scenario: Create a new permissions_granted row (the group has no access to the item's parents, but has full access to the item itself)
    Given I am the user with id "21"
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated       | can_grant_view_generated | is_owner_generated |
      | 21       | 100     | solution                 | solution                 | 1                  |
      | 21       | 101     | none                     | none                     | 0                  |
      | 21       | 102     | none                     | none                     | 0                  |
      | 21       | 103     | none                     | none                     | 0                  |
      | 31       | 100     | content_with_descendants | none                     | 0                  |
      | 31       | 101     | content_with_descendants | none                     | 0                  |
      | 31       | 102     | content_with_descendants | none                     | 0                  |
      | 31       | 103     | content_with_descendants | none                     | 0                  |
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | can_view                 | can_grant_view | is_owner | giver_group_id | latest_update_on    |
      | 21       | 100     | solution                 | solution       | 1        | 23             | 2019-05-30 11:00:00 |
      | 31       | 100     | content_with_descendants | none           | 0        | 23             | 2019-05-30 11:00:00 |
    When I send a PUT request to "/groups/31/items/100" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response should be "updated"
    And the table "permissions_granted" should be:
      | group_id | item_id | can_view                 | is_owner | giver_group_id | TIMESTAMPDIFF(SECOND, latest_update_on, NOW()) < 3 |
      | 21       | 100     | solution                 | 1        | 23             | 0                                                  |
      | 25       | 100     | content                  | 0        | 23             | 0                                                  |
      | 31       | 100     | content_with_descendants | 0        | 23             | 0                                                  |
      | 31       | 100     | solution                 | 0        | 21             | 1                                                  |
    And the table "permissions_generated" should be:
      | group_id | item_id | can_view_generated | is_owner_generated |
      | 21       | 100     | solution           | 1                  |
      | 21       | 101     | info               | 0                  |
      | 21       | 102     | none               | 0                  |
      | 21       | 103     | none               | 0                  |
      | 25       | 100     | content            | 0                  |
      | 25       | 101     | info               | 0                  |
      | 25       | 102     | none               | 0                  |
      | 25       | 103     | none               | 0                  |
      | 31       | 100     | solution           | 0                  |
      | 31       | 101     | info               | 0                  |
      | 31       | 102     | none               | 0                  |
      | 31       | 103     | none               | 0                  |

  Scenario: Create a new permissions_granted row (the group has no access to the item's parents, but has 'content' access to the item itself)
    Given I am the user with id "21"
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | is_owner_generated |
      | 21       | 100     | solution           | solution                 | 1                  |
      | 21       | 101     | none               | none                     | 0                  |
      | 21       | 102     | none               | none                     | 0                  |
      | 21       | 103     | none               | none                     | 0                  |
      | 31       | 100     | content            | none                     | 0                  |
      | 31       | 101     | content            | none                     | 0                  |
      | 31       | 102     | content            | none                     | 0                  |
      | 31       | 103     | content            | none                     | 0                  |
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | can_view | can_grant_view | is_owner | giver_group_id | latest_update_on    |
      | 21       | 100     | none     | solution       | 1        | 23             | 2019-05-30 11:00:00 |
      | 31       | 100     | content  | none           | 0        | 23             | 2019-05-30 11:00:00 |
    When I send a PUT request to "/groups/31/items/100" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response should be "updated"
    And the table "permissions_granted" should be:
      | group_id | item_id | can_view | is_owner | giver_group_id | TIMESTAMPDIFF(SECOND, latest_update_on, NOW()) < 3 |
      | 21       | 100     | none     | 1        | 23             | 0                                                  |
      | 25       | 100     | content  | 0        | 23             | 0                                                  |
      | 31       | 100     | content  | 0        | 23             | 0                                                  |
      | 31       | 100     | solution | 0        | 21             | 1                                                  |
    And the table "permissions_generated" should be:
      | group_id | item_id | can_view_generated | is_owner_generated |
      | 21       | 100     | solution           | 1                  |
      | 21       | 101     | info               | 0                  |
      | 21       | 102     | none               | 0                  |
      | 21       | 103     | none               | 0                  |
      | 25       | 100     | content            | 0                  |
      | 25       | 101     | info               | 0                  |
      | 25       | 102     | none               | 0                  |
      | 25       | 103     | none               | 0                  |
      | 31       | 100     | solution           | 0                  |
      | 31       | 101     | info               | 0                  |
      | 31       | 102     | none               | 0                  |
      | 31       | 103     | none               | 0                  |

  Scenario: Create a new permissions_granted row (the group has no access to the item's parents, but has info access to the item itself)
    Given I am the user with id "21"
    And the database table 'permissions_generated' has also the following rows:
      | group_id | item_id | can_view_generated | can_grant_view_generated | is_owner_generated |
      | 21       | 100     | solution           | solution                 | 1                  |
      | 21       | 101     | none               | none                     | 0                  |
      | 21       | 102     | none               | none                     | 0                  |
      | 21       | 103     | none               | none                     | 0                  |
      | 31       | 100     | info               | none                     | 0                  |
      | 31       | 101     | info               | none                     | 0                  |
      | 31       | 102     | info               | none                     | 0                  |
      | 31       | 103     | info               | none                     | 0                  |
    And the database table 'permissions_granted' has also the following rows:
      | group_id | item_id | can_view | can_grant_view | is_owner | giver_group_id | latest_update_on    |
      | 21       | 100     | none     | solution       | 1        | 23             | 2019-05-30 11:00:00 |
      | 31       | 100     | info     | none           | 0        | 23             | 2019-05-30 11:00:00 |
    When I send a PUT request to "/groups/31/items/100" with the following body:
    """
    {
      "can_view": "solution"
    }
    """
    Then the response should be "updated"
    And the table "permissions_granted" should be:
      | group_id | item_id | can_view | is_owner | giver_group_id | TIMESTAMPDIFF(SECOND, latest_update_on, NOW()) < 3 |
      | 21       | 100     | none     | 1        | 23             | 0                                                  |
      | 25       | 100     | content  | 0        | 23             | 0                                                  |
      | 31       | 100     | info     | 0        | 23             | 0                                                  |
      | 31       | 100     | solution | 0        | 21             | 1                                                  |
    And the table "permissions_generated" should be:
      | group_id | item_id | can_view_generated | is_owner_generated |
      | 21       | 100     | solution           | 1                  |
      | 21       | 101     | info               | 0                  |
      | 21       | 102     | none               | 0                  |
      | 21       | 103     | none               | 0                  |
      | 25       | 100     | content            | 0                  |
      | 25       | 101     | info               | 0                  |
      | 25       | 102     | none               | 0                  |
      | 25       | 103     | none               | 0                  |
      | 31       | 100     | solution           | 0                  |
      | 31       | 101     | info               | 0                  |
      | 31       | 102     | none               | 0                  |
      | 31       | 103     | none               | 0                  |
