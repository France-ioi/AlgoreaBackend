package golang

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSet(t *testing.T) {
	set1 := NewSet[string]()
	assert.NotNil(t, set1)
	assert.False(t, set1.isImmutable)
	assert.Empty(t, set1.data)

	set2 := NewSet[int]()
	assert.NotNil(t, set2)
	assert.False(t, set2.isImmutable)
	assert.Empty(t, set2.data)

	set3 := NewSet("a", "b", "c")
	assert.NotNil(t, set3)
	assert.False(t, set3.isImmutable)
	assert.Equal(t, map[string]struct{}{"a": {}, "b": {}, "c": {}}, set3.data)
}

func TestSet_Add(t *testing.T) {
	set1 := NewSet[string]().Add("a", "b")
	assert.False(t, set1.isImmutable)
	assert.Equal(t, map[string]struct{}{"a": {}, "b": {}}, set1.data)

	set2 := NewSet[int](1)
	set2.Add(2)
	assert.False(t, set1.isImmutable)
	assert.Equal(t, map[int]struct{}{1: {}, 2: {}}, set2.data)
}

func TestSet_Clone(t *testing.T) {
	set := NewSet[string]("a", "b")
	clone := set.Clone()
	assert.False(t, clone.isImmutable)
	assert.Equal(t, map[string]struct{}{"a": {}, "b": {}}, clone.data)

	set.isImmutable = true
	clone = set.Clone()
	assert.False(t, clone.isImmutable)
	assert.Equal(t, map[string]struct{}{"a": {}, "b": {}}, clone.data)
}

func TestSet_Contains(t *testing.T) {
	assert.True(t, NewSet[string]("a", "b").Contains("a"))
	assert.False(t, NewSet[string]("a", "b").Contains("c"))
	assert.True(t, NewSet[int](1, 2).Contains(1))
	assert.False(t, NewSet[int](1, 2).Contains(3))
}

func TestSet_IsEmpty(t *testing.T) {
	assert.True(t, NewSet[string]().IsEmpty())
	assert.True(t, NewSet[string]().MarkImmutable().IsEmpty())
	assert.False(t, NewSet[string]("a").IsEmpty())
	assert.False(t, NewSet[string]("a").MarkImmutable().IsEmpty())
}

func TestSet_IsImmutable(t *testing.T) {
	set := NewSet[string]()
	set.isImmutable = true
	assert.True(t, set.IsImmutable())
	set.isImmutable = false
	assert.False(t, set.IsImmutable())
}

func TestSet_MarkImmutable(t *testing.T) {
	set := NewSet[string]()
	assert.False(t, set.isImmutable)
	assert.NotPanics(t, func() { set.Add("a") })
	assert.NotPanics(t, func() { set.Remove("a") })
	assert.NotPanics(t, func() { set.MergeWith(NewSet("a")) })

	assert.Equal(t, set, set.MarkImmutable())
	assert.True(t, set.isImmutable)
	assert.Panics(t, func() { set.Add("a") })
	assert.Panics(t, func() { set.Remove("a") })
	assert.Panics(t, func() { set.MergeWith(NewSet("a")) })

	assert.True(t, NewSet("a", "b").MarkImmutable().IsImmutable())
}

func TestSet_MergeWith(t *testing.T) {
	set1 := NewSet[string]("a", "b")
	set2 := NewSet[string]("b", "c")
	assert.Equal(t, map[string]struct{}{"a": {}, "b": {}, "c": {}}, set1.MergeWith(set2).data)
	assert.Equal(t, map[string]struct{}{"a": {}, "b": {}, "c": {}}, set1.data)
	assert.Equal(t, map[string]struct{}{"b": {}, "c": {}}, set2.data)
}

func TestSet_Remove(t *testing.T) {
	set := NewSet[string]("a", "b")
	assert.Equal(t, map[string]struct{}{"a": {}, "b": {}}, set.Remove("c").data)
	assert.Equal(t, map[string]struct{}{"a": {}, "b": {}}, set.data)
	assert.Equal(t, map[int]struct{}{}, NewSet[int]().Remove(0).data)
}

func TestSet_Size(t *testing.T) {
	assert.Equal(t, 0, NewSet[string]().Size())
	assert.Equal(t, 2, NewSet[string]("a", "b").Size())
	assert.Equal(t, 3, NewSet[string]("a", "b", "c").Size())
}

func TestSet_Values(t *testing.T) {
	assert.Equal(t, []string{}, NewSet[string]().Values())

	values := NewSet[string]("a", "b").Values()
	sort.Strings(values)
	assert.Equal(t, []string{"a", "b"}, values)
}
