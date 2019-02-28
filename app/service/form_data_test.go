package service

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestFormData_ParseJSONRequestData(t *testing.T) {
	tests := []struct {
		name                string
		definitionStructure interface{}
		json                string
		wantErr             string
		wantFieldErrors     FieldErrors
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
			FieldErrors{"id": {"expected type 'int64', got unconvertible type 'string'"}},
		},
		{
			"unexpected field",
			&struct {
				ID int64 `json:"id"`
			}{},
			`{"my_id":"123"}`,
			"invalid input data",
			FieldErrors{"my_id": {"unexpected field"}},
		},
		{
			"field ignored by json",
			&struct {
				Name string `json:"-" gorm:"column:sName"`
			}{},
			`{"Name":"test"}`,
			"invalid input data",
			FieldErrors{"Name": {"unexpected field"}},
		},
		{
			"decoder error for a field",
			&struct {
				Time time.Time `json:"time"`
			}{},
			`{"time":"123"}`,
			"invalid input data",
			FieldErrors{"time": {"decoding error: parsing time \"123\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"123\" as \"2006\""}},
		},
		{
			"multiple errors",
			&struct {
				ID int64 `json:"id" valid:"required"`
			}{},
			`{"my_id":1, "id":0}`,
			"invalid input data",
			FieldErrors{
				"id":    {"non zero value required"},
				"my_id": {"unexpected field"},
			},
		},
		{
			"nested structure",
			&struct {
				Struct struct {
					Name        string `json:"name" valid:"required"`
					OtherStruct struct {
						Name string `json:"name" valid:"required"`
					} `json:"other_struct" valid:"required"`
					OtherStruct2 struct {
						Name string `json:"name" valid:"required"`
					} `valid:"required"`
				} `json:"struct" valid:"required"`
			}{},
			`{"id":0}`,
			"invalid input data",
			FieldErrors{
				"id":                       {"unexpected field"},
				"struct":                   {"non zero value required"},
				"struct.name":              {"non zero value required"},
				"struct.other_struct":      {"non zero value required"},
				"struct.other_struct.name": {"non zero value required"},
				"struct.OtherStruct2":      {"non zero value required"},
				"struct.OtherStruct2.name": {"non zero value required"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormData(tt.definitionStructure)
			req, _ := http.NewRequest("POST", "/", strings.NewReader(tt.json))
			err := f.ParseJSONRequestData(req)
			if tt.wantErr != "" {
				assert.Equal(t, tt.wantErr, err.Error())
			} else {
				assert.Nil(t, err)
			}
			if tt.wantFieldErrors != nil {
				assert.IsType(t, FieldErrors{}, err)
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
				ID   int64  `json:"id"`
				Name string `json:"name"`
			}{},
			`{"id":123, "name":"John"}`,
			map[string]interface{}{"ID": int64(123), "Name": "John"},
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
				Name        string  `json:"name" gorm:"column:sName"`
				Description *string `json:"description" gorm:"column:sDescription"`
				Number      int64   `json:"number" gorm:"column:iNumber"`
				Number2     *int64  `json:"number2" gorm:"column:iNumber2"`
			}{},
			`{"name": null, "description": null, "number": null, "number2": null}`,
			map[string]interface{}{"sName": "", "sDescription": (*string)(nil), "iNumber": int64(0), "iNumber2": (*int64)(nil)},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormData(tt.definitionStructure)
			req, _ := http.NewRequest("POST", "/", strings.NewReader(tt.json))
			assert.Nil(t, f.ParseJSONRequestData(req))

			got := f.ConstructMapForDB()
			assert.Equal(t, tt.want, got)
		})
	}
}
