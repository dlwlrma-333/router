package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func migrateGroupModelProvidersWithDB(tx *gorm.DB) error {
	// Historical migration kept only for version continuity.
	return nil
}

func backfillGroupModelRouteProviderFromChannelModelsWithDB(tx *gorm.DB) error {
	if tx == nil {
		return fmt.Errorf("database handle is nil")
	}
	if err := migrateGroupModelRoutesTableWithDB(tx); err != nil {
		return err
	}

	rows := make([]GroupModelRoute, 0)
	groupCol := `"group"`
	if err := tx.
		Select(groupCol, "model", "channel_id", "upstream_model", "provider", "enabled", "priority").
		Where("channel_id <> ''").
		Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	candidates := make([]string, 0, len(rows)*2)
	for _, row := range rows {
		if modelName := strings.TrimSpace(row.Model); modelName != "" {
			candidates = append(candidates, modelName)
		}
		if upstreamModel := strings.TrimSpace(row.UpstreamModel); upstreamModel != "" {
			candidates = append(candidates, upstreamModel)
		}
	}
	providerByModel, err := LoadUniqueProviderMapByModelsWithDB(tx, candidates)
	if err != nil {
		return err
	}

	for _, row := range rows {
		resolvedProvider := ResolveGroupModelRouteProviderWithDB(tx, providerByModel, row.ChannelId, row.Model, row.UpstreamModel)
		resolvedProvider = NormalizeGroupModelRouteProvider(resolvedProvider)
		currentProvider := NormalizeGroupModelRouteProvider(row.Provider)
		if currentProvider == resolvedProvider {
			continue
		}
		if err := tx.Model(&GroupModelRoute{}).
			Where(groupCol+" = ? AND model = ? AND channel_id = ?", strings.TrimSpace(row.Group), strings.TrimSpace(row.Model), strings.TrimSpace(row.ChannelId)).
			Update("provider", resolvedProvider).Error; err != nil {
			return err
		}
	}
	return nil
}

func dropGroupModelProvidersTableWithDB(tx *gorm.DB) error {
	if tx == nil {
		return fmt.Errorf("database handle is nil")
	}
	if tx.Migrator().HasTable("group_model_providers") {
		return tx.Migrator().DropTable("group_model_providers")
	}
	return nil
}

func ResolveGroupModelRouteProviderWithDB(tx *gorm.DB, providerByModel map[string]string, channelID string, modelName string, upstreamModel string) string {
	normalizedChannelID := strings.TrimSpace(channelID)
	normalizedModelName := strings.TrimSpace(modelName)
	if tx != nil && normalizedChannelID != "" && normalizedModelName != "" {
		row := ChannelModel{}
		if err := tx.
			Select("provider", "model", "upstream_model").
			Where("channel_id = ? AND model = ?", normalizedChannelID, normalizedModelName).
			Take(&row).Error; err == nil {
			provider := NormalizeGroupModelRouteProvider(row.Provider)
			if provider != "" {
				return provider
			}
			return NormalizeGroupModelRouteProvider(ResolveProviderFromCatalogMap(providerByModel, row.Model, row.UpstreamModel))
		}
	}
	return NormalizeGroupModelRouteProvider(ResolveProviderFromCatalogMap(providerByModel, modelName, upstreamModel))
}
