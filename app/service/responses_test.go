package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/render"
	assertlib "github.com/stretchr/testify/assert"
)

func httpResponseForResponse(renderer render.Renderer) *httptest.ResponseRecorder {
	var fn AppHandler = func(respW http.ResponseWriter, req *http.Request) *APIError {
		_ = render.Render(respW, req, renderer)
		return NoError
	}
	handler := http.HandlerFunc(fn.ServeHTTP)

	req, _ := http.NewRequest("GET", "/dummy", http.NoBody)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)
	return recorder
}

func TestCreationSuccess(t *testing.T) {
	assert := assertlib.New(t)

	data := struct {
		ItemID int64 `json:"id"`
	}{42}

	recorder := httpResponseForResponse(CreationSuccess(data))
	assert.Equal(`{"success":true,"message":"created","data":{"id":42}}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusCreated, recorder.Code)
}

func TestUpdateSuccess(t *testing.T) {
	assert := assertlib.New(t)

	data := struct {
		Info string `json:"info"`
	}{"some info"}

	recorder := httpResponseForResponse(UpdateSuccess(data))
	assert.Equal(`{"success":true,"message":"updated","data":{"info":"some info"}}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestDeletionSuccess(t *testing.T) {
	assert := assertlib.New(t)

	data := struct {
		Info string `json:"info"`
	}{"some info"}

	recorder := httpResponseForResponse(DeletionSuccess(data))
	assert.Equal(`{"success":true,"message":"deleted","data":{"info":"some info"}}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestUnchangedSuccess(t *testing.T) {
	assert := assertlib.New(t)

	recorder := httpResponseForResponse(UnchangedSuccess(http.StatusResetContent))
	assert.Equal(`{"success":true,"message":"unchanged","data":{"changed":false}}`+"\n", recorder.Body.String())
	assert.Equal(http.StatusResetContent, recorder.Code)
}

func TestResponse_Render(t *testing.T) {
	response := &Response[*struct{}]{HTTPStatusCode: http.StatusOK, Message: "", Success: true}
	recorder := httpResponseForResponse(response)
	assertlib.Equal(t, `{"success":true,"message":"success"}`+"\n", recorder.Body.String())
	assertlib.Equal(t, http.StatusOK, recorder.Code)
}

func TestCheckInt64JsonHasStringTag(t *testing.T) {
	type structTypeInvalid struct {
		A  bool
		ID int64 `json:"id"`
	}
	type structTypeValid struct {
		A  bool
		ID int64 `json:"id,string"`
	}
	type structSliceTypeInvalid struct {
		B     bool
		Slice []structTypeInvalid
	}
	type structSliceTypeValid struct {
		B     bool
		Slice []structTypeValid
	}
	type structMapTypeInvalid struct {
		B   bool
		Map map[string]structTypeInvalid
	}
	type structMapTypeValid struct {
		B   bool
		Map map[string]structTypeValid
	}

	tests := []struct {
		name                string
		definitionStructure interface{}
		shouldPanic         bool
	}{
		{
			name: "int64/uint64 without JSON tag should pass",
			definitionStructure: &struct {
				A   bool
				ID  int64
				UID uint64
			}{},
			shouldPanic: false,
		},
		{
			name: "int64 with JSON tag, without string, should panic",
			definitionStructure: &struct {
				A  bool
				ID int64 `json:"id"`
			}{},
			shouldPanic: true,
		},
		{
			name: "uint64 with JSON tag, without string, should panic",
			definitionStructure: &struct {
				A  bool
				ID uint64 `json:"id"`
			}{},
			shouldPanic: true,
		},
		{
			name: "int64 in sub-struct with JSON tag, without string, should panic",
			definitionStructure: &struct {
				A      bool
				Inside struct {
					Inside2 struct {
						B  bool
						ID int64 `json:"id"`
					}
				}
			}{},
			shouldPanic: true,
		},
		{
			name: "int64/uint64 with JSON tag ignored, should pass",
			definitionStructure: &struct {
				A   bool
				ID  int64  `json:"-"`
				UID uint64 `json:"-"`
			}{},
			shouldPanic: false,
		},
		{
			name: "int64/uint64 with JSON tag, with string, should pass",
			definitionStructure: &struct {
				A   bool
				ID  int64  `json:"id,string"`
				ID2 uint64 `json:"id2,string"`
			}{},
			shouldPanic: false,
		},
		{
			name: "should panic with top-level slice containing invalid structures",
			definitionStructure: []structTypeInvalid{
				{
					A:  false,
					ID: 1,
				},
				{
					A:  false,
					ID: 2,
				},
			},
			shouldPanic: true,
		},
		{
			name: "should pass with top-level slice containing valid structures",
			definitionStructure: []structTypeValid{
				{
					A:  false,
					ID: 1,
				},
				{
					A:  false,
					ID: 2,
				},
			},
			shouldPanic: false,
		},
		{
			name: "should panic with sub-level slice containing invalid structures",
			definitionStructure: structSliceTypeInvalid{
				B: false,
				Slice: []structTypeInvalid{
					{
						A:  false,
						ID: 1,
					},
					{
						A:  false,
						ID: 2,
					},
				},
			},
			shouldPanic: true,
		},
		{
			name: "should pass with sub-level slice containing valid structures",
			definitionStructure: structSliceTypeValid{
				B: false,
				Slice: []structTypeValid{
					{
						A:  false,
						ID: 1,
					},
					{
						A:  false,
						ID: 2,
					},
				},
			},
			shouldPanic: false,
		},
		{
			name: "should panic with top-level map containing invalid structures",
			definitionStructure: map[string]interface{}{
				"a": structTypeInvalid{
					A:  false,
					ID: 1,
				},
				"b": structTypeInvalid{
					A:  false,
					ID: 2,
				},
			},
			shouldPanic: true,
		},
		{
			name: "should pass with top-level map containing valid structures",
			definitionStructure: map[string]interface{}{
				"a": structTypeValid{
					A:  false,
					ID: 1,
				},
				"b": structTypeValid{
					A:  false,
					ID: 2,
				},
			},
			shouldPanic: false,
		},
		{
			name: "should panic with sub-level map containing invalid structures",
			definitionStructure: structMapTypeInvalid{
				B: false,
				Map: map[string]structTypeInvalid{
					"a": {
						A:  false,
						ID: 1,
					},
					"b": {
						A:  false,
						ID: 2,
					},
				},
			},
			shouldPanic: true,
		},
		{
			name: "should pass with sub-level map containing valid structures",
			definitionStructure: structMapTypeValid{
				B: false,
				Map: map[string]structTypeValid{
					"a": {
						A:  false,
						ID: 1,
					},
					"b": {
						A:  false,
						ID: 2,
					},
				},
			},
			shouldPanic: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assertlib.Panics(t, func() { CheckInt64JsonHasStringTag(tt.definitionStructure) })
			} else {
				assertlib.NotPanics(t, func() { CheckInt64JsonHasStringTag(tt.definitionStructure) })
			}
		})
	}
}
