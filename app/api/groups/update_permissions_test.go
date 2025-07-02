package groups

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/France-ioi/AlgoreaBackend/v2/app/database"
	"github.com/France-ioi/AlgoreaBackend/v2/app/service"
)

var (
	canViewValues      = [...]string{"none", "info", "content", "content_with_descendants", "solution"}
	canGrantViewValues = [...]string{"none", "enter", "content", "content_with_descendants", "solution", "solution_with_grant"}
	canWatchValues     = [...]string{"none", "result", "answer", "answer_with_grant"}
	canEditValues      = [...]string{"none", "children", "all", "all_with_grant"}
)

func Test_checkIfPossibleToModifyCanView(t *testing.T) {
	type args struct {
		viewPermissionToSet string
		currentPermissions  *userPermissions
		managerPermissions  *managerGeneratedPermissions
	}
	type testStruct struct {
		name string
		args args
		want bool
	}
	var tests []testStruct

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	for _, currentCanView := range canViewValues {
		var newValueIsGreater bool
		for _, newCanViewValue := range canViewValues {
			if newValueIsGreater {
				for _, canGrantViewValue := range canGrantViewValues {
					tests = append(tests, testStruct{
						name: fmt.Sprintf("requires the manager to have can_grant_view_generated >= new value (%s -> %s, %s)",
							currentCanView, newCanViewValue, canGrantViewValue),
						args: args{
							viewPermissionToSet: newCanViewValue,
							currentPermissions:  &userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(currentCanView)},
							managerPermissions: &managerGeneratedPermissions{
								CanGrantViewGeneratedValue: permissionGrantedStore.GrantViewIndexByName(canGrantViewValue),
							},
						},
						want: permissionGrantedStore.ViewIndexByName(newCanViewValue) <= permissionGrantedStore.GrantViewIndexByName(canGrantViewValue),
					})
				}
			} else {
				tests = append(tests, testStruct{
					name: fmt.Sprintf("allows setting a lower value or the same value (%s -> %s)", currentCanView, newCanViewValue),
					args: args{
						viewPermissionToSet: newCanViewValue,
						currentPermissions:  &userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(currentCanView)},
					},
					want: true,
				})
			}
			if newCanViewValue == currentCanView {
				newValueIsGreater = true
			}
		}
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, checkIfPossibleToModifyCanView(
				tt.args.viewPermissionToSet, tt.args.currentPermissions, tt.args.managerPermissions, dataStore))
		})
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanGrantView_AllowsSettingLowerOrSameValue(t *testing.T) {
	testCheckerAllowsSettingLowerOrSameValue(t, canGrantViewValues[:],
		checkIfPossibleToModifyCanGrantView,
		func(value interface{}, permissionGrantedStore *database.PermissionGrantedStore) *userPermissions {
			return &userPermissions{CanGrantViewValue: permissionGrantedStore.GrantViewIndexByName(value.(string))}
		})
}

