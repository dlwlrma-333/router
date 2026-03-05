package model

import (
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type GroupChannelBindingItem struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Type    int    `json:"type"`
	Status  int    `json:"status"`
	Models  string `json:"models"`
	Bound   bool   `json:"bound"`
	Updated int64  `json:"updated_at"`
}

func ListGroupChannelBindings(group string) ([]GroupChannelBindingItem, error) {
	groupName := strings.TrimSpace(group)
	if groupName == "" {
		return nil, fmt.Errorf("分组名称不能为空")
	}

	channels := make([]Channel, 0)
	if err := DB.
		Select("id", "name", "type", "status", "models", "created_time").
		Order("created_time desc").
		Find(&channels).Error; err != nil {
		return nil, err
	}

	boundIDs := make([]string, 0)
	groupCol := `"group"`
	if err := DB.Model(&Ability{}).
		Distinct("channel_id").
		Where(groupCol+" = ?", groupName).
		Pluck("channel_id", &boundIDs).Error; err != nil {
		return nil, err
	}
	boundSet := make(map[string]struct{}, len(boundIDs))
	for _, id := range boundIDs {
		normalized := strings.TrimSpace(id)
		if normalized == "" {
			continue
		}
		boundSet[normalized] = struct{}{}
	}

	items := make([]GroupChannelBindingItem, 0, len(channels))
	for _, channel := range channels {
		_, bound := boundSet[channel.Id]
		items = append(items, GroupChannelBindingItem{
			Id:      channel.Id,
			Name:    strings.TrimSpace(channel.Name),
			Type:    channel.Type,
			Status:  channel.Status,
			Models:  strings.TrimSpace(channel.Models),
			Bound:   bound,
			Updated: channel.CreatedTime,
		})
	}
	return items, nil
}

func ReplaceGroupChannelBindings(group string, channelIDs []string) error {
	groupName := strings.TrimSpace(group)
	if groupName == "" {
		return fmt.Errorf("分组名称不能为空")
	}

	groupCatalog := GroupCatalog{}
	if err := DB.Where("name = ?", groupName).First(&groupCatalog).Error; err != nil {
		return err
	}

	normalizedChannelIDs := normalizeChannelIDList(channelIDs)

	channelsByID := make(map[string]Channel, len(normalizedChannelIDs))
	if len(normalizedChannelIDs) > 0 {
		channels := make([]Channel, 0)
		if err := DB.
			Select("id", "name", "status", "models", "priority").
			Where("id IN ?", normalizedChannelIDs).
			Find(&channels).Error; err != nil {
			return err
		}
		for _, channel := range channels {
			channelsByID[channel.Id] = channel
		}
		if len(channelsByID) != len(normalizedChannelIDs) {
			missing := make([]string, 0)
			for _, id := range normalizedChannelIDs {
				if _, ok := channelsByID[id]; !ok {
					missing = append(missing, id)
				}
			}
			sort.Strings(missing)
			return fmt.Errorf("渠道不存在: %s", strings.Join(missing, ", "))
		}
	}

	abilities := make([]Ability, 0)
	for _, id := range normalizedChannelIDs {
		channel := channelsByID[id]
		models := normalizeModelNames(strings.Split(channel.Models, ","))
		for _, modelName := range models {
			abilities = append(abilities, Ability{
				Group:     groupName,
				Model:     modelName,
				ChannelId: channel.Id,
				Enabled:   channel.Status == ChannelStatusEnabled,
				Priority:  channel.Priority,
			})
		}
	}

	groupCol := `"group"`
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(groupCol+" = ?", groupName).Delete(&Ability{}).Error; err != nil {
			return err
		}
		if len(abilities) == 0 {
			return nil
		}
		return tx.Create(&abilities).Error
	})
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
