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

// Success or failure
// swagger:response publishedOrFailedResponse
type publishedOrFailedResponse struct {
	// in: body
	Body struct {
		// enum: published,failed
		// required: true
		Message string `json:"message"`
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

// OK.
// Success response with the requested answer.
// Note: when we retrieve the current answer but there is no answer, only `{"type": null}` is returned.
// swagger:response itemAnswerGetResponse
type itemAnswerGetResponse struct {
	// description: The returned answer
	// in:body
	Body struct {
		// required:true
		ID int64 `json:"id,string"`
		// required:true
		AuthorID int64 `json:"author_id,string"`
		// required:true
		ItemID int64 `json:"item_id,string"`
		// Nullable
		// format:integer
		// required:true
		AttemptID *string `json:"attempt_id,string"`
		// Can be `null` when there is no applicable existing answer for the user.
		// e.g., No answer when we try to retrieve the current answer.
		// Nullable
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
		CreatedAt time.Time `json:"created_at"`
		// Nullable
		// required:true
		Score *float32 `json:"score"`
		// Nullable
		// required:true
		GradedAt *time.Time `json:"graded_at"`
	}
}

// Created. The group has successfully entered the contest.
// swagger:response itemEnterResponse
type itemEnterResponse struct {
	// in:body
	Body struct {
		// enum: created
		// required: true
		Message string `json:"message"`
		// true
		// required: true
		Success bool `json:"success"`
		// required: true
		Data struct {
			// Nullable
			// pattern: ^\d{1,3}:[0-5]?\d:[0-5]?\d$
			// example: 838:59:59
			// required: true
			Duration *string `json:"duration"`
			// required: true
			EnteredAt database.Time `json:"entered_at"`
		} `json:"data"`
	}
}
