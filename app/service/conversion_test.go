package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertSliceOfMapsFromDBToJSON(t *testing.T) {
	tests := []struct {
		name  string
		dbMap []map[string]interface{}
		want  []map[string]interface{}
	}{
		{
			"nested structures",
			[]map[string]interface{}{{
				"User__ID":            int64(1),
				"Item__String__Title": "Chapter 1",
				"Item__String__ID":    "2",
			}},
			[]map[string]interface{}{
				{
					"User": map[string]interface{}{"ID": "1"},
					"Item": map[string]interface{}{"String": map[string]interface{}{"Title": "Chapter 1", "ID": "2"}},
				},
			},
		},
		{
			"keeps nil fields",
			[]map[string]interface{}{{"TheGreatestUser": nil, "otherField": 1}},
			[]map[string]interface{}{{"TheGreatestUser": nil, "otherField": 1}},
		},
		{
			"replaces empty sub-maps with nils",
			[]map[string]interface{}{{"the_greatest_user": nil, "empty_sub_map__field1": nil, "empty_sub_map__field2": nil}},
			[]map[string]interface{}{{"the_greatest_user": nil, "empty_sub_map": nil}},
		},
		{
			"converts int64 into string",
			[]map[string]interface{}{{
				"int64":             int64(123),
				"int32":             int32(1234),
				"nbCorrectionsRead": int64(12345),
				"iGrade":            int64(-1),
			}}, // gorm returns numbers as int64
			[]map[string]interface{}{{
				"int64":             "123",
				"int32":             int32(1234),
				"nbCorrectionsRead": "12345",
				"iGrade":            "-1",
			}},
		},
		{
			"handles datetime",
			[]map[string]interface{}{{
				"my_date":   "2019-05-30 11:00:00",
				"null_date": nil,
			}},
			[]map[string]interface{}{{
				"my_date":   "2019-05-30T11:00:00Z",
				"null_date": nil,
			}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertSliceOfMapsFromDBToJSON(tt.dbMap)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertSliceOfMapsFromDBToJSON_PanicsWhenDatetimeIsInvalid(t *testing.T) {
	assert.Panics(t, func() {
		ConvertSliceOfMapsFromDBToJSON([]map[string]interface{}{{"some_date": "1234:13:05 24:60:60"}})
	})
}
