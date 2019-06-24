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
