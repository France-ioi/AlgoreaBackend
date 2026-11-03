package payloads

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/v2/app/payloadstest"
)

func TestPayloads_ParseMap(t *testing.T) {
	tests := []struct {
		name      string
		raw       map[string]any
		want      any
		wantError error
	}{
		{
			name: "task token",
			raw:  payloadstest.TaskPayloadFromAlgoreaPlatform(),
			want: &TaskToken{
				Date:               "02-05-2019",
				UserID:             "556371821693219925",
				AttemptID:          "556371821693219925/100",
				ItemURL:            "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
				LocalItemID:        "901756573345831409",
				PlatformName:       "test_dmitry",
				RandomSeed:         "556371821693219925",
				HintsGivenCount:    new("0"),
				HintsAllowed:       new(false),
				HintPossible:       new(true),
				AccessSolutions:    new(true),
				ReadAnswers:        new(true),
				Login:              new("test"),
				SubmissionPossible: new(true),
				SupportedLangProg:  new("*"),
				IsAdmin:            new(false),
				Converted: TaskTokenConverted{
					UserID:        556371821693219925,
					LocalItemID:   901756573345831409,
					AttemptID:     100,
					ParticipantID: 556371821693219925,
				},
			},
		},
		{
			name: "answer token",
			raw:  payloadstest.AnswerPayloadFromAlgoreaPlatform(),
			want: &AnswerToken{
				Date:            "02-05-2019",
				UserID:          "556371821693219925",
				AttemptID:       "100/0",
				ItemURL:         "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
				LocalItemID:     "901756573345831409",
				PlatformName:    "test_dmitry",
				RandomSeed:      "556371821693219925",
				HintsGivenCount: "0",
				Answer: "{\"idSubmission\":\"899146309203855074\",\"langProg\":\"python\"," +
					"\"sourceCode\":\"print(min(int(input()), int(input()), int(input())))\"}",
				UserAnswerID: "251510027138726857",
				Converted: AnswerTokenConverted{
					UserID:        556371821693219925,
					UserAnswerID:  251510027138726857,
					AttemptID:     0,
					ParticipantID: 100,
					LocalItemID:   901756573345831409,
				},
			},
		},
		{
			name: "hint token",
			raw:  payloadstest.HintPayloadFromTaskPlatform(),
			want: &HintToken{
				Date:      "02-05-2019",
				UserID:    "556371821693219925",
				ItemURL:   "http://taskplatform.mblockelet.info/task.html?taskId=212873689338185696",
				AskedHint: formdata.AnythingFromString("1"),
				Converted: HintTokenConverted{
					UserID: 556371821693219925,
				},
			},
		},
		{
			name: "hint token with a complex askedHint",
			raw: map[string]any{
				"itemUrl":   "http://taskplatform.mblockelet.info/task.html?taskId=212873689338185696",
				"idUser":    "556371821693219925",
				"askedHint": formdata.AnythingFromString(`{"rotorIndex":0,"cellRank":1}`),
				"date":      "02-05-2019",
			},
			want: &HintToken{
				Date:      "02-05-2019",
				UserID:    "556371821693219925",
				ItemURL:   "http://taskplatform.mblockelet.info/task.html?taskId=212873689338185696",
				AskedHint: formdata.AnythingFromString(`{"rotorIndex":0,"cellRank":1}`),
				Converted: HintTokenConverted{
					UserID: 556371821693219925,
				},
			},
		},
		{
			name: "invalid task token",
			raw:  map[string]any{"date": "abcdef"},
			want: &TaskToken{Date: "abcdef"},
			wantError: errors.New(
				"invalid TaskToken: invalid input data"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := reflect.New(reflect.TypeOf(tt.want).Elem()).Interface()
			err := ParseMap(tt.raw, target)
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.want, target)
		})
	}
}

func TestConvertIntoMap(t *testing.T) {
	type nestedStruct struct {
		Field string `json:"field"`
	}
	type testStruct struct {
		notExported   string // unexported field, must be skipped regardless of any tag
		Normal        string `json:"normal"`
		WithoutTag    string
		Skipped       string        `json:"-"`
		AnotherNormal string        `json:"another_normal"`
		Struct        *nestedStruct `json:"struct"`
		OmitEmpty     *string       `json:"omit_empty,omitempty"`
	}
	got := ConvertIntoMap(&testStruct{
		notExported:   "notExported value",
		Normal:        "Normal value",
		WithoutTag:    "WithoutTag value",
		Skipped:       "Skipped value",
		AnotherNormal: "AnotherNormal value",
		Struct: &nestedStruct{
			Field: "Field value",
		},
	})
	assert.Equal(t, map[string]any{
		"normal":         "Normal value",
		"another_normal": "AnotherNormal value",
		"struct": map[string]any{
			"field": "Field value",
		},
	}, got)
}

// TestConvertIntoMap_SkipsUnexportedFieldsWithJSONTag specifically exercises the
// `!fieldValue.CanInterface()` branch in ConvertIntoMap: a field that is unexported
// but DOES carry a json tag (so the early `name == "-"` skip does not fire and we
// reach the CanInterface() guard). The struct is built with reflect.StructOf at
// runtime to avoid the `go vet` "struct field has json tag but is not exported"
// diagnostic that fires on a statically-declared equivalent.
func TestConvertIntoMap_SkipsUnexportedFieldsWithJSONTag(t *testing.T) {
	pkgPath := reflect.TypeFor[Binder]().PkgPath()
	typ := reflect.StructOf([]reflect.StructField{
		{
			Name:    "notExported",
			PkgPath: pkgPath,
			Type:    reflect.TypeFor[string](),
			Tag:     `json:"not_exported"`,
		},
	})
	got := ConvertIntoMap(reflect.New(typ).Interface())
	assert.Equal(t, map[string]any{}, got)
}
