Feature: Export the current progress of users on a subset of items as CSV (groupUserProgressCSV)
  Background:
    Given the database has the following table 'groups':
      | id | type    | name           |
      | 1  | Base    | Root 1         |
      | 3  | Base    | Root 2         |
      | 4  | Club    | Parent         |
      | 11 | Class   | Our Class      |
      | 12 | Class   | Other Class    |
      | 13 | Class   | Special Class  |
      | 14 | Team    | Super Team     |
      | 15 | Team    | Our Team       |
      | 16 | Team    | First Team     |
      | 17 | Other   | A custom group |
      | 18 | Club    | Our Club       |
      | 19 | Club    | Another Club   |
      | 20 | Friends | My Friends     |
      | 21 | User    | owner          |
      | 51 | User    | johna          |
      | 53 | User    | johnb          |
      | 55 | User    | johnc          |
      | 57 | User    | johnd          |
      | 59 | User    | johne          |
      | 61 | User    | janea          |
      | 63 | User    | janeb          |
      | 65 | User    | janec          |
      | 67 | User    | janed          |
      | 69 | User    | janee          |
    And the database has the following table 'users':
      | login | group_id | first_name  | last_name | default_language |
      | owner | 21       | Jean-Michel | Blanquer  | en               |
      | johna | 51       | John        | Adams     | fr               |
      | johnb | 53       | John        | Black     | fr               |
      | johnc | 55       | John        | Cook      | fr               |
      | johnd | 57       | John        | null      | fr               |
      | johne | 59       | null        | Eliot     | fr               |
      | janea | 61       | null        | null      | fr               |
      | janeb | 63       | null        | null      | fr               |
      | janec | 65       | null        | null      | fr               |
      | janed | 67       | null        | null      | fr               |
      | janee | 69       | Jane        | Ebbot     | fr               |
    And the database has the following table 'group_managers':
      | group_id | manager_id | can_watch_members |
      | 1        | 21         | true              |
      | 19       | 4          | true              |
      | 51       | 4          | true              |
    And the database has the following table 'groups_groups':
      | parent_group_id | child_group_id | personal_info_view_approved_at |
      | 1               | 11             | null                           |
      | 1               | 67             | null                           |
      | 3               | 13             | null                           |
      | 4               | 21             | null                           |
      | 11              | 14             | null                           |
      | 11              | 17             | null                           |
      | 11              | 18             | null                           |
      | 11              | 59             | 2019-05-30 11:00:00            |
      | 11              | 63             | null                           |
      | 11              | 65             | null                           |
      | 13              | 15             | null                           |
      | 13              | 16             | null                           |
      | 13              | 69             | null                           |
      | 14              | 51             | 2019-05-30 11:00:00            |
      | 14              | 53             | 2019-05-30 11:00:00            |
      | 14              | 55             | null                           |
      | 15              | 57             | null                           |
      | 15              | 59             | null                           |
      | 15              | 61             | null                           |
      | 16              | 63             | null                           |
      | 16              | 65             | null                           |
      | 16              | 67             | null                           |
      | 19              | 69             | null                           |
      | 20              | 21             | null                           |
    And the groups ancestors are computed
    And the database has the following table 'items':
      | id   | type    | default_language_tag |
      | 200  | Chapter | fr                   |
      | 210  | Chapter | fr                   |
      | 211  | Task    | fr                   |
      | 212  | Task    | fr                   |
      | 213  | Task    | fr                   |
      | 214  | Task    | fr                   |
      | 215  | Task    | fr                   |
      | 216  | Task    | fr                   |
      | 217  | Task    | fr                   |
      | 218  | Task    | fr                   |
      | 219  | Task    | fr                   |
      | 220  | Chapter | fr                   |
      | 221  | Task    | fr                   |
      | 222  | Task    | fr                   |
      | 223  | Task    | fr                   |
      | 224  | Task    | fr                   |
      | 225  | Task    | fr                   |
      | 226  | Task    | fr                   |
      | 227  | Task    | fr                   |
      | 228  | Task    | fr                   |
      | 229  | Task    | fr                   |
      | 300  | Course  | fr                   |
      | 310  | Chapter | fr                   |
      | 311  | Task    | fr                   |
      | 312  | Task    | fr                   |
      | 313  | Task    | fr                   |
      | 314  | Task    | fr                   |
      | 315  | Task    | fr                   |
      | 316  | Task    | fr                   |
      | 317  | Task    | fr                   |
      | 318  | Task    | fr                   |
      | 319  | Task    | fr                   |
      | 400  | Chapter | fr                   |
      | 410  | Chapter | fr                   |
      | 411  | Task    | fr                   |
      | 412  | Task    | fr                   |
      | 413  | Task    | fr                   |
      | 414  | Task    | fr                   |
      | 415  | Task    | fr                   |
      | 416  | Task    | fr                   |
      | 417  | Task    | fr                   |
      | 418  | Task    | fr                   |
      | 419  | Task    | fr                   |
      | 1010 | Chapter | fr                   |
    And the database has the following table 'items_strings':
      | item_id | language_tag | title         |
      | 200     | fr           | Chapitre 200  |
      | 210     | fr           | Chapitre 210  |
      | 211     | fr           | Item 211      |
      | 212     | fr           | Item 212      |
      | 213     | fr           | Item 213      |
      | 214     | fr           | Item 214      |
      | 215     | fr           | Item 215      |
      | 216     | fr           | Item 216      |
      | 217     | fr           | Item 217      |
      | 218     | fr           | Item 218      |
      | 219     | fr           | Item 219      |
      | 220     | fr           | Chapitre 220  |
      | 220     | en           | Chapter 220   |
      | 221     | fr           | Item 221      |
      | 222     | fr           | Item 222      |
      | 223     | fr           | Item 223      |
      | 224     | fr           | Item 224      |
      | 225     | fr           | Item 225      |
      | 226     | fr           | Item 226      |
      | 227     | fr           | Item 227      |
      | 228     | fr           | Item 228      |
      | 229     | fr           | Item 229      |
      | 300     | fr           | Cours 300     |
      | 310     | fr           | Chapitre 310  |
      | 311     | fr           | Item 311      |
      | 312     | fr           | Item 312      |
      | 313     | fr           | Item 313      |
      | 314     | fr           | Item 314      |
      | 315     | fr           | Item 315      |
      | 316     | fr           | Item 316      |
      | 317     | fr           | Item 317      |
      | 318     | fr           | Item 318      |
      | 319     | fr           | Item 319      |
      | 400     | fr           | Chapitre 400  |
      | 410     | fr           | Chapitre 410  |
      | 411     | fr           | Item 411      |
      | 412     | fr           | Item 412      |
      | 413     | fr           | Item 413      |
      | 414     | fr           | Item 414      |
      | 415     | fr           | Item 415      |
      | 416     | fr           | Item 416      |
      | 417     | fr           | Item 417      |
      | 418     | fr           | Item 418      |
      | 419     | fr           | Item 419      |
      | 1010    | fr           | Chapitre 1010 |
    And the database has the following table 'items_items':
      | parent_item_id | child_item_id | child_order |
      | 200            | 210           | 0           |
      | 200            | 220           | 1           |
      | 210            | 211           | 0           |
      | 210            | 212           | 1           |
      | 210            | 213           | 2           |
      | 210            | 214           | 3           |
      | 210            | 215           | 4           |
      | 210            | 216           | 5           |
      | 210            | 217           | 6           |
      | 210            | 218           | 7           |
      | 210            | 219           | 8           |
      | 220            | 221           | 0           |
      | 220            | 222           | 1           |
      | 220            | 223           | 2           |
      | 220            | 224           | 3           |
      | 220            | 225           | 4           |
      | 220            | 226           | 5           |
      | 220            | 227           | 6           |
      | 220            | 228           | 7           |
      | 220            | 229           | 8           |
      | 300            | 310           | 0           |
      | 310            | 311           | 0           |
      | 310            | 312           | 1           |
      | 310            | 313           | 2           |
      | 310            | 314           | 3           |
      | 310            | 315           | 4           |
      | 310            | 316           | 5           |
      | 310            | 317           | 6           |
      | 310            | 318           | 7           |
      | 310            | 319           | 8           |
      | 400            | 410           | 0           |
      | 410            | 411           | 0           |
      | 410            | 412           | 1           |
      | 410            | 413           | 2           |
      | 410            | 414           | 3           |
      | 410            | 415           | 4           |
      | 410            | 416           | 5           |
      | 410            | 417           | 6           |
      | 410            | 418           | 7           |
      | 410            | 419           | 8           |
    And the database has the following table 'permissions_generated':
      | group_id | item_id | can_view_generated       | can_watch_generated |
      | 21       | 210     | none                     | result              |
      | 21       | 211     | info                     | none                |
      | 20       | 212     | content                  | none                |
      | 21       | 213     | content_with_descendants | none                |
      | 20       | 214     | info                     | none                |
      | 21       | 215     | content                  | none                |
      | 20       | 216     | none                     | none                |
      | 21       | 217     | none                     | none                |
      | 20       | 218     | none                     | none                |
      | 21       | 219     | none                     | none                |
      | 20       | 220     | none                     | answer              |
      | 20       | 221     | info                     | none                |
      | 21       | 222     | content                  | none                |
      | 20       | 223     | content_with_descendants | none                |
      | 21       | 224     | info                     | none                |
      | 20       | 225     | content                  | none                |
      | 21       | 226     | none                     | none                |
      | 20       | 227     | none                     | none                |
      | 21       | 228     | none                     | none                |
      | 20       | 229     | none                     | none                |
      | 4        | 310     | none                     | none                |
      | 20       | 310     | none                     | result              |
      | 21       | 311     | info                     | none                |
      | 20       | 312     | content                  | none                |
      | 21       | 313     | content_with_descendants | none                |
      | 20       | 314     | info                     | none                |
      | 21       | 315     | content                  | none                |
      | 20       | 316     | none                     | none                |
      | 21       | 317     | none                     | none                |
      | 20       | 318     | none                     | none                |
      | 21       | 319     | none                     | none                |
      | 20       | 411     | info                     | none                |
      | 21       | 412     | content                  | none                |
      | 20       | 413     | content_with_descendants | none                |
      | 21       | 414     | info                     | none                |
      | 20       | 415     | content                  | none                |
      | 21       | 416     | none                     | none                |
      | 20       | 417     | none                     | none                |
      | 21       | 418     | none                     | none                |
      | 20       | 419     | none                     | none                |
      | 4        | 1010    | none                     | answer_with_grant   |
    And the database has the following table 'attempts':
      | id | participant_id | created_at          |
      | 0  | 14             | 2017-05-29 06:38:38 |
      | 1  | 14             | 2017-05-29 06:38:38 |
      | 2  | 14             | 2017-05-29 06:38:38 |
      | 3  | 14             | 2017-05-29 06:38:38 |
      | 0  | 15             | 2017-03-29 06:38:38 |
      | 0  | 16             | 2018-12-01 00:00:00 |
      | 1  | 67             | 2019-01-01 00:00:00 |
      | 0  | 67             | 2017-05-29 06:38:38 |
      | 4  | 14             | 2017-05-29 06:38:38 |
      | 1  | 15             | 2017-03-29 06:38:38 |
      | 5  | 14             | 2017-05-29 06:38:38 |
    And the database has the following table 'results':
      | attempt_id | participant_id | item_id | started_at          | score_computed | score_obtained_at   | hints_cached | submissions | validated_at        | latest_activity_at  |
      | 0          | 14             | 210     | 2017-05-29 05:38:38 | 50             | 2017-05-29 06:38:38 | 115          | 127         | null                | 2019-05-29 06:38:38 |
      | 0          | 14             | 211     | 2017-05-29 06:38:38 | 0              | 2017-05-29 06:38:38 | 100          | 100         | 2017-05-30 06:38:38 | 2018-05-30 06:38:38 | # latest_activity_at for 51, 211 comes from this line (the last activity is made by a team)
      | 1          | 14             | 211     | 2017-05-29 06:38:38 | 40             | 2017-05-29 06:38:38 | 2            | 3           | 2017-05-29 06:38:58 | 2018-05-29 06:38:38 | # min(validated_at) for 51, 211 comes from this line (from a team)
      | 2          | 14             | 211     | 2017-05-29 06:38:38 | 50             | 2017-05-29 06:38:38 | 3            | 4           | 2017-05-31 06:58:38 | 2018-05-28 06:38:38 | # hints_cached & submissions for 51, 211 come from this line (the best attempt is made by a team)
      | 3          | 14             | 211     | 2017-05-29 06:38:38 | 50             | 2017-05-30 06:38:38 | 10           | 20          | null                | 2018-05-27 06:38:38 |
      | 0          | 15             | 211     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                | 2018-05-26 06:38:38 |
      | 0          | 15             | 212     | 2017-03-29 06:38:38 | 0              | null                | 0            | 0           | null                | 2018-05-25 06:38:38 |
      | 0          | 16             | 212     | 2018-12-01 00:00:00 | 10             | 2017-05-30 06:38:38 | 0            | 0           | null                | 2019-06-01 00:00:00 | # started_at for 67, 212 & 63, 212 comes from this line (the first attempt is started by a team)
      | 0          | 67             | 212     | 2019-01-01 00:00:00 | 20             | 2017-06-30 06:38:38 | 1            | 2           | null                | 2019-06-01 00:00:00 | # hints_cached & submissions for 67, 212 come from this line (the best attempt is made by a user)
      | 1          | 67             | 212     | 2019-01-01 00:00:00 | 10             | 2017-05-30 06:38:38 | 6            | 7           | null                | 2019-07-01 00:00:00 | # latest_activity_at for 67, 212 comes from this line (the last activity is made by a user)
      | 0          | 67             | 213     | 2018-11-01 00:00:00 | 0              | null                | 0            | 0           | null                | 2018-11-01 00:00:00 | # started_at for 67, 213 comes from this line (the first attempt is started by a user)
      | 0          | 67             | 214     | 2017-05-29 06:38:38 | 15             | 2017-05-29 06:38:48 | 10           | 11          | 2017-05-29 06:38:48 | 2017-05-30 06:38:48 | # min(validated_at) for 67, 214 comes from this line (from a user)
      | 0          | 67             | 215     | 3018-11-01 00:00:00 | 0              | null                | 0            | 0           | null                | 2018-11-01 00:00:00 | # started_at for 67, 213 comes from this line (the first attempt is started by a user)
      | 4          | 14             | 211     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                | 2018-05-24 06:38:38 |
      | 1          | 15             | 211     | 2017-04-29 06:38:38 | 0              | null                | 0            | 0           | null                | 2018-05-23 06:38:38 |
      | 1          | 15             | 212     | 2017-03-29 06:38:38 | 100            | null                | 0            | 0           | null                | 2018-05-22 06:38:38 |
      | 5          | 14             | 211     | 2017-05-29 06:38:38 | 0              | null                | 0            | 0           | null                | 2018-05-21 06:38:38 |
      | 5          | 14             | 1010    | 2017-05-29 06:38:38 | 10             | null                | 0            | 0           | null                | 2018-05-21 06:38:38 |

  Scenario: Get progress of users
    Given I am the user with id "21"
    When I send a GET request to "/groups/1/user-progress-csv?parent_item_ids=210,220"
    Then the response code should be 200
    And the response header "Content-Type" should be "text/csv"
    And the response header "Content-Disposition" should be "attachment; filename=users_progress_for_group_1_and_child_items_of_210_220.csv"
    And the response body should be:
    """
    Login;First name;Last name;Chapitre 210;1. Item 211;2. Item 212;3. Item 213;4. Item 214;5. Item 215;Chapter 220;1. Item 221;2. Item 222;3. Item 223;4. Item 224;5. Item 225
    janeb;;;;;10;;;;;;;;;
    janec;;;;;10;;;;;;;;;
    janed;;;;;20;0;15;0;;;;;;
    johna;Adams;John;50;50;;;;;;;;;;
    johnb;Black;John;50;50;;;;;;;;;;
    johnc;;;50;50;;;;;;;;;;
    johne;Eliot;;;0;100;;;;;;;;;

    """


  Scenario: Get progress of the first user for all the visible items (also checks the limit)
    Given I am the user with id "21"
    When I send a GET request to "/groups/1/user-progress-csv?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response header "Content-Type" should be "text/csv"
    And the response header "Content-Disposition" should be "attachment; filename=users_progress_for_group_1_and_child_items_of_210_220_310.csv"
    And the response body should be:
    """
    Login;First name;Last name;Chapitre 210;1. Item 211;2. Item 212;3. Item 213;4. Item 214;5. Item 215;Chapter 220;1. Item 221;2. Item 222;3. Item 223;4. Item 224;5. Item 225;Chapitre 310;1. Item 311;2. Item 312;3. Item 313;4. Item 314;5. Item 315
    janeb;;;;;10;;;;;;;;;;;;;;;
    janec;;;;;10;;;;;;;;;;;;;;;
    janed;;;;;20;0;15;0;;;;;;;;;;;;
    johna;Adams;John;50;50;;;;;;;;;;;;;;;;
    johnb;Black;John;50;50;;;;;;;;;;;;;;;;
    johnc;;;50;50;;;;;;;;;;;;;;;;
    johne;Eliot;;;0;100;;;;;;;;;;;;;;;

    """

  Scenario: No users
    Given I am the user with id "21"
    When I send a GET request to "/groups/17/user-progress-csv?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response header "Content-Type" should be "text/csv"
    And the response header "Content-Disposition" should be "attachment; filename=users_progress_for_group_17_and_child_items_of_210_220_310.csv"
    And the response body should be:
    """
    Login;First name;Last name;Chapitre 210;1. Item 211;2. Item 212;3. Item 213;4. Item 214;5. Item 215;Chapter 220;1. Item 221;2. Item 222;3. Item 223;4. Item 224;5. Item 225;Chapitre 310;1. Item 311;2. Item 312;3. Item 313;4. Item 314;5. Item 315

    """

  Scenario: No visible child items
    Given I am the user with id "21"
    When I send a GET request to "/groups/1/user-progress-csv?parent_item_ids=1010"
    Then the response code should be 200
    And the response header "Content-Type" should be "text/csv"
    And the response header "Content-Disposition" should be "attachment; filename=users_progress_for_group_1_and_child_items_of_1010.csv"
    And the response body should be:
    """
    Login;First name;Last name;Chapitre 1010
    janeb;;;
    janec;;;
    janed;;;
    johna;Adams;John;10
    johnb;Black;John;10
    johnc;;;10
    johne;Eliot;;

    """

  Scenario: The input group_id is a user
    Given I am the user with id "21"
    When I send a GET request to "/groups/51/user-progress-csv?parent_item_ids=210,220,310"
    Then the response code should be 200
    And the response header "Content-Type" should be "text/csv"
    And the response header "Content-Disposition" should be "attachment; filename=users_progress_for_group_51_and_child_items_of_210_220_310.csv"
    And the response body should be:
    """
    Login;First name;Last name;Chapitre 210;1. Item 211;2. Item 212;3. Item 213;4. Item 214;5. Item 215;Chapter 220;1. Item 221;2. Item 222;3. Item 223;4. Item 224;5. Item 225;Chapitre 310;1. Item 311;2. Item 312;3. Item 313;4. Item 314;5. Item 315

    """

  Scenario: The input group_id is a team
    Given I am the user with id "21"
    When I send a GET request to "/groups/14/user-progress-csv?parent_item_ids=210"
    Then the response code should be 200
    And the response header "Content-Type" should be "text/csv"
    And the response header "Content-Disposition" should be "attachment; filename=users_progress_for_group_14_and_child_items_of_210.csv"
    And the response body should be:
    """
    Login;First name;Last name;Chapitre 210;1. Item 211;2. Item 212;3. Item 213;4. Item 214;5. Item 215
    johna;Adams;John;50;50;;;;
    johnb;Black;John;50;50;;;;
    johnc;;;50;50;;;;

    """

  Scenario: Users are direct members of the input group_id which is not a team
    Given I am the user with id "21"
    When I send a GET request to "/groups/19/user-progress-csv?parent_item_ids=210"
    Then the response code should be 200
    And the response header "Content-Type" should be "text/csv"
    And the response header "Content-Disposition" should be "attachment; filename=users_progress_for_group_19_and_child_items_of_210.csv"
    And the response body should be:
    """
    Login;First name;Last name;Chapitre 210;1. Item 211;2. Item 212;3. Item 213;4. Item 214;5. Item 215
    janee;;;;;;;;

    """

  Scenario: No parent item ids given
    Given I am the user with id "21"
    When I send a GET request to "/groups/1/user-progress-csv?parent_item_ids="
    Then the response code should be 200
    And the response header "Content-Type" should be "text/csv"
    And the response header "Content-Disposition" should be "attachment; filename=users_progress_for_group_1_and_child_items_of_.csv"
    And the response body should be:
    """
    Login;First name;Last name

    """
