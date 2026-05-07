package model

import (
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type GroupChannelBindingItem struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Status   int    `json:"status"`
	Models   string `json:"models"`
	Bound    bool   `json:"bound"`
	Priority *int64 `json:"priority,omitempty"`
	Updated  int64  `json:"updated_at"`
}

func ListGroupChannelBindings(groupID string) ([]GroupChannelBindingItem, error) {
	if strings.TrimSpace(groupID) == "" {
		return nil, fmt.Errorf("分组 ID 不能为空")
	}
	return listGroupChannelBindingsWithDB(DB, groupID, true)
}

func listGroupChannelBindingsWithDB(db *gorm.DB, groupID string, enabledOnly bool) ([]GroupChannelBindingItem, error) {
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	groupID = strings.TrimSpace(groupID)

	channels := make([]Channel, 0)
	query := db.
		Select("id", "name", "protocol", "status", "created_time").
		Order("created_time desc")
	if enabledOnly {
		query = query.Where("status = ?", ChannelStatusEnabled)
	}
	if err := query.Find(&channels).Error; err != nil {
		return nil, err
	}
	channelRefs := make([]*Channel, 0, len(channels))
	for i := range channels {
		channelRefs = append(channelRefs, &channels[i])
	}
	if err := HydrateChannelsWithModels(db, channelRefs); err != nil {
		return nil, err
	}

	bindingRows, err := listGroupChannelBindingRowsWithDB(db, groupID)
	if err != nil {
		return nil, err
	}
	boundSet := make(map[string]struct{}, len(bindingRows))
	priorityByChannelID := make(map[string]*int64, len(bindingRows))
	updatedByChannelID := make(map[string]int64, len(bindingRows))
	for _, row := range bindingRows {
		normalized := strings.TrimSpace(row.ChannelId)
		if normalized == "" {
			continue
		}
		boundSet[normalized] = struct{}{}
		priority := row.Priority
		priorityByChannelID[normalized] = &priority
		updatedByChannelID[normalized] = row.UpdatedAt
	}

	items := make([]GroupChannelBindingItem, 0, len(channels))
	for _, channel := range channels {
		channel.NormalizeIdentity()
		channelID := strings.TrimSpace(channel.Id)
		if channelID == "" {
			continue
		}
		_, bound := boundSet[channelID]
		items = append(items, GroupChannelBindingItem{
			Id:       channelID,
			Name:     channel.DisplayName(),
			Protocol: channel.GetProtocol(),
			Status:   channel.Status,
			Models:   strings.TrimSpace(channel.Models),
			Bound:    bound,
			Priority: resolveGroupChannelBindingPriority(bound, priorityByChannelID[channelID], channel.Priority),
			Updated:  resolveGroupChannelBindingUpdatedAt(bound, updatedByChannelID[channelID], channel.CreatedTime),
		})
	}
	return items, nil
}

func ReplaceGroupChannelBindings(groupID string, channelIDs []string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return replaceGroupChannelBindingsWithDB(tx, groupID, channelIDs)
	})
}

func ReplaceGroupChannelBindingsWithItems(groupID string, items []GroupChannelBindingItem) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		return replaceGroupChannelBindingsWithItemsDB(tx, groupID, items)
	})
}

func replaceGroupChannelBindingsWithDB(db *gorm.DB, groupID string, channelIDs []string) error {
	items := make([]GroupChannelBindingItem, 0, len(channelIDs))
	for _, channelID := range normalizeChannelIDList(channelIDs) {
		items = append(items, GroupChannelBindingItem{
			Id:    channelID,
			Bound: true,
		})
	}
	return replaceGroupChannelBindingsWithItemsDB(db, groupID, items)
}

func replaceGroupChannelBindingsWithItemsDB(db *gorm.DB, groupID string, items []GroupChannelBindingItem) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return fmt.Errorf("分组 ID 不能为空")
	}

	groupCatalog, err := getGroupCatalogByIDWithDB(db, groupID)
	if err != nil {
		return err
	}
	groupID = groupCatalog.Id

	normalizedItems := normalizeGroupChannelBindingItems(items)
	if err := replaceGroupChannelBindingRowsWithItemsDB(db, groupID, normalizedItems); err != nil {
		return err
	}
	normalizedChannelIDs, err := listGroupBoundChannelIDsWithDB(db, groupID)
	if err != nil {
		return err
	}
	priorityByChannelID, err := listGroupChannelBindingPriorityByChannelWithDB(db, groupID)
	if err != nil {
		return err
	}
	channelsByID := make(map[string]Channel, len(normalizedChannelIDs))
	if len(normalizedChannelIDs) > 0 {
		enabledChannels, err := loadEnabledChannelsByIDWithDB(db, normalizedChannelIDs)
		if err != nil {
			return err
		}
		for channelID, channel := range enabledChannels {
			channelsByID[channelID] = *channel
		}
	}

	groupCol := `"group"`
	groupModels, err := listGroupModelRowsWithDB(db, groupID, true)
	if err != nil {
		return err
	}
	routes := make([]GroupModelRoute, 0)
	for _, id := range normalizedChannelIDs {
		channel, ok := channelsByID[id]
		if !ok {
			continue
		}
		channelAbilities := SyncGroupModelRoutesForChannel(groupID, &channel, groupModels, priorityByChannelID[id])
		if priority, ok := priorityByChannelID[id]; ok {
			for idx := range channelAbilities {
				channelAbilities[idx].Priority = helperInt64Pointer(priority)
			}
		}
		routes = append(routes, channelAbilities...)
	}
	routes = normalizeGroupModelRouteRowsPreserveOrder(routes)

	if err := db.Where(groupCol+" = ?", groupID).Delete(&GroupModelRoute{}).Error; err != nil {
		return err
	}
	if len(routes) > 0 {
		if err := db.Create(&routes).Error; err != nil {
			return err
		}
	}
	if _, err := buildGroupModelRouteProviderMap(routes); err != nil {
		return err
	}
	return nil
}

