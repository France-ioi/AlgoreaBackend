package service

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/render"

	"github.com/France-ioi/AlgoreaBackend/v2/app/appenv"
)

// Response is used for generating non-data responses, i.e. on error or on POST/PUT/PATCH/DELETE request.
type Response struct {
	HTTPStatusCode int         `json:"-"`
	Success        bool        `json:"success"`
	Message        string      `json:"message"`
	Data           interface{} `json:"data,omitempty"`
}

// Render generates the HTTP response from Response.
func (resp *Response) Render(_ http.ResponseWriter, r *http.Request) error {
	if resp.Success && resp.Message == "" {
		resp.Message = "success"
	}
	render.Status(r, resp.HTTPStatusCode)
	return nil
}

// CreationSuccess generated a success response for a POST creation.
func CreationSuccess(data interface{}) render.Renderer {
	return &Response{
		HTTPStatusCode: http.StatusCreated,
		Success:        true,
		Message:        "created",
		Data:           data,
	}
}

// UpdateSuccess generated a success response for a PUT updating.
func UpdateSuccess(data interface{}) render.Renderer {
	return &Response{
		HTTPStatusCode: http.StatusOK,
		Success:        true,
		Message:        "updated",
		Data:           data,
	}
}

// DeletionSuccess generated a success response for a DELETE deletion.
func DeletionSuccess(data interface{}) render.Renderer {
	return &Response{
		HTTPStatusCode: http.StatusOK,
		Success:        true,
		Message:        "deleted",
		Data:           data,
	}
}

// UnchangedSuccess generated a success response for a POST/PUT/DELETE action if no data have been modified.
func UnchangedSuccess(httpStatus int) render.Renderer {
	return &Response{
		HTTPStatusCode: httpStatus,
		Success:        true,
		Message:        "unchanged",
		Data: map[string]bool{
			"changed": false,
		},
	}
}

// AppResponder serializes the output of services.
// Whenever in test env, it also checks that all int64/uint64 are serialized as strings,
// otherwise, it breaks javascript that tries to convert them as floats.
func AppResponder(w http.ResponseWriter, r *http.Request, v interface{}) {
	if appenv.IsEnvTest() {
		CheckInt64JsonHasStringTag(v)
	}

	render.DefaultResponder(w, r, v)
}

// CheckInt64JsonHasStringTag checks that all fields of obj, recursively,
// which are int64/uint64, and are exported as JSON, also have the "string" tag.
// Otherwise, it breaks javascript that tries to convert them as floats.
// This function is only intended to be used in test env, and panics if the checks fail.
func CheckInt64JsonHasStringTag(obj interface{}) {
	checkInt64JsonHasStringTag(reflect.ValueOf(obj))
}

func checkInt64JsonHasStringTag(val reflect.Value) {
	val = getElem(val)

	switch {
	case val.Kind() == reflect.Slice:
		for j := 0; j < val.Len(); j++ {
			checkInt64JsonHasStringTag(val.Index(j))
		}
	case val.Kind() == reflect.Map:
		for _, k := range val.MapKeys() {
			checkInt64JsonHasStringTag(val.MapIndex(k))
		}
	default:
		checkInt64JsonHasStringTagInFields(val)
	}
}

func checkInt64JsonHasStringTagInFields(val reflect.Value) {
	if val.Kind() != reflect.Struct {
		return
	}

	rt := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fieldType := rt.Field(i)

		panicIfInt64JsonHasStringTag(&fieldType)

		valField := getElem(val.Field(i))
		if valField.Kind() == reflect.Struct || valField.Kind() == reflect.Map || valField.Kind() == reflect.Slice {
			checkInt64JsonHasStringTag(valField)
		}
	}
}

func panicIfInt64JsonHasStringTag(fieldType *reflect.StructField) {
	name := fieldType.Name
	typ := fieldType.Type.Name()
	jsonTag := fieldType.Tag.Get("json")

	if (typ == "int64" || typ == "uint64") && jsonTag != "" && jsonTag != "-" && !strings.Contains(jsonTag, ",string") {
		panic(name + " is of type int64 but json metadata doesn't contain \",string\". This might break javascript.")
	}
}

func getElem(val reflect.Value) reflect.Value {
	if val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		return val.Elem()
	}

	return val
}
