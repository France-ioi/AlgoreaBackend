// Package rand provides utilities to generate random data.
package rand

import (
	mr "math/rand"
	"sync"
	"unsafe"
)

var (
	globalRand = mr.New(mr.NewSource(1))
	globalLock sync.Mutex
)

// Seed uses the provided seed value to initialize the generator to a deterministic state.
func Seed(s int64) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalRand.Seed(s)
}

// Int31n returns, as an int32, a non-negative pseudo-random number in [0,n). It panics if n <= 0.
func Int31n(n int32) int32 {
	globalLock.Lock()
	defer globalLock.Unlock()
	return globalRand.Int31n(n)
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func Int63() int64 {
	globalLock.Lock()
	defer globalLock.Unlock()
	return globalRand.Int63()
}

// Float64 returns, as a float64, a pseudo-random number in [0.0,1.0).
func Float64() float64 {
	globalLock.Lock()
	defer globalLock.Unlock()
	return globalRand.Float64()
}

// String returns a random string of n characters.
func String(n int) string {
	globalLock.Lock()
	defer globalLock.Unlock()

	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[globalRand.Intn(len(letters))]
	}
	return string(s)
}

const rngLen = 607

type rngSource struct {
	_ int           // index into vec
	_ int           // index into vec
	_ [rngLen]int64 // current feedback register
}

type sourceInterface struct {
	typ unsafe.Pointer
	val *rngSource
}

type rnd struct {
	src   sourceInterface
	src64 sourceInterface
}

// GetSource returns a copy of the current source of the random number generator.
func GetSource() interface{} {
	globalLock.Lock()
	defer globalLock.Unlock()

	source := ((*rnd)(unsafe.Pointer(globalRand))).src64.val //nolint:gosec // G103: Valid use of unsafe call

	sourceCopy := &rngSource{}
	*sourceCopy = *source
	return sourceCopy
}

// SetSource sets the source of the random number generator to a copy of the given source.
func SetSource(newSource interface{}) {
	globalLock.Lock()
	defer globalLock.Unlock()

	source := ((*rnd)(unsafe.Pointer(globalRand))).src64.val //nolint:gosec // G103: Valid use of unsafe call
	*source = *newSource.(*rngSource)
}
