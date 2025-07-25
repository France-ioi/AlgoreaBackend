package payloads

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/formdata"
)

func TestHintToken_UnmarshalJSON_InvalidJSON(t *testing.T) {
	tt := &HintToken{}
	err := tt.UnmarshalJSON([]byte("[]"))
	require.Error(t, err)
	assert.Equal(t, "json: cannot unmarshal array into Go value of type map[string]*formdata.Anything", err.Error())
}

func TestHintToken_UnmarshalJSON_WrongUserID(t *testing.T) {
	tt := &HintToken{}
	err := tt.UnmarshalJSON([]byte(`{"idUser":"abc"}`))
	require.Error(t, err)
	assert.Equal(t, "wrong idUser", err.Error())
}

func TestHintToken_UnmarshalJSON(t *testing.T) {
	tt := &HintToken{}
	err := tt.UnmarshalJSON([]byte(`{"idUser":"10", "askedHint":   false}`))
	require.NoError(t, err)
	assert.Equal(t, &HintToken{
		UserID:    "10",
		AskedHint: formdata.AnythingFromString("false"),
		Converted: HintTokenConverted{
			UserID: 10,
		},
	}, tt)
}

func TestHintToken_MarshalJSON(t *testing.T) {
	tt := &HintToken{
		UserID:      "10",
		AttemptID:   "100",
		LocalItemID: "200",
		AskedHint:   formdata.AnythingFromString("false"),
	}
	result, err := json.Marshal(ConvertIntoMap(tt))
	require.NoError(t, err)
	assert.JSONEq(t, `{"askedHint":false,"date":"","idAttempt":"100","idItemLocal":"200","idUser":"10","itemUrl":""}`, string(result))
}
