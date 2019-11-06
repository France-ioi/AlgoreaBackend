package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var canViewValues = []string{"none", "info", "content", "content_with_descendants", "solution"}

func TestItemAccessDetails_IsGrayed(t *testing.T) {
	for _, canView := range canViewValues {
		canView := canView
		t.Run(canView, func(t *testing.T) {
			assert.Equal(t, canView == "info", (&ItemAccessDetails{CanView: canView}).IsGrayed())
		})
	}
}

func TestItemAccessDetails_IsForbidden(t *testing.T) {
	for _, canView := range canViewValues {
		canView := canView
		t.Run(canView, func(t *testing.T) {
			assert.Equal(t, canView == "none", (&ItemAccessDetails{CanView: canView}).IsForbidden())
		})
	}
}
