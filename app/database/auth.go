package database

// This file describe the interfaces related to the auth package

// AuthUser represents an authenticated user from the app.
// Typically it is used to check the permissions and authorizations
type AuthUser interface {
	SelfGroupID() int64
}
