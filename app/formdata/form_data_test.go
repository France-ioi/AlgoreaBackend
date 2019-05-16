package formdata_test

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/mapstructure"

	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/payloads"
)

func TestFormData_ParseJSONRequestData(t *testing.T) {
	tests := []struct {
		name                string
		definitionStructure interface{}
		json                string
		wantErr             string
		wantFieldErrors     formdata.FieldErrors
	}{
		{
			"simple",
			&struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			}{},
			`{"id":123, "name":"John"}`,
			"",
			nil,
		},
		{
			"invalid JSON",
			&struct{}{},
			`{id:123, name:"John"}`,
			"invalid character 'i' looking for beginning of object key string",
			nil,
		},
		{
			"wrong value for a field",
			&struct {
				ID int64 `json:"id"`
			}{},
			`{"id":"123"}`,
			"invalid input data",
			formdata.FieldErrors{"id": {"expected type 'int64', got unconvertible type 'string'"}},
		},
		{
			"null value for a not-null field",
			&struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			}{},
			`{"id":null, "name":null}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":   {"should not be null (expected type: int64)"},
				"name": {"should not be null (expected type: string)"},
			},
		},
		{
			"unexpected field",
			&struct {
				ID int64 `json:"id"`
			}{},
			`{"my_id":"123"}`,
			"invalid input data",
			formdata.FieldErrors{"my_id": {"unexpected field"}},
		},
		{
			"field ignored by json",
			&struct {
				Name string `json:"-" gorm:"column:sName"`
			}{},
			`{"Name":"test"}`,
			"invalid input data",
			formdata.FieldErrors{"Name": {"unexpected field"}},
		},
		{
			"decoder error for a field",
			&struct {
				Time time.Time `json:"time"`
			}{},
			`{"time":"123"}`,
			"invalid input data",
			formdata.FieldErrors{
				"time": {"decoding error: parsing time \"123\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"123\" as \"2006\""},
			},
		},
		{
			"multiple errors",
			&struct {
				ID *int64 `json:"id" valid:"required"`
			}{},
			`{"my_id":1, "id":null}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":    {"non zero value required"},
				"my_id": {"unexpected field"},
			},
		},
		{
			"nested structure",
			&struct {
				Struct struct {
					Name        string `json:"name" valid:"required"`
					OtherStruct *struct {
						Name *string `json:"name" valid:"required"`
					} `json:"other_struct" valid:"required"`
					OtherStruct2 *struct {
						Name *string `json:"name" valid:"required"`
					} `valid:"required"`
				} `json:"struct" valid:"required"`
			}{},
			`{"id":null, "struct":{"other_struct":null, "OtherStruct2": null}}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":                  {"unexpected field"},
				"struct.other_struct": {"non zero value required"},
				"struct.OtherStruct2": {"non zero value required"},
			},
		},
		{
			"nested structure2",
			&struct {
				Struct struct {
					Name        *string `json:"name" valid:"required"`
					OtherStruct struct {
						Name *string `json:"name" valid:"required"`
					} `json:"other_struct" valid:"required"`
					OtherStruct2 struct {
						Name *string `json:"name" valid:"required"`
					} `valid:"required"`
				} `json:"struct" valid:"required"`
			}{},
			`{"id":null, "struct":{"name":null, "other_struct":{"name":null}, "OtherStruct2":{"name":null}}}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":                       {"unexpected field"},
				"struct.name":              {"non zero value required"},
				"struct.other_struct.name": {"non zero value required"},
				"struct.OtherStruct2.name": {"non zero value required"},
			},
		},
		{
			"ignores errors in fields that are not given",
			&struct {
				ID   *int64  `json:"id" valid:"required"`
				Name *string `json:"name" valid:"required"`
			}{},
			`{}`,
			"",
			nil,
		},
		{
			"rare errors (unsupported type)",
			&struct {
				Field chan bool `json:"field"`
			}{},
			`{"field":"value"}`,
			"invalid input data",
			formdata.FieldErrors{"": {"field: unsupported type: chan"}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			f := formdata.NewFormData(tt.definitionStructure)
			req, _ := http.NewRequest("POST", "/", strings.NewReader(tt.json))
			err := f.ParseJSONRequestData(req)
			if tt.wantErr != "" {
				assert.NotNil(t, err, "Should produce an error, but it did not")
				if err != nil {
					assert.Equal(t, tt.wantErr, err.Error())
				}
			} else {
				assert.Nil(t, err)
			}
			if tt.wantFieldErrors != nil {
				assert.IsType(t, formdata.FieldErrors{}, err)
				assert.Equal(t, tt.wantFieldErrors, err)
			}
		})
	}
}

