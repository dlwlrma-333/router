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
	APIKey       string                     `json:"api_key,omitempty"`
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

	items, err := loadModelProviderCatalogFromTable(db)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		items = buildDefaultModelProviderCatalogMigration(helper.GetTimestamp())
		logger.SysLogf("migration: initialized model provider catalog with %d default providers", len(items))
	} else {
		items, err = normalizeModelProviderCatalogItems(items, false)
		if err != nil {
			return err
		}
	}

	return saveModelProviderCatalogToTable(db, items)
}

func normalizeModelProviderCatalogItems(items []modelProviderCatalogMigrationItem, mergeWithCodeDefaults bool) ([]modelProviderCatalogMigrationItem, error) {
	now := helper.GetTimestamp()
	indexByProvider := make(map[string]int, len(items))
	normalized := make([]modelProviderCatalogMigrationItem, 0, len(items))

	for _, item := range items {
		provider := commonutils.NormalizeModelProvider(item.Provider)
		if provider == "" {
			provider = commonutils.NormalizeModelProvider(item.Name)
		}
		if provider == "" {
			continue
		}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			name = provider
		}
		source := strings.TrimSpace(strings.ToLower(item.Source))
		if source == "" {
			source = "manual"
		}
		details := make([]ModelProviderModelDetail, 0, len(item.ModelDetails)+len(item.Models))
		details = append(details, item.ModelDetails...)
		for _, modelName := range item.Models {
			details = append(details, ModelProviderModelDetail{Model: strings.TrimSpace(modelName)})
		}
		details = MergeModelProviderDetails(provider, details, item.Models, false, now)
		entry := modelProviderCatalogMigrationItem{
			Provider:     provider,
			Name:         name,
			Models:       ModelProviderModelNames(details),
			ModelDetails: details,
			BaseURL:      strings.TrimSpace(item.BaseURL),
			APIKey:       strings.TrimSpace(item.APIKey),
			SortOrder:    normalizeModelProviderSortOrderValue(item.SortOrder),
			Source:       source,
			UpdatedAt:    item.UpdatedAt,
		}

		if idx, ok := indexByProvider[provider]; ok {
			existing := normalized[idx]
			existing.ModelDetails = MergeModelProviderDetails(
				provider,
				append(existing.ModelDetails, entry.ModelDetails...),
				append(existing.Models, entry.Models...),
				false,
				now,
			)
			existing.Models = ModelProviderModelNames(existing.ModelDetails)
			if existing.Name == existing.Provider && entry.Name != entry.Provider {
				existing.Name = entry.Name
			}
			if existing.BaseURL == "" && entry.BaseURL != "" {
				existing.BaseURL = entry.BaseURL
			}
			if entry.BaseURL != "" && entry.Source != "default" {
				existing.BaseURL = entry.BaseURL
			}
			if existing.APIKey == "" && entry.APIKey != "" {
				existing.APIKey = entry.APIKey
			}
			if entry.APIKey != "" && entry.Source != "default" {
				existing.APIKey = entry.APIKey
			}
			if entry.SortOrder > 0 && entry.Source != "default" {
				existing.SortOrder = entry.SortOrder
			}
			if existing.SortOrder <= 0 && entry.SortOrder > 0 {
				existing.SortOrder = entry.SortOrder
			}
			if entry.UpdatedAt > existing.UpdatedAt {
				existing.UpdatedAt = entry.UpdatedAt
			}
			existing.Source = entry.Source
			normalized[idx] = existing
			continue
		}

		indexByProvider[provider] = len(normalized)
		normalized = append(normalized, entry)
	}

	if mergeWithCodeDefaults {
		normalized = reconcileWithCodeDefaults(normalized, now)
	}
	normalized = finalizeModelProviderCatalogSortOrders(normalized)
	return normalized, nil
}

