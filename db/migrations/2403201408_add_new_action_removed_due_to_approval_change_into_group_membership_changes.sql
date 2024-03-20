-- +migrate Up
ALTER TABLE `group_membership_changes`
  MODIFY `action`
    ENUM('invitation_created','invitation_withdrawn','invitation_refused','invitation_accepted',
      'join_request_created','join_request_withdrawn','join_request_refused','join_request_accepted',
      'leave_request_created','leave_request_withdrawn','leave_request_refused','leave_request_accepted',
      'left','removed','joined_by_code','added_directly','expired','joined_by_badge', 'removed_due_to_approval_change') DEFAULT NULL;

-- +migrate Down
UPDATE `group_membership_changes` SET `action` = 'removed' WHERE `action` = 'removed_due_to_approval_change';
ALTER TABLE `group_membership_changes`
  MODIFY `action`
    ENUM('invitation_created','invitation_withdrawn','invitation_refused','invitation_accepted',
      'join_request_created','join_request_withdrawn','join_request_refused','join_request_accepted',
      'leave_request_created','leave_request_withdrawn','leave_request_refused','leave_request_accepted',
      'left','removed','joined_by_code','added_directly','expired','joined_by_badge') DEFAULT NULL;
