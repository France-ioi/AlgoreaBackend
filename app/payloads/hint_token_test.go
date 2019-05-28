package payloads

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/app/formdata"
)

func TestHintToken_UnmarshalJSON_InvalidJSON(t *testing.T) {
	tt := &HintToken{}
	err := tt.UnmarshalJSON([]byte("[]"))
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, "json: cannot unmarshal array into Go value of type map[string]formdata.Anything", err.Error())
	}
}

func TestHintToken_UnmarshalJSON_WrongUserID(t *testing.T) {
	tt := &HintToken{}
	err := tt.UnmarshalJSON([]byte(`{"idUser":"abc"}`))
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, "wrong idUser", err.Error())
	}
}

func TestHintToken_UnmarshalJSON(t *testing.T) {
	tt := &HintToken{}
	err := tt.UnmarshalJSON([]byte(`{"idUser":"10", "askedHint":   false}`))
	assert.NoError(t, err)
	assert.Equal(t, &HintToken{
		UserID:    ptrString("10"),
		AskedHint: *formdata.AnythingFromString("false"),
		Converted: HintTokenConverted{
			UserID: ptrInt64(10),
		},
	}, tt)
}

func TestHintToken_MarshalJSON(t *testing.T) {
	tt := &HintToken{
		UserID:    ptrString("10"),
		AttemptID: "100",
		AskedHint: *formdata.AnythingFromString("false"),
	}
	result, err := json.Marshal(ConvertIntoMap(tt))
	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"askedHint":false,"date":"","idAttempt":"100","idUser":"10"}`), result)
}
