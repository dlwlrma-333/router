package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/yeying-community/router/common/helper"
	"gorm.io/gorm"
)

const GroupQuotaDailyCountersTableName = "group_quota_daily_counters"

type GroupQuotaDailyCounter struct {
	GroupID       string `json:"group_id" gorm:"primaryKey;type:char(36)"`
	BizDate       string `json:"biz_date" gorm:"primaryKey;type:varchar(10)"`
	ReservedQuota int64  `json:"reserved_quota" gorm:"type:bigint;not null;default:0"`
	ConsumedQuota int64  `json:"consumed_quota" gorm:"type:bigint;not null;default:0"`
	UpdatedAt     int64  `json:"updated_at" gorm:"bigint;index"`
}

func (GroupQuotaDailyCounter) TableName() string {
	return GroupQuotaDailyCountersTableName
}

type GroupDailyQuotaReservation struct {
	GroupID       string
	BizDate       string
	ReservedQuota int64
}

func (reservation GroupDailyQuotaReservation) Active() bool {
	return strings.TrimSpace(reservation.GroupID) != "" &&
		strings.TrimSpace(reservation.BizDate) != "" &&
		reservation.ReservedQuota > 0
}

func businessDateByTimezone(now time.Time, timezone string) string {
	locationName := normalizeGroupQuotaResetTimezone(timezone)
	location, err := time.LoadLocation(locationName)
	if err != nil {
		location = time.FixedZone(DefaultGroupQuotaResetTimezone, 8*3600)
	}
	return now.In(location).Format("2006-01-02")
}

func ReserveGroupDailyQuotaWithDB(db *gorm.DB, groupID string, quota int64) (GroupDailyQuotaReservation, bool, error) {
	if db == nil {
		return GroupDailyQuotaReservation{}, false, fmt.Errorf("database handle is nil")
	}
	normalizedGroupID := strings.TrimSpace(groupID)
	normalizedQuota := normalizeGroupDailyQuotaLimit(quota)
	if normalizedGroupID == "" || normalizedQuota <= 0 {
		return GroupDailyQuotaReservation{}, true, nil
	}
	policy := GetGroupDailyQuotaPolicy(normalizedGroupID)
	if policy.Limit <= 0 {
		return GroupDailyQuotaReservation{}, true, nil
	}
	now := time.Now()
	bizDate := businessDateByTimezone(now, policy.Timezone)
	updatedAt := helper.GetTimestamp()
	result := db.Exec(
		`INSERT INTO group_quota_daily_counters (group_id, biz_date, reserved_quota, consumed_quota, updated_at)
		 VALUES (?, ?, ?, 0, ?)
		 ON CONFLICT (group_id, biz_date)
		 DO UPDATE
		 SET reserved_quota = group_quota_daily_counters.reserved_quota + EXCLUDED.reserved_quota,
		     updated_at = EXCLUDED.updated_at
		 WHERE (group_quota_daily_counters.consumed_quota + group_quota_daily_counters.reserved_quota + EXCLUDED.reserved_quota) <= ?`,
		normalizedGroupID,
		bizDate,
		normalizedQuota,
		updatedAt,
		policy.Limit,
	)
	if result.Error != nil {
		return GroupDailyQuotaReservation{}, false, result.Error
	}
	if result.RowsAffected == 0 {
		return GroupDailyQuotaReservation{}, false, nil
	}
	return GroupDailyQuotaReservation{
		GroupID:       normalizedGroupID,
		BizDate:       bizDate,
		ReservedQuota: normalizedQuota,
	}, true, nil
}

func ReserveGroupDailyQuota(groupID string, quota int64) (GroupDailyQuotaReservation, bool, error) {
	return ReserveGroupDailyQuotaWithDB(DB, groupID, quota)
}

func ReleaseGroupDailyQuotaReservationWithDB(db *gorm.DB, reservation GroupDailyQuotaReservation) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	if !reservation.Active() {
		return nil
	}
	result := db.Exec(
		`UPDATE group_quota_daily_counters
		 SET reserved_quota = GREATEST(reserved_quota - ?, 0),
		     updated_at = ?
		 WHERE group_id = ? AND biz_date = ?`,
		reservation.ReservedQuota,
		helper.GetTimestamp(),
		reservation.GroupID,
		reservation.BizDate,
	)
	return result.Error
}

func ReleaseGroupDailyQuotaReservation(reservation GroupDailyQuotaReservation) error {
	return ReleaseGroupDailyQuotaReservationWithDB(DB, reservation)
}

func SettleGroupDailyQuotaReservationWithDB(db *gorm.DB, reservation GroupDailyQuotaReservation, consumedQuota int64) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	if !reservation.Active() {
		return nil
	}
	consumed := consumedQuota
	if consumed < 0 {
		consumed = 0
	}
	now := helper.GetTimestamp()
	result := db.Exec(
		`INSERT INTO group_quota_daily_counters (group_id, biz_date, reserved_quota, consumed_quota, updated_at)
		 VALUES (?, ?, 0, ?, ?)
		 ON CONFLICT (group_id, biz_date)
		 DO UPDATE
		 SET reserved_quota = GREATEST(group_quota_daily_counters.reserved_quota - ?, 0),
		     consumed_quota = group_quota_daily_counters.consumed_quota + EXCLUDED.consumed_quota,
		     updated_at = EXCLUDED.updated_at`,
		reservation.GroupID,
		reservation.BizDate,
		consumed,
		now,
		reservation.ReservedQuota,
	)
	return result.Error
}

func SettleGroupDailyQuotaReservation(reservation GroupDailyQuotaReservation, consumedQuota int64) error {
	return SettleGroupDailyQuotaReservationWithDB(DB, reservation, consumedQuota)
}
