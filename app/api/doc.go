// Package api AlgoreaBackend API.
// API for the Algorea backend.
//
//     Schemes: http, https
//     Host: localhost
//     BasePath: /
//     Version: 0.0.1
//     License: MIT http://opensource.org/licenses/MIT
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
// swagger:meta
package api

import (
	"github.com/France-ioi/AlgoreaBackend/app/api/groups"
)

// Not actually a response, just a hack to get go-swagger to include definitions
//
// swagger:response parameterBodies
type paramBodies struct { //nolint: deadcode
	// in: body
	GroupUpdateInput groups.GroupUpdateInput
}
