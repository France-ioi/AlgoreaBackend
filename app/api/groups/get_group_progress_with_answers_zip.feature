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
