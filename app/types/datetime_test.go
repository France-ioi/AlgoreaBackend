package types

import (
	"encoding/json"
	"testing"
	"time"

	assertlib "github.com/stretchr/testify/assert"
)

type SampleDTInput struct {
	CreatedAt   RequiredDatetime
	ValidFrom   NullableDatetime
	ModifiedAt  OptionalDatetime
	AccessUntil OptNullDatetime
}

func (v *SampleDTInput) validate() error {
	return Validate([]string{"createdAt", "validFrom", "optionalDatetime", "optNullDatetime"},
		&v.CreatedAt, &v.ValidFrom, &v.ModifiedAt, &v.AccessUntil)
}

func TestNewDT(t *testing.T) {
	assert := assertlib.New(t)

	value := time.Date(2001, time.February, 3, 5, 6, 7, 890000000, time.UTC)
	n := NewDatetime(value)
	assert.Equal(value, n.Value)
	assert.True(n.Set)
	assert.False(n.Null)
}

func TestDTValid(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "CreatedAt": "2001-02-03T05:06:07.89Z", "ValidFrom": "2002-01-01T23:11:11.000000001+02:00", "ModifiedAt": "2001-09-02T12:30:00Z", "AccessUntil": "2042-12-31T23:59:59Z" }`
	input := &SampleDTInput{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.Equal(time.Date(2001, time.February, 3, 5, 6, 7, 890000000, time.UTC), input.CreatedAt.Value)
	assert.Equal(time.Date(2002, time.January, 1, 21, 11, 11, 1, time.UTC), input.ValidFrom.Value.In(time.UTC))
	assert.Equal(time.Date(2001, time.September, 2, 12, 30, 0, 0, time.UTC), input.ModifiedAt.Value)
	assert.Equal(time.Date(2042, time.December, 31, 23, 59, 59, 0, time.UTC), input.AccessUntil.Value)
	assert.NoError(input.validate())
}

func TestDTWithNonDT(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "CreatedAt": "2001", "ValidFrom": "2002-01-01T23:11:11Z", "ModifiedAt": "2001-09-02T12:30Z", "AccessUntil": "2042-12-31T23:59:59Z" }`
	input := &SampleDTInput{}
	assert.Error(json.Unmarshal([]byte(jsonInput), &input))
}

func TestDTWithNull(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "CreatedAt": null, "ValidFrom": null, "ModifiedAt": null, "AccessUntil": null }`
	input := &SampleDTInput{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.Error(input.CreatedAt.Validate())
	assert.NoError(input.ValidFrom.Validate())
	assert.Error(input.ModifiedAt.Validate())
	assert.NoError(input.AccessUntil.Validate())
	assert.Error(input.validate())
}

func TestDTWithNotSet(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{}`
	input := &SampleDTInput{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.Error(input.CreatedAt.Validate())
	assert.Error(input.ValidFrom.Validate())
	assert.NoError(input.ModifiedAt.Validate())
	assert.NoError(input.AccessUntil.Validate())
	assert.Error(input.validate())
}
