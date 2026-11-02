package monkey

import (
	"reflect"
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