func normalizeGroupChannelBindingItems(items []GroupChannelBindingItem) []GroupChannelBindingItem {
	if len(items) == 0 {
		return []GroupChannelBindingItem{}
	}
	result := make([]GroupChannelBindingItem, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		channelID := strings.TrimSpace(item.Id)
		if channelID == "" {
			continue
		}
		if _, ok := seen[channelID]; ok {
			continue
		}
		seen[channelID] = struct{}{}
		result = append(result, GroupChannelBindingItem{
			Id:       channelID,
			Bound:    item.Bound,
			Priority: helperInt64Pointer(item.Priority),
		})
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Id < result[j].Id
	})
	return result
}

func resolveGroupChannelBindingPriority(bound bool, abilityPriority *int64, channelPriority *int64) *int64 {
	if bound && abilityPriority != nil {
		return helperInt64Pointer(abilityPriority)
	}
	return helperInt64Pointer(channelPriority)
}

func resolveGroupChannelBindingUpdatedAt(bound bool, bindingUpdatedAt int64, fallback int64) int64 {
	if bound && bindingUpdatedAt > 0 {
		return bindingUpdatedAt
	}
	return fallback
}

func loadEnabledChannelsByIDWithDB(db *gorm.DB, channelIDs []string) (map[string]*Channel, error) {
	if db == nil {
		return nil, fmt.Errorf("database handle is nil")
	}
	normalizedChannelIDs := normalizeChannelIDList(channelIDs)
	if len(normalizedChannelIDs) == 0 {
		return map[string]*Channel{}, nil
	}

	channels := make([]Channel, 0, len(normalizedChannelIDs))
	if err := db.
		Select("id", "name", "protocol", "status", "priority", "created_time").
		Where("id IN ?", normalizedChannelIDs).
		Find(&channels).Error; err != nil {
		return nil, err
	}

	channelsByID := make(map[string]*Channel, len(channels))
	disabled := make([]string, 0)
	for i := range channels {
		channel := &channels[i]
		channel.NormalizeIdentity()
		channelID := strings.TrimSpace(channel.Id)
		if channelID == "" {
			continue
		}
		channelsByID[channelID] = channel
		if channel.Status != ChannelStatusEnabled {
			disabled = append(disabled, channelID)
		}
	}

	if len(channelsByID) != len(normalizedChannelIDs) {
		missing := make([]string, 0)
		for _, channelID := range normalizedChannelIDs {
			if _, ok := channelsByID[channelID]; !ok {
				missing = append(missing, channelID)
			}
		}
		sort.Strings(missing)
		return nil, fmt.Errorf("渠道不存在: %s", strings.Join(missing, ", "))
	}
	if len(disabled) > 0 {
		sort.Strings(disabled)
		return nil, fmt.Errorf("渠道未启用，不能绑定到分组: %s", strings.Join(disabled, ", "))
	}

	channelRefs := make([]*Channel, 0, len(channels))
	for i := range channels {
		channelRefs = append(channelRefs, &channels[i])
	}
	if err := HydrateChannelsWithModels(db, channelRefs); err != nil {
		return nil, err
	}
	return channelsByID, nil
}

func normalizeChannelIDList(ids []string) []string {
	if len(ids) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(ids))
	result := make([]string, 0, len(ids))
	for _, item := range ids {
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
	sort.Strings(result)
	return result
}

func normalizeModelNames(models []string) []string {
	if len(models) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(models))
	result := make([]string, 0, len(models))
	for _, item := range models {
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
	sort.Strings(result)
	return result
}

func SyncGroupModelRoutesForChannel(groupID string, channel *Channel, groupModels []GroupModel, channelPriority *int64) []GroupModelRoute {
	if channel == nil || len(groupModels) == 0 {
		return nil
	}
	catalog := buildGroupModelConfigChannelCatalog(channel)
	result := make([]GroupModelRoute, 0, len(groupModels))
	seenGroupModelRouteKeys := make(map[string]struct{}, len(groupModels))
	priority := helperInt64Pointer(channel.Priority)
	if channelPriority != nil {
		priority = helperInt64Pointer(channelPriority)
	}
	for _, groupModel := range groupModels {
		modelName := strings.TrimSpace(groupModel.Model)
		if modelName == "" || !groupModel.Enabled {
			continue
		}
		upstream, ok := catalog.aliasToUpstream[modelName]
		if !ok || strings.TrimSpace(upstream) == "" {
			continue
		}
		key := modelName + "::" + strings.TrimSpace(channel.Id)
		if _, ok := seenGroupModelRouteKeys[key]; ok {
			continue
		}
		seenGroupModelRouteKeys[key] = struct{}{}
		provider := NormalizeGroupModelRouteProvider(groupModel.Provider)
		if provider == "" {
			provider = NormalizeGroupModelRouteProvider(catalog.ResolveProvider(GroupModelConfigItem{Model: modelName}, upstream))
		}
		result = append(result, GroupModelRoute{
			Group:         strings.TrimSpace(groupID),
			Model:         modelName,
			ChannelId:     strings.TrimSpace(channel.Id),
			UpstreamModel: NormalizeGroupModelRouteUpstreamModel(modelName, upstream),
			Provider:      provider,
			Enabled:       channel.Status == ChannelStatusEnabled,
			Priority:      priority,
		})
	}
	return result
}