func TestFormData_ParseMapData(t *testing.T) {
	tests := []struct {
		name                string
		definitionStructure interface{}
		sourceMap           map[string]interface{}
		wantErr             string
		wantFieldErrors     formdata.FieldErrors
	}{
		{
			"simple",
			&struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			}{},
			map[string]interface{}{"id": 123, "name": "John"},
			"",
			nil,
		},
		{
			"wrong value for a field",
			&struct {
				ID int64 `json:"id"`
			}{},
			map[string]interface{}{"id": "123"},
			"invalid input data",
			formdata.FieldErrors{"id": {"expected type 'int64', got unconvertible type 'string'"}},
		},
		{
			"null value for a not-null field",
			&struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			}{},
			map[string]interface{}{"id": nil, "name": nil},
			"invalid input data",
			formdata.FieldErrors{
				"id":   {"should not be null (expected type: int64)"},
				"name": {"should not be null (expected type: string)"},
			},
		},
		{
			"unexpected field",
			&struct {
				ID int64 `json:"id"`
			}{},
			map[string]interface{}{"my_id": "123"},
			"invalid input data",
			formdata.FieldErrors{"my_id": {"unexpected field"}},
		},
		{
			"field ignored by json",
			&struct {
				Name string `json:"-" gorm:"column:sName"`
			}{},
			map[string]interface{}{"Name": "test"},
			"invalid input data",
			formdata.FieldErrors{"Name": {"unexpected field"}},
		},
		{
			"decoder error for a field",
			&struct {
				Time time.Time `json:"time"`
			}{},
			map[string]interface{}{"time": "123"},
			"invalid input data",
			formdata.FieldErrors{
				"time": {"decoding error: parsing time \"123\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"123\" as \"2006\""},
			},
		},
		{
			"multiple errors",
			&struct {
				ID *int64 `json:"id" valid:"required"`
			}{},
			map[string]interface{}{"my_id": 1, "id": nil},
			"invalid input data",
			formdata.FieldErrors{
				"id":    {"non zero value required"},
				"my_id": {"unexpected field"},
			},
		},
		{
			"nested structure",
			&struct {
				Struct struct {
					Name        string `json:"name" valid:"required"`
					OtherStruct *struct {
						Name *string `json:"name" valid:"required"`
					} `json:"other_struct" valid:"required"`
					OtherStruct2 *struct {
						Name *string `json:"name" valid:"required"`
					} `valid:"required"`
				} `json:"struct" valid:"required"`
			}{},
			map[string]interface{}{"id": nil, "struct": map[string]interface{}{"other_struct": nil, "OtherStruct2": nil}},
			"invalid input data",
			formdata.FieldErrors{
				"id":                  {"unexpected field"},
				"struct.other_struct": {"non zero value required"},
				"struct.OtherStruct2": {"non zero value required"},
			},
		},
		{
			"nested structure2",
			&struct {
				Struct struct {
					Name        *string `json:"name" valid:"required"`
					OtherStruct struct {
						Name *string `json:"name" valid:"required"`
					} `json:"other_struct" valid:"required"`
					OtherStruct2 struct {
						Name *string `json:"name" valid:"required"`
					} `valid:"required"`
				} `json:"struct" valid:"required"`
			}{},
			map[string]interface{}{
				"id": nil,
				"struct": map[string]interface{}{
					"name": nil,
					"other_struct": map[string]interface{}{
						"name": nil,
					},
					"OtherStruct2": map[string]interface{}{
						"name": nil,
					},
				},
			},
			"invalid input data",
			formdata.FieldErrors{
				"id":                       {"unexpected field"},
				"struct.name":              {"non zero value required"},
				"struct.other_struct.name": {"non zero value required"},
				"struct.OtherStruct2.name": {"non zero value required"},
			},
		},
		{
			"ignores errors in fields that are not given",
			&struct {
				ID   *int64  `json:"id" valid:"required"`
				Name *string `json:"name" valid:"required"`
			}{},
			map[string]interface{}{},
			"",
			nil,
		},
		{
			"rare errors (unsupported type)",
			&struct {
				Field chan bool `json:"field"`
			}{},
			map[string]interface{}{"field": "value"},
			"invalid input data",
			formdata.FieldErrors{"": {"field: unsupported type: chan"}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			f := formdata.NewFormData(tt.definitionStructure)
			err := f.ParseMapData(tt.sourceMap)
			if tt.wantErr != "" {
				assert.NotNil(t, err, "Should produce an error, but it did not")
				if err != nil {
					assert.Equal(t, tt.wantErr, err.Error())
				}
			} else {
				assert.Nil(t, err)
			}
			if tt.wantFieldErrors != nil {
				assert.IsType(t, formdata.FieldErrors{}, err)
				assert.Equal(t, tt.wantFieldErrors, err)
			}
		})
	}
}

