package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"gorm.io/gorm"
)

var (
	groupBillingRatioLock sync.RWMutex
	groupBillingRatioMap  = map[string]float64{}
)

func normalizeGroupBillingRatio(value float64) float64 {
	if value < 0 {
		return 1
	}
	return value
}

func buildGroupBillingRatioMap(rows []GroupCatalog) map[string]float64 {
	ratios := make(map[string]float64, len(rows))
	for _, row := range rows {
		groupID := strings.TrimSpace(row.Id)
		if groupID == "" {
			continue
		}
		ratios[groupID] = normalizeGroupBillingRatio(row.BillingRatio)
	}
	return ratios
}

func setGroupBillingRatioRuntime(ratios map[string]float64) {
	groupBillingRatioLock.Lock()
	groupBillingRatioMap = ratios
	groupBillingRatioLock.Unlock()
}

func GetGroupBillingRatio(id string) float64 {
	groupID := strings.TrimSpace(id)
	if groupID == "" {
		return 1
	}
	groupBillingRatioLock.RLock()
	ratio, ok := groupBillingRatioMap[groupID]
	groupBillingRatioLock.RUnlock()
	if !ok {
		return 1
	}
	return normalizeGroupBillingRatio(ratio)
}

func syncGroupBillingRatiosRuntimeWithDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	rows, err := listGroupCatalogWithDB(db)
	if err != nil {
		return err
	}
	setGroupBillingRatioRuntime(buildGroupBillingRatioMap(rows))
	return nil
}

func syncGroupBillingRatiosFromJSONWithDB(db *gorm.DB, raw string) error {
	if db == nil {
		return nil
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		trimmed = "{}"
	}
	ratios := make(map[string]float64)
	if err := json.Unmarshal([]byte(trimmed), &ratios); err != nil {
		return err
	}
	rows, err := listGroupCatalogWithDB(db)
	if err != nil {
		return err
	}
	for _, row := range rows {
		groupID := strings.TrimSpace(row.Id)
		if groupID == "" {
			continue
		}
		nextRatio := 1.0
		value, ok := ratios[groupID]
		if !ok {
			value, ok = ratios[strings.TrimSpace(row.Name)]
		}
		if ok {
			if value < 0 {
				return fmt.Errorf("group %s billing ratio cannot be negative", groupID)
			}
			nextRatio = value
		}
		if row.BillingRatio == nextRatio {
			continue
		}
		if err := db.Model(&GroupCatalog{}).
			Where("id = ?", row.Id).
			Update("billing_ratio", nextRatio).Error; err != nil {
			return err
		}
	}
	return nil
}

func backfillGroupBillingRatiosFromLegacyOptionWithDB(db *gorm.DB) error {
	if db == nil || !db.Migrator().HasTable(&Option{}) {
		return nil
	}
	option := Option{}
	if err := db.Where("key = ?", "GroupRatio").First(&option).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	return syncGroupBillingRatiosFromJSONWithDB(db, option.Value)
}
