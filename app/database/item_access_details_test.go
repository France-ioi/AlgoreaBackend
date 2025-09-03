package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func canViewValues() []string {
	return []string{"none", "info", "content", "content_with_descendants", "solution"}
}

func TestItemAccessDetails_IsInfo(t *testing.T) {
	for _, canView := range canViewValues() {
		canView := canView
		t.Run(canView, func(t *testing.T) {
			assert.Equal(t, canView == "info", (&ItemAccessDetails{CanView: canView}).IsInfo())
		})
	}
}

func TestItemAccessDetails_IsForbidden(t *testing.T) {
	for _, canView := range canViewValues() {
		canView := canView
		t.Run(canView, func(t *testing.T) {
			assert.Equal(t, canView == "none", (&ItemAccessDetails{CanView: canView}).IsForbidden())
		})
	}
}
