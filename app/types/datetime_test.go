package types

import (
	"encoding/json"
	"testing"
	"time"

	assertlib "github.com/stretchr/testify/assert"
)

type SampleDTInput struct {
	Required         RequiredDatetime
	Nullable         NullableDatetime
	Optional         OptionalDatetime
	OptionalNullable OptNullDatetime
}

func (v *SampleDTInput) Validate() error {
	return Validate([]string{"Required", "Nullable", "Optional", "OptionalNullable"},
		&v.Required, &v.Nullable, &v.Optional, &v.OptionalNullable)
}

func TestDTValid(t *testing.T) {
	assert := assertlib.New(t)

	jsonInput := `{ "Required": "2001-02-03T05:06:07.89Z", "Nullable": "2002-01-01T23:11:11.000000001+02:00", ` +
		`"Optional": "2001-09-02T12:30:00Z", "OptionalNullable": "2042-12-31T23:59:59Z" }`
	input := &SampleDTInput{}
	assert.NoError(json.Unmarshal([]byte(jsonInput), &input))
	assert.Equal(time.Date(2001, time.February, 3, 5, 6, 7, 890000000, time.UTC), input.Required.Value)
	assert.Equal(time.Date(2002, time.January, 1, 21, 11, 11, 1, time.UTC), input.Nullable.Value.(time.Time).In(time.UTC))
	assert.Equal(time.Date(2001, time.September, 2, 12, 30, 0, 0, time.UTC), input.Optional.Value)
	assert.Equal(time.Date(2042, time.December, 31, 23, 59, 59, 0, time.UTC), input.OptionalNullable.Value)
	assert.NoError(input.Validate())
}
