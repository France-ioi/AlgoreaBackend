package database

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/France-ioi/AlgoreaBackend/v2/app/structures"
)

func TestRawGeneratedPermissionFields_AsItemPermissions(t *testing.T) {
	tests := []struct {
		name string
		raw  *RawGeneratedPermissionFields
		want *structures.ItemPermissions
	}{
		{
			name: "one",
			raw: &RawGeneratedPermissionFields{
				CanViewGeneratedValue:      3,
				CanGrantViewGeneratedValue: 6,
				CanWatchGeneratedValue:     2,
				CanEditGeneratedValue:      4,
				IsOwnerGenerated:           false,
			},
			want: &structures.ItemPermissions{
				CanView:      "content",
				CanGrantView: "solution_with_grant",
				CanWatch:     "result",
				CanEdit:      "all_with_grant",
				IsOwner:      false,
			},
		},
		{
			name: "two",
			raw: &RawGeneratedPermissionFields{
				CanViewGeneratedValue:      4,
				CanGrantViewGeneratedValue: 5,
				CanWatchGeneratedValue:     3,
				CanEditGeneratedValue:      3,
				IsOwnerGenerated:           true,
			},
			want: &structures.ItemPermissions{
				CanView:      "content_with_descendants",
				CanGrantView: "solution",
				CanWatch:     "answer",
				CanEdit:      "all",
				IsOwner:      true,
			},
		},
	}
	db, mock := NewDBMock()
	ClearAllDBEnums()
	MockDBEnumQueries(mock)
	defer ClearAllDBEnums()
	permissionGrantedStore := NewDataStore(db).PermissionsGranted()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.raw.AsItemPermissions(permissionGrantedStore))
		})
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}
