package items

import (
	"time"

	"github.com/France-ioi/AlgoreaBackend/app/database"
	"github.com/France-ioi/AlgoreaBackend/app/structures"
)

// RawItemResultFields represents DB data fields for item results used by itemNavigationView & itemChildrenView
type RawItemResultFields struct {
	// results
	AttemptID        *int64
	ScoreComputed    float32
	Validated        bool
	StartedAt        *database.Time
	LatestActivityAt database.Time
	EndedAt          *database.Time

	// attempts
	AttemptAllowsSubmissionsUntil database.Time
}

func (raw *RawItemResultFields) asItemResult() *structures.ItemResult {
	if raw.AttemptID == nil {
		return nil
	}
	return &structures.ItemResult{
		AttemptID:                     *raw.AttemptID,
		ScoreComputed:                 raw.ScoreComputed,
		Validated:                     raw.Validated,
		StartedAt:                     (*time.Time)(raw.StartedAt),
		LatestActivityAt:              time.Time(raw.LatestActivityAt),
		EndedAt:                       (*time.Time)(raw.EndedAt),
		AttemptAllowsSubmissionsUntil: time.Time(raw.AttemptAllowsSubmissionsUntil),
	}
}
