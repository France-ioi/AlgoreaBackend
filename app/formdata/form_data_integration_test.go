package formdata_test

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/France-ioi/mapstructure"
	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
)

type NamedStruct struct {
	ID   *int64  `json:"id" validate:"min=1"`
	Name *string `json:"name" validate:"min=1"` // length >= 1
}

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
				ID int32 `json:"id"`
			}{},
			`{"id":"123"}`,
			"invalid input data",
			formdata.FieldErrors{"id": {"expected type 'int32', got unconvertible type 'string'"}},
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
				ID *int64 `json:"id" validate:"set"`
			}{},
			`{"my_id":1, "id":null}`,
			"invalid input data",
			formdata.FieldErrors{
				"my_id": {"unexpected field"},
			},
		},
		{
			"nested structure",
			&struct {
				Struct struct {
					Name        string `json:"name" validate:"set"`
					OtherStruct *struct {
						Name *string `json:"name" validate:"set"`
					} `json:"other_struct" validate:"set"`
					OtherStruct2 *struct {
						Name *string `json:"name" validate:"set"`
					} `validate:"set"`
				} `json:"struct" validate:"set"`
			}{},
			`{"id":null, "struct":{"other_struct":null, "OtherStruct2": null}}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":          {"unexpected field"},
				"struct.name": {"missing field"},
			},
		},
		{
			"nested structure with squash",
			&struct {
				Struct struct {
					Name        string `json:"name" validate:"set"`
					OtherStruct struct {
						Name *string `json:"name1" validate:"set"`
					} `json:"other_struct,squash" validate:"set"` //nolint:staticcheck SA5008: unknown JSON option "squash"
					OtherStruct2 struct {
						Name *string `json:"name2" validate:"set"`
					} `json:"other_struct2,squash" validate:"set"` //nolint:staticcheck SA5008: unknown JSON option "squash"
				} `json:"struct" validate:"set"`
			}{},
			`{"id":null, "struct":{"name1": "my name"}}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":           {"unexpected field"},
				"struct.name":  {"missing field"},
				"struct.name2": {"missing field"},
			},
		},
		{
			"nested structure with squash and invalid fields",
			&struct {
				Struct struct {
					Name string `json:"name" validate:"min=1"`
				} `json:"struct,squash"` //nolint:staticcheck SA5008: unknown JSON option "squash"
			}{},
			`{"name": ""}}`,
			"invalid input data",
			formdata.FieldErrors{
				"name": {"name must be at least 1 character in length"},
			},
		},
		{
			"nested structure2",
			&struct {
				Struct struct {
					Name        *string `json:"name" validate:"set"`
					OtherStruct struct {
						Name *string `json:"name" validate:"set"`
					} `json:"other_struct" validate:"set"`
					OtherStruct2 struct {
						Name *string `json:"name" validate:"set"`
					} `json:"other_struct2" validate:"set"`
				} `json:"struct" validate:"set"`
			}{},
			`{"id":null, "struct":{"name":null, "other_struct":{}, "other_struct2":{}}}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":                        {"unexpected field"},
				"struct.other_struct.name":  {"missing field"},
				"struct.other_struct2.name": {"missing field"},
			},
		},
		{
			"nested structure with nils",
			&struct {
				Struct struct {
					Name        *string `json:"name" validate:"set"`
					OtherStruct *struct {
						Name *string `json:"name" validate:"set"`
					} `json:"other_struct" validate:"set"`
					OtherStruct2 *struct {
						Name *string `json:"name" validate:"set"`
					} `json:"other_struct2" validate:"set"`
				} `json:"struct" validate:"set"`
			}{},
			`{"struct":{"name":null, "other_struct":null, "other_struct2":null}}`,
			"",
			nil,
		},
		{
			"ignores errors in fields that are not given",
			&struct {
				ID   *int64  `json:"id" validate:"min=1"`
				Name *string `json:"name" validate:"min=1"` // length >= 1
			}{},
			`{}`,
			"",
			nil,
		},
		{
			"ignores errors related to scalar fields that are not given, but should be set",
			&struct {
				ID   int64  `json:"id" validate:"set,min=1"`
				Name string `json:"name" validate:"set,min=1"` // length >= 1
			}{},
			`{}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":   {"missing field"},
				"name": {"missing field"},
			},
		},
		{
			"validates fields with empty values",
			&struct {
				ID   int64  `json:"id" validate:"set,min=1"`
				Name string `json:"name" validate:"set,min=1"` // length >= 1
			}{},
			`{"id":0, "name":""}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":   {"id must be 1 or greater"},
				"name": {"name must be at least 1 character in length"},
			},
		},
		{
			"validates pointer fields with empty values",
			&struct {
				ID   *int64  `json:"id" validate:"set,min=1"`
				Name *string `json:"name" validate:"set,min=1"` // length >= 1
			}{},
			`{"id":0, "name":""}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":   {"id must be 1 or greater"},
				"name": {"name must be at least 1 character in length"},
			},
		},
		{
			"validates pointer fields with null values",
			&struct {
				ID   *int64  `json:"id" validate:"set,min=1"`
				Name *string `json:"name" validate:"set,min=1"` // length >= 1
			}{},
			`{"id":null, "name":null}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":   {"id must be 1 or greater"},
				"name": {"name must be at least 1 character in length"},
			},
		},
		{
			"null validator ignores fields that are not given",
			&struct {
				ID *int64 `json:"id" validate:"null"`
			}{},
			`{}`,
			"",
			nil,
		},
		{
			"null validator requires fields to be null",
			&struct {
				ID *int64 `json:"id" validate:"null"`
			}{},
			`{"id":1234}`,
			"invalid input data",
			formdata.FieldErrors{
				"id": {"should be null"},
			},
		},
		{
			"null accepts null values",
			&struct {
				ID *int64 `json:"id" validate:"null"`
			}{},
			`{"id":null}`,
			"",
			nil,
		},
		{
			"named structure",
			&NamedStruct{},
			`{"id":0,"name":""}`,
			"invalid input data",
			formdata.FieldErrors{
				"id":   {"id must be 1 or greater"},
				"name": {"name must be at least 1 character in length"},
			},
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
		{
			"custom error messages",
			&struct {
				Date     string `json:"date" validate:"dmy-date"`
				Duration string `json:"duration" validate:"duration"`
			}{},
			`{"date":"value","duration":"another value"}`,
			"invalid input data",
			formdata.FieldErrors{"date": {"should be dd-mm-yyyy"}, "duration": {"invalid duration"}},
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
				ID int32 `json:"id"`
			}{},
			map[string]interface{}{"id": "123"},
			"invalid input data",
			formdata.FieldErrors{"id": {"expected type 'int32', got unconvertible type 'string'"}},
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
				ID *int64 `json:"id" validate:"set"`
			}{},
			map[string]interface{}{"my_id": 1, "id": nil},
			"invalid input data",
			formdata.FieldErrors{
				"my_id": {"unexpected field"},
			},
		},
		{
			"nested structure",
			&struct {
				Struct struct {
					Name        string `json:"name" validate:"set"`
					OtherStruct *struct {
						Name *string `json:"name" validate:"set"`
					} `json:"other_struct" validate:"set"`
					OtherStruct2 *struct {
						Name *string `json:"name" validate:"set"`
					} `json:"other_struct2" validate:"set"`
				} `json:"struct" validate:"set"`
			}{},
			map[string]interface{}{"id": nil, "struct": map[string]interface{}{"other_struct": nil, "other_struct2": nil}},
			"invalid input data",
			formdata.FieldErrors{
				"id":          {"unexpected field"},
				"struct.name": {"missing field"},
			},
		},
		{
			"nested structure2",
			&struct {
				Struct struct {
					Name        *string `json:"name" validate:"set"`
					OtherStruct struct {
						Name *string `json:"name" validate:"set"`
					} `json:"other_struct" validate:"set"`
					OtherStruct2 struct {
						Name *string `json:"name" validate:"set"`
					} `json:"other_struct2" validate:"set"`
				} `json:"struct" validate:"set"`
			}{},
			map[string]interface{}{
				"id": nil,
				"struct": map[string]interface{}{
					"name": nil,
					"other_struct": map[string]interface{}{
						"name": nil,
					},
					"other_struct2": map[string]interface{}{
						"name": nil,
					},
				},
			},
			"invalid input data",
			formdata.FieldErrors{
				"id": {"unexpected field"},
			},
		},
		{
			"ignores errors in fields that are not given",
			&struct {
				ID   *int64  `json:"id" validate:"min=1"`
				Name *string `json:"name" validate:"min=1"` // length >= 1
			}{},
			map[string]interface{}{},
			"",
			nil,
		},
		{
			"runs validators for pointers",
			&struct {
				ID   *int64  `json:"id" validate:"min=1"`
				Name *string `json:"name" validate:"min=1"` // length >= 1
			}{},
			map[string]interface{}{"id": 0, "name": ""},
			"invalid input data",
			formdata.FieldErrors{
				"id":   []string{"id must be 1 or greater"},
				"name": []string{"name must be at least 1 character in length"},
			},
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
				ID           int64   `json:"id2"`
				Name         string  `json:"name"`
				NullableName *string `json:"nullable_name"`
			}{},
			`{"id2": 123, "name": "John", "nullable_name": "Paul"}`,
			map[string]interface{}{
				"id": int64(123), "name": "John", "nullable_name": func() *string { s := "Paul"; return &s }(),
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
			"skips fields without json tag",
			&struct {
				Name string
			}{},
			`{"Name":"Test"}`,
			map[string]interface{}{},
		},
		{
			"skips unexported attributes",
			&struct {
				name string `json:"name" gorm:"column:name"` //nolint:govet
			}{},
			`{"name":"Test"}`,
			map[string]interface{}{},
		},
		{
			"skips fields ignored by json",
			&struct {
				Name        string `json:"-" sql:"column:sName"`
				Description string `sql:"column:sDescription"`
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
					Name        string `json:"name" validate:"set" sql:"column:structs.sName"`
					OtherStruct struct {
						Name string `json:"name" validate:"set" sql:"column:structs.otherStructs.sName"`
					} `json:"other_struct" validate:"set"`
				} `json:"struct" validate:"set"`
			}{},
			`{"struct":{"name":"John Doe", "other_struct": {"name": "Still John Doe"}}}`,
			map[string]interface{}{"structs.sName": "John Doe", "structs.otherStructs.sName": "Still John Doe"},
		},
		{
			"timestamp",
			&struct {
				Time time.Time `json:"time"`
			}{},
			`{"time": "2019-05-30T11:00:00Z"}`,
			map[string]interface{}{
				"time": time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC),
			},
		},
		{
			"structure with squash",
			&struct {
				Struct struct {
					Name        string `json:"name" validate:"set" sql:"column:structs.sName"`
					OtherStruct struct {
						Name string `json:"name" validate:"set" sql:"column:structs.otherStructs.sName"`
					} `json:"other_struct" validate:"set"`
				} `json:"struct,squash"` //nolint:staticcheck SA5008: unknown JSON option "squash"
			}{},
			`{"name":"John Doe", "other_struct": {"name": "Still John Doe"}}`,
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

func TestFormData_ConstructPartialMapForDB(t *testing.T) {
	tests := []struct {
		name                string
		definitionStructure interface{}
		json                string
		want                map[string]interface{}
	}{
		{
			"structure",
			&struct {
				Struct struct {
					Name        string `json:"name" validate:"set" sql:"column:structs.sName"`
					OtherStruct struct {
						Name string `json:"name" validate:"set" sql:"column:structs.otherStructs.sName"`
					} `json:"other_struct" validate:"set"`
				} `json:"struct" validate:"set"`
			}{},
			`{"struct":{"name":"John Doe", "other_struct": {"name": "Still John Doe"}}}`,
			map[string]interface{}{"structs.sName": "John Doe", "structs.otherStructs.sName": "Still John Doe"},
		},
		{
			"pointer to structure",
			&struct {
				Struct *struct {
					Name        string `json:"name" validate:"set" sql:"column:structs.sName"`
					OtherStruct struct {
						Name string `json:"name" validate:"set" sql:"column:structs.otherStructs.sName"`
					} `json:"other_struct" validate:"set"`
				} `json:"struct" validate:"set"`
			}{},
			`{"struct":{"name":"John Doe", "other_struct": {"name": "Still John Doe"}}}`,
			map[string]interface{}{"structs.sName": "John Doe", "structs.otherStructs.sName": "Still John Doe"},
		},
		{
			"structure with squash",
			&struct {
				Struct struct {
					Name        string `json:"name" validate:"set" sql:"column:structs.sName"`
					OtherStruct struct {
						Name string `json:"name" validate:"set" sql:"column:structs.otherStructs.sName"`
					} `json:"other_struct" validate:"set"`
				} `json:"struct,squash"` //nolint:staticcheck SA5008: unknown JSON option "squash"
			}{},
			`{"name":"John Doe", "other_struct": {"name": "Still John Doe"}}`,
			map[string]interface{}{"structs.sName": "John Doe", "structs.otherStructs.sName": "Still John Doe"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			f := formdata.NewFormData(tt.definitionStructure)
			req, _ := http.NewRequest("POST", "/", strings.NewReader(tt.json))
			assert.Nil(t, f.ParseJSONRequestData(req))

			got := f.ConstructPartialMapForDB("Struct")
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
			name:     "string to anything (serialize)",
			typeFrom: reflect.TypeOf("string"),
			typeTo:   reflect.TypeOf(formdata.Anything{}),
			data:     "string",
			want:     *formdata.AnythingFromBytes([]byte(`"string"`)),
		},
		{
			name:     "[]byte to anything (just copy)",
			typeFrom: reflect.TypeOf([]byte("null")),
			typeTo:   reflect.TypeOf(formdata.Anything{}),
			data:     []byte("null"),
			want:     *formdata.AnythingFromBytes([]byte(`null`)),
		},
		{
			name:     "int to anything (serialize)",
			typeFrom: reflect.TypeOf(int(1)),
			typeTo:   reflect.TypeOf(formdata.Anything{}),
			data:     int(1),
			want:     *formdata.AnythingFromBytes([]byte("1")),
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

func Test_stringToInt64HookFunc(t *testing.T) {
	tests := []struct {
		name     string
		typeFrom reflect.Type
		typeTo   reflect.Type
		data     interface{}
		want     interface{}
		wantErr  error
	}{
		{
			name:     "string to int64 (parse)",
			typeFrom: reflect.TypeOf("string"),
			typeTo:   reflect.TypeOf(int64(0)),
			data:     "1234",
			want:     int64(1234),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			hook := formdata.StringToInt64HookFunc()
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
