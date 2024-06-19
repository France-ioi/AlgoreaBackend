package items

import (
	"github.com/France-ioi/AlgoreaBackend/app/database"
)

// RawWatchedGroupStatFields represents DB data fields for watched group stats used by itemNavigationView & itemChildrenView.
type RawWatchedGroupStatFields struct {
	CanWatchForGroupResults  bool
	WatchedGroupCanView      int
	WatchedGroupAvgScore     float32
	WatchedGroupAllValidated bool
}

func (stat *RawWatchedGroupStatFields) asItemWatchedGroupStat(
	watchedGroupIDIsSet bool, permissionGrantedStore *database.PermissionGrantedStore,
) *itemWatchedGroupStat {
	if !watchedGroupIDIsSet {
		return nil
	}
	result := &itemWatchedGroupStat{
		CanView: permissionGrantedStore.ViewNameByIndex(stat.WatchedGroupCanView),
	}
	if stat.CanWatchForGroupResults {
		result.AvgScore = &stat.WatchedGroupAvgScore
		result.AllValidated = &stat.WatchedGroupAllValidated
	}
	return result
}
