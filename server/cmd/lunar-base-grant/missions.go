// Mission actions for the lunar-base-grant shim:
//   - set_missions  set/clear one or many user_missions rows inside a single
//                   UpdateUser transaction (Mission Editor)
//
// user_missions is a plain status table, but it is still written through
// lunar-tear's own store (LoadUser -> mutate -> diffAndSave) rather than with
// raw SQL: the game client reads IUserMission rows with
// (missionId, startDatetime, progressValue, missionProgressStatusType,
// clearDatetime, latestVersion) and only understands rows the server itself
// can produce. Writing SQL directly (as an earlier version of the editor did)
// can persist states the server never writes — e.g. a row whose status is 0
// ("Unknown") — and it races a running server's WAL. Going through UpdateUser
// keeps every write on the same code path the game server uses.

package main

import (
	"errors"
	"fmt"
	"time"

	"lunar-tear/server/internal/model"
	"lunar-tear/server/internal/store"
)

// missionSpec is one mission's target state for the `set_missions` action.
type missionSpec struct {
	MissionID int32 `json:"mission_id"`
	Status    int32 `json:"status"`
	Progress  int32 `json:"progress"`
}

// runSetMissions applies a batch of mission status/progress changes in one
// UpdateUser transaction. Status 0 ("not started") is persisted by *deleting*
// the row: the server never stores status 0, and "no row" is the canonical
// not-started state the client already handles for every mission.
func runSetMissions(req *request) (int, error) {
	if len(req.Missions) == 0 {
		return 0, errors.New("missions list is empty")
	}
	for _, m := range req.Missions {
		if m.MissionID <= 0 {
			return 0, fmt.Errorf("invalid mission_id %d", m.MissionID)
		}
		switch model.MissionProgressStatusType(m.Status) {
		case model.MissionProgressStatusTypeUnknown,
			model.MissionProgressStatusTypeInProgress,
			model.MissionProgressStatusTypeClear,
			model.MissionProgressStatusTypeRewardReceived:
		default:
			return 0, fmt.Errorf("invalid status %d for mission %d", m.Status, m.MissionID)
		}
		if m.Progress < 0 {
			return 0, fmt.Errorf("negative progress for mission %d", m.MissionID)
		}
	}

	db, st, err := openDB(req.DBPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	now := time.Now().UnixMilli()
	applied := 0
	_, err = st.UpdateUser(req.UserID, func(u *store.UserState) {
		u.EnsureMaps()
		// Heal saves written by the old raw-SQL editor: status 0 is the
		// server's "Unknown" sentinel and is never a persisted state, so
		// purge any leftover rows carrying it before applying this batch.
		for id, row := range u.Missions {
			if row.MissionProgressStatusType == int32(model.MissionProgressStatusTypeUnknown) {
				delete(u.Missions, id)
			}
		}
		for _, m := range req.Missions {
			if m.Status == int32(model.MissionProgressStatusTypeUnknown) {
				if _, ok := u.Missions[m.MissionID]; ok {
					delete(u.Missions, m.MissionID)
					applied++
				}
				continue
			}
			row, exists := u.Missions[m.MissionID]
			row.MissionId = m.MissionID
			if !exists {
				// New rows start now; existing rows keep their original
				// start_datetime (same as the server's own transitions).
				row.StartDatetime = now
			}
			row.ProgressValue = m.Progress
			row.MissionProgressStatusType = m.Status
			if m.Status >= int32(model.MissionProgressStatusTypeClear) {
				row.ClearDatetime = now
			} else {
				row.ClearDatetime = 0
			}
			row.LatestVersion = now
			u.Missions[m.MissionID] = row
			applied++
		}
	})
	if err != nil {
		return 0, fmt.Errorf("set missions: %w", err)
	}
	return applied, nil
}
