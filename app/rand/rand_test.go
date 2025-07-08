package rand

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat64(t *testing.T) {
	Seed(1)
	assert.Equal(t, 0.6046602879796196, Float64())
	assert.Equal(t, 0.9405090880450124, Float64())
	Seed(2)
	assert.Equal(t, 0.16729663442585624, Float64())
	assert.Equal(t, 0.2650543054337802, Float64())
	Seed(1)
	assert.Equal(t, 0.6046602879796196, Float64())
}

func TestInt31n(t *testing.T) {
	Seed(1)
	assert.Equal(t, int32(1298498081), Int31n(math.MaxInt32))
	assert.Equal(t, int32(2019727887), Int31n(math.MaxInt32))
	Seed(2)
	assert.Equal(t, int32(359266786), Int31n(math.MaxInt32))
	assert.Equal(t, int32(569199786), Int31n(math.MaxInt32))
	Seed(1)
	assert.Equal(t, int32(1298498081), Int31n(math.MaxInt32))
	Seed(1)
	assert.Equal(t, int32(939984059), Int31n(1298498081))
}

func TestInt63(t *testing.T) {
	Seed(1)
	assert.Equal(t, int64(5577006791947779410), Int63())
	assert.Equal(t, int64(8674665223082153551), Int63())
	Seed(2)
	assert.Equal(t, int64(1543039099823358511), Int63())
	assert.Equal(t, int64(2444694468985893231), Int63())
	Seed(1)
	assert.Equal(t, int64(5577006791947779410), Int63())
}

func TestGetSource_SetSource(t *testing.T) {
	Seed(1)
	Int63()
	Float64()
	oldSource := GetSource()
	values := make([]interface{}, 0, 4)
	values = append(values, Float64(), Int63(), Float64(), Int63())
	SetSource(oldSource)
	newValues := make([]interface{}, 0, 4)
	newValues = append(newValues, Float64(), Int63(), Float64(), Int63())
	assert.Equal(t, values, newValues)
	SetSource(oldSource)
	newValues = []interface{}{Float64(), Int63(), Float64(), Int63()}
	assert.Equal(t, values, newValues)
}