func testCheckerAllowsSettingLowerOrSameValue(
	t *testing.T, values []string,
	funcToCheck interface{}, currentPermissionsGenerator func(interface{}, *database.PermissionGrantedStore) *userPermissions,
) {
	t.Helper()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)

	for _, currentValue := range values {
		currentValue := currentValue
		want := true
		for _, newValue := range values {
			newValue := newValue
			t.Run(fmt.Sprintf("%s -> %s", currentValue, newValue), func(t *testing.T) {
				assert.Equal(t, want,
					reflect.ValueOf(funcToCheck).Call([]reflect.Value{
						reflect.ValueOf(newValue), reflect.ValueOf(currentPermissionsGenerator(currentValue, dataStore.PermissionsGranted())),
						reflect.ValueOf((*managerGeneratedPermissions)(nil)), reflect.ValueOf(dataStore),
					})[0].Interface())
			})
			if newValue == currentValue {
				want = false
			}
		}
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanGrantView_RequiresCanViewBeGreaterOrEqualToCanGrantView(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	for currentCanGrantViewIndex, currentCanGrantView := range canGrantViewValues {
		currentCanGrantView := currentCanGrantView
		for newCanGrantViewIndex := currentCanGrantViewIndex + 1; newCanGrantViewIndex < len(canGrantViewValues); newCanGrantViewIndex++ {
			newCanGrantViewValue := canGrantViewValues[newCanGrantViewIndex]
			if newCanGrantViewValue == solutionWithGrant {
				break
			}
			for _, currentCanViewValue := range canViewValues {
				currentCanViewValue := currentCanViewValue
				t.Run(fmt.Sprintf("%s -> %s, %s", currentCanGrantView, newCanGrantViewValue, currentCanViewValue),
					func(t *testing.T) {
						assert.Equal(t,
							permissionGrantedStore.GrantViewIndexByName(newCanGrantViewValue) <= permissionGrantedStore.ViewIndexByName(currentCanViewValue),
							checkIfPossibleToModifyCanGrantView(
								newCanGrantViewValue,
								&userPermissions{
									CanViewValue:      permissionGrantedStore.ViewIndexByName(currentCanViewValue),
									CanGrantViewValue: permissionGrantedStore.GrantViewIndexByName(currentCanGrantView),
								},
								&managerGeneratedPermissions{
									CanGrantViewGeneratedValue: permissionGrantedStore.GrantViewIndexByName(solutionWithGrant),
								}, dataStore))
					})
			}
		}
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanGrantView_SolutionWithGrant(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	assert.Equal(t, true, checkIfPossibleToModifyCanGrantView(
		solutionWithGrant,
		&userPermissions{
			CanGrantViewValue: permissionGrantedStore.GrantViewIndexByName(solution),
			CanViewValue:      permissionGrantedStore.ViewIndexByName(solution),
		},
		&managerGeneratedPermissions{
			IsOwnerGenerated:           true,
			CanGrantViewGeneratedValue: permissionGrantedStore.GrantViewIndexByName(solutionWithGrant),
		}, dataStore))

	assert.Equal(t, false, checkIfPossibleToModifyCanGrantView(
		solutionWithGrant,
		&userPermissions{
			CanGrantViewValue: permissionGrantedStore.GrantViewIndexByName(solution),
			CanViewValue:      permissionGrantedStore.ViewIndexByName(content),
		},
		&managerGeneratedPermissions{
			IsOwnerGenerated:           true,
			CanGrantViewGeneratedValue: permissionGrantedStore.GrantViewIndexByName(solutionWithGrant),
		}, dataStore))

	assert.Equal(t, false, checkIfPossibleToModifyCanGrantView(
		solutionWithGrant,
		&userPermissions{
			CanGrantViewValue: permissionGrantedStore.GrantViewIndexByName(solution),
			CanViewValue:      permissionGrantedStore.ViewIndexByName(solution),
		},
		&managerGeneratedPermissions{
			IsOwnerGenerated:           false,
			CanGrantViewGeneratedValue: permissionGrantedStore.GrantViewIndexByName(solutionWithGrant),
		}, dataStore))

	for _, canGrantView := range canGrantViewValues {
		canGrantView := canGrantView
		t.Run(canGrantView, func(t *testing.T) {
			assert.Equal(t, canGrantView == solutionWithGrant, checkIfPossibleToModifyCanGrantView(
				solutionWithGrant,
				&userPermissions{
					CanGrantViewValue: permissionGrantedStore.GrantViewIndexByName(solution),
					CanViewValue:      permissionGrantedStore.ViewIndexByName(solution),
				},
				&managerGeneratedPermissions{
					IsOwnerGenerated:           true,
					CanGrantViewGeneratedValue: permissionGrantedStore.GrantViewIndexByName(canGrantView),
				}, dataStore))
		})
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanGrantView_RequiresManagerToHaveSolutionWithGrantPermission(t *testing.T) {
	testCheckerRequiresManagerToHaveSpecificPermission(
		t, canGrantViewValues[:], solutionWithGrant, solution,
		checkIfPossibleToModifyCanGrantView,
		func(newValue, managerValue string, permissionGrantedStore *database.PermissionGrantedStore) *managerGeneratedPermissions {
			return &managerGeneratedPermissions{
				IsOwnerGenerated:           newValue == solutionWithGrant,
				CanGrantViewGeneratedValue: permissionGrantedStore.GrantViewIndexByName(managerValue),
			}
		})
}

func testCheckerRequiresManagerToHaveSpecificPermission(
	t *testing.T, values []string, requiredPermission, userCanViewValue string,
	funcToCheck func(string, *userPermissions, *managerGeneratedPermissions, *database.DataStore) bool,
	managerPermissionsGenerator func(
		newValue, managerValue string, permissionGrantedStore *database.PermissionGrantedStore) *managerGeneratedPermissions,
) {
	t.Helper()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	for _, newValue := range values {
		newValue := newValue
		for _, managerValue := range values {
			managerValue := managerValue
			t.Run(fmt.Sprintf("=> %s, %s", newValue, managerValue), func(t *testing.T) {
				assert.Equal(t, managerValue == requiredPermission,
					funcToCheck(
						newValue,
						&userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(userCanViewValue)},
						managerPermissionsGenerator(newValue, managerValue, permissionGrantedStore), dataStore))
			})
		}
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanWatch_AllowsSettingLowerOrSameValue(t *testing.T) {
	testCheckerAllowsSettingLowerOrSameValue(t, canWatchValues[:],
		checkIfPossibleToModifyCanWatch,
		func(value interface{}, permissionGrantedStore *database.PermissionGrantedStore) *userPermissions {
			return &userPermissions{CanWatchValue: permissionGrantedStore.WatchIndexByName(value.(string))}
		})
}

func Test_checkIfPossibleToModifyCanWatch_RequiresCanViewBeGreaterOrEqualToContent(t *testing.T) {
	testCheckerRequiresCanViewBeGreaterOrEqualToContent(t, canWatchValues[:], answerWithGrant,
		checkIfPossibleToModifyCanWatch,
		func(value interface{}, viewValue string, permissionGrantedStore *database.PermissionGrantedStore) *userPermissions {
			return &userPermissions{
				CanViewValue:  permissionGrantedStore.ViewIndexByName(viewValue),
				CanWatchValue: permissionGrantedStore.WatchIndexByName(value.(string)),
			}
		},
		func(permissionGrantedStore *database.PermissionGrantedStore) *managerGeneratedPermissions {
			return &managerGeneratedPermissions{
				CanWatchGeneratedValue: permissionGrantedStore.WatchIndexByName(answerWithGrant),
			}
		})
}

func testCheckerRequiresCanViewBeGreaterOrEqualToContent(
	t *testing.T, values []string, stopValue string, funcToCheck interface{},
	currentPermissionsGenerator func(
		value interface{}, viewValue string, permissionGrantedStore *database.PermissionGrantedStore) *userPermissions,
	managerPermissionsGenerator func(
		permissionGrantedStore *database.PermissionGrantedStore) *managerGeneratedPermissions,
) {
	t.Helper()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	for currentIndex, currentValue := range values {
		currentValue := currentValue
		for newIndex := currentIndex + 1; newIndex < len(values); newIndex++ {
			newValue := values[newIndex]
			if newValue == stopValue {
				break
			}
			for _, currentCanViewValue := range canViewValues {
				currentCanViewValue := currentCanViewValue
				t.Run(fmt.Sprintf("%s -> %s, %s",
					currentValue, newValue, currentCanViewValue), func(t *testing.T) {
					assert.Equal(t,
						permissionGrantedStore.ViewIndexByName(content) <= permissionGrantedStore.ViewIndexByName(currentCanViewValue),
						reflect.ValueOf(funcToCheck).Call([]reflect.Value{
							reflect.ValueOf(newValue),
							reflect.ValueOf(currentPermissionsGenerator(currentValue, currentCanViewValue, dataStore.PermissionsGranted())),
							reflect.ValueOf(managerPermissionsGenerator(permissionGrantedStore)), reflect.ValueOf(dataStore),
						})[0].Interface())
				})
			}
		}
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanWatch_AnswerWithGrant(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	for _, test := range []struct {
		canView          string
		isOwnerGenerated bool
		want             bool
	}{
		{"content", true, true},
		{"info", true, false},
		{"content", false, false},
	} {
		test := test
		t.Run(fmt.Sprintf("%s (can_view=%s, managerIsOwner=%v)", answerWithGrant, test.canView, test.isOwnerGenerated),
			func(t *testing.T) {
				assert.Equal(t, test.want, checkIfPossibleToModifyCanWatch(
					answerWithGrant,
					&userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(test.canView)},
					&managerGeneratedPermissions{
						IsOwnerGenerated:       test.isOwnerGenerated,
						CanWatchGeneratedValue: permissionGrantedStore.WatchIndexByName(answerWithGrant),
					}, dataStore))
			})
	}

	for _, canWatch := range canWatchValues {
		canWatch := canWatch
		t.Run(canWatch, func(t *testing.T) {
			assert.Equal(t, canWatch == answerWithGrant, checkIfPossibleToModifyCanWatch(
				answerWithGrant,
				&userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(content)},
				&managerGeneratedPermissions{
					IsOwnerGenerated:       true,
					CanWatchGeneratedValue: permissionGrantedStore.WatchIndexByName(canWatch),
				}, dataStore))
		})
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanWatch_RequiresManagerToHaveAnswerWithGrantPermission(t *testing.T) {
	testCheckerRequiresManagerToHaveSpecificPermission(
		t, canWatchValues[:], answerWithGrant, content,
		checkIfPossibleToModifyCanWatch,
		func(newValue, managerValue string, permissionGrantedStore *database.PermissionGrantedStore) *managerGeneratedPermissions {
			return &managerGeneratedPermissions{
				IsOwnerGenerated:       newValue == answerWithGrant,
				CanWatchGeneratedValue: permissionGrantedStore.WatchIndexByName(managerValue),
			}
		})
}

func Test_checkIfPossibleToModifyCanEdit_AllowsSettingLowerOrSameValue(t *testing.T) {
	testCheckerAllowsSettingLowerOrSameValue(t, canEditValues[:],
		checkIfPossibleToModifyCanEdit,
		func(value interface{}, permissionGrantedStore *database.PermissionGrantedStore) *userPermissions {
			return &userPermissions{CanEditValue: permissionGrantedStore.EditIndexByName(value.(string))}
		})
}

func Test_checkIfPossibleToModifyCanEdit_RequiresCanViewBeGreaterOrEqualToContent(t *testing.T) {
	testCheckerRequiresCanViewBeGreaterOrEqualToContent(t, canEditValues[:], allWithGrant,
		checkIfPossibleToModifyCanEdit,
		func(value interface{}, viewValue string, permissionGrantedStore *database.PermissionGrantedStore) *userPermissions {
			return &userPermissions{
				CanViewValue: permissionGrantedStore.ViewIndexByName(viewValue),
				CanEditValue: permissionGrantedStore.EditIndexByName(value.(string)),
			}
		},
		func(permissionGrantedStore *database.PermissionGrantedStore) *managerGeneratedPermissions {
			return &managerGeneratedPermissions{
				CanEditGeneratedValue: permissionGrantedStore.EditIndexByName(allWithGrant),
			}
		})
}

func Test_checkIfPossibleToModifyCanEdit_AllWithGrant(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	assert.Equal(t, true, checkIfPossibleToModifyCanEdit(
		allWithGrant,
		&userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(content)},
		&managerGeneratedPermissions{
			IsOwnerGenerated:      true,
			CanEditGeneratedValue: permissionGrantedStore.EditIndexByName(allWithGrant),
		}, dataStore))

	assert.Equal(t, false, checkIfPossibleToModifyCanEdit(
		allWithGrant,
		&userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(info)},
		&managerGeneratedPermissions{
			IsOwnerGenerated:      true,
			CanEditGeneratedValue: permissionGrantedStore.EditIndexByName(allWithGrant),
		}, dataStore))

	assert.Equal(t, false, checkIfPossibleToModifyCanEdit(
		allWithGrant,
		&userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(content)},
		&managerGeneratedPermissions{
			IsOwnerGenerated:      false,
			CanEditGeneratedValue: permissionGrantedStore.EditIndexByName(allWithGrant),
		}, dataStore))

	for _, canEdit := range canEditValues {
		canEdit := canEdit
		t.Run(canEdit, func(t *testing.T) {
			assert.Equal(t, canEdit == allWithGrant, checkIfPossibleToModifyCanEdit(
				allWithGrant,
				&userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(content)},
				&managerGeneratedPermissions{
					IsOwnerGenerated:      true,
					CanEditGeneratedValue: permissionGrantedStore.EditIndexByName(canEdit),
				}, dataStore))
		})
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanEdit_RequiresManagerToHaveAllWithGrantPermission(t *testing.T) {
	testCheckerRequiresManagerToHaveSpecificPermission(
		t, canEditValues[:], allWithGrant, content,
		checkIfPossibleToModifyCanEdit,
		func(newValue, managerValue string, permissionGrantedStore *database.PermissionGrantedStore) *managerGeneratedPermissions {
			return &managerGeneratedPermissions{
				IsOwnerGenerated:      newValue == allWithGrant,
				CanEditGeneratedValue: permissionGrantedStore.EditIndexByName(managerValue),
			}
		})
}

func Test_checkIfPossibleToModifyCanMakeSessionOfficial_AllowsSettingLowerOrSameValue(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)

	assert.Equal(t, true, checkIfPossibleToModifyCanMakeSessionOfficial(
		false, &userPermissions{CanMakeSessionOfficial: false}, &managerGeneratedPermissions{}, dataStore))
	assert.Equal(t, true, checkIfPossibleToModifyCanMakeSessionOfficial(
		false, &userPermissions{CanMakeSessionOfficial: true}, &managerGeneratedPermissions{}, dataStore))
	assert.Equal(t, true, checkIfPossibleToModifyCanMakeSessionOfficial(
		true, &userPermissions{CanMakeSessionOfficial: true}, &managerGeneratedPermissions{}, dataStore))
	assert.Equal(t, false, checkIfPossibleToModifyCanMakeSessionOfficial(
		true, &userPermissions{CanMakeSessionOfficial: false}, &managerGeneratedPermissions{}, dataStore))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanMakeSessionOfficial_RequiresCanViewBeGreaterOrEqualToInfo(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	for _, currentCanViewValue := range canViewValues {
		currentCanViewValue := currentCanViewValue
		t.Run(currentCanViewValue, func(t *testing.T) {
			assert.Equal(t,
				permissionGrantedStore.ViewIndexByName(info) <= permissionGrantedStore.ViewIndexByName(currentCanViewValue),
				checkIfPossibleToModifyCanMakeSessionOfficial(
					true,
					&userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(currentCanViewValue)},
					&managerGeneratedPermissions{IsOwnerGenerated: true}, dataStore))
		})
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanMakeSessionOfficial_RequiresManagerToBeOwner(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	assert.Equal(t, true, checkIfPossibleToModifyCanMakeSessionOfficial(
		true,
		&userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(info)},
		&managerGeneratedPermissions{IsOwnerGenerated: true}, dataStore))
	assert.Equal(t, false, checkIfPossibleToModifyCanMakeSessionOfficial(
		true,
		&userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(info)},
		&managerGeneratedPermissions{IsOwnerGenerated: false}, dataStore))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanEnterFrom_AllowsSettingGreaterOrSameValue(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)

	tm := time.Date(2019, 5, 30, 11, 0, 0, 1, time.UTC)
	tmPlus := tm.Add(time.Nanosecond)
	assert.Equal(t, true, checkIfPossibleToModifyCanEnterFrom(
		tmPlus, &userPermissions{CanEnterFrom: database.Time(tmPlus)}, &managerGeneratedPermissions{}, dataStore))
	assert.Equal(t, true, checkIfPossibleToModifyCanEnterFrom(
		tm, &userPermissions{CanEnterFrom: database.Time(tm)}, &managerGeneratedPermissions{}, dataStore))
	assert.Equal(t, false, checkIfPossibleToModifyCanEnterFrom(
		tm, &userPermissions{CanEnterFrom: database.Time(tmPlus)}, &managerGeneratedPermissions{}, dataStore))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanEnterFrom_RequiresManagerToHaveCanGrantViewGreaterOrEqualToEnter(t *testing.T) {
	tm := time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC)
	tmPlus := tm.Add(time.Nanosecond)
	testCheckerRequiresManagerToHaveCanGrantViewGreaterOrEqualToEnter(
		t, tm, checkIfPossibleToModifyCanEnterFrom,
		func() *userPermissions {
			return &userPermissions{CanEnterFrom: database.Time(tmPlus)}
		})
}

func testCheckerRequiresManagerToHaveCanGrantViewGreaterOrEqualToEnter(
	t *testing.T, value, funcToCheck interface{}, currentPermissionsGenerator func() *userPermissions,
) {
	t.Helper()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	for _, canGrantView := range canGrantViewValues {
		canGrantView := canGrantView
		t.Run(canGrantView, func(t *testing.T) {
			assert.Equal(t,
				permissionGrantedStore.GrantViewIndexByName(enter) <= permissionGrantedStore.GrantViewIndexByName(canGrantView),
				reflect.ValueOf(funcToCheck).Call([]reflect.Value{
					reflect.ValueOf(value), reflect.ValueOf(currentPermissionsGenerator()),
					reflect.ValueOf(&managerGeneratedPermissions{
						CanGrantViewGeneratedValue: permissionGrantedStore.GrantViewIndexByName(canGrantView),
					}), reflect.ValueOf(dataStore),
				})[0].Interface())
		})
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanEnterUntil_AllowsSettingSmallerOrSameValue(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)

	assert.Equal(t, true, checkIfPossibleToModifyCanEnterUntil(
		time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC),
		&userPermissions{CanEnterUntil: database.Time(time.Date(2019, 5, 30, 11, 0, 0, 1, time.UTC))},
		&managerGeneratedPermissions{}, dataStore))
	assert.Equal(t, true, checkIfPossibleToModifyCanEnterUntil(
		time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC),
		&userPermissions{CanEnterUntil: database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))},
		&managerGeneratedPermissions{}, dataStore))
	assert.Equal(t, false, checkIfPossibleToModifyCanEnterUntil(
		time.Date(2019, 5, 30, 11, 0, 0, 1, time.UTC),
		&userPermissions{CanEnterUntil: database.Time(time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC))},
		&managerGeneratedPermissions{}, dataStore))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_checkIfPossibleToModifyCanEnterUntil_RequiresManagerToHaveCanGrantViewGreaterOrEqualToEnter(t *testing.T) {
	tm := time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC)
	tmPlus := tm.Add(time.Nanosecond)
	testCheckerRequiresManagerToHaveCanGrantViewGreaterOrEqualToEnter(
		t, tmPlus, checkIfPossibleToModifyCanEnterUntil,
		func() *userPermissions {
			return &userPermissions{CanEnterFrom: database.Time(tm)}
		})
}

func Test_IsOwnerValidator_AllowsSettingIsOwnerToFalseOrSameValue(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	dataStore := database.NewDataStore(db)

	currentPermissions := &userPermissions{}
	dataMap, modified, err := parsePermissionsInputData(dataStore, &managerGeneratedPermissions{}, currentPermissions,
		&database.User{}, 0, map[string]interface{}{"is_owner": false})
	assert.Nil(t, err)
	assert.False(t, modified)
	assert.Equal(t, &userPermissions{}, currentPermissions)
	assert.Equal(t, map[string]interface{}{"is_owner": false}, dataMap)

	currentPermissions = &userPermissions{IsOwner: true}
	dataMap, modified, err = parsePermissionsInputData(dataStore, &managerGeneratedPermissions{},
		currentPermissions, &database.User{}, 0, map[string]interface{}{"is_owner": false})
	assert.Nil(t, err)
	assert.True(t, modified)
	assert.Equal(t, map[string]interface{}{"is_owner": false}, dataMap)

	currentPermissions = &userPermissions{IsOwner: true}
	dataMap, modified, err = parsePermissionsInputData(dataStore, &managerGeneratedPermissions{},
		currentPermissions, &database.User{}, 0, map[string]interface{}{"is_owner": true})
	assert.Nil(t, err)
	assert.False(t, modified)
	assert.Equal(t, map[string]interface{}{"is_owner": true}, dataMap)

	currentPermissions = &userPermissions{}
	_, _, err = parsePermissionsInputData(dataStore, &managerGeneratedPermissions{},
		currentPermissions, &database.User{}, 0, map[string]interface{}{"is_owner": true})
	require.IsType(t, (*service.APIError)(nil), err)
	var apiError *service.APIError
	assert.True(t, errors.As(err, &apiError))
	assert.Equal(t, http.StatusBadRequest, apiError.HTTPStatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_IsOwnerValidator_RequiresManagerToBeOwnerToMakeSomebodyAnOwner(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	dataStore := database.NewDataStore(db)

	currentPermissions := &userPermissions{}
	_, _, err := parsePermissionsInputData(dataStore,
		&managerGeneratedPermissions{IsOwnerGenerated: false}, currentPermissions,
		&database.User{}, 0, map[string]interface{}{"is_owner": true})
	require.IsType(t, (*service.APIError)(nil), err)
	var apiError *service.APIError
	assert.True(t, errors.As(err, &apiError))
	assert.Equal(t, http.StatusBadRequest, apiError.HTTPStatusCode)

	currentPermissions = &userPermissions{}
	dataMap, modified, err := parsePermissionsInputData(dataStore,
		&managerGeneratedPermissions{IsOwnerGenerated: true}, currentPermissions,
		&database.User{}, 0, map[string]interface{}{"is_owner": true})
	assert.Nil(t, err)
	assert.True(t, modified)
	assert.Equal(t, map[string]interface{}{"is_owner": true}, dataMap)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CanViewValidator_SetsModifiedFlagAndUpdatesCurrentPermissions(t *testing.T) {
	testValidatorSetsModifiedFlagAndUpdatesCurrentPermissions(t, canViewValues[:], "can_view", checkIfPossibleToModifyCanView,
		func(value interface{}, permissionGrantedStore *database.PermissionGrantedStore) *userPermissions {
			return &userPermissions{CanViewValue: permissionGrantedStore.ViewIndexByName(value.(string))}
		}, true)
}

func testValidatorSetsModifiedFlagAndUpdatesCurrentPermissions(
	t *testing.T, values interface{}, fieldName string, checkFunc interface{},
	currentPermissionsGenerator func(interface{}, *database.PermissionGrantedStore) *userPermissions,
	mockEnums bool,
) {
	t.Helper()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	if mockEnums {
		database.ClearAllDBEnums()
		database.MockDBEnumQueries(mock)
		defer database.ClearAllDBEnums()
	}
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	pg := monkey.Patch(checkFunc,
		reflect.MakeFunc(reflect.ValueOf(checkFunc).Type(), func([]reflect.Value) []reflect.Value {
			return []reflect.Value{reflect.ValueOf(true)}
		}).Interface())
	defer pg.Unpatch()

	reflValues := reflect.ValueOf(values)
	for currentIndex := 0; currentIndex < reflValues.Len(); currentIndex++ {
		currentValue := reflValues.Index(currentIndex).Interface()
		for newIndex := 0; newIndex < reflValues.Len(); newIndex++ {
			newValue := reflValues.Index(newIndex).Interface()
			t.Run(fmt.Sprintf("%s -> %s", currentValue, newValue), func(t *testing.T) {
				currentPermissions := currentPermissionsGenerator(currentValue, permissionGrantedStore)
				dataMap, modified, err := parsePermissionsInputData(dataStore,
					&managerGeneratedPermissions{}, currentPermissions, &database.User{}, 0,
					map[string]interface{}{fieldName: newValue})
				assert.Nil(t, err)
				assert.Equal(t, newValue != currentValue, modified)
				assert.Equal(t, map[string]interface{}{fieldName: newValue}, dataMap)
				assert.Equal(t, currentPermissionsGenerator(newValue, permissionGrantedStore), currentPermissions)
			})
		}
	}

	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CanViewValidator_FailsWhenCheckReturnsFalse(t *testing.T) {
	testValidatorFailsWhenCheckReturnsFalse(t, map[string]interface{}{"can_view": "info"}, checkIfPossibleToModifyCanView)
}

func testValidatorFailsWhenCheckReturnsFalse(t *testing.T, parsedBody map[string]interface{}, checkFunc interface{}) {
	t.Helper()

	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	dataStore := database.NewDataStore(db)

	pg := monkey.Patch(checkFunc,
		reflect.MakeFunc(reflect.ValueOf(checkFunc).Type(), func([]reflect.Value) []reflect.Value {
			return []reflect.Value{reflect.ValueOf(false)}
		}).Interface())
	defer pg.Unpatch()

	_, _, err := parsePermissionsInputData(dataStore,
		&managerGeneratedPermissions{}, &userPermissions{}, &database.User{}, 0, parsedBody)
	require.IsType(t, (*service.APIError)(nil), err)
	var apiError *service.APIError
	assert.True(t, errors.As(err, &apiError))
	assert.Equal(t, http.StatusBadRequest, apiError.HTTPStatusCode)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CanGrantViewValidator_SetsModifiedFlagAndUpdatesCurrentPermissions(t *testing.T) {
	testValidatorSetsModifiedFlagAndUpdatesCurrentPermissions(t, canGrantViewValues[:], "can_grant_view",
		checkIfPossibleToModifyCanGrantView,
		func(value interface{}, permissionGrantedStore *database.PermissionGrantedStore) *userPermissions {
			return &userPermissions{CanGrantViewValue: permissionGrantedStore.GrantViewIndexByName(value.(string))}
		}, true)
}

func Test_CanGrantViewValidator_FailsWhenCheckReturnsFalse(t *testing.T) {
	testValidatorFailsWhenCheckReturnsFalse(t, map[string]interface{}{"can_grant_view": "info"}, checkIfPossibleToModifyCanGrantView)
}

func Test_CanWatchValidator_SetsModifiedFlagAndUpdatesCurrentPermissions(t *testing.T) {
	testValidatorSetsModifiedFlagAndUpdatesCurrentPermissions(t, canWatchValues[:], "can_watch",
		checkIfPossibleToModifyCanWatch,
		func(value interface{}, permissionGrantedStore *database.PermissionGrantedStore) *userPermissions {
			return &userPermissions{CanWatchValue: permissionGrantedStore.WatchIndexByName(value.(string))}
		}, true)
}

func Test_CanWatchValidator_FailsWhenCheckReturnsFalse(t *testing.T) {
	testValidatorFailsWhenCheckReturnsFalse(t, map[string]interface{}{"can_watch": "result"}, checkIfPossibleToModifyCanWatch)
}

func Test_CanEditValidator_SetsModifiedFlagAndUpdatesCurrentPermissions(t *testing.T) {
	testValidatorSetsModifiedFlagAndUpdatesCurrentPermissions(t, canEditValues[:], "can_edit",
		checkIfPossibleToModifyCanEdit,
		func(value interface{}, permissionGrantedStore *database.PermissionGrantedStore) *userPermissions {
			return &userPermissions{CanEditValue: permissionGrantedStore.EditIndexByName(value.(string))}
		}, true)
}

func Test_CanEditValidator_FailsWhenCheckReturnsFalse(t *testing.T) {
	testValidatorFailsWhenCheckReturnsFalse(t, map[string]interface{}{"can_edit": "all"}, checkIfPossibleToModifyCanEdit)
}

func Test_CanMakeSessionOfficialValidator_SetsModifiedFlagAndUpdatesCurrentPermissions(t *testing.T) {
	testValidatorSetsModifiedFlagAndUpdatesCurrentPermissions(t, []bool{false, true}, "can_make_session_official",
		checkIfPossibleToModifyCanMakeSessionOfficial,
		func(value interface{}, _ *database.PermissionGrantedStore) *userPermissions {
			return &userPermissions{CanMakeSessionOfficial: value.(bool)}
		}, false)
}

func Test_CanMakeSessionOfficialValidator_FailsWhenCheckReturnsFalse(t *testing.T) {
	testValidatorFailsWhenCheckReturnsFalse(
		t, map[string]interface{}{"can_make_session_official": false}, checkIfPossibleToModifyCanMakeSessionOfficial)
}

func Test_CanEnterFromValidator_SetsModifiedFlag(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	dataStore := database.NewDataStore(db)

	pg := monkey.Patch(checkIfPossibleToModifyCanEnterFrom,
		func(time.Time, *userPermissions, *managerGeneratedPermissions, *database.DataStore) bool { return true })
	defer pg.Unpatch()

	tm := time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC)
	tmPlus := tm.Add(time.Second)
	currentPermissions := &userPermissions{CanEnterFrom: database.Time(tmPlus)}
	dataMap, modified, err := parsePermissionsInputData(dataStore,
		&managerGeneratedPermissions{}, currentPermissions, &database.User{}, 0,
		map[string]interface{}{"can_enter_from": "2019-05-30T11:00:00Z"})
	assert.Nil(t, err)
	assert.True(t, modified)
	assert.Equal(t, map[string]interface{}{"can_enter_from": tm}, dataMap)

	dataMap, modified, err = parsePermissionsInputData(dataStore,
		&managerGeneratedPermissions{}, currentPermissions, &database.User{}, 0,
		map[string]interface{}{"can_enter_from": "2019-05-30T11:00:01Z"})
	assert.Nil(t, err)
	assert.False(t, modified)
	assert.Equal(t, map[string]interface{}{"can_enter_from": tmPlus}, dataMap)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CanEnterFromValidator_FailsWhenCheckReturnsFalse(t *testing.T) {
	testValidatorFailsWhenCheckReturnsFalse(
		t, map[string]interface{}{"can_enter_from": "2019-05-30T11:00:00Z"}, checkIfPossibleToModifyCanEnterFrom)
}

func Test_CanEnterUntilValidator_SetsModifiedFlag(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	dataStore := database.NewDataStore(db)

	pg := monkey.Patch(checkIfPossibleToModifyCanEnterUntil,
		func(time.Time, *userPermissions, *managerGeneratedPermissions, *database.DataStore) bool { return true })
	defer pg.Unpatch()

	currentPermissions := &userPermissions{
		CanEnterUntil: database.Time(time.Date(2019, 5, 30, 11, 0, 1, 0, time.UTC)),
	}
	dataMap, modified, err := parsePermissionsInputData(dataStore,
		&managerGeneratedPermissions{}, currentPermissions, &database.User{}, 0,
		map[string]interface{}{"can_enter_until": "2019-05-30T11:00:00Z"})
	assert.Nil(t, err)
	assert.True(t, modified)
	assert.Equal(t, map[string]interface{}{
		"can_enter_until": time.Date(2019, 5, 30, 11, 0, 0, 0, time.UTC),
	}, dataMap)

	dataMap, modified, err = parsePermissionsInputData(dataStore,
		&managerGeneratedPermissions{}, currentPermissions, &database.User{}, 0,
		map[string]interface{}{"can_enter_until": "2019-05-30T11:00:01Z"})
	assert.Nil(t, err)
	assert.False(t, modified)
	assert.Equal(t, map[string]interface{}{
		"can_enter_until": time.Date(2019, 5, 30, 11, 0, 1, 0, time.UTC),
	}, dataMap)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func Test_CanEnterUntilValidator_FailsWhenCheckReturnsFalse(t *testing.T) {
	testValidatorFailsWhenCheckReturnsFalse(
		t, map[string]interface{}{"can_enter_until": "2019-05-30T11:00:00Z"}, checkIfPossibleToModifyCanEnterUntil)
}

func Test_parsePermissionsInputData_ChecksCanViewFirstAndUsesItsNewValue(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	currentPermissions := &userPermissions{}
	dataMap, modified, err := parsePermissionsInputData(dataStore,
		&managerGeneratedPermissions{
			IsOwnerGenerated:           true,
			CanGrantViewGeneratedValue: permissionGrantedStore.GrantViewIndexByName(solutionWithGrant),
			CanWatchGeneratedValue:     permissionGrantedStore.WatchIndexByName(answerWithGrant),
			CanEditGeneratedValue:      permissionGrantedStore.EditIndexByName(allWithGrant),
		}, currentPermissions, &database.User{}, 0,
		map[string]interface{}{
			"can_view":                  "solution",
			"can_grant_view":            "solution_with_grant",
			"can_watch":                 "answer_with_grant",
			"can_edit":                  "all_with_grant",
			"can_make_session_official": true,
		})
	assert.Nil(t, err)
	assert.True(t, modified)
	assert.Equal(t, &userPermissions{
		CanViewValue:           permissionGrantedStore.ViewIndexByName(solution),
		CanGrantViewValue:      permissionGrantedStore.GrantViewIndexByName(solutionWithGrant),
		CanWatchValue:          permissionGrantedStore.WatchIndexByName(answerWithGrant),
		CanEditValue:           permissionGrantedStore.EditIndexByName(allWithGrant),
		CanMakeSessionOfficial: true,
	}, currentPermissions)
	assert.Equal(t, map[string]interface{}{
		"can_view":                  "solution",
		"can_grant_view":            "solution_with_grant",
		"can_watch":                 "answer_with_grant",
		"can_edit":                  "all_with_grant",
		"can_make_session_official": true,
	}, dataMap)

	assert.NoError(t, mock.ExpectationsWereMet())
}

type correctPermissionsDataMapTest struct {
	canView         string
	value           interface{}
	expectedDataMap map[string]interface{}
}

func Test_correctPermissionsDataMap(t *testing.T) {
	db, mock := database.NewDBMock()
	defer func() { _ = db.Close() }()
	database.ClearAllDBEnums()
	database.MockDBEnumQueries(mock)
	defer database.ClearAllDBEnums()
	dataStore := database.NewDataStore(db)
	permissionGrantedStore := dataStore.PermissionsGranted()

	tests := []struct {
		permission          string
		userPermissionsFunc func(value interface{}, canView string) *userPermissions
		tests               []correctPermissionsDataMapTest
	}{
		{
			"can_grant_view", func(value interface{}, canView string) *userPermissions {
				return &userPermissions{
					CanViewValue:      permissionGrantedStore.ViewIndexByName(canView),
					CanGrantViewValue: permissionGrantedStore.GrantViewIndexByName(value.(string)),
				}
			},
			[]correctPermissionsDataMapTest{
				{"none", "none", map[string]interface{}{}},
				{"none", "enter", map[string]interface{}{"can_grant_view": "none"}},
				{"none", "content", map[string]interface{}{"can_grant_view": "none"}},
				{"none", "content_with_descendants", map[string]interface{}{"can_grant_view": "none"}},
				{"none", "solution", map[string]interface{}{"can_grant_view": "none"}},
				{"none", "solution_with_grant", map[string]interface{}{"can_grant_view": "none"}},
				{"info", "none", map[string]interface{}{}},
				{"info", "enter", map[string]interface{}{}},
				{"info", "content", map[string]interface{}{"can_grant_view": "enter"}},
				{"info", "content_with_descendants", map[string]interface{}{"can_grant_view": "enter"}},
				{"info", "solution", map[string]interface{}{"can_grant_view": "enter"}},
				{"info", "solution_with_grant", map[string]interface{}{"can_grant_view": "enter"}},
				{"content", "none", map[string]interface{}{}},
				{"content", "enter", map[string]interface{}{}},
				{"content", "content", map[string]interface{}{}},
				{"content", "content_with_descendants", map[string]interface{}{"can_grant_view": "content"}},
				{"content", "solution", map[string]interface{}{"can_grant_view": "content"}},
				{"content", "solution_with_grant", map[string]interface{}{"can_grant_view": "content"}},
				{"content_with_descendants", "none", map[string]interface{}{}},
				{"content_with_descendants", "enter", map[string]interface{}{}},
				{"content_with_descendants", "content", map[string]interface{}{}},
				{"content_with_descendants", "content_with_descendants", map[string]interface{}{}},
				{"content_with_descendants", "solution", map[string]interface{}{"can_grant_view": "content_with_descendants"}},
				{"content_with_descendants", "solution_with_grant", map[string]interface{}{"can_grant_view": "content_with_descendants"}},
				{"solution", "none", map[string]interface{}{}},
				{"solution", "enter", map[string]interface{}{}},
				{"solution", "content", map[string]interface{}{}},
				{"solution", "content_with_descendants", map[string]interface{}{}},
				{"solution", "solution", map[string]interface{}{}},
				{"solution", "solution_with_grant", map[string]interface{}{}},
			},
		},
		{"can_watch", func(value interface{}, canView string) *userPermissions {
			return &userPermissions{
				CanViewValue:  permissionGrantedStore.ViewIndexByName(canView),
				CanWatchValue: permissionGrantedStore.WatchIndexByName(value.(string)),
			}
		}, generateCorrectPermissionsDataMapTestsForWatchOrEdit("can_watch")},
		{"can_edit", func(value interface{}, canView string) *userPermissions {
			return &userPermissions{
				CanViewValue: permissionGrantedStore.ViewIndexByName(canView),
				CanEditValue: permissionGrantedStore.EditIndexByName(value.(string)),
			}
		}, generateCorrectPermissionsDataMapTestsForWatchOrEdit("can_edit")},
		{
			"can_make_session_official", func(value interface{}, canView string) *userPermissions {
				return &userPermissions{
					CanViewValue:           permissionGrantedStore.ViewIndexByName(canView),
					CanMakeSessionOfficial: value.(bool),
				}
			},
			[]correctPermissionsDataMapTest{
				{"none", false, map[string]interface{}{}},
				{"none", true, map[string]interface{}{"can_make_session_official": false}},
				{"info", false, map[string]interface{}{}},
				{"info", true, map[string]interface{}{}},
				{"content", false, map[string]interface{}{}},
				{"content", true, map[string]interface{}{}},
				{"content_with_descendants", false, map[string]interface{}{}},
				{"content_with_descendants", true, map[string]interface{}{}},
				{"solution", false, map[string]interface{}{}},
				{"solution", true, map[string]interface{}{}},
			},
		},
	}

	for _, ts := range tests {
		ts := ts
		t.Run(ts.permission, func(t *testing.T) {
			for _, tt := range ts.tests {
				tt := tt
				t.Run(fmt.Sprintf("%s=%v, can_view=%s", ts.permission, tt.value, tt.canView), func(t *testing.T) {
					dataMap := make(map[string]interface{})
					correctPermissionsDataMap(dataStore, dataMap, ts.userPermissionsFunc(tt.value, tt.canView))
					assert.Equal(t, tt.expectedDataMap, dataMap)
				})
			}
		})
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func generateCorrectPermissionsDataMapTestsForWatchOrEdit(permission string) []correctPermissionsDataMapTest {
	var value2, value3, value4 string
	switch permission {
	case "can_watch":
		value2 = "result"
		value3 = "answer"
		value4 = "answer_with_grant"
	case "can_edit":
		value2 = "children"
		value3 = "all"
		value4 = "all_with_grant"
	}
	return []correctPermissionsDataMapTest{
		{"none", "none", map[string]interface{}{}},
		{"none", value2, map[string]interface{}{permission: "none"}},
		{"none", value3, map[string]interface{}{permission: "none"}},
		{"none", value4, map[string]interface{}{permission: "none"}},
		{"info", "none", map[string]interface{}{}},
		{"info", value2, map[string]interface{}{permission: "none"}},
		{"info", value3, map[string]interface{}{permission: "none"}},
		{"info", value4, map[string]interface{}{permission: "none"}},
		{"content", "none", map[string]interface{}{}},
		{"content", value2, map[string]interface{}{}},
		{"content", value3, map[string]interface{}{}},
		{"content", value4, map[string]interface{}{}},
		{"content_with_descendants", "none", map[string]interface{}{}},
		{"content_with_descendants", value2, map[string]interface{}{}},
		{"content_with_descendants", value3, map[string]interface{}{}},
		{"content_with_descendants", value4, map[string]interface{}{}},
		{"solution", "none", map[string]interface{}{}},
		{"solution", value2, map[string]interface{}{}},
		{"solution", value3, map[string]interface{}{}},
		{"solution", value4, map[string]interface{}{}},
	}
}
