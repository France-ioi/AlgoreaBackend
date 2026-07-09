Feature: Export the current progress of a group with answers on a subset of items as a ZIP file (groupGroupProgressWithAnswersZIP)
  Scenario: Should include empty group_progress.csv when no parent item ids are given
    Given I am @Teacher
    And there is a group @Classroom
    And I am a manager of the group @Classroom and can watch for submissions from the group and its descendants
    When I send a GET request to "/groups/@Classroom/group-progress-with-answers-zip?parent_item_ids="
    Then the response code should be 200
    And the response header "Content-Type" should be "application/zip"
    And the response header "Content-Disposition" should be "attachment; filename=groups_progress_with_answers_for_group-@Classroom-and_child_items_of-.zip"
    And the response should be a ZIP file containing the following files:
      """
        [
          {
            "filename": "group_progress.csv",
            "content": "Group name\n"
          }
        ]
      """

  Scenario: Should export nested progress, submissions and CSV content
    Given I am the user with id "21"
    And the database has the following table "groups":
      | id | type  | name       |
      | 1  | Base  | Root 1     |
      | 11 | Class | Our Class  |
      | 15 | Team  | Our Team   |
    And the database has the following users:
      | group_id | login | default_language |
      | 21       | owner | en               |
      | 57       | johnd | fr               |
    And the database has the following table "group_managers":
      | group_id | manager_id | can_watch_members |
      | 1        | 21         | true              |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id | is_team_membership |
      | 1               | 11             | 0                  |
      | 11              | 15             | 0                  |
      | 15              | 57             | 1                  |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id  | type    | default_language_tag |
      | 210 | Chapter | fr                   |
      | 211 | Task    | fr                   |
      | 212 | Task    | fr                   |
      | 213 | Task    | fr                   |
      | 220 | Chapter | fr                   |
    And the database has the following table "items_strings":
      | item_id | language_tag | title        |
      | 210     | fr           | Chapitre 210 |
      | 211     | fr           | Item 211     |
      | 212     | fr           | Item 212     |
      | 213     | fr           | SubTask 213  |
      | 220     | fr           | Chapitre 220 |
      | 220     | en           | Chapter 220  |
    And the database has the following table "items_items":
      | parent_item_id | child_item_id | child_order |
      | 210            | 211           | 1           |
      | 210            | 212           | 1           |
      | 211            | 213           | 2           |
    And the database has the following table "permissions_generated":
      | group_id | item_id | can_view_generated | can_watch_generated |
      | 21       | 210     | info               | answer              |
      | 21       | 211     | info               | none                |
      | 21       | 212     | info               | none                |
      | 21       | 213     | info               | none                |
      | 21       | 220     | info               | answer              |
    And the database has the following table "attempts":
      | id | participant_id | created_at          |
      | 3  | 57             | 2019-01-01 00:00:00 |
    And the database has the following table "results":
      | attempt_id | participant_id | item_id | started_at          | score_computed | score_obtained_at   | hints_cached | submissions | validated_at        | latest_activity_at  |
      | 3          | 57             | 213     | 2019-01-01 00:00:00 | 15             | 2019-01-02 00:00:00 | 2            | 3           | 2019-01-03 00:00:00 | 2019-01-04 00:00:00 |
    And the results are computed
    And the database has the following table "answers":
      | id   | author_id | participant_id | attempt_id | item_id | type       | answer      | created_at          |
      | 9001 | 57        | 57             | 3          | 213     | Submission | first try   | 2019-01-01 01:00:00 |
      | 9002 | 57        | 57             | 3          | 213     | Submission | second try  | 2019-01-01 02:00:00 |
    And the database has the following table "gradings":
      | answer_id | score | graded_at           |
      | 9001      | 5     | 2019-01-01 01:30:00 |
      | 9002      | 15    | 2019-01-01 02:30:00 |
    When I send a GET request to "/groups/11/group-progress-with-answers-zip?parent_item_ids=210,220"
    Then the response code should be 200
    And the response header "Content-Type" should be "application/zip"
    And the response header "Content-Disposition" should be "attachment; filename=groups_progress_with_answers_for_group-11-and_child_items_of-210-220.zip"
    And the response should be a ZIP file containing the following files:
      """
        [
          {
            "filename": "group_progress.csv",
            "content": "Group name;Chapitre 210;1. Item 211;2. Item 212\n"
          },
          {
            "filename": "0-Chapitre 210-210/submissions/johnd/data.json",
            "content": "{\"hints_requested\":0,\"latest_activity_at\":null,\"score\":0,\"submissions\":0,\"time_spent\":0,\"validated\":false}"
          },
          {
            "filename": "0-Chapitre 210-210/1-Item 211-211/submissions/johnd/data.json",
            "content": "{\"hints_requested\":0,\"latest_activity_at\":null,\"score\":0,\"submissions\":0,\"time_spent\":0,\"validated\":false}"
          },
          {
            "filename": "0-Chapitre 210-210/1-Item 211-211/2-SubTask 213-213/submissions/johnd/data.json",
            "content": "{\"hints_requested\":2,\"latest_activity_at\":\"2019-01-04T00:00:00Z\",\"score\":15,\"submissions\":3,\"time_spent\":172800,\"validated\":true}"
          },
          {
            "filename": "0-Chapitre 210-210/1-Item 211-211/2-SubTask 213-213/submissions/johnd/0-3-5-9001.txt",
            "content": "first try"
          },
          {
            "filename": "0-Chapitre 210-210/1-Item 211-211/2-SubTask 213-213/submissions/johnd/1-3-15-9002.txt",
            "content": "second try"
          },
          {
            "filename": "0-Chapitre 210-210/1-Item 212-212/submissions/johnd/data.json",
            "content": "{\"hints_requested\":0,\"latest_activity_at\":null,\"score\":0,\"submissions\":0,\"time_spent\":0,\"validated\":false}"
          },
          {
            "filename": "0-Chapter 220-220/submissions/johnd/data.json",
            "content": "{\"hints_requested\":0,\"latest_activity_at\":null,\"score\":0,\"submissions\":0,\"time_spent\":0,\"validated\":false}"
          }
        ]
      """
