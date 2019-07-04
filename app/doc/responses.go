package doc

import "time"

// These definitions are unused by code, just used to generate documentation

// The request has successfully updated the object
// swagger:response updatedResponse
type updatedResponse struct {
	// in: body
	Body struct {
		// "updated"
		// enum: updated
		// required: true
		Message string `json:"message"`
		// true
		// required: true
		Success bool `json:"success"`
	}
}

type CreatedResponse struct {
	// "created"
	// enum: created
	// required: true
	Message string `json:"message"`
	// true
	// required: true
	Success bool `json:"success"`
}

// OK. Success response with the requested answer
// swagger:response itemAnswerGetResponse
type itemAnswerGetResponse struct {
	// description: The returned answer
	// in:body
	Body struct {
		// required:true
		ID int64 `json:"id,string"`
		// required:true
		UserID int64 `json:"user_id,string"`
		// required:true
		ItemID int64 `json:"item_id,string"`
		// Nullable
		// format:integer
		// required:true
		AttemptID *string `json:"attempt_id,string"`
		// required:true
		// enum:Submission,Saved,Current
		Type string `json:"type"`
		// Nullable
		// required:true
		State *string `json:"state"`
		// Nullable
		// required:true
		Answer *string `json:"answer"`
		// required:true
		SubmissionDate time.Time `json:"submission_date"`
		// Nullable
		// required:true
		Score *float32 `json:"score"`
		// Nullable
		// required:true
		Validated *bool `json:"validated"`
		// Nullable
		// required:true
		GradingDate *time.Time `json:"grading_date"`
		// Nullable
		// format:integer
		// required:true
		UserGraderID *string `json:"user_grader_id"`
	}
}

// OK. Success response with groups progress on items
// For all children of items in the parent_item_id list, display the result for each direct child
// of the given group_id whose type is not in (Team,UserSelf). Values are averages of all the group's
// "end-members" where “end-member” defined as descendants of the group which are either
// 1) teams or
// 2) users who descend from the input group not only through teams (one or more).
// swagger:response groupsGetGroupProgressResponse
type groupsGetGroupProgressResponse struct {
	// in: body
	Body []struct {
		// The child’s `group_id`
		// required:true
		GroupID int64 `json:"group_id,string"`
		// required:true
		ItemID int64 `json:"item_id,string"`
		// Average score of all "end-members".
		// The score of an "end-member" is the max of his `groups_attempt.iScore` or 0 if no attempts.
		// required:true
		AverageScore float32 `json:"average_score"`
		// % (float [0,1]) of "end-members" who have validated the task.
		// An "end-member" has validated a task if one of his attempts has `groups_attempts.bValidated` = 1.
		// No attempts for an "end-member" is considered as not validated.
		// required:true
		ValidationRate float32 `json:"validation_rate"`
		// Average number of hints requested by each "end-member".
		// The number of hints requested of an "end-member" is the `groups_attempts.nbHintsCached`
		// of the attempt with the best score
		// (if several with the same score, we use the first attempt chronologically on `sBestAnswerDate`).
		// required:true
		AvgHintsRequested float32 `json:"avg_hints_requested"`
		// Average number of submissions made by each "end-member".
		// The number of submissions made by an "end-member" is the `groups_attempts.nbSubmissionsAttempts`.
		// of the attempt with the best score
		// (if several with the same score, we use the first attempt chronologically on `sBestAnswerDate`).
		// required:true
		AvgSubmissionsAttempts float32 `json:"avg_submissions_attempts"`
		// Average time spent among all the "end-members" (in seconds). The time spent by an "end-member" is computed as:
		//
		//   1) if no attempts yet: 0
		//
		//   2) if one attempt validated: min(`sValidationDate`) - min(`sStartDate`)
		//     (i.e., time between the first time it started one (any) attempt
		//      and the time he first validated the task)
		//
		//   3) if no attempts validated: `now` - min(`sStartDate`)
		// required:true
		AvgTimeSpent float32 `json:"avg_time_spent"`
	}
}
