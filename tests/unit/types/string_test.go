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

	json_input := `{ "Title": "The Pragmatic Programmer", "Description": "From Journeyman to Master", "Author": "Andy Hunt", "LastReader": "John Doe" }`
	input := &SampleStrInput{}
	if err := json.Unmarshal([]byte(json_input), &input); err != nil {
		t.Error(err)
	}
	if input.Title.Value != "The Pragmatic Programmer" {
		t.Errorf("invalid decoded value: %s", input.Title.Value)
	}
	if input.Description.Value != "From Journeyman to Master" {
		t.Errorf("invalid decoded value: %s", input.Description.Value)
	}
	if input.Author.Value != "Andy Hunt" {
		t.Errorf("invalid decoded value: %s", input.Author.Value)
	}
	if input.LastReader.Value != "John Doe" {
		t.Errorf("invalid decoded value: %s", input.LastReader.Value)
	}
	if err := input.validate(); err != nil {
		t.Error(err)
	}
}

func TestStrWithNonStr(t *testing.T) {
	json_input := `{ "Title": 1234, "Description": "From Journeyman to Master", "Author": "Andy Hunt", "LastReader": "John Doe" }`
	input := &SampleStrInput{}
	if err := json.Unmarshal([]byte(json_input), &input); err == nil {
		t.Errorf("was expecting a decoding error, got: %s", input.Title.Value)
	}
}

func TestStrWithDefault(t *testing.T) {

	json_input := `{ "Title": "", "Description": "", "Author": "", "LastReader": "" }`
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
