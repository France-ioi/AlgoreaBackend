// +build !unit

package api_test

import (
	"testing"

	"github.com/France-ioi/AlgoreaBackend/testhelpers"
)

func TestBDD(t *testing.T) {
	testhelpers.RunGodogTests(t)
}
