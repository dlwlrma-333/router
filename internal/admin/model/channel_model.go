package model

import (
	"fmt"
	"strings"

	"github.com/yeying-community/router/common/helper"
	"gorm.io/gorm"
)

const (
	ChannelModelsTableName = "channel_models"
)

type ChannelModel struct {
	ChannelId string `json:"channel_id" gorm:"primaryKey;type:char(36);index"`
	Model     string `json:"model" gorm:"primaryKey;type:varchar(255)"`
	Selected  bool   `json:"selected" gorm:"default:true;index"`
	SortOrder int    `json:"sort_order" gorm:"default:0"`
	UpdatedAt int64  `json:"updated_at" gorm:"bigint"`
}

func (ChannelModel) TableName() string {
	return ChannelModelsTableName
}

func NormalizeChannelModelIDsPreserveOrder(modelIDs []string) []string {
	return normalizeTrimmedValuesPreserveOrder(modelIDs)
}

func ParseChannelModelCSV(models string) []string {
	if strings.TrimSpace(models) == "" {
		return []string{}
	}
	return NormalizeChannelModelIDsPreserveOrder(strings.FieldsFunc(models, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	}))
}

func JoinChannelModelCSV(modelIDs []string) string {
	return strings.Join(NormalizeChannelModelIDsPreserveOrder(modelIDs), ",")
}

func HydrateChannelWithModels(db *gorm.DB, channel *Channel) error {
	if channel == nil {
		return nil
	}
	return HydrateChannelsWithModels(db, []*Channel{channel})
}

func HydrateChannelsWithModels(db *gorm.DB, channels []*Channel) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	channelIDs := make([]string, 0, len(channels))
	normalizedChannels := make([]*Channel, 0, len(channels))
	for _, channel := range channels {
		if channel == nil {
			continue
		}
		channel.Id = strings.TrimSpace(channel.Id)
		if channel.Id == "" {
			channel.SetSelectedModelIDs(nil)
			channel.SetAvailableModelIDs(nil)
			continue
		}
		channelIDs = append(channelIDs, channel.Id)
		normalizedChannels = append(normalizedChannels, channel)
	}
	if len(normalizedChannels) == 0 {
		return nil
	}

	rowsByChannelID, err := loadChannelModelRowsByChannelIDs(db, channelIDs)
	if err != nil {
		return err
	}
	for _, channel := range normalizedChannels {
		applyChannelModelRows(channel, rowsByChannelID[channel.Id])
	}
	return nil
}

func ListSelectedChannelModelIDsByChannelIDWithDB(db *gorm.DB, channelID string) ([]string, error) {
	rows, err := listChannelModelRowsByChannelIDWithDB(db, channelID)
	if err != nil {
		return nil, err
	}
	modelIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		if !row.Selected {
			continue
		}
		modelIDs = append(modelIDs, row.Model)
	}
	return NormalizeChannelModelIDsPreserveOrder(modelIDs), nil
}

func ListAvailableChannelModelIDsByChannelIDWithDB(db *gorm.DB, channelID string) ([]string, error) {
	rows, err := listChannelModelRowsByChannelIDWithDB(db, channelID)
	if err != nil {
		return nil, err
	}
	modelIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		modelIDs = append(modelIDs, row.Model)
	}
	return NormalizeChannelModelIDsPreserveOrder(modelIDs), nil
}

func SyncFetchedChannelModelsWithDB(db *gorm.DB, channelID string, modelIDs []string) error {
	return replaceChannelModelRowsWithDB(db, channelID, modelIDs, buildChannelModelSelectionSet(modelIDs))
}

func ReplaceChannelSelectedModelsWithDB(db *gorm.DB, channelID string, selected []string) error {
	existingRows, err := listChannelModelRowsByChannelIDWithDB(db, channelID)
	if err != nil {
		return err
	}
	available := make([]string, 0, len(existingRows)+len(selected))
	seen := make(map[string]struct{}, len(existingRows)+len(selected))
	for _, row := range existingRows {
		if _, ok := seen[row.Model]; ok {
			continue
		}
		seen[row.Model] = struct{}{}
		available = append(available, row.Model)
	}
	for _, modelID := range NormalizeChannelModelIDsPreserveOrder(selected) {
		if _, ok := seen[modelID]; ok {
			continue
		}
		seen[modelID] = struct{}{}
		available = append(available, modelID)
	}
	return replaceChannelModelRowsWithDB(db, channelID, available, buildChannelModelSelectionSet(selected))
}

