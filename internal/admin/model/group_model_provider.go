package model

import (
	"fmt"
	"sort"
	"strings"
)

func NormalizeGroupModelProviderValue(provider string) string {
	return NormalizeGroupModelRouteProvider(provider)
}

func ListGroupModelProviderMapByModels(groupID string, modelNames []string) (map[string]string, error) {
	normalizedGroupID := strings.TrimSpace(groupID)
	normalizedModelNames := NormalizeChannelModelIDsPreserveOrder(modelNames)
	result := make(map[string]string, len(normalizedModelNames))
	if normalizedGroupID == "" || len(normalizedModelNames) == 0 {
		return result, nil
	}

	rows := make([]GroupModel, 0, len(normalizedModelNames))
	groupCol := `"group"`
	if err := DB.
		Select(groupCol, "model", "provider").
		Where(groupCol+" = ?", normalizedGroupID).
		Where("model IN ?", normalizedModelNames).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return buildGroupModelProviderMap(rows)
}

func buildGroupModelRouteProviderMap(rows []GroupModelRoute) (map[string]string, error) {
	groupModels := make([]GroupModel, 0, len(rows))
	for _, row := range rows {
		groupModels = append(groupModels, GroupModel{
			Group:    strings.TrimSpace(row.Group),
			Model:    strings.TrimSpace(row.Model),
			Provider: NormalizeGroupModelRouteProvider(row.Provider),
			Enabled:  row.Enabled,
		})
	}
	return buildGroupModelProviderMap(groupModels)
}

func buildGroupModelProviderMap(rows []GroupModel) (map[string]string, error) {
	if len(rows) == 0 {
		return map[string]string{}, nil
	}
	modelOrder := make([]string, 0, len(rows))
	seenModel := make(map[string]struct{}, len(rows))
	candidatesByModel := make(map[string]map[string]struct{}, len(rows))
	for _, row := range rows {
		modelName := strings.TrimSpace(row.Model)
		if modelName == "" {
			continue
		}
		if _, ok := seenModel[modelName]; !ok {
			seenModel[modelName] = struct{}{}
			modelOrder = append(modelOrder, modelName)
		}
		provider := NormalizeGroupModelRouteProvider(row.Provider)
		if provider == "" {
			continue
		}
		if _, ok := candidatesByModel[modelName]; !ok {
			candidatesByModel[modelName] = make(map[string]struct{}, 1)
		}
		candidatesByModel[modelName][provider] = struct{}{}
	}
	providerByModel := make(map[string]string, len(modelOrder))
	for _, modelName := range modelOrder {
		candidateSet := candidatesByModel[modelName]
		if len(candidateSet) == 0 {
			providerByModel[modelName] = ""
			continue
		}
		providers := make([]string, 0, len(candidateSet))
		for provider := range candidateSet {
			providers = append(providers, provider)
		}
		sort.Strings(providers)
		if len(providers) > 1 {
			return nil, fmt.Errorf("同一分组模型仅允许一个供应商: %s (%s)", modelName, strings.Join(providers, " / "))
		}
		providerByModel[modelName] = providers[0]
	}
	return providerByModel, nil
}
