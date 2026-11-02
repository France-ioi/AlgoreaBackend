// Package monkey is a bou.ke/monkey-compatible facade over bytedance/mockey.
// bou.ke/monkey instance-method patches break on Go 1.26+; mockey supports 1.26.
package monkey

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/bytedance/mockey"
)

// PatchGuard mirrors bou.ke/monkey.PatchGuard: Unpatch temporarily disables the
// patch so the original can be called; Restore re-enables it.
type PatchGuard struct {
	mocker    *mockey.Mocker
	targetKey uintptr
	active    bool
}

// Unpatch temporarily disables the patch (original is callable again).
func (p *PatchGuard) Unpatch() {
	if p == nil || p.mocker == nil || !p.active {
		return
	}
	p.mocker.UnPatch()
	p.active = false
}

// Restore re-enables a previously Unpatch()'d guard.
// No-op if the guard was replaced by a later Patch of the same target.
func (p *PatchGuard) Restore() {
	if p == nil || p.mocker == nil || p.active {
		return
	}
	p.mocker.Patch()
	p.active = true
}

//nolint:gochecknoglobals // tracks active patches so UnpatchAll can clear them (same as bou.ke/monkey)
var (
	guardsMutex sync.Mutex
	guards      []*PatchGuard
)

func targetKeyOf(target any) uintptr {
	return reflect.ValueOf(target).Pointer()
}

// unpatchExistingLocked drops any prior *active* mock of the same target so
// re-Patch behaves like bou.ke/monkey (replace) instead of mockey's re-mock panic.
// Temporarily Unpatch()'d guards are left alone so nested Patch+Restore (e.g.
// MockDBTime inside an Exec patch) can Restore the outer guard afterwards.
// Permanently replaced guards are marked dead (mocker=nil) so Restore is a no-op.
func unpatchExistingLocked(key uintptr) {
	kept := guards[:0]
	for _, guard := range guards {
		if guard == nil || guard.targetKey != key {
			kept = append(kept, guard)
			continue
		}
		if !guard.active {
			kept = append(kept, guard)
			continue
		}
		if guard.mocker != nil {
			guard.mocker.UnPatch()
		}
		guard.active = false
		guard.mocker = nil
	}
	guards = kept
}

func buildMock(target, replacement any) (mocker *mockey.Mocker) {
	// Prefer safe Mock; fall back to MockUnsafe only when mockey rejects a function
	// as too short (one-line getters) — same cases bou.ke/monkey accepted.
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		msg := fmt.Sprint(recovered)
		if !strings.Contains(msg, "too short") {
			panic(recovered)
		}
		mocker = mockey.MockUnsafe(target).To(replacement).Build()
	}()
	mocker = mockey.Mock(target).To(replacement).Build()
	return mocker
}

func register(key uintptr, target, replacement any) *PatchGuard {
	guardsMutex.Lock()
	unpatchExistingLocked(key)
	guardsMutex.Unlock()

	g := &PatchGuard{mocker: buildMock(target, replacement), targetKey: key, active: true}

	guardsMutex.Lock()
	guards = append(guards, g)
	guardsMutex.Unlock()
	return g
}

// Patch replaces a function with replacement (same signature as bou.ke/monkey.Patch).
func Patch(target, replacement any) *PatchGuard {
	return register(targetKeyOf(target), target, replacement)
}

// PatchInstanceMethod replaces an instance method (same signature as bou.ke/monkey.PatchInstanceMethod).
func PatchInstanceMethod(target reflect.Type, methodName string, replacement any) *PatchGuard {
	var instance any
	switch target.Kind() {
	case reflect.Ptr:
		instance = reflect.New(target.Elem()).Interface()
	default:
		instance = reflect.New(target).Elem().Interface()
	}
	method := mockey.GetMethod(instance, methodName)
	return register(targetKeyOf(method), method, replacement)
}

// UnpatchAll permanently disables every patch created through this package.
func UnpatchAll() {
	guardsMutex.Lock()
	current := guards
	guards = nil
	guardsMutex.Unlock()
	for i := len(current) - 1; i >= 0; i-- {
		guard := current[i]
		if guard != nil && guard.mocker != nil && guard.active {
			guard.mocker.UnPatch()
			guard.active = false
		}
	}
}