func DeleteChannelModelsByChannelIDWithDB(db *gorm.DB, channelID string) error {
	return DeleteChannelModelsByChannelIDsWithDB(db, []string{channelID})
}

func DeleteChannelModelsByChannelIDsWithDB(db *gorm.DB, channelIDs []string) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	normalizedIDs := normalizeTrimmedValuesPreserveOrder(channelIDs)
	if len(normalizedIDs) == 0 {
		return nil
	}
	return db.Where("channel_id IN ?", normalizedIDs).Delete(&ChannelModel{}).Error
}

func EnsureChannelTestModelWithDB(db *gorm.DB, channelID string) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	normalizedChannelID := strings.TrimSpace(channelID)
	if normalizedChannelID == "" {
		return nil
	}

	type channelTestModelRecord struct {
		TestModel string `gorm:"column:test_model"`
	}

	record := channelTestModelRecord{}
	if err := db.Model(&Channel{}).
		Select("test_model").
		Where("id = ?", normalizedChannelID).
		Take(&record).Error; err != nil {
		return err
	}

	selectedModelIDs, err := ListSelectedChannelModelIDsByChannelIDWithDB(db, normalizedChannelID)
	if err != nil {
		return err
	}
	current := strings.TrimSpace(record.TestModel)
	for _, modelID := range selectedModelIDs {
		if modelID == current {
			return nil
		}
	}

	next := ""
	if len(selectedModelIDs) > 0 {
		next = selectedModelIDs[0]
	}
	if current == next {
		return nil
	}
	return db.Model(&Channel{}).
		Where("id = ?", normalizedChannelID).
		Update("test_model", next).Error
}

func loadChannelModelRowsByChannelIDs(db *gorm.DB, channelIDs []string) (map[string][]ChannelModel, error) {
	rowsByChannelID := make(map[string][]ChannelModel)
	normalizedIDs := normalizeTrimmedValuesPreserveOrder(channelIDs)
	if len(normalizedIDs) == 0 {
		return rowsByChannelID, nil
	}
	rows := make([]ChannelModel, 0)
	if err := db.
		Where("channel_id IN ?", normalizedIDs).
		Order("channel_id asc, sort_order asc, model asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		rowsByChannelID[row.ChannelId] = append(rowsByChannelID[row.ChannelId], row)
	}
	return rowsByChannelID, nil
}

func normalizeTrimmedValuesPreserveOrder(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, item := range values {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func listChannelModelRowsByChannelIDWithDB(db *gorm.DB, channelID string) ([]ChannelModel, error) {
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	normalizedChannelID := strings.TrimSpace(channelID)
	if normalizedChannelID == "" {
		return []ChannelModel{}, nil
	}
	rows := make([]ChannelModel, 0)
	if err := db.
		Where("channel_id = ?", normalizedChannelID).
		Order("sort_order asc, model asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func applyChannelModelRows(channel *Channel, rows []ChannelModel) {
	if channel == nil {
		return
	}
	available := make([]string, 0, len(rows))
	selected := make([]string, 0, len(rows))
	for _, row := range rows {
		available = append(available, row.Model)
		if row.Selected {
			selected = append(selected, row.Model)
		}
	}
	channel.SetAvailableModelIDs(available)
	channel.SetSelectedModelIDs(selected)
}

func buildChannelModelSelectionSet(modelIDs []string) map[string]struct{} {
	normalized := NormalizeChannelModelIDsPreserveOrder(modelIDs)
	set := make(map[string]struct{}, len(normalized))
	for _, modelID := range normalized {
		set[modelID] = struct{}{}
	}
	return set
}

func replaceChannelModelRowsWithDB(db *gorm.DB, channelID string, available []string, selectedSet map[string]struct{}) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	normalizedChannelID := strings.TrimSpace(channelID)
	if normalizedChannelID == "" {
		return nil
	}
	normalizedAvailable := NormalizeChannelModelIDsPreserveOrder(available)
	now := helper.GetTimestamp()
	rows := make([]ChannelModel, 0, len(normalizedAvailable))
	for idx, modelID := range normalizedAvailable {
		_, selected := selectedSet[modelID]
		rows = append(rows, ChannelModel{
			ChannelId: normalizedChannelID,
			Model:     modelID,
			Selected:  selected,
			SortOrder: idx + 1,
			UpdatedAt: now,
		})
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("channel_id = ?", normalizedChannelID).Delete(&ChannelModel{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}
