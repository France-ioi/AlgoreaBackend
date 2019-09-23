Feature: Change item access rights for a group
  Background:
    Given the database has the following table 'users':
      | id | login | self_group_id | owned_group_id | first_name  | last_name |
      | 1  | owner | 21            | 22             | Jean-Michel | Blanquer  |
      | 2  | user  | 23            | 24             | John        | Doe       |
      | 3  | jane  | 31            | 32             | Jane        | Doe       |
    And the database has the following table 'groups':
      | id | name        | type      |
      | 21 | owner       | UserSelf  |
      | 22 | owner-admin | UserAdmin |
      | 23 | user        | UserSelf  |
      | 24 | user-admin  | UserAdmin |
      | 25 | some class  | Class     |
      | 31 | jane        | UserSelf  |
      | 32 | jane-admin  | UserAdmin |
    And the database has the following table 'groups_ancestors':
      | ancestor_group_id | child_group_id | is_self |
      | 21                | 21             | 1       |
      | 22                | 22             | 1       |
      | 22                | 23             | 0       |
      | 22                | 31             | 0       |
      | 23                | 23             | 1       |
      | 24                | 24             | 1       |
      | 25                | 23             | 0       |
      | 25                | 25             | 1       |
      | 31                | 31             | 1       |
      | 32                | 32             | 1       |
    And the database has the following table 'items':
      | id  |
      | 100 |
      | 101 |
      | 102 |
      | 103 |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | always_visible | access_restricted | child_order |
      | 100            | 101           | true           | true              | 0           |
      | 101            | 102           | false          | false             | 0           |
      | 102            | 103           | false          | false             | 0           |
    And the database has the following table 'items_ancestors':
      | ancestor_item_id | child_item_id |
      | 100              | 101           |
      | 100              | 102           |
      | 100              | 103           |
      | 101              | 102           |
      | 101              | 103           |
      | 102              | 103           |
    And the database has the following table 'groups_items':
      | group_id | item_id | full_access_date | cached_full_access_date | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_solutions_date | cached_access_solutions_date | manager_access | cached_manager_access | access_reason                                  | creator_user_id |
      | 25       | 100     | null             | null                    | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | null                  | null                         | 0              | 0                     | the parent item is visible to the user's class | 2               |
      | 25       | 101     | null             | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0              | 0                     | null                                           | 2               |
      | 25       | 102     | null             | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0              | 0                     | null                                           | 2               |
      | 25       | 103     | null             | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0              | 0                     | null                                           | 2               |

  Scenario: Create a new groups_items row (manager access)
    Given I am the user with id "1"
    And the database table 'groups_items' has also the following rows:
      | group_id | item_id | manager_access | cached_manager_access | access_reason                            | creator_user_id |
      | 21       | 100     | 1              | 1                     | the admin can manage the item's ancestor | 2               |
      | 21       | 101     | 0              | 1                     | null                                     | 2               |
      | 21       | 102     | 0              | 1                     | null                                     | 2               |
      | 21       | 103     | 0              | 1                     | null                                     | 2               |
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | group_id | item_id | full_access_date    | cached_full_access_date | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_solutions_date | cached_access_solutions_date | cached_manager_access | access_reason                                  | creator_user_id |
      | 21       | 100     | null                | null                    | null                | null                       | null                      | null                  | null                         | 1                     | the admin can manage the item's ancestor       | 2               |
      | 21       | 101     | null                | null                    | null                | null                       | null                      | null                  | null                         | 1                     | null                                           | 2               |
      | 21       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 1                     | null                                           | 2               |
      | 21       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 1                     | null                                           | 2               |
      | 23       | 102     | 2019-04-06 09:26:40 | 2019-04-06 09:26:40     | 2019-03-06 09:26:40 | 2019-03-06 09:26:40        | null                      | 2019-05-06 09:26:40   | 2019-05-06 09:26:40          | 0                     | the user really needs this access              | 1               |
      | 23       | 103     | null                | 2019-04-06 09:26:40     | null                | 2019-03-06 09:26:40        | null                      | null                  | 2019-05-06 09:26:40          | 0                     | null                                           | 1               |
      | 25       | 100     | null                | null                    | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | null                  | null                         | 0                     | the parent item is visible to the user's class | 2               |
      | 25       | 101     | null                | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0                     | null                                           | 2               |
      | 25       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | null                                           | 2               |
      | 25       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | null                                           | 2               |

  Scenario: Create a new groups_items row (owner access on the item's ancestor)
    Given I am the user with id "1"
    And the database table 'groups_items' has also the following rows:
      | group_id | item_id | owner_access | access_reason                                | creator_user_id |
      | 21       | 100     | 1            | the admin is an owner of the item's ancestor | 2               |
      | 21       | 101     | 0            | null                                         | 2               |
      | 21       | 102     | 0            | null                                         | 2               |
      | 21       | 103     | 0            | null                                         | 2               |
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | group_id | item_id | full_access_date    | cached_full_access_date | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_solutions_date | cached_access_solutions_date | manager_access | cached_manager_access | owner_access | access_reason                                  | creator_user_id |
      | 21       | 100     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0              | 0                     | 1            | the admin is an owner of the item's ancestor   | 2               |
      | 21       | 101     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0              | 0                     | 0            | null                                           | 2               |
      | 21       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0              | 0                     | 0            | null                                           | 2               |
      | 21       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0              | 0                     | 0            | null                                           | 2               |
      | 23       | 102     | 2019-04-06 09:26:40 | 2019-04-06 09:26:40     | 2019-03-06 09:26:40 | 2019-03-06 09:26:40        | null                      | 2019-05-06 09:26:40   | 2019-05-06 09:26:40          | 0              | 0                     | 0            | the user really needs this access              | 1               |
      | 23       | 103     | null                | 2019-04-06 09:26:40     | null                | 2019-03-06 09:26:40        | null                      | null                  | 2019-05-06 09:26:40          | 0              | 0                     | 0            | null                                           | 1               |
      | 25       | 100     | null                | null                    | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | null                  | null                         | 0              | 0                     | 0            | the parent item is visible to the user's class | 2               |
      | 25       | 101     | null                | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0              | 0                     | 0            | null                                           | 2               |
      | 25       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0              | 0                     | 0            | null                                           | 2               |
      | 25       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0              | 0                     | 0            | null                                           | 2               |

  Scenario: Create a new groups_items row (owner access on the item)
    Given I am the user with id "1"
    And the database table 'groups_items' has also the following rows:
      | group_id | item_id | owner_access | access_reason                     | creator_user_id |
      | 21       | 102     | 1            | the admin is an owner of the item | 2               |
      | 21       | 103     | 0            | null                              | 2               |
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | group_id | item_id | full_access_date    | cached_full_access_date | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_solutions_date | cached_access_solutions_date | cached_manager_access | owner_access | access_reason                                  | creator_user_id |
      | 21       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 1            | the admin is an owner of the item              | 2               |
      | 21       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 23       | 102     | 2019-04-06 09:26:40 | 2019-04-06 09:26:40     | 2019-03-06 09:26:40 | 2019-03-06 09:26:40        | null                      | 2019-05-06 09:26:40   | 2019-05-06 09:26:40          | 0                     | 0            | the user really needs this access              | 1               |
      | 23       | 103     | null                | 2019-04-06 09:26:40     | null                | 2019-03-06 09:26:40        | null                      | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                           | 1               |
      | 25       | 100     | null                | null                    | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | null                  | null                         | 0                     | 0            | the parent item is visible to the user's class | 2               |
      | 25       | 101     | null                | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |

  Scenario: Update an existing groups_items row
    Given I am the user with id "1"
    And the database table 'groups_items' has also the following rows:
      | group_id | item_id | full_access_date | cached_full_access_date | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_solutions_date | cached_access_solutions_date | manager_access | cached_manager_access | access_reason                            | creator_user_id |
      | 21       | 100     | null             | null                    | null                | null                       | null                      | null                  | null                         | 1              | 1                     | the admin can manage the item's ancestor | 2               |
      | 21       | 101     | null             | null                    | null                | null                       | null                      | null                  | null                         | 0              | 1                     | null                                     | 2               |
      | 21       | 102     | null             | null                    | null                | null                       | null                      | null                  | null                         | 0              | 1                     | null                                     | 2               |
      | 23       | 102     | null             | null                    | null                | null                       | null                      | null                  | null                         | 0              | 0                     | null                                     | 2               |
      | 23       | 103     | null             | null                    | null                | null                       | null                      | null                  | null                         | 0              | 0                     | null                                     | 2               |
    When I send a PUT request to "/groups/23/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | group_id | item_id | full_access_date    | cached_full_access_date | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_solutions_date | cached_access_solutions_date | cached_manager_access | access_reason                                  | creator_user_id |
      | 21       | 100     | null                | null                    | null                | null                       | null                      | null                  | null                         | 1                     | the admin can manage the item's ancestor       | 2               |
      | 21       | 101     | null                | null                    | null                | null                       | null                      | null                  | null                         | 1                     | null                                           | 2               |
      | 21       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 1                     | null                                           | 2               |
      | 21       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 1                     | null                                           | 2               |
      | 23       | 102     | 2019-04-06 09:26:40 | 2019-04-06 09:26:40     | 2019-03-06 09:26:40 | 2019-03-06 09:26:40        | null                      | 2019-05-06 09:26:40   | 2019-05-06 09:26:40          | 0                     | the user really needs this access              | 2               |
      | 23       | 103     | null                | 2019-04-06 09:26:40     | null                | 2019-03-06 09:26:40        | null                      | null                  | 2019-05-06 09:26:40          | 0                     | null                                           | 2               |
      | 25       | 100     | null                | null                    | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | null                  | null                         | 0                     | the parent item is visible to the user's class | 2               |
      | 25       | 101     | null                | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0                     | null                                           | 2               |
      | 25       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | null                                           | 2               |
      | 25       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | null                                           | 2               |

  Scenario: Create a new groups_items row (the group has only partial access on the item's parent)
    Given I am the user with id "1"
    And the database table 'groups_items' has also the following rows:
      | group_id | item_id | owner_access | partial_access_date | cached_partial_access_date | access_reason                                     | creator_user_id |
      | 21       | 102     | 1            | null                | null                       | the admin is an owner of the item                 | 2               |
      | 21       | 103     | 0            | null                | null                       | null                                              | 2               |
      | 31       | 101     | 0            | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | the group has partial access to the item's parent | 2               |
      | 31       | 102     | 0            | null                | null                       | null                                              | 2               |
      | 31       | 103     | 0            | null                | null                       | null                                              | 2               |
    When I send a PUT request to "/groups/31/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | group_id | item_id | full_access_date    | cached_full_access_date | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_solutions_date | cached_access_solutions_date | cached_manager_access | owner_access | access_reason                                     | creator_user_id |
      | 21       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 1            | the admin is an owner of the item                 | 2               |
      | 21       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                              | 2               |
      | 25       | 100     | null                | null                    | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | null                  | null                         | 0                     | 0            | the parent item is visible to the user's class    | 2               |
      | 25       | 101     | null                | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0                     | 0            | null                                              | 2               |
      | 25       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                              | 2               |
      | 25       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                              | 2               |
      | 31       | 101     | null                | null                    | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | null                  | null                         | 0                     | 0            | the group has partial access to the item's parent | 2               |
      | 31       | 102     | 2019-04-06 09:26:40 | 2019-04-06 09:26:40     | 2019-03-06 09:26:40 | 2019-01-06 09:26:40        | null                      | 2019-05-06 09:26:40   | 2019-05-06 09:26:40          | 0                     | 0            | the user really needs this access                 | 2               |
      | 31       | 103     | null                | 2019-04-06 09:26:40     | null                | 2019-01-06 09:26:40        | null                      | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                              | 2               |

  Scenario: Create a new groups_items row (the group has only full access on the item's parent)
    Given I am the user with id "1"
    And the database table 'groups_items' has also the following rows:
      | group_id | item_id | owner_access | full_access_date    | cached_full_access_date | access_reason                                  | creator_user_id |
      | 21       | 102     | 1            | null                | null                    | the admin is an owner of the item              | 2               |
      | 21       | 103     | 0            | null                | null                    | null                                           | 2               |
      | 31       | 101     | 0            | 2019-01-06 09:26:40 | 2019-01-06 09:26:40     | the group has full access to the item's parent | 2               |
      | 31       | 102     | 0            | null                | 2019-01-06 09:26:40     | null                                           | 2               |
      | 31       | 103     | 0            | null                | 2019-01-06 09:26:40     | null                                           | 2               |
    When I send a PUT request to "/groups/31/items/102" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | group_id | item_id | full_access_date    | cached_full_access_date | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_solutions_date | cached_access_solutions_date | cached_manager_access | owner_access | access_reason                                  | creator_user_id |
      | 21       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 1            | the admin is an owner of the item              | 2               |
      | 21       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 100     | null                | null                    | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | null                  | null                         | 0                     | 0            | the parent item is visible to the user's class | 2               |
      | 25       | 101     | null                | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 31       | 101     | 2019-01-06 09:26:40 | 2019-01-06 09:26:40     | null                | null                       | null                      | null                  | null                         | 0                     | 0            | the group has full access to the item's parent | 2               |
      | 31       | 102     | 2019-04-06 09:26:40 | 2019-01-06 09:26:40     | 2019-03-06 09:26:40 | 2019-03-06 09:26:40        | null                      | 2019-05-06 09:26:40   | 2019-05-06 09:26:40          | 0                     | 0            | the user really needs this access              | 2               |
      | 31       | 103     | null                | 2019-01-06 09:26:40     | null                | 2019-03-06 09:26:40        | null                      | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                           | 2               |


  Scenario: Create a new groups_items row (the group has no access to the item's parents, but has full access to the item itself)
    Given I am the user with id "1"
    And the database table 'groups_items' has also the following rows:
      | group_id | item_id | owner_access | full_access_date    | cached_full_access_date | access_reason                                  | creator_user_id |
      | 21       | 100     | 1            | null                | null                    | the admin is an owner of the item              | 2               |
      | 21       | 101     | 0            | null                | null                    | null                                           | 2               |
      | 21       | 102     | 0            | null                | null                    | null                                           | 2               |
      | 21       | 103     | 0            | null                | null                    | null                                           | 2               |
      | 31       | 100     | 0            | 2019-01-06 09:26:40 | 2019-01-06 09:26:40     | the group has full access to the item's parent | 2               |
      | 31       | 101     | 0            | null                | 2019-01-06 09:26:40     | null                                           | 2               |
      | 31       | 102     | 0            | null                | 2019-01-06 09:26:40     | null                                           | 2               |
      | 31       | 103     | 0            | null                | 2019-01-06 09:26:40     | null                                           | 2               |
    When I send a PUT request to "/groups/31/items/100" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | group_id | item_id | full_access_date    | cached_full_access_date | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_solutions_date | cached_access_solutions_date | cached_manager_access | owner_access | access_reason                                  | creator_user_id |
      | 21       | 100     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 1            | the admin is an owner of the item              | 2               |
      | 21       | 101     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 21       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 21       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 100     | null                | null                    | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | null                  | null                         | 0                     | 0            | the parent item is visible to the user's class | 2               |
      | 25       | 101     | null                | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 31       | 100     | 2019-04-06 09:26:40 | 2019-04-06 09:26:40     | 2019-03-06 09:26:40 | 2019-03-06 09:26:40        | null                      | 2019-05-06 09:26:40   | 2019-05-06 09:26:40          | 0                     | 0            | the user really needs this access              | 2               |
      | 31       | 101     | null                | 2019-04-06 09:26:40     | null                | null                       | 2019-03-06 09:26:40       | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                           | 2               |
      | 31       | 102     | null                | 2019-04-06 09:26:40     | null                | null                       | null                      | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                           | 2               |
      | 31       | 103     | null                | 2019-04-06 09:26:40     | null                | null                       | null                      | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                           | 2               |

  Scenario: Create a new groups_items row (the group has no access to the item's parents, but has partial access to the item itself)
    Given I am the user with id "1"
    And the database table 'groups_items' has also the following rows:
      | group_id | item_id | owner_access | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_reason                                     | creator_user_id |
      | 21       | 100     | 1            | null                | null                       | null                      | the admin is an owner of the item                 | 2               |
      | 21       | 101     | 0            | null                | null                       | null                      | null                                              | 2               |
      | 21       | 102     | 0            | null                | null                       | null                      | null                                              | 2               |
      | 21       | 103     | 0            | null                | null                       | null                      | null                                              | 2               |
      | 31       | 100     | 0            | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | the group has partial access to the item's parent | 2               |
      | 31       | 101     | 0            | null                | null                       | 2019-01-06 09:26:40       | null                                              | 2               |
      | 31       | 102     | 0            | null                | null                       | 2019-01-06 09:26:40       | null                                              | 2               |
      | 31       | 103     | 0            | null                | null                       | 2019-01-06 09:26:40       | null                                              | 2               |
    When I send a PUT request to "/groups/31/items/100" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | group_id | item_id | full_access_date    | cached_full_access_date | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_solutions_date | cached_access_solutions_date | cached_manager_access | owner_access | access_reason                                  | creator_user_id |
      | 21       | 100     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 1            | the admin is an owner of the item              | 2               |
      | 21       | 101     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 21       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 21       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 100     | null                | null                    | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | null                  | null                         | 0                     | 0            | the parent item is visible to the user's class | 2               |
      | 25       | 101     | null                | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 31       | 100     | 2019-04-06 09:26:40 | 2019-04-06 09:26:40     | 2019-03-06 09:26:40 | 2019-03-06 09:26:40        | null                      | 2019-05-06 09:26:40   | 2019-05-06 09:26:40          | 0                     | 0            | the user really needs this access              | 2               |
      | 31       | 101     | null                | 2019-04-06 09:26:40     | null                | null                       | 2019-03-06 09:26:40       | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                           | 2               |
      | 31       | 102     | null                | 2019-04-06 09:26:40     | null                | null                       | null                      | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                           | 2               |
      | 31       | 103     | null                | 2019-04-06 09:26:40     | null                | null                       | null                      | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                           | 2               |

  Scenario: Create a new groups_items row (the group has no access to the item's parents, but has grayed access to the item itself)
    Given I am the user with id "1"
    And the database table 'groups_items' has also the following rows:
      | group_id | item_id | owner_access | cached_grayed_access_date | access_reason                                    | creator_user_id |
      | 21       | 100     | 1            | null                      | the admin is an owner of the item                | 2               |
      | 21       | 101     | 0            | null                      | null                                             | 2               |
      | 21       | 102     | 0            | null                      | null                                             | 2               |
      | 21       | 103     | 0            | null                      | null                                             | 2               |
      | 31       | 100     | 0            | 2019-01-06 09:26:40       | the group has grayed access to the item's parent | 2               |
      | 31       | 101     | 0            | 2019-01-06 09:26:40       | null                                             | 2               |
      | 31       | 102     | 0            | 2019-01-06 09:26:40       | null                                             | 2               |
      | 31       | 103     | 0            | 2019-01-06 09:26:40       | null                                             | 2               |
    When I send a PUT request to "/groups/31/items/100" with the following body:
    """
    {
      "partial_access_date": "2019-03-06T09:26:40Z",
      "full_access_date": "2019-04-06T09:26:40Z",
      "access_solutions_date": "2019-05-06T09:26:40Z",
      "access_reason": "the user really needs this access"
    }
    """
    Then the response should be "updated"
    And the table "groups_items" should be:
      | group_id | item_id | full_access_date    | cached_full_access_date | partial_access_date | cached_partial_access_date | cached_grayed_access_date | access_solutions_date | cached_access_solutions_date | cached_manager_access | owner_access | access_reason                                  | creator_user_id |
      | 21       | 100     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 1            | the admin is an owner of the item              | 2               |
      | 21       | 101     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 21       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 21       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 100     | null                | null                    | 2019-01-06 09:26:40 | 2019-01-06 09:26:40        | null                      | null                  | null                         | 0                     | 0            | the parent item is visible to the user's class | 2               |
      | 25       | 101     | null                | null                    | null                | null                       | 2019-01-06 09:26:40       | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 102     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 25       | 103     | null                | null                    | null                | null                       | null                      | null                  | null                         | 0                     | 0            | null                                           | 2               |
      | 31       | 100     | 2019-04-06 09:26:40 | 2019-04-06 09:26:40     | 2019-03-06 09:26:40 | 2019-03-06 09:26:40        | null                      | 2019-05-06 09:26:40   | 2019-05-06 09:26:40          | 0                     | 0            | the user really needs this access              | 2               |
      | 31       | 101     | null                | 2019-04-06 09:26:40     | null                | null                       | 2019-03-06 09:26:40       | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                           | 2               |
      | 31       | 102     | null                | 2019-04-06 09:26:40     | null                | null                       | null                      | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                           | 2               |
      | 31       | 103     | null                | 2019-04-06 09:26:40     | null                | null                       | null                      | null                  | 2019-05-06 09:26:40          | 0                     | 0            | null                                           | 2               |
