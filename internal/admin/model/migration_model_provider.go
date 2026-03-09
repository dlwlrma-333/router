package model

import (
	"sort"
	"strings"

	"github.com/yeying-community/router/common/helper"
	"github.com/yeying-community/router/common/logger"
	commonutils "github.com/yeying-community/router/common/utils"
	"gorm.io/gorm"
)

type modelProviderCatalogMigrationItem struct {
	Provider     string                     `json:"provider"`
	Name         string                     `json:"name,omitempty"`
	Models       []string                   `json:"models"`
	ModelDetails []ModelProviderModelDetail `json:"model_details,omitempty"`
	BaseURL      string                     `json:"base_url,omitempty"`
	SortOrder    int                        `json:"sort_order,omitempty"`
	Source       string                     `json:"source,omitempty"`
	UpdatedAt    int64                      `json:"updated_at,omitempty"`
}

func normalizeModelProviderSortOrderValue(sortOrder int) int {
	if sortOrder > 0 {
		return sortOrder
	}
	return 0
}

func finalizeModelProviderCatalogSortOrders(items []modelProviderCatalogMigrationItem) []modelProviderCatalogMigrationItem {
	sort.SliceStable(items, func(i, j int) bool {
		leftOrder := normalizeModelProviderSortOrderValue(items[i].SortOrder)
		rightOrder := normalizeModelProviderSortOrderValue(items[j].SortOrder)
		if leftOrder > 0 && rightOrder > 0 {
			if leftOrder != rightOrder {
				return leftOrder < rightOrder
			}
			return items[i].Provider < items[j].Provider
		}
		if leftOrder > 0 {
			return true
		}
		if rightOrder > 0 {
			return false
		}
		return items[i].Provider < items[j].Provider
	})

	nextOrder := 10
	for i := range items {
		order := normalizeModelProviderSortOrderValue(items[i].SortOrder)
		if order > 0 {
			items[i].SortOrder = order
			if order >= nextOrder {
				nextOrder = order + 10
			}
			continue
		}
		items[i].SortOrder = nextOrder
		nextOrder += 10
	}
	return items
}

func syncModelProviderCatalogWithDB(db *gorm.DB) error {
	if err := db.AutoMigrate(&ModelProvider{}, &ModelProviderModel{}); err != nil {
		return err
	}
	var count int64
	if err := db.Model(&ModelProvider{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	items := buildDefaultModelProviderCatalogMigration(helper.GetTimestamp())
	logger.SysLogf("migration: initialized model provider catalog with %d default providers", len(items))
	return saveModelProviderCatalogToTable(db, items)
}

func saveModelProviderCatalogToTable(db *gorm.DB, items []modelProviderCatalogMigrationItem) error {
	now := helper.GetTimestamp()
	items = finalizeModelProviderCatalogSortOrders(items)
	providerRows := make([]ModelProvider, 0, len(items))
	modelRows := make([]ModelProviderModel, 0)
	for _, item := range items {
		provider := commonutils.NormalizeModelProvider(item.Provider)
		if provider == "" {
			continue
		}
		details := MergeModelProviderDetails(provider, item.ModelDetails, item.Models, false, now)
		updatedAt := item.UpdatedAt
		if updatedAt == 0 {
			updatedAt = now
		}
		source := strings.TrimSpace(strings.ToLower(item.Source))
		if source == "" {
			source = "manual"
		}
		providerRows = append(providerRows, ModelProvider{
			Id:        provider,
			Name:      strings.TrimSpace(item.Name),
			BaseURL:   strings.TrimSpace(item.BaseURL),
			SortOrder: item.SortOrder,
			Source:    source,
			UpdatedAt: updatedAt,
		})
		modelRows = append(modelRows, BuildModelProviderModelRows(provider, details, now)...)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&ModelProviderModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("1 = 1").Delete(&ModelProvider{}).Error; err != nil {
			return err
		}
		if len(providerRows) > 0 {
			if err := tx.Create(&providerRows).Error; err != nil {
				return err
			}
		}
		if len(modelRows) > 0 {
			if err := tx.Create(&modelRows).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func buildDefaultModelProviderCatalogMigration(now int64) []modelProviderCatalogMigrationItem {
	seeds := BuildDefaultModelProviderCatalogSeeds(now)
	items := make([]modelProviderCatalogMigrationItem, 0, len(seeds))
	for _, seed := range seeds {
		items = append(items, modelProviderCatalogMigrationItem{
			Provider:     seed.Provider,
			Name:         seed.Name,
			Models:       ModelProviderModelNames(seed.ModelDetails),
			ModelDetails: seed.ModelDetails,
			BaseURL:      seed.BaseURL,
			SortOrder:    seed.SortOrder,
			Source:       "default",
			UpdatedAt:    now,
		})
	}
	return items
}
