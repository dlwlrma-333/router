package model

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

const (
	DefaultGroupQuotaResetTimezone = "Asia/Shanghai"
)

type GroupDailyQuotaPolicy struct {
	Limit    int64
	Timezone string
}

var (
	groupDailyQuotaPolicyLock sync.RWMutex
	groupDailyQuotaPolicyMap  = map[string]GroupDailyQuotaPolicy{}
)

func normalizeGroupDailyQuotaLimit(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}

func normalizeGroupQuotaResetTimezone(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return DefaultGroupQuotaResetTimezone
	}
	if _, err := time.LoadLocation(normalized); err != nil {
		return DefaultGroupQuotaResetTimezone
	}
	return normalized
}

func ValidateGroupQuotaResetTimezone(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return DefaultGroupQuotaResetTimezone, nil
	}
	if _, err := time.LoadLocation(normalized); err != nil {
		return "", fmt.Errorf("重置时区不合法")
	}
	return normalized, nil
}

func buildGroupDailyQuotaPolicyMap(rows []GroupCatalog) map[string]GroupDailyQuotaPolicy {
	policies := make(map[string]GroupDailyQuotaPolicy, len(rows))
	for _, row := range rows {
		groupID := strings.TrimSpace(row.Id)
		if groupID == "" {
			continue
		}
		policies[groupID] = GroupDailyQuotaPolicy{
			Limit:    normalizeGroupDailyQuotaLimit(row.DailyQuotaLimit),
			Timezone: normalizeGroupQuotaResetTimezone(row.QuotaResetTimezone),
		}
	}
	return policies
}

func setGroupDailyQuotaPoliciesRuntime(policies map[string]GroupDailyQuotaPolicy) {
	groupDailyQuotaPolicyLock.Lock()
	groupDailyQuotaPolicyMap = policies
	groupDailyQuotaPolicyLock.Unlock()
}

func GetGroupDailyQuotaPolicy(id string) GroupDailyQuotaPolicy {
	groupID := strings.TrimSpace(id)
	if groupID == "" {
		return GroupDailyQuotaPolicy{
			Limit:    0,
			Timezone: DefaultGroupQuotaResetTimezone,
		}
	}
	groupDailyQuotaPolicyLock.RLock()
	policy, ok := groupDailyQuotaPolicyMap[groupID]
	groupDailyQuotaPolicyLock.RUnlock()
	if !ok {
		return GroupDailyQuotaPolicy{
			Limit:    0,
			Timezone: DefaultGroupQuotaResetTimezone,
		}
	}
	return GroupDailyQuotaPolicy{
		Limit:    normalizeGroupDailyQuotaLimit(policy.Limit),
		Timezone: normalizeGroupQuotaResetTimezone(policy.Timezone),
	}
}

func syncGroupDailyQuotaPoliciesRuntimeWithDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	rows, err := listGroupCatalogWithDB(db)
	if err != nil {
		return err
	}
	setGroupDailyQuotaPoliciesRuntime(buildGroupDailyQuotaPolicyMap(rows))
	return nil
}

func syncGroupRuntimeCachesWithDB(db *gorm.DB) error {
	if err := syncGroupBillingRatiosRuntimeWithDB(db); err != nil {
		return err
	}
	return syncGroupDailyQuotaPoliciesRuntimeWithDB(db)
}
