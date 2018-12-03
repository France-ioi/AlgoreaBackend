package types

import (
  "encoding/json"
  "testing"

  assert_lib "github.com/stretchr/testify/assert"
)

type SampleStrInput struct {
  Title       RequiredString
  Description NullableString
  Author      OptionalString
  LastReader  OptNullString
}

func (v *SampleStrInput) validate() error {
  return Validate(&v.Title, &v.Description, &v.Author, &v.LastReader)
}

func TestStrValid(t *testing.T) {
  assert := assert_lib.New(t)

  jsonInput := `{ "Title": "The Pragmatic Programmer", "Description": "From Journeyman to Master", "Author": "Andy Hunt", "LastReader": "John Doe" }`
  input := &SampleStrInput{}
  assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
  assert.Equal("The Pragmatic Programmer", input.Title.Value)
  assert.Equal("From Journeyman to Master", input.Description.Value)
  assert.Equal("Andy Hunt", input.Author.Value)
  assert.Equal("John Doe", input.LastReader.Value)
  assert.NoError(input.validate())
}

func TestStrWithNonStr(t *testing.T) {
  assert := assert_lib.New(t)

  jsonInput := `{ "Title": 1234, "Description": "From Journeyman to Master", "Author": "Andy Hunt", "LastReader": "John Doe" }`
  input := &SampleStrInput{}
  assert.Error(json.Unmarshal([]byte(jsonInput), &input))
}

func TestStrWithDefault(t *testing.T) {
  assert := assert_lib.New(t)

  jsonInput := `{ "Title": "", "Description": "", "Author": "", "LastReader": "" }`
  input := &SampleStrInput{}
  assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
  assert.NoError(input.validate())
}

func TestStrWithNull(t *testing.T) {
  assert := assert_lib.New(t)

  jsonInput := `{ "Title": null, "Description": null, "Author": null, "LastReader": null }`
  input := &SampleStrInput{}
  assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
  assert.Error(input.Title.Validate())
  assert.NoError(input.Description.Validate())
  assert.Error(input.Author.Validate())
  assert.NoError(input.LastReader.Validate())
  assert.Error(input.validate())
}

func TestStrWithNotSet(t *testing.T) {
  assert := assert_lib.New(t)

  jsonInput := `{}`
  input := &SampleStrInput{}
  assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
  assert.Error(input.Title.Validate())
  assert.Error(input.Description.Validate())
  assert.NoError(input.Author.Validate())
  assert.NoError(input.LastReader.Validate())
  assert.Error(input.validate())
}
