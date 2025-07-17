package formdata

import (
	"encoding/json"
	"sync/atomic"
	"unsafe"

	"github.com/France-ioi/AlgoreaBackend/v2/golang"
)

// Anything represents a value of any type serialized as JSON.
type Anything struct {
	raw *[]byte
}

// AnythingFromString creates an instance of Anything with data from the given string.
func AnythingFromString(s string) *Anything {
	return &Anything{raw: golang.Ptr([]byte(s))}
}

// AnythingFromBytes creates an instance of Anything with data from the given bytes slice.
func AnythingFromBytes(bytes []byte) *Anything {
	return &Anything{raw: golang.Ptr(bytes)}
}

// Bytes returns stored bytes.
func (a *Anything) Bytes() []byte {
	if a == nil {
		return nil
	}
	//nolint:gosec // here we use atomic operations to access the raw pointer safely
	rawPtr := (*[]byte)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&a.raw))))
	if rawPtr == nil {
		return nil
	}
	return *rawPtr
}

// UnmarshalJSON of Anything just copies the JSON data.
func (a *Anything) UnmarshalJSON(raw []byte) error {
	newRaw := make([]byte, len(raw))
	copy(newRaw, raw)
	//nolint:gosec // here we use atomic operations to update the raw pointer safely
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&a.raw)), unsafe.Pointer(&newRaw))
	return nil
}

// MarshalJSON of Anything copies the stored JSON data back.
func (a *Anything) MarshalJSON() ([]byte, error) {
	if a == nil {
		return []byte("null"), nil
	}
	//nolint:gosec // here we use atomic operations to access the raw pointer safely
	rawPtr := (*[]byte)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&a.raw))))
	if rawPtr == nil || len(*rawPtr) == 0 {
		return []byte("null"), nil
	}
	return *rawPtr, nil
}

var (
	_ = json.Unmarshaler(&Anything{})
	_ = json.Marshaler(&Anything{})
)
