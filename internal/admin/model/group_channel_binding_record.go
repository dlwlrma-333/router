package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yeying-community/router/common/helper"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	GroupChannelBindingsTableName = "group_channel_bindings"
)

type GroupChannelBinding struct {
	Group     string `json:"group" gorm:"column:group;primaryKey;type:varchar(32);autoIncrement:false"`
	ChannelId string `json:"channel_id" gorm:"primaryKey;type:varchar(64);autoIncrement:false;index"`
	Enabled   bool   `json:"enabled" gorm:"not null;default:true;index"`
	Priority  int64  `json:"priority" gorm:"bigint;not null;default:0;index"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt int64  `json:"updated_at" gorm:"bigint;index"`
}

func (GroupChannelBinding) TableName() string {
	return GroupChannelBindingsTableName
}

func listGroupChannelBindingRowsWithDB(db *gorm.DB, groupID string) ([]GroupChannelBinding, error) {
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	groupCatalog, err := resolveGroupCatalogByReferenceWithDB(db, groupID)
	if err != nil {
		return nil, err
	}
	groupCol := `"group"`
	rows := make([]GroupChannelBinding, 0)
	if err := db.
		Where(groupCol+" = ?", groupCatalog.Id).
		Order("priority desc, channel_id asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func listGroupBoundChannelIDsWithDB(db *gorm.DB, groupID string) ([]string, error) {
	rows, err := listGroupChannelBindingRowsWithDB(db, groupID)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		if !row.Enabled {
			continue
		}
		channelID := strings.TrimSpace(row.ChannelId)
		if channelID == "" {
			continue
		}
		result = append(result, channelID)
	}
	return normalizeChannelIDList(result), nil
}

func listGroupChannelBindingPriorityByChannelWithDB(db *gorm.DB, groupID string) (map[string]*int64, error) {
	result := make(map[string]*int64)
	rows, err := listGroupChannelBindingRowsWithDB(db, groupID)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		channelID := strings.TrimSpace(row.ChannelId)
		if channelID == "" || !row.Enabled {
			continue
		}
		priority := row.Priority
		result[channelID] = &priority
	}
	return result, nil
}

func ListGroupChannelBindingPriorityByChannelWithDB(db *gorm.DB, groupID string) (map[string]*int64, error) {
	return listGroupChannelBindingPriorityByChannelWithDB(db, groupID)
}

func replaceGroupChannelBindingRowsWithItemsDB(db *gorm.DB, groupID string, items []GroupChannelBindingItem) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	groupCatalog, err := resolveGroupCatalogByReferenceWithDB(db, groupID)
	if err != nil {
		return err
	}
	groupID = groupCatalog.Id

	normalizedItems := normalizeGroupChannelBindingItems(items)
	now := helper.GetTimestamp()
	channelIDs := make([]string, 0, len(normalizedItems))
	for _, item := range normalizedItems {
		if !item.Bound {
			continue
		}
		channelIDs = append(channelIDs, item.Id)
	}
	channelIDs = normalizeChannelIDList(channelIDs)
	channelsByID, err := loadChannelsByIDWithDB(db, channelIDs)
	if err != nil {
		return err
	}
	existingRows, err := listGroupChannelBindingRowsWithDB(db, groupID)
	if err != nil {
		return err
	}
	existingByChannelID := make(map[string]GroupChannelBinding, len(existingRows))
	for _, row := range existingRows {
		channelID := strings.TrimSpace(row.ChannelId)
		if channelID == "" {
			continue
		}
		existingByChannelID[channelID] = row
	}

	rows := make([]GroupChannelBinding, 0, len(channelIDs))
	for _, item := range normalizedItems {
		if !item.Bound {
			continue
		}
		channel, ok := channelsByID[item.Id]
		if !ok {
			return fmt.Errorf("渠道不存在: %s", item.Id)
		}
		if channel.Status != ChannelStatusEnabled {
			return fmt.Errorf("渠道未启用: %s", item.Id)
		}
		existing, hasExisting := existingByChannelID[item.Id]
		priority := resolveGroupChannelBindingPriority(true, item.Priority, channel.Priority)
		createdAt := now
		if hasExisting && existing.CreatedAt > 0 {
			createdAt = existing.CreatedAt
		}
		rows = append(rows, GroupChannelBinding{
			Group:     groupID,
			ChannelId: item.Id,
			Enabled:   true,
			Priority:  toSafeGroupChannelBindingPriority(priority),
			CreatedAt: createdAt,
			UpdatedAt: now,
		})
	}

	groupCol := `"group"`
	if err := db.Where(groupCol+" = ?", groupID).Delete(&GroupChannelBinding{}).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	return db.Create(&rows).Error
}

func syncGroupChannelBindingRowsByChannelIDsDB(db *gorm.DB, groupID string, channelIDs []string) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	groupCatalog, err := resolveGroupCatalogByReferenceWithDB(db, groupID)
	if err != nil {
		return err
	}
	groupID = groupCatalog.Id

	normalizedChannelIDs := normalizeChannelIDList(channelIDs)
	if len(normalizedChannelIDs) == 0 {
		groupCol := `"group"`
		return db.Where(groupCol+" = ?", groupID).Delete(&GroupChannelBinding{}).Error
	}
	existingRows, err := listGroupChannelBindingRowsWithDB(db, groupID)
	if err != nil {
		return err
	}
	existingByChannelID := make(map[string]GroupChannelBinding, len(existingRows))
	for _, row := range existingRows {
		channelID := strings.TrimSpace(row.ChannelId)
		if channelID == "" {
			continue
		}
		existingByChannelID[channelID] = row
	}
	channelsByID, err := loadChannelsByIDWithDB(db, normalizedChannelIDs)
	if err != nil {
		return err
	}
	now := helper.GetTimestamp()
	rows := make([]GroupChannelBinding, 0, len(normalizedChannelIDs))
	for _, channelID := range normalizedChannelIDs {
		channel, ok := channelsByID[channelID]
		if !ok {
			return fmt.Errorf("渠道不存在: %s", channelID)
		}
		if channel.Status != ChannelStatusEnabled {
			return fmt.Errorf("渠道未启用: %s", channelID)
		}
		existing, hasExisting := existingByChannelID[channelID]
		priority := resolveGroupChannelBindingPriority(true, nil, channel.Priority)
		if hasExisting {
			priority = helperInt64Pointer(&existing.Priority)
		}
		createdAt := now
		if hasExisting && existing.CreatedAt > 0 {
			createdAt = existing.CreatedAt
		}
		rows = append(rows, GroupChannelBinding{
			Group:     groupID,
			ChannelId: channelID,
			Enabled:   true,
			Priority:  toSafeGroupChannelBindingPriority(priority),
			CreatedAt: createdAt,
			UpdatedAt: now,
		})
	}
	groupCol := `"group"`
	if err := db.Where(groupCol+" = ?", groupID).Delete(&GroupChannelBinding{}).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group"}, {Name: "channel_id"}},
		UpdateAll: true,
	}).Create(&rows).Error
}

func migrateGroupChannelBindingsWithDB(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	if err := db.AutoMigrate(&GroupChannelBinding{}); err != nil {
		return err
	}
	type sourceRow struct {
		Group     string `gorm:"column:group"`
		ChannelId string `gorm:"column:channel_id"`
		Priority  *int64 `gorm:"column:priority"`
	}
	groupCol := `"group"`
	sourceRows := make([]sourceRow, 0)
	if err := db.Model(&GroupModelRoute{}).
		Select(groupCol + " as \"group\", channel_id, MAX(priority) AS priority").
		Where("channel_id <> ''").
		Group(groupCol + ", channel_id").
		Find(&sourceRows).Error; err != nil {
		return err
	}
	if len(sourceRows) == 0 {
		return nil
	}

	groupOrder := make([]string, 0)
	groupChannelIDs := make(map[string][]string)
	priorityByGroupChannel := make(map[string]*int64, len(sourceRows))
	seenGroup := make(map[string]struct{})
	for _, row := range sourceRows {
		groupID := strings.TrimSpace(row.Group)
		channelID := strings.TrimSpace(row.ChannelId)
		if groupID == "" || channelID == "" {
			continue
		}
		if _, ok := seenGroup[groupID]; !ok {
			seenGroup[groupID] = struct{}{}
			groupOrder = append(groupOrder, groupID)
		}
		groupChannelIDs[groupID] = append(groupChannelIDs[groupID], channelID)
		priorityByGroupChannel[groupID+"::"+channelID] = helperInt64Pointer(row.Priority)
	}
	sort.Strings(groupOrder)
	if len(groupOrder) == 0 {
		return nil
	}

	for _, groupID := range groupOrder {
		channelIDs := normalizeChannelIDList(groupChannelIDs[groupID])
		if err := backfillGroupChannelBindingRowsFromGroupModelRouteDB(db, groupID, channelIDs, priorityByGroupChannel); err != nil {
			return err
		}
	}
	return nil
}

func backfillGroupChannelBindingRowsFromGroupModelRouteDB(db *gorm.DB, groupID string, channelIDs []string, priorityByGroupChannel map[string]*int64) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	groupCatalog, err := resolveGroupCatalogByReferenceWithDB(db, groupID)
	if err != nil {
		return err
	}
	normalizedChannelIDs := normalizeChannelIDList(channelIDs)
	if len(normalizedChannelIDs) == 0 {
		return nil
	}
	channelsByID, err := loadChannelsByIDWithDB(db, normalizedChannelIDs)
	if err != nil {
		return err
	}
	existingRows, err := listGroupChannelBindingRowsWithDB(db, groupCatalog.Id)
	if err != nil {
		return err
	}
	existingByChannelID := make(map[string]GroupChannelBinding, len(existingRows))
	for _, row := range existingRows {
		channelID := strings.TrimSpace(row.ChannelId)
		if channelID == "" {
			continue
		}
		existingByChannelID[channelID] = row
	}
	now := helper.GetTimestamp()
	rows := make([]GroupChannelBinding, 0, len(normalizedChannelIDs))
	for _, channelID := range normalizedChannelIDs {
		channel, ok := channelsByID[channelID]
		if !ok {
			continue
		}
		existing, hasExisting := existingByChannelID[channelID]
		priority := resolveGroupChannelBindingPriority(true, priorityByGroupChannel[groupID+"::"+channelID], channel.Priority)
		createdAt := now
		if hasExisting && existing.CreatedAt > 0 {
			createdAt = existing.CreatedAt
		}
		rows = append(rows, GroupChannelBinding{
			Group:     groupCatalog.Id,
			ChannelId: channelID,
			Enabled:   channel.Status == ChannelStatusEnabled,
			Priority:  toSafeGroupChannelBindingPriority(priority),
			CreatedAt: createdAt,
			UpdatedAt: now,
		})
	}
	groupCol := `"group"`
	if err := db.Where(groupCol+" = ?", groupCatalog.Id).Delete(&GroupChannelBinding{}).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	return db.Create(&rows).Error
}

func loadChannelsByIDWithDB(db *gorm.DB, channelIDs []string) (map[string]*Channel, error) {
	result := make(map[string]*Channel)
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	normalizedChannelIDs := normalizeChannelIDList(channelIDs)
	if len(normalizedChannelIDs) == 0 {
		return result, nil
	}
	rows := make([]Channel, 0, len(normalizedChannelIDs))
	if err := db.Where("id IN ?", normalizedChannelIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].NormalizeIdentity()
		channelID := strings.TrimSpace(rows[i].Id)
		if channelID == "" {
			continue
		}
		result[channelID] = &rows[i]
	}
	return result, nil
}

func toSafeGroupChannelBindingPriority(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
