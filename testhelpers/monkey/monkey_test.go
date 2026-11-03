package monkey

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	origValue    = "orig"
	patchedValue = "patched"
)

func shortTarget() string { return origValue }

func longEnoughTarget() string {
	s := origValue
	s += "-suffix"
	return s
}

type patchable struct{}

func (p *patchable) Value() string {
	s := origValue
	s += "-method"
	return s
}

func TestPatch_ReplaceOnRepatchMarksOldGuardDead(t *testing.T) {
	defer UnpatchAll()

	first := Patch(longEnoughTarget, func() string { return "first" })
	require.Equal(t, "first", longEnoughTarget())

	second := Patch(longEnoughTarget, func() string { return "second" })
	require.Equal(t, "second", longEnoughTarget())

	// Replaced (still-active) guard must not re-attach an untracked patch.
	first.Restore()
	assert.Equal(t, "second", longEnoughTarget())
	assert.Nil(t, first.mocker)

	second.Unpatch()
	assert.Equal(t, "orig-suffix", longEnoughTarget())
	second.Restore()
	assert.Equal(t, "second", longEnoughTarget())
}

func TestPatch_NestedUnpatchAllowsInnerPatchAndOuterRestore(t *testing.T) {
	defer UnpatchAll()

	outer := Patch(longEnoughTarget, func() string { return "outer" })
	require.Equal(t, "outer", longEnoughTarget())

	outer.Unpatch()
	inner := Patch(longEnoughTarget, func() string { return "inner" })
	assert.Equal(t, "inner", longEnoughTarget())
	inner.Unpatch()

	outer.Restore()
	assert.Equal(t, "outer", longEnoughTarget())
	assert.NotNil(t, outer.mocker)
}

func TestPatch_UnpatchRestoreToggle(t *testing.T) {
	defer UnpatchAll()

	guard := Patch(longEnoughTarget, func() string { return patchedValue })
	require.Equal(t, patchedValue, longEnoughTarget())

	guard.Unpatch()
	assert.Equal(t, "orig-suffix", longEnoughTarget())

	guard.Restore()
	assert.Equal(t, patchedValue, longEnoughTarget())
}

func TestUnpatchAll_ClearsActivePatches(t *testing.T) {
	Patch(longEnoughTarget, func() string { return patchedValue })
	require.Equal(t, patchedValue, longEnoughTarget())

	UnpatchAll()
	assert.Equal(t, "orig-suffix", longEnoughTarget())
	assert.Empty(t, guards)
}

func TestPatch_TooShortFallsBackToMockUnsafe(t *testing.T) {
	defer UnpatchAll()

	// One-liner is too short for safe Mock; MockUnsafe must succeed.
	guard := Patch(shortTarget, func() string { return "unsafe" })
	require.NotNil(t, guard)
	assert.Equal(t, "unsafe", shortTarget())
}

func TestPatchInstanceMethod_PatchesMethod(t *testing.T) {
	defer UnpatchAll()

	guard := PatchInstanceMethod(reflect.TypeFor[*patchable](), "Value", func(*patchable) string {
		return patchedValue
	})
	require.NotNil(t, guard)
	assert.Equal(t, patchedValue, (&patchable{}).Value())
}

func TestPatchGuard_NilAndIdempotent(t *testing.T) {
	var nilGuard *PatchGuard
	nilGuard.Unpatch()
	nilGuard.Restore()

	defer UnpatchAll()
	guard := Patch(longEnoughTarget, func() string { return patchedValue })
	require.Equal(t, patchedValue, longEnoughTarget())

	// Already active: Restore is a no-op.
	guard.Restore()
	assert.Equal(t, patchedValue, longEnoughTarget())

	guard.Unpatch()
	assert.Equal(t, "orig-suffix", longEnoughTarget())
	// Already inactive: Unpatch is a no-op.
	guard.Unpatch()
	assert.Equal(t, "orig-suffix", longEnoughTarget())
}

func TestPatchGuard_ConcurrentUnpatchRestore(t *testing.T) {
	defer UnpatchAll()

	guard := Patch(longEnoughTarget, func() string { return patchedValue })
	require.Equal(t, patchedValue, longEnoughTarget())

	const goroutines = 8
	const iterations = 200
	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutines)
	for range goroutines {
		go func() {
			defer waitGroup.Done()
			for range iterations {
				// Do not call the patched target here: concurrent Unpatch/Restore
				// rewrites its prologue; a torn fetch can SIGSEGV. This test only
				// needs races on active bookkeeping.
				guard.Unpatch()
				guard.Restore()
			}
		}()
	}
	waitGroup.Wait()

	assert.Equal(t, patchedValue, longEnoughTarget())
}

func TestPatchInstanceMethodWithOrigin_CallsRealMethod(t *testing.T) {
	defer UnpatchAll()

	var origin func(*patchable) string
	guard := PatchInstanceMethodWithOrigin(reflect.TypeFor[*patchable](), "Value",
		func(p *patchable) string {
			return origin(p) + "-via-origin"
		},
		&origin,
	)
	require.NotNil(t, guard)
	require.NotNil(t, origin)
	assert.Equal(t, "orig-method-via-origin", (&patchable{}).Value())
}
