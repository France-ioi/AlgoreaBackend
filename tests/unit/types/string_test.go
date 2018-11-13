package types_test

import (
	"encoding/json"
	"testing"

	t "github.com/France-ioi/AlgoreaBackend/app/types"
)

type SampleStrInput struct {
	Title       t.RequiredString
	Description t.NullableString
	Author      t.OptionalString
	LastReader  t.OptNullString
}

func (v *SampleStrInput) validate() error {
	return t.Validate(&v.Title, &v.Description, &v.Author, &v.LastReader)
}

func TestStrValid(t *testing.T) {

	json_input := `{ "Title": 2147483645, "Description": 22, "Author": -1, "LastReader": 7 }`
	input := &SampleStrInput{}
	if err := json.Unmarshal([]byte(json_input), &input); err != nil {
		t.Error(err)
	}
	if input.Title.Value != 2147483645 {
		t.Errorf("invalid decoded value: %d", input.Title.Value)
	}
	if input.Description.Value != 22 {
		t.Errorf("invalid decoded value: %d", input.Description.Value)
	}
	if input.Author.Value != -1 {
		t.Errorf("invalid decoded value: %d", input.Author.Value)
	}
	if input.LastReader.Value != 7 {
		t.Errorf("invalid decoded value: %d", input.LastReader.Value)
	}
	if err := input.validate(); err != nil {
		t.Error(err)
	}
}

func TestStrWithNonStr(t *testing.T) {
	json_input := `{ "Title": "not an int", "Description": 22, "Author": -1, "LastReader": 7 }`
	input := &SampleStrInput{}
	if err := json.Unmarshal([]byte(json_input), &input); err == nil {
		t.Errorf("was expecting a decoding error, got: %d", input.Title.Value)
	}
}

func TestStrWithDefault(t *testing.T) {

	json_input := `{ "Title": 0, "Description": 0, "Author": 0, "LastReader": 0 }`
	input := &SampleStrInput{}
	if err := json.Unmarshal([]byte(json_input), &input); err != nil {
		t.Error(err)
	}
	if err := input.validate(); err != nil {
		t.Error(err)
	}
}

func TestStrWithNull(t *testing.T) {

	json_input := `{ "Title": null, "Description": null, "Author": null, "LastReader": null }`
	input := &SampleStrInput{}
	if err := json.Unmarshal([]byte(json_input), &input); err != nil {
		t.Error(err)
	}
	if input.Title.Validate() == nil { // should NOT be valid
		t.Error("was expecting a validation error")
	}
	if err := input.Description.Validate(); err != nil { // should be valid
		t.Error(err)
	}
	if input.Author.Validate() == nil { // should NOT be valid
		t.Error("was expecting a validation error")
	}
	if err := input.LastReader.Validate(); err != nil { // should be valid
		t.Error(err)
	}
	if input.validate() == nil {
		t.Error("was expecting a validation error")
	}
}

func TestStrWithNotSet(t *testing.T) {

	json_input := `{}`
	input := &SampleStrInput{}
	if err := json.Unmarshal([]byte(json_input), &input); err != nil {
		t.Error(err)
	}
	if input.Title.Validate() == nil { // should NOT be valid
		t.Error("was expecting a validation error")
	}
	if input.Description.Validate() == nil { // should NOT be valid
		t.Error("was expecting a validation error")
	}
	if err := input.Author.Validate(); err != nil { // should be valid
		t.Error(err)
	}
	if err := input.LastReader.Validate(); err != nil { // should be valid
		t.Error(err)
	}
	if input.validate() == nil {
		t.Error("was expecting a validation error")
	}
}
