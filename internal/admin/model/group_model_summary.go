package model

import (
	"fmt"
	"sort"
	"strings"
)

type GroupModelSummaryChannel struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Status   int    `json:"status"`
}

type GroupModelSummaryItem struct {
	Model    string                     `json:"model"`
	Channels []GroupModelSummaryChannel `json:"channels"`
}

func ListGroupModelSummaries(groupID string) ([]GroupModelSummaryItem, error) {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, fmt.Errorf("分组 ID 不能为空")
	}
	groupCatalog, err := getGroupCatalogByIDWithDB(DB, groupID)
	if err != nil {
		return nil, err
	}

	groupModels, err := listGroupModelRowsWithDB(DB, groupCatalog.Id, true)
	if err != nil {
		return nil, err
	}
	if len(groupModels) == 0 {
		return []GroupModelSummaryItem{}, nil
	}

	routes := make([]GroupModelRoute, 0)
	groupCol := `"group"`
	if err := DB.
		Where(groupCol+" = ?", groupCatalog.Id).
		Order("model asc, priority desc, channel_id asc").
		Find(&routes).Error; err != nil {
		return nil, err
	}

	channelIDSet := make(map[string]struct{}, len(routes))
	channelIDs := make([]string, 0, len(routes))
	for _, row := range routes {
		channelID := strings.TrimSpace(row.ChannelId)
		if channelID == "" {
			continue
		}
		if _, ok := channelIDSet[channelID]; ok {
			continue
		}
		channelIDSet[channelID] = struct{}{}
		channelIDs = append(channelIDs, channelID)
	}
	sort.Strings(channelIDs)

	channels := make([]Channel, 0, len(channelIDs))
	if len(channelIDs) > 0 {
		if err := DB.
			Select("id", "name", "protocol", "status").
			Where("id IN ?", channelIDs).
			Where("status = ?", ChannelStatusEnabled).
			Find(&channels).Error; err != nil {
			return nil, err
		}
	}

	channelsByID := make(map[string]GroupModelSummaryChannel, len(channels))
	for _, channel := range channels {
		channel.NormalizeIdentity()
		channelID := strings.TrimSpace(channel.Id)
		if channelID == "" {
			continue
		}
		channelsByID[channelID] = GroupModelSummaryChannel{
			Id:       channelID,
			Name:     channel.DisplayName(),
			Protocol: channel.GetProtocol(),
			Status:   channel.Status,
		}
	}

	summaryByModel := make(map[string]*GroupModelSummaryItem, len(groupModels))
	modelOrder := make([]string, 0, len(groupModels))
	modelChannelSeen := make(map[string]map[string]struct{}, len(groupModels))
	for _, row := range groupModels {
		modelName := strings.TrimSpace(row.Model)
		if modelName == "" {
			continue
		}
		if _, ok := summaryByModel[modelName]; ok {
			continue
		}
		summaryByModel[modelName] = &GroupModelSummaryItem{
			Model:    modelName,
			Channels: make([]GroupModelSummaryChannel, 0),
		}
		modelChannelSeen[modelName] = make(map[string]struct{})
		modelOrder = append(modelOrder, modelName)
	}
	for _, row := range routes {
		modelName := strings.TrimSpace(row.Model)
		channelID := strings.TrimSpace(row.ChannelId)
		if modelName == "" || channelID == "" {
			continue
		}
		channel, ok := channelsByID[channelID]
		if !ok {
			continue
		}
		if _, ok := summaryByModel[modelName]; !ok {
			continue
		}
		if _, ok := modelChannelSeen[modelName][channelID]; ok {
			continue
		}
		modelChannelSeen[modelName][channelID] = struct{}{}
		summaryByModel[modelName].Channels = append(summaryByModel[modelName].Channels, channel)
	}

	result := make([]GroupModelSummaryItem, 0, len(modelOrder))
	for _, modelName := range modelOrder {
		item := summaryByModel[modelName]
		if item == nil {
			continue
		}
		sort.Slice(item.Channels, func(i, j int) bool {
			left := item.Channels[i]
			right := item.Channels[j]
			if left.Name != right.Name {
				return left.Name < right.Name
			}
			return left.Id < right.Id
		})
		result = append(result, *item)
	}
	return result, nil
}
