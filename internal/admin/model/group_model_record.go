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
	GroupModelsTableName = "group_models"
)

type GroupModel struct {
	Group     string `json:"group" gorm:"column:group;primaryKey;type:varchar(32);autoIncrement:false"`
	Model     string `json:"model" gorm:"primaryKey;type:varchar(255);autoIncrement:false"`
	Provider  string `json:"provider" gorm:"type:varchar(128);default:'';index"`
	Enabled   bool   `json:"enabled" gorm:"not null;default:true;index"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;index"`
	UpdatedAt int64  `json:"updated_at" gorm:"bigint;index"`
}

func (GroupModel) TableName() string {
	return GroupModelsTableName
}

func listGroupModelRowsWithDB(db *gorm.DB, groupID string, enabledOnly bool) ([]GroupModel, error) {
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	groupCatalog, err := resolveGroupCatalogByReferenceWithDB(db, groupID)
	if err != nil {
		return nil, err
	}
	groupCol := `"group"`
	rows := make([]GroupModel, 0)
	query := db.Where(groupCol+" = ?", groupCatalog.Id).Order("model asc")
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func ListGroupModelRowsByDB(db *gorm.DB, groupID string) ([]GroupModel, error) {
	return listGroupModelRowsWithDB(db, groupID, true)
}

func listGroupModelNamesWithDB(db *gorm.DB, groupID string, enabledOnly bool) ([]string, error) {
	rows, err := listGroupModelRowsWithDB(db, groupID, enabledOnly)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		modelName := strings.TrimSpace(row.Model)
		if modelName == "" {
			continue
		}
		result = append(result, modelName)
	}
	return NormalizeChannelModelIDsPreserveOrder(result), nil
}

func ListGroupModelNamesByDB(db *gorm.DB, groupID string) ([]string, error) {
	return listGroupModelNamesWithDB(db, groupID, true)
}

func RebuildGroupModelsFromRoutesWithDB(db *gorm.DB, groupID string) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	groupCatalog, err := resolveGroupCatalogByReferenceWithDB(db, groupID)
	if err != nil {
		return err
	}
	groupID = groupCatalog.Id

	groupCol := `"group"`
	routeRows := make([]GroupModelRoute, 0)
	if err := db.
		Where(groupCol+" = ? AND enabled = ?", groupID, true).
		Order("model asc, channel_id asc").
		Find(&routeRows).Error; err != nil {
		return err
	}

	nextRows := make([]GroupModel, 0)
	indexByModel := make(map[string]int)
	for _, route := range routeRows {
		modelName := strings.TrimSpace(route.Model)
		if modelName == "" {
			continue
		}
		provider := NormalizeGroupModelRouteProvider(route.Provider)
		if idx, ok := indexByModel[modelName]; ok {
			if nextRows[idx].Provider == "" {
				nextRows[idx].Provider = provider
			}
			continue
		}
		indexByModel[modelName] = len(nextRows)
		nextRows = append(nextRows, GroupModel{
			Group:    groupID,
			Model:    modelName,
			Provider: provider,
			Enabled:  true,
		})
	}

	return replaceGroupModelsWithDB(db, groupID, nextRows)
}

func replaceGroupModelsWithDB(db *gorm.DB, groupID string, rows []GroupModel) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	groupCatalog, err := resolveGroupCatalogByReferenceWithDB(db, groupID)
	if err != nil {
		return err
	}
	groupID = groupCatalog.Id
	if _, err := buildGroupModelProviderMap(rows); err != nil {
		return err
	}
	normalizedRows := normalizeGroupModelRows(groupID, rows)
	now := helper.GetTimestamp()
	existingRows, err := listGroupModelRowsWithDB(db, groupID, false)
	if err != nil {
		return err
	}
	existingByModel := make(map[string]GroupModel, len(existingRows))
	for _, row := range existingRows {
		modelName := strings.TrimSpace(row.Model)
		if modelName == "" {
			continue
		}
		existingByModel[modelName] = row
	}

	for i := range normalizedRows {
		row := &normalizedRows[i]
		row.Group = groupID
		row.Provider = NormalizeGroupModelRouteProvider(row.Provider)
		row.Enabled = row.Enabled
		row.UpdatedAt = now
		if existing, ok := existingByModel[row.Model]; ok && existing.CreatedAt > 0 {
			row.CreatedAt = existing.CreatedAt
		} else {
			row.CreatedAt = now
		}
	}

	groupCol := `"group"`
	if err := db.Where(groupCol+" = ?", groupID).Delete(&GroupModel{}).Error; err != nil {
		return err
	}
	if len(normalizedRows) == 0 {
		return nil
	}
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group"}, {Name: "model"}},
		UpdateAll: true,
	}).Create(&normalizedRows).Error
}

func normalizeGroupModelRows(groupID string, rows []GroupModel) []GroupModel {
	if len(rows) == 0 {
		return []GroupModel{}
	}
	merged := make(map[string]GroupModel, len(rows))
	for _, row := range rows {
		modelName := strings.TrimSpace(row.Model)
		if modelName == "" {
			continue
		}
		existing, ok := merged[modelName]
		if !ok {
			merged[modelName] = GroupModel{
				Group:    strings.TrimSpace(groupID),
				Model:    modelName,
				Provider: NormalizeGroupModelRouteProvider(row.Provider),
				Enabled:  row.Enabled,
			}
			continue
		}
		existing.Enabled = existing.Enabled || row.Enabled
		if existing.Provider == "" {
			existing.Provider = NormalizeGroupModelRouteProvider(row.Provider)
		}
		merged[modelName] = existing
	}
	modelNames := make([]string, 0, len(merged))
	for modelName := range merged {
		modelNames = append(modelNames, modelName)
	}
	sort.Strings(modelNames)
	result := make([]GroupModel, 0, len(modelNames))
	for _, modelName := range modelNames {
		result = append(result, merged[modelName])
	}
	return result
}

func migrateGroupModelsWithDB(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	if err := migrateGroupModelRoutesTableWithDB(db); err != nil {
		return err
	}
	if err := db.AutoMigrate(&GroupModel{}); err != nil {
		return err
	}

	type sourceRow struct {
		Group      string `gorm:"column:group"`
		Model      string `gorm:"column:model"`
		Provider   string `gorm:"column:provider"`
		EnabledInt int    `gorm:"column:enabled_int"`
	}
	groupCol := `"group"`
	rows := make([]sourceRow, 0)
	if err := db.Model(&GroupModelRoute{}).
		Select(groupCol + " as \"group\", model, provider, MAX(CASE WHEN enabled THEN 1 ELSE 0 END) as enabled_int").
		Where("channel_id <> ''").
		Group(groupCol + ", model, provider").
		Order(groupCol + " asc, model asc").
		Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	grouped := make(map[string][]GroupModel)
	order := make([]string, 0)
	seenGroup := make(map[string]struct{})
	for _, row := range rows {
		groupID := strings.TrimSpace(row.Group)
		modelName := strings.TrimSpace(row.Model)
		if groupID == "" || modelName == "" {
			continue
		}
		if _, ok := seenGroup[groupID]; !ok {
			seenGroup[groupID] = struct{}{}
			order = append(order, groupID)
		}
		grouped[groupID] = append(grouped[groupID], GroupModel{
			Group:    groupID,
			Model:    modelName,
			Provider: NormalizeGroupModelRouteProvider(row.Provider),
			Enabled:  row.EnabledInt > 0,
		})
	}
	sort.Strings(order)
	for _, groupID := range order {
		if err := replaceGroupModelsWithDB(db, groupID, grouped[groupID]); err != nil {
			return err
		}
	}
	return nil
}