func TestFormData_ConstructMapForDB(t *testing.T) {
	tests := []struct {
		name                string
		definitionStructure interface{}
		json                string
		want                map[string]interface{}
	}{
		{
			"simple",
			&struct {
				ID           int64   `json:"id"`
				Name         string  `json:"name"`
				NullableName *string `json:"nullable_name"`
			}{},
			`{"id": 123, "name": "John", "nullable_name": "Paul"}`,
			map[string]interface{}{
				"ID": int64(123), "Name": "John", "NullableName": func() *string { s := "Paul"; return &s }(),
			},
		},
		{
			"sql and gorm tags",
			&struct {
				ID    int64  `json:"id" sql:"column:id"`
				Name  string `json:"name" gorm:"column:sName"`
				Skip  string `json:"skip1" gorm:"-"`
				Skip2 string `json:"skip2" sql:"-"`
			}{},
			`{"id":123, "name":"John", "skip1":"skip", "skip2":"skip"}`,
			map[string]interface{}{"id": int64(123), "sName": "John"},
		},
		{
			"skips missing fields",
			&struct {
				Name        string `json:"name" gorm:"column:sName"`
				Description string `json:"description" gorm:"column:sDescription"`
			}{},
			`{}`,
			map[string]interface{}{},
		},
		{
			"skips unexported fields",
			&struct {
				name string
			}{},
			`{"name":"test"}`,
			map[string]interface{}{},
		},
		{
			"skips fields ignored by json",
			&struct {
				Name string `json:"-" sql:"column:sName"`
			}{},
			`{}`,
			map[string]interface{}{},
		},
		{
			"keeps nulls",
			&struct {
				Description *string `json:"description" gorm:"column:sDescription"`
				Text        *string `json:"text" gorm:"column:sText"`
				Number2     *int64  `json:"number2" gorm:"column:iNumber2"`
				Number3     *int64  `json:"number3" gorm:"column:iNumber3"`
			}{},
			`{"description": null, "text": "", "number2": null, "number3": 0}`,
			map[string]interface{}{
				"sDescription": (*string)(nil), "sText": func() *string { s := ""; return &s }(),
				"iNumber2": (*int64)(nil), "iNumber3": func() *int64 { n := int64(0); return &n }(),
			},
		},
		{
			"nested structure",
			&struct {
				Struct struct {
					Name        string `json:"name" valid:"required" sql:"column:structs.sName"`
					OtherStruct struct {
						Name string `json:"name" valid:"required" sql:"column:structs.otherStructs.sName"`
					} `json:"other_struct" valid:"required"`
				} `json:"struct" valid:"required"`
			}{},
			`{"struct":{"name":"John Doe", "other_struct": {"name": "Still John Doe"}}}`,
			map[string]interface{}{"structs.sName": "John Doe", "structs.otherStructs.sName": "Still John Doe"},
		},
		{
			"sql vs gorm: the last gorm column name wins",
			&struct {
				Name string `json:"name" sql:"column:name_sql1;column:name_sql2" gorm:"column:name_gorm1;column:name_gorm2"`
			}{},
			`{"name":"John"}`,
			map[string]interface{}{"name_gorm2": "John"},
		},
		{
			"sql vs gorm: gorm '-' skips the field",
			&struct {
				Name string `json:"name" sql:"column:name_sql" gorm:"-"`
			}{},
			`{"name":"John"}`,
			map[string]interface{}{},
		},
		{
			"sql vs gorm: sql '-' skips the field",
			&struct {
				Name string `json:"name" sql:"-" gorm:"column:name_gorm"`
			}{},
			`{"name":"John"}`,
			map[string]interface{}{},
		},
		{
			"several fields with the same column name: the last one wins",
			&struct {
				Name1 string `json:"name1" sql:"column:name"`
				Name2 string `json:"name2" sql:"column:name"`
			}{},
			`{"name1":"John", "name2":"Paul"}`,
			map[string]interface{}{"name": "Paul"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			f := formdata.NewFormData(tt.definitionStructure)
			req, _ := http.NewRequest("POST", "/", strings.NewReader(tt.json))
			assert.Nil(t, f.ParseJSONRequestData(req))

			got := f.ConstructMapForDB()
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_toAnythingHookFunc(t *testing.T) {
	tests := []struct {
		name     string
		typeFrom reflect.Type
		typeTo   reflect.Type
		data     interface{}
		want     interface{}
		wantErr  error
	}{
		{
			name:     "string to anything (just convert to bytes)",
			typeFrom: reflect.TypeOf("string"),
			typeTo:   reflect.TypeOf(payloads.Anything{}),
			data:     "string",
			want:     []byte("string"),
		},
		{
			name:     "int to anything (serialize)",
			typeFrom: reflect.TypeOf(int(1)),
			typeTo:   reflect.TypeOf(payloads.Anything{}),
			data:     int(1),
			want:     []byte("1"),
		},
		{
			name:     "int to string (does nothing)",
			typeFrom: reflect.TypeOf(int(1)),
			typeTo:   reflect.TypeOf("1"),
			data:     int(1),
			want:     int(1),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			hook := formdata.ToAnythingHookFunc()
			converted, err := mapstructure.DecodeHookExec(hook, tt.typeFrom, tt.typeTo, tt.data)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, converted)
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}
