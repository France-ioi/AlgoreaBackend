package database

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestNow(t *testing.T) {
	expectedExpr := gorm.Expr("123")
	nowExpr = expectedExpr
	defer func() { nowExpr = gorm.Expr("NOW()") }()
	assert.Equal(t, expectedExpr, Now())
}

func TestMockNow(t *testing.T) {
	expectedExpr := gorm.Expr("?", "2017-01-02T12:34:56Z")
	oldNow := MockNow("2017-01-02T12:34:56Z")
	defer func() { nowExpr = gorm.Expr("NOW()") }()
	assert.Equal(t, expectedExpr, Now())
	assert.Equal(t, gorm.Expr("NOW()"), oldNow)
}

func TestRestoreNow(t *testing.T) {
	expectedExpr := gorm.Expr("NOW()")
	nowExpr = gorm.Expr("123")
	defer func() { nowExpr = expectedExpr }()
	RestoreNow(expectedExpr)
	assert.Equal(t, expectedExpr, nowExpr)
}
