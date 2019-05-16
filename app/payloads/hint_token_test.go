package payloads

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHintToken_UnmarshalJSON_InvalidJSON(t *testing.T) {
	tt := &HintToken{}
	err := tt.UnmarshalJSON([]byte("[]"))
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, "json: cannot unmarshal array into Go value of type map[string]payloads.Anything", err.Error())
	}
}

func TestHintToken_UnmarshalJSON_WrongUserID(t *testing.T) {
	tt := &HintToken{}
	err := tt.UnmarshalJSON([]byte("{}"))
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
		UserID:    "10",
		AskedHint: Anything("false"),
		Converted: HintTokenConverted{
			UserID: 10,
		},
	}, tt)
}
