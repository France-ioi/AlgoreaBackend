// Package rand provides utilities to generate random data.
package rand

import (
	mr "math/rand"
	"sync"
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
