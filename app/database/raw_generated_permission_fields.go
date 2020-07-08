package database

import (
	"github.com/France-ioi/AlgoreaBackend/app/structures"
)

// RawGeneratedPermissionFields represents DB data fields for item permissions used by item-related services
type RawGeneratedPermissionFields struct {
	CanViewGeneratedValue      int
	CanGrantViewGeneratedValue int
	CanWatchGeneratedValue     int
	CanEditGeneratedValue      int
	IsOwnerGenerated           bool
}

// AsItemPermissions converts RawGeneratedPermissionFields into structures.ItemPermissions
func (raw *RawGeneratedPermissionFields) AsItemPermissions(
	permissionGrantedStore *PermissionGrantedStore) *structures.ItemPermissions {
	return &structures.ItemPermissions{
		CanView:      permissionGrantedStore.ViewNameByIndex(raw.CanViewGeneratedValue),
		CanGrantView: permissionGrantedStore.GrantViewNameByIndex(raw.CanGrantViewGeneratedValue),
		CanWatch:     permissionGrantedStore.WatchNameByIndex(raw.CanWatchGeneratedValue),
		CanEdit:      permissionGrantedStore.EditNameByIndex(raw.CanEditGeneratedValue),
		IsOwner:      raw.IsOwnerGenerated,
	}
}
