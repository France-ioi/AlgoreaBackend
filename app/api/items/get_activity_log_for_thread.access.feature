Feature: Get activity log for a thread - access
  Background:
    Given the database has the following table "items":
      | id     | default_language_tag |
      | @Item1 | fr                   |
      | @Item2 | fr                   |

  Scenario Outline: User has can_view>=content on the item and the thread.participant_id = authenticated user's self group
    Given I am @User
    And I can view <view_permission> of the item @Item1
    And there is a thread with "item_id=@Item1,participant_id=@User,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1000h")}}"
    When I send a GET request to "/items/@Item1/participant/@User/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """
  Examples:
    | view_permission          |
    | content                  |
    | content_with_descendants |
    | solution                 |

  Scenario: User has can_view>=content on the item (via an ancestor group) and the threads.participant_id = authenticated user's self group
    Given I am @User
    And I am a member of the group @ParentGroup
    And the group @ParentGroup can view content of the item @Item1
    And there is a thread with "item_id=@Item1,participant_id=@User,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1000h")}}"
    When I send a GET request to "/items/@Item1/participant/@User/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: User has can_view>=content on the item and the user is a team member of threads.participant_id
    Given I am @User
    And I can view content of the item @Item2
    And there is a team @Team
    And I am a member of the group @Team
    And there is a thread with "item_id=@Item2,participant_id=@Team,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1000h")}}"
    When I send a GET request to "/items/@Item2/participant/@Team/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: One of the user's teams has can_view>=content on the item and the user is a team member of threads.participant_id
    Given I am @User
    And there is a team @Team1
    And there is a team @Team2
    And I am a member of the group @Team1
    And I am a member of the group @Team2
    And the group @Team1 can view content of the item @Item2
    And there is a thread with "item_id=@Item2,participant_id=@Team2,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1000h")}}"
    When I send a GET request to "/items/@Item2/participant/@Team2/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: One of the user's teams has can_view>=content (via an ancestor) on the item and the user is a team member of threads.participant_id
    Given I am @User
    And there is a team @Team1
    And there is a team @Team2
    And I am a member of the group @Team1
    And I am a member of the group @Team2
    And the group @Team1 is a child of the group @TeamParent
    And the group @TeamParent can view content of the item @Item2
    And there is a thread with "item_id=@Item2,participant_id=@Team2,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1000h")}}"
    When I send a GET request to "/items/@Item2/participant/@Team2/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario: User can view content and has can_watch>=answer on the threads.item_id
    Given I am @User
    And I can view content of the item @Item2
    And I have the watch permission set to "answer" on the item @Item2
    And there is a thread with "item_id=@Item2,participant_id=@Participant,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1000h")}}"
    When I send a GET request to "/items/@Item2/participant/@Participant/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """

  Scenario Outline: User can view content and has can_watch>=answer (via an ancestor) on the threads.item_id
    Given I am @User
    And I am a member of the group @ChildGroupAbleToWatch
    And the group @ChildGroupAbleToWatch is a child of the group @GroupAbleToWatch
    And the group @GroupAbleToWatch has the watch permission set to "<watch_permission>" on the item @Item2
    And the group @ChildGroupAbleToWatch can view content of the item @Item2
    And there is a thread with "item_id=@Item2,participant_id=@Participant,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1000h")}}"
    When I send a GET request to "/items/@Item2/participant/@Participant/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """
  Examples:
    | watch_permission  |
    | answer            |
    | answer_with_grant |

  Scenario Outline: User can view content and can_watch=result on the item and is a thread reader, and has a validated result on the item
    Given I am @User
    And I have the watch permission set to "result" on the item @Item2
    And I can view content of the item @Item2
    And I have a validated result on the item @Item2
    And I am a member of the group @Helper
    And there is a thread with "item_id=@Item2,participant_id=@Participant,helper_group_id=@Helper,status=<thread_status>,latest_update_at=<thread_latest_update_at>"
    When I send a GET request to "/items/@Item2/participant/@Participant/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """
  Examples:
    | thread_status           | thread_latest_update_at        |
    | waiting_for_participant | 2020-05-30 12:00:00            |
    | waiting_for_trainer     | 2020-05-30 12:00:00            |
    | closed                  | {{relativeTimeDB("-335h59m")}} |

  Scenario Outline: User can view content and has can_watch=result on the item (via an ancestor group) and is a thread reader, and has a validated result on the item
    Given I am @User
    And I am a member of the group @ParentGroup
    And the group @ParentGroup has the watch permission set to "result" on the item @Item2
    And the group @ParentGroup can view content of the item @Item2
    And I am a member of the group @Helper
    And there is a thread with "item_id=@Item2,participant_id=@Participant,helper_group_id=@Helper,status=<thread_status>,latest_update_at=<thread_latest_update_at>"
    And I have a validated result on the item @Item2
    When I send a GET request to "/items/@Item2/participant/@Participant/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """
  Examples:
    | thread_status           | thread_latest_update_at        |
    | waiting_for_participant | 2020-05-30 12:00:00            |
    | waiting_for_trainer     | 2020-05-30 12:00:00            |
    | closed                  | {{relativeTimeDB("-335h59m")}} |

  Scenario Outline: User can view content and has can_watch=result on the item and is a thread reader, and one of the user's teams has a validated result on the item
    Given I am @User
    And I have the watch permission set to "result" on the item @Item2
    And I can view content of the item @Item2
    And I am a member of the group @Helper
    And there is a thread with "item_id=@Item2,participant_id=@Participant,helper_group_id=@Helper,status=<thread_status>,latest_update_at=<thread_latest_update_at>"
    And there is a team @Team1
    And there is a team @Team2
    And I am a member of the group @Team1
    And I am a member of the group @Team2
    And the group @Team2 has a validated result on the item @Item2
    When I send a GET request to "/items/@Item2/participant/@Participant/thread/log"
    Then the response code should be 200
    And the response body should be, in JSON:
    """
    []
    """
  Examples:
    | thread_status           | thread_latest_update_at        |
    | waiting_for_participant | 2020-05-30 12:00:00            |
    | waiting_for_trainer     | 2020-05-30 12:00:00            |
    | closed                  | {{relativeTimeDB("-335h59m")}} |
