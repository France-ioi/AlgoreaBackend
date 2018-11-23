package types_test // nolint

import (
  "encoding/json"
  "testing"

  t "github.com/France-ioi/AlgoreaBackend/app/types"
)

type SampleIntInput struct {
  ID       t.RequiredInt64
  ChildID  t.NullableInt64
  Order    t.OptionalInt64
  ParentID t.OptNullInt64
}

func (v *SampleIntInput) validate() error {
  return t.Validate(&v.ID, &v.ChildID, &v.Order, &v.ParentID)
}

func TestIntValid(t *testing.T) {

  json_input := `{ "ID": 2147483645, "ChildID": 22, "Order": -1, "ParentID": 7 }`
  input := &SampleIntInput{}
  if err := json.Unmarshal([]byte(json_input), &input); err != nil {
    t.Error(err)
  }
  if input.ID.Value != 2147483645 {
    t.Errorf("invalid decoded value: %d", input.ID.Value)
  }
  if input.ChildID.Value != 22 {
    t.Errorf("invalid decoded value: %d", input.ChildID.Value)
  }
  if input.Order.Value != -1 {
    t.Errorf("invalid decoded value: %d", input.Order.Value)
  }
  if input.ParentID.Value != 7 {
    t.Errorf("invalid decoded value: %d", input.ParentID.Value)
  }
  if err := input.validate(); err != nil {
    t.Error(err)
  }
}

func TestIntWithNonInt(t *testing.T) {
  json_input := `{ "ID": "not an int", "ChildID": 22, "Order": -1, "ParentID": 7 }`
  input := &SampleIntInput{}
  if err := json.Unmarshal([]byte(json_input), &input); err == nil {
    t.Errorf("was expecting a decoding error, got: %d", input.ID.Value)
  }
}

func TestIntWithDefault(t *testing.T) {

  json_input := `{ "ID": 0, "ChildID": 0, "Order": 0, "ParentID": 0 }`
  input := &SampleIntInput{}
  if err := json.Unmarshal([]byte(json_input), &input); err != nil {
    t.Error(err)
  }
  if err := input.validate(); err != nil {
    t.Error(err)
  }
}

func TestIntWithNull(t *testing.T) {

  json_input := `{ "ID": null, "ChildID": null, "Order": null, "ParentID": null }`
  input := &SampleIntInput{}
  if err := json.Unmarshal([]byte(json_input), &input); err != nil {
    t.Error(err)
  }
  if input.ID.Validate() == nil { // should NOT be valid
    t.Error("was expecting a validation error")
  }
  if err := input.ChildID.Validate(); err != nil { // should be valid
    t.Error(err)
  }
  if input.Order.Validate() == nil { // should NOT be valid
    t.Error("was expecting a validation error")
  }
  if err := input.ParentID.Validate(); err != nil { // should be valid
    t.Error(err)
  }
  if input.validate() == nil {
    t.Error("was expecting a validation error")
  }
}

func TestIntWithNotSet(t *testing.T) {

  json_input := `{}`
  input := &SampleIntInput{}
  if err := json.Unmarshal([]byte(json_input), &input); err != nil {
    t.Error(err)
  }
  if input.ID.Validate() == nil { // should NOT be valid
    t.Error("was expecting a validation error")
  }
  if input.ChildID.Validate() == nil { // should NOT be valid
    t.Error("was expecting a validation error")
  }
  if err := input.Order.Validate(); err != nil { // should be valid
    t.Error(err)
  }
  if err := input.ParentID.Validate(); err != nil { // should be valid
    t.Error(err)
  }
  if input.validate() == nil {
    t.Error("was expecting a validation error")
  }
}
