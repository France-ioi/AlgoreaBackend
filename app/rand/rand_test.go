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

func TestString(t *testing.T) {
	Seed(1)
	assert.Equal(t, "", String(0))
	assert.Equal(t, "BpLnfgDsc2", String(10))
	Seed(2)
	assert.Equal(t, "KSiOW4eQ7s", String(10))
	assert.Equal(t, "klpgstrQZt", String(10))
	Seed(1)
	assert.Equal(t, "BpLnfgDsc2W", String(11))
}
