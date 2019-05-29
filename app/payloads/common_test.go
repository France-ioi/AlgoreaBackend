package payloads

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/formdata"
	"github.com/France-ioi/AlgoreaBackend/app/payloadstest"
)

func TestPayloads_ParseMap(t *testing.T) {
	var tests = []struct {
		name      string
		raw       map[string]interface{}
		want      interface{}
		wantError error
	}{
		{
			name: "task token",
			raw:  payloadstest.TaskPayloadFromAlgoreaPlatform,
			want: &TaskToken{
				Date:               "02-05-2019",
				UserID:             "556371821693219925",
				AttemptID:          "100",
				ItemURL:            "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
				LocalItemID:        "901756573345831409",
				PlatformName:       "test_dmitry",
				RandomSeed:         "556371821693219925",
				HintsGivenCount:    ptrString("0"),
				HintsAllowed:       ptrBool(false),
				HintPossible:       ptrBool(true),
				AccessSolutions:    ptrBool(true),
				ReadAnswers:        ptrBool(true),
				Login:              ptrString("test"),
				SubmissionPossible: ptrBool(true),
				SupportedLangProg:  ptrString("*"),
				IsAdmin:            ptrBool(false),
				Converted: TaskTokenConverted{
					UserID:      556371821693219925,
					LocalItemID: 901756573345831409,
					AttemptID:   100,
				},
			},
		},
		{
			name: "answer token",
			raw:  payloadstest.AnswerPayloadFromAlgoreaPlatform,
			want: &AnswerToken{
				Date:            "02-05-2019",
				UserID:          "556371821693219925",
				AttemptID:       "100",
				ItemURL:         "http://taskplatform.mblockelet.info/task.html?taskId=403449543672183936",
				LocalItemID:     "901756573345831409",
				PlatformName:    "test_dmitry",
				RandomSeed:      "556371821693219925",
				HintsGivenCount: "0",
				Answer: "{\"idSubmission\":\"899146309203855074\",\"langProg\":\"python\"," +
					"\"sourceCode\":\"print(min(int(input()), int(input()), int(input())))\"}",
				UserAnswerID: "251510027138726857",
			},
		},
		{
			name: "hint token",
			raw:  payloadstest.HintPayloadFromTaskPlatform,
			want: &HintToken{
				Date:      "02-05-2019",
				UserID:    "556371821693219925",
				ItemURL:   "http://taskplatform.mblockelet.info/task.html?taskId=212873689338185696",
				AskedHint: *formdata.AnythingFromString("1"),
				Converted: HintTokenConverted{
					UserID: 556371821693219925,
				},
			},
		},
		{
			name: "hint token with a complex askedHint",
			raw: map[string]interface{}{
				"itemUrl":   "http://taskplatform.mblockelet.info/task.html?taskId=212873689338185696",
				"idUser":    "556371821693219925",
				"askedHint": *formdata.AnythingFromString(`{"rotorIndex":0,"cellRank":1}`),
				"date":      "02-05-2019",
			},
			want: &HintToken{
				Date:      "02-05-2019",
				UserID:    "556371821693219925",
				ItemURL:   "http://taskplatform.mblockelet.info/task.html?taskId=212873689338185696",
				AskedHint: *formdata.AnythingFromString(`{"rotorIndex":0,"cellRank":1}`),
				Converted: HintTokenConverted{
					UserID: 556371821693219925,
				},
			},
		},
		{
			name: "invalid task token",
			raw:  map[string]interface{}{"date": "abcdef"},
			want: &TaskToken{Date: "abcdef"},
			wantError: errors.New(
				"invalid TaskToken: invalid input data"),
		},
	}
	for _, tt := range tests {
		tt := tt
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
		notExported   string `json:"not_exported"` // nolint:govet
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
	assert.Equal(t, map[string]interface{}{
		"normal":         "Normal value",
		"another_normal": "AnotherNormal value",
		"struct": map[string]interface{}{
			"field": "Field value",
		},
	}, got)
}

func ptrString(s string) *string { return &s }