func loadModelProviderCatalogFromTable(db *gorm.DB) ([]modelProviderCatalogMigrationItem, error) {
	detailsByProvider, err := LoadModelProviderModelDetailsMap(db)
	if err != nil {
		return nil, err
	}

	rows := make([]ModelProvider, 0)
	if err := db.Order("sort_order asc, provider asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]modelProviderCatalogMigrationItem, 0, len(rows))
	for _, row := range rows {
		provider := commonutils.NormalizeModelProvider(row.Provider)
		if provider == "" {
			continue
		}
		details := MergeModelProviderDetails(provider, detailsByProvider[provider], nil, false, helper.GetTimestamp())
		items = append(items, modelProviderCatalogMigrationItem{
			Provider:     provider,
			Name:         strings.TrimSpace(row.Name),
			Models:       ModelProviderModelNames(details),
			ModelDetails: details,
			BaseURL:      strings.TrimSpace(row.BaseURL),
			APIKey:       strings.TrimSpace(row.APIKey),
			SortOrder:    normalizeModelProviderSortOrderValue(row.SortOrder),
			Source:       strings.TrimSpace(strings.ToLower(row.Source)),
			UpdatedAt:    row.UpdatedAt,
		})
	}
	return items, nil
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
			Provider:  provider,
			Name:      strings.TrimSpace(item.Name),
			BaseURL:   strings.TrimSpace(item.BaseURL),
			APIKey:    strings.TrimSpace(item.APIKey),
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

func reconcileWithCodeDefaults(items []modelProviderCatalogMigrationItem, now int64) []modelProviderCatalogMigrationItem {
	defaults := buildDefaultModelProviderCatalogMigration(now)
	defaultByProvider := make(map[string]modelProviderCatalogMigrationItem, len(defaults))
	for _, item := range defaults {
		defaultByProvider[item.Provider] = item
	}

	result := make(map[string]modelProviderCatalogMigrationItem, len(items)+len(defaults))
	for _, item := range defaults {
		result[item.Provider] = item
	}

	for _, item := range items {
		provider := commonutils.NormalizeModelProvider(item.Provider)
		if provider == "" {
			continue
		}
		item.Provider = provider
		item.ModelDetails = MergeModelProviderDetails(provider, item.ModelDetails, item.Models, false, now)
		item.Models = ModelProviderModelNames(item.ModelDetails)

		if seededItem, ok := defaultByProvider[provider]; ok {
			merged := seededItem
			if strings.TrimSpace(item.Name) != "" && item.Name != provider {
				merged.Name = strings.TrimSpace(item.Name)
			}
			if strings.TrimSpace(item.BaseURL) != "" {
				merged.BaseURL = strings.TrimSpace(item.BaseURL)
			}
			if strings.TrimSpace(item.APIKey) != "" {
				merged.APIKey = strings.TrimSpace(item.APIKey)
			}
			if item.SortOrder > 0 {
				merged.SortOrder = item.SortOrder
			}
			if item.UpdatedAt > 0 {
				merged.UpdatedAt = item.UpdatedAt
			}
			if item.Source != "default" {
				merged.Source = item.Source
			}
			merged.ModelDetails = MergeModelProviderDetails(
				provider,
				append(seededItem.ModelDetails, item.ModelDetails...),
				append(seededItem.Models, item.Models...),
				false,
				now,
			)
			merged.Models = ModelProviderModelNames(merged.ModelDetails)
			result[provider] = merged
			continue
		}
		if item.Source == "" {
			item.Source = "manual"
		}
		item.SortOrder = normalizeModelProviderSortOrderValue(item.SortOrder)
		result[provider] = item
	}

	mergedItems := make([]modelProviderCatalogMigrationItem, 0, len(result))
	for _, item := range result {
		item.ModelDetails = MergeModelProviderDetails(item.Provider, item.ModelDetails, item.Models, false, now)
		item.Models = ModelProviderModelNames(item.ModelDetails)
		if item.Name == "" {
			item.Name = item.Provider
		}
		if item.Source == "" {
			item.Source = "manual"
		}
		item.SortOrder = normalizeModelProviderSortOrderValue(item.SortOrder)
		mergedItems = append(mergedItems, item)
	}
	return finalizeModelProviderCatalogSortOrders(mergedItems)
}
