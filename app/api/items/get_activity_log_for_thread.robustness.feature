Feature: Get activity log for a thread - robustness
  Background:
    Given the database has the following users:
      | group_id | login |
      | 11       | jdoe  |
      | 14       | jane  |
    And the database has the following table "groups":
      | id | name | type |
      | 13 | team | Team |
    And the database has the following table "groups_groups":
      | parent_group_id | child_group_id |
      | 13              | 14             |
    And the groups ancestors are computed
    And the database has the following table "items":
      | id  | entry_participant_type | default_language_tag |
      | 200 | User                   | fr                   |
      | 210 | Team                   | fr                   |
      | 220 | Team                   | fr                   |

  Scenario: Wrong item_id
    Given I am the user with id "11"
    When I send a GET request to "/items/abc/participant/11/thread/log"
    Then the response code should be 400
    And the response error message should contain "Wrong value for item_id (should be int64)"

  Scenario: Wrong participant_id
    Given I am the user with id "11"
    When I send a GET request to "/items/200/participant/abc/thread/log"
    Then the response code should be 400
    And the response error message should contain "Wrong value for participant_id (should be int64)"

  Scenario: User doesn't exist
    Given I am the user with id "404"
    When I send a GET request to "/items/200/participant/11/thread/log"
    Then the response code should be 401
    And the response error message should contain "Invalid access token"

  Scenario: No access to the thread (the user is the participant, can_view<content)
    Given I am @User
    And there is a thread with "item_id=200,participant_id=@User,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    And I can view info of the item 200
    When I send a GET request to "/items/200/participant/@User/thread/log"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the thread (the user is a member of the participant group, can_view<content)
    Given I am @User
    And I am a member of the group @Participant
    And there is a thread with "item_id=200,participant_id=@Participant,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    And the group @User can view info of the item 200
    When I send a GET request to "/items/200/participant/@Participant/thread/log"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the thread (can_view>=content, but the user is not a member of the participant group)
    Given I am @User
    And there is a group @Participant
    And there is a thread with "item_id=200,participant_id=@Participant,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    And the group @User can view content of the item 200
    When I send a GET request to "/items/200/participant/@Participant/thread/log"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No thread (the user can view content and has can_watch>=answer, but the thread doesn't exist)
    Given I am @User
    And I have the watch permission set to "answer" on the item 200
    And I can view content of the item 200
    And there is a user @Participant
    And there is a thread with "item_id=200,participant_id=@User,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDBMs("-1h")}}"
    And there is a thread with "item_id=210,participant_id=@Participant,helper_group_id=@Helper,status=waiting_for_participant"
    When I send a GET request to "/items/200/participant/@Participant/thread/log"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the thread (the user is from the helper group with can_view>=content and can_watch>=result, has a validated result, but the thread has been expired)
    Given I am @User
    And I am a member of the group @Helper
    And I have the watch permission set to "result" on the item 200
    And I can view content of the item 200
    And I have a validated result on the item 200
    And there is a user @Participant
    And there is a thread with "item_id=200,participant_id=@Participant,helper_group_id=@Helper,status=closed,latest_update_at={{relativeTimeDB("-336h")}}"
    When I send a GET request to "/items/200/participant/@Participant/thread/log"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the thread (the user is from the helper group with can_view=content and can_watch>=result, but has a not validated result, although the thread has not been expired)
    Given I am @User
    And I am a member of the group @Helper
    And I have the watch permission set to "result" on the item 200
    And I can view content of the item 200
    And the database table "results" also has the following rows:
      | attempt_id | participant_id | item_id | validated_at |
      | 2          | @User          | 200     | null         |
    And there is a user @Participant
    And there is a thread with "item_id=200,participant_id=@Participant,helper_group_id=@Helper,status=waiting_for_participant"
    When I send a GET request to "/items/200/participant/@Participant/thread/log"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"

  Scenario: No access to the thread (the user is from the helper group with can_view=content and can_watch>=result and has a validated results,
            but there is no thread for the participant-item pair for this helper group)
    Given I am @User
    And I am a member of the group @Helper
    And I have the watch permission set to "result" on the item 200
    And I can view content of the item 200
    And I have a validated result on the item 200
    And there is a user @Participant
    And there is a thread with "item_id=200,participant_id=@User,helper_group_id=@Helper,status=waiting_for_participant"
    And there is a thread with "item_id=200,participant_id=@Participant,helper_group_id=@Participant,status=waiting_for_participant"
    And there is a thread with "item_id=210,participant_id=@Participant,helper_group_id=@Helper,status=waiting_for_participant"
    When I send a GET request to "/items/200/participant/@Participant/thread/log"
    Then the response code should be 403
    And the response error message should contain "Insufficient access rights"
