package main

import (
	"errors"
	"fmt"
	"time"

	"lunar-tear/server/internal/masterdata"
	"lunar-tear/server/internal/masterdata/memorydb"
	"lunar-tear/server/internal/store"
	"lunar-tear/server/internal/utils"
)

// playerExpCurve loads the cumulative player-exp thresholds (index = level)
// from the encrypted master data. Level and exp are a single pair — the game
// re-derives the level from exp with LevelAndCap on every quest — so the shim
// derives the missing side instead of ever storing a mismatched pair.
func playerExpCurve(masterDataPath string) ([]int32, error) {
	if masterDataPath == "" {
		return nil, errors.New("master_data_path required to edit level/exp")
	}
	if err := memorydb.Init(masterDataPath); err != nil {
		return nil, fmt.Errorf("init master data: %w", err)
	}
	rows, err := utils.ReadTable[masterdata.EntityMNumericalParameterMap]("m_numerical_parameter_map")
	if err != nil {
		return nil, fmt.Errorf("load exp curve: %w", err)
	}
	curve := masterdata.BuildExpThresholds(rows, 1)
	if len(curve) < 2 {
		return nil, errors.New("player exp curve is empty")
	}
	return curve, nil
}

// levelFromExp mirrors gameutil.LevelAndCap for the level side only.
func levelFromExp(curve []int32, exp int32) int32 {
	level := int32(1)
	for i := 1; i < len(curve); i++ {
		if exp >= curve[i] {
			level = int32(i)
		} else {
			break
		}
	}
	return level
}

// runSetUserInfo updates the account-level fields the Profile page can edit:
// display name, message, level, exp and paid/free gems. Only the fields that
// are actually present in the request change (nil = leave untouched); the whole
// batch runs in one UpdateUser transaction so lunar-tear's own save path
// persists it with the usual version bookkeeping.
func runSetUserInfo(req *request) (int, error) {
	if req.UserID <= 0 {
		return 0, errors.New("user_id must be positive")
	}
	if req.Name == nil && req.Message == nil && req.Level == nil && req.Exp == nil &&
		req.PaidGem == nil && req.FreeGem == nil {
		return 0, errors.New("no user-info fields to update")
	}
	for name, v := range map[string]*int32{
		"level": req.Level, "exp": req.Exp, "paid_gem": req.PaidGem, "free_gem": req.FreeGem,
	} {
		if v != nil && *v < 0 {
			return 0, fmt.Errorf("%s must not be negative", name)
		}
	}

	db, st, err := openDB(req.DBPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Level/exp are validated before opening the transaction so a bad pair can
	// never be written.
	var newLevel, newExp int32
	hasLevelExp := req.Level != nil || req.Exp != nil
	if hasLevelExp {
		curve, curveErr := playerExpCurve(req.MasterDataPath)
		if curveErr != nil {
			return 0, curveErr
		}
		switch {
		case req.Exp != nil:
			// Experience is authoritative: derive the level from it.
			newExp = *req.Exp
			if newExp > curve[len(curve)-1] {
				newExp = curve[len(curve)-1]
			}
			newLevel = levelFromExp(curve, newExp)
		default:
			newLevel = *req.Level
			if newLevel < 1 {
				newLevel = 1
			}
			if max := int32(len(curve) - 1); newLevel > max {
				return 0, fmt.Errorf("level must be between 1 and %d", max)
			}
			newExp = curve[newLevel]
		}
	}

	now := time.Now().UnixMilli()
	applied := 0
	_, err = st.UpdateUser(req.UserID, func(u *store.UserState) {
		if req.Name != nil {
			u.Profile.Name = *req.Name
			u.Profile.NameUpdateDatetime = now
			applied++
		}
		if req.Message != nil {
			u.Profile.Message = *req.Message
			u.Profile.MessageUpdateDatetime = now
			applied++
		}
		if req.PaidGem != nil {
			u.Gem.PaidGem = *req.PaidGem
			applied++
		}
		if req.FreeGem != nil {
			u.Gem.FreeGem = *req.FreeGem
			applied++
		}
		if hasLevelExp {
			// Always write the pair together, and bump the row version so any
			// version-based client sync picks the edit up on next login.
			u.Status.Level = newLevel
			u.Status.Exp = newExp
			u.Status.LatestVersion = now
			applied++
		}
	})
	if err != nil {
		return 0, fmt.Errorf("set user info: %w", err)
	}
	return applied, nil
}
