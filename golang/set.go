package golang

// Set is a simple implementation of set of values of type T.
// It allows to mark the set as immutable to prevent further modifications.
// It is not thread-safe.
type Set[T comparable] struct {
	data        map[T]struct{}
	isImmutable bool
}

// NewSet creates a new set with the given values.
func NewSet[T comparable](values ...T) *Set[T] {
	set := &Set[T]{data: make(map[T]struct{}, len(values))}
	for _, value := range values {
		set.data[value] = struct{}{}
	}
	return set
}

// MarkImmutable marks the set as immutable.
func (s *Set[T]) MarkImmutable() *Set[T] {
	s.isImmutable = true
	return s
}

func (s *Set[T]) mustBeMutable() {
	if s.isImmutable {
		panic("cannot modify an immutable set")
	}
}

// Add adds the given values to the set.
func (s *Set[T]) Add(value ...T) *Set[T] {
	s.mustBeMutable()
	for _, value := range value {
		s.data[value] = struct{}{}
	}
	return s
}

// Remove removes the given values from the set.
func (s *Set[T]) Remove(value T) *Set[T] {
	s.mustBeMutable()
	delete(s.data, value)
	return s
}

// Contains returns true if the set contains the given value.
func (s *Set[T]) Contains(value T) bool {
	_, ok := s.data[value]
	return ok
}

// IsEmpty returns true if the set is empty.
func (s *Set[T]) IsEmpty() bool {
	return len(s.data) == 0
}

// IsImmutable returns true if the set is immutable.
func (s *Set[T]) IsImmutable() bool {
	return s.isImmutable
}

// Size returns the number of elements in the set.
func (s *Set[T]) Size() int {
	return len(s.data)
}

// MergeWith adds all values from the other set to the set.
func (s *Set[T]) MergeWith(other *Set[T]) *Set[T] {
	s.mustBeMutable()
	for value := range other.data {
		s.data[value] = struct{}{}
	}
	return s
}

// Values returns all values in the set as a slice.
// The order of the values is not guaranteed.
func (s *Set[T]) Values() []T {
	values := make([]T, 0, len(s.data))
	for value := range s.data {
		values = append(values, value)
	}
	return values
}

// Clone returns a mutable shallow copy of the set.
func (s *Set[T]) Clone() *Set[T] {
	clone := NewSet[T]()
	clone.data = make(map[T]struct{}, len(s.data))
	for value := range s.data {
		clone.data[value] = struct{}{}
	}
	return clone
}
