Feature: Export the current progress of a group with answers on a subset of items as a ZIP file (groupGroupProgressWithAnswersZIP)
  Scenario: Should include empty group_progress.csv when no parent item ids are given
    Given I am @Teacher
    And there is a group @Classroom
    And I am a manager of the group @Classroom and can watch its members
    When I send a GET request to "/groups/@Classroom/group-progress-with-answers-zip?parent_item_ids="
    Then the response code should be 200
    And the response header "Content-Type" should be "application/zip"
    And the response header "Content-Disposition" should be "attachment; filename=groups_progress_with_answers_for_group-@Classroom-and_child_items_of-.zip"
    And the response should be a ZIP file containing the following files:
      """
        [
          {
            "filename": "group_progress.csv",
            "content": ""
          }
        ]
      """

  Scenario: Should contain only the given item when it has no children, without submissions when there is no none
    Given I am @Teacher
    And there are the following groups:
      | group      | members  |
      | @Classroom | @Student |
    And I am a manager of the group @Classroom and can watch its members
    And there are the following items:
      | item  | type |
      | @Item | Task |
    And there are the following item strings:
      | item  | language_tag | title      |
      | @Item | fr           | item title |
    And there are the following item permissions:
      | item  |  | group    | can_watch | can_view |
      | @Item |  | @Teacher | answer    |          |
      | @Item |  | @Student |           | content  |
    When I send a GET request to "/groups/@Classroom/group-progress-with-answers-zip?parent_item_ids=@Item"
    Then the response code should be 200
    And the response header "Content-Type" should be "application/zip"
    And the response header "Content-Disposition" should be "attachment; filename=groups_progress_with_answers_for_group-@Classroom-and_child_items_of-@Item.zip"
    And the response should be a ZIP file containing the following files:
      """
        [
          {
            "filename": "group_progress.csv",
            "content": "Group name\n"
          },
          {
            "filename": "@Item-item title/submissions/Student/data.json",
            "content": ""
          }
        ]
      """
