//go:build tools

// Package tools is a build-time dependency stub for the swagger CLI tool.
//
// Why this exists:
//
//	`go install github.com/France-ioi/go-swagger/cmd/swagger@<rev>` resolves
//	dependencies from that revision's go.mod, which transitively pins
//	golang.org/x/tools v0.21.0. That version of x/tools contains a compile-time
//	size assertion on the stdlib's token.FileSet layout
//	(internal/tokeninternal/tokeninternal.go: `var _ [-delta * delta]int`).
//	The layout changed in Go 1.23+, so the assertion fires as
//	"invalid array length -delta * delta (constant -256 of type int64)" and
//	swagger fails to compile.
//
//	Building swagger from inside this module instead lets us pin
//	golang.org/x/tools >= v0.26.0 (the first release that no longer trips on
//	Go 1.25's token.FileSet layout) via this directory's go.mod.
//
// Usage (from the repo root):
//
//	cd tools/swagger && go install github.com/France-ioi/go-swagger/cmd/swagger
package tools

import _ "github.com/France-ioi/go-swagger/cmd/swagger"
