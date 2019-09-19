package doc

import (
	"time"

	"github.com/France-ioi/AlgoreaBackend/app/database"
)

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

// The request has succeeded. The `data.changed` shows if the object has been updated.
// swagger:response updatedOrUnchangedResponse
type updatedOrUnchangedResponse struct {
	// in: body
	Body struct {
		// enum: updated,unchanged
		// required: true
		Message string `json:"message"`
		// true
		// required: true
		Success bool `json:"success"`
		// required: true
		Data struct {
			// required: true
			Changed bool `json:"changed"`
		} `json:"data"`
	}
}

// Success
// swagger:response successResponse
type successResponse struct {
	// in: body
	Body struct {
		// "success"
		// enum: success
		// required: true
		Message string `json:"message"`
		// true
		// required: true
		Success bool `json:"success"`
	}
}

// The request has successfully deleted the object
// swagger:response deletedResponse
type deletedResponse struct {
	// in: body
	Body struct {
		// "deleted"
		// enum: deleted
		// required: true
		Message string `json:"message"`
		// true
		// required: true
		Success bool `json:"success"`
	}
}

// The request has succeeded. The `data.changed` shows if the object has been deleted.
// swagger:response deletedOrUnchangedResponse
type deletedOrUnchangedResponse struct {
	// in: body
	Body struct {
		// enum: deleted,unchanged
		// required: true
		Message string `json:"message"`
		// true
		// required: true
		Success bool `json:"success"`
		// required: true
		Data struct {
			// required: true
			Changed bool `json:"changed"`
		} `json:"data"`
	}
}

// Created. Success response with the created object's id.
// swagger:response createdWithIDResponse
type createdWithIDResponse struct {
	// in: body
	Body struct {
		// enum: created
		// required: true
		Message string `json:"message"`
		// true
		// required: true
		Success bool `json:"success"`
		// required: true
		Data struct {
			// required: true
			ID int64 `json:"id,string"`
		} `json:"data"`
	}
}

// The request has succeeded. The `data.changed` shows if the object has been created.
// swagger:response createdOrUnchangedResponse
type createdOrUnchangedResponse struct {
	// in: body
	Body struct {
		// enum: created,unchanged
		// required: true
		Message string `json:"message"`
		// true
		// required: true
		Success bool `json:"success"`
		// required: true
		Data struct {
			// required: true
			Changed bool `json:"changed"`
		} `json:"data"`
	}
}

// OK. Success response with the per-group update statuses
// swagger:response
type updatedGroupRelationsResponse struct {
	// in:body
	Body struct {
		// "updated"
		// enum: updated
		// required: true
		Message string `json:"message"`
		// true
		// required: true
		Success bool `json:"success"`
		// `group_id` -> `result`
		// required: true
		Data map[string]database.GroupGroupTransitionResult `json:"data"`
	}
}

// enum: [cycle, invalid, success, unchanged, not_found]
type loginTransitionResult string

// Created. Success response with the per-login results
// swagger:response createdLoginRelationsResponse
type createdLoginRelationsResponse struct {
	// in:body
	Body struct {
		// "created"
		// enum: created
		// required: true
		Message string `json:"message"`
		// true
		// enum: true
		// required: true
		Success bool `json:"success"`
		// `login` -> `result`
		// required: true
		Data map[string]loginTransitionResult `json:"data"`
	}
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
