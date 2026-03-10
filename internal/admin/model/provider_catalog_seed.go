package model

import (
	"strings"

	"github.com/yeying-community/router/common/helper"
	"github.com/yeying-community/router/common/logger"
	commonutils "github.com/yeying-community/router/common/utils"
	"gorm.io/gorm"
)

func normalizeProviderSortOrderValue(sortOrder int) int {
	if sortOrder > 0 {
		return sortOrder
	}
	return 0
}

func ensureProviderCatalogSeededWithDB(db *gorm.DB) error {
	if err := db.AutoMigrate(&Provider{}, &ProviderModel{}); err != nil {
		return err
	}
	var count int64
	if err := db.Model(&Provider{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	seeds := BuildDefaultProviderCatalogSeeds(helper.GetTimestamp())
	logger.SysLogf("migration: initialized model provider catalog with %d default providers", len(seeds))
	return saveProviderCatalogSeedsToTable(db, seeds)
}

func saveProviderCatalogSeedsToTable(db *gorm.DB, seeds []ProviderCatalogSeed) error {
	now := helper.GetTimestamp()
	providerRows := make([]Provider, 0, len(seeds))
	modelRows := make([]ProviderModel, 0)
	for _, seed := range seeds {
		provider := commonutils.NormalizeProvider(seed.Provider)
		if provider == "" {
			continue
		}
		details := normalizeDefaultProviderSeedModelDetails(provider, seed.ModelDetails, now)
		providerRows = append(providerRows, Provider{
			Id:        provider,
			Name:      strings.TrimSpace(seed.Name),
			BaseURL:   strings.TrimSpace(seed.BaseURL),
			SortOrder: normalizeProviderSortOrderValue(seed.SortOrder),
			Source:    "default",
			UpdatedAt: now,
		})
		modelRows = append(modelRows, BuildProviderModelRows(provider, details, now)...)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&ProviderModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("1 = 1").Delete(&Provider{}).Error; err != nil {
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
