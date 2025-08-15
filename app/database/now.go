package database

import "github.com/jinzhu/gorm"

//nolint:gochecknoglobals // we use nowExpr instead of just gorm.Expr() for testing purposes
var nowExpr = gorm.Expr("NOW()")

// Now returns a DB expression that returns current DB time (it is usually gorm.Expr("NOW()")).
func Now() *gorm.SqlExpr {
	return nowExpr
}

// MockNow changes the DB expression for getting current DB time so that it will return the given timestamp.
func MockNow(timestamp string) (oldNow *gorm.SqlExpr) {
	oldNow = nowExpr
	nowExpr = gorm.Expr("?", timestamp)
	return
}

// RestoreNow sets the DB expression for getting current DB time to its default value (gorm.Expr("NOW()")).
func RestoreNow(oldNow *gorm.SqlExpr) {
	nowExpr = oldNow
}
