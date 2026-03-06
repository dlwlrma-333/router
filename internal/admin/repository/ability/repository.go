package ability

import (
	"context"
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/yeying-community/router/internal/admin/model"
)

func init() {
	model.BindAbilityRepository(model.AbilityRepository{
		GetRandomSatisfiedChannel: GetRandomSatisfiedChannel,
		AddAbilities:              AddAbilities,
		DeleteAbilities:           DeleteAbilities,
		UpdateAbilities:           UpdateAbilities,
		UpdateAbilityStatus:       UpdateAbilityStatus,
		GetTopChannelByModel:      GetTopChannelByModel,
		GetGroupModels:            GetGroupModels,
	})
}

func GetRandomSatisfiedChannel(group string, modelName string, ignoreFirstPriority bool) (*model.Channel, error) {
	ability := model.Ability{}
	groupCol := `"group"`
	trueVal := "true"

	var channelQuery *gorm.DB
	if ignoreFirstPriority {
		channelQuery = model.DB.Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, modelName)
	} else {
		maxPrioritySubQuery := model.DB.Model(&model.Ability{}).Select("MAX(priority)").Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, modelName)
		channelQuery = model.DB.Where(groupCol+" = ? and model = ? and enabled = "+trueVal+" and priority = (?)", group, modelName, maxPrioritySubQuery)
	}
	if err := channelQuery.Order("RANDOM()").First(&ability).Error; err != nil {
		return nil, err
	}
	channel := model.Channel{Id: ability.ChannelId}
	err := model.DB.First(&channel, "id = ?", ability.ChannelId).Error
	return &channel, err
}

func AddAbilities(channel *model.Channel) error {
	// Channel-group bindings are managed centrally in group management.
	// Channel creation no longer auto-generates abilities.
	if channel == nil {
		return nil
	}
	return nil
}

func listBoundGroupsByChannelID(channelID string) ([]string, error) {
	groupCol := `"group"`
	groups := make([]string, 0)
	err := model.DB.Model(&model.Ability{}).
		Distinct(groupCol).
		Where("channel_id = ?", channelID).
		Pluck(groupCol, &groups).Error
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(groups))
	seen := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		normalized := strings.TrimSpace(group)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result, nil
}

func buildAbilitiesForChannel(channel *model.Channel, groups []string) []model.Ability {
	if channel == nil || len(groups) == 0 {
		return nil
	}
	models := channel.SelectedModelIDs()
	abilities := make([]model.Ability, 0, len(models)*len(groups))
	for _, modelName := range models {
		normalizedModel := strings.TrimSpace(modelName)
		if normalizedModel == "" {
			continue
		}
		for _, group := range groups {
			normalizedGroup := strings.TrimSpace(group)
			if normalizedGroup == "" {
				continue
			}
			ability := model.Ability{
				Group:     normalizedGroup,
				Model:     normalizedModel,
				ChannelId: channel.Id,
				Enabled:   channel.Status == model.ChannelStatusEnabled,
				Priority:  channel.Priority,
			}
			abilities = append(abilities, ability)
		}
	}
	return abilities
}

func DeleteAbilities(channel *model.Channel) error {
	return model.DB.Where("channel_id = ?", channel.Id).Delete(&model.Ability{}).Error
}

func UpdateAbilities(channel *model.Channel) error {
	if channel == nil {
		return nil
	}
	groups, err := listBoundGroupsByChannelID(channel.Id)
	if err != nil {
		return err
	}
	abilities := buildAbilitiesForChannel(channel, groups)
	err = DeleteAbilities(channel)
	if err != nil {
		return err
	}
	if len(abilities) == 0 {
		return nil
	}
	return model.DB.Create(&abilities).Error
}

func UpdateAbilityStatus(channelId string, status bool) error {
	return model.DB.Model(&model.Ability{}).Where("channel_id = ?", channelId).Select("enabled").Update("enabled", status).Error
}

func GetTopChannelByModel(group string, modelName string) (*model.Channel, error) {
	groupCol := `"group"`
	trueVal := "true"

	ability := model.Ability{}
	err := model.DB.Where(groupCol+" = ? and model = ? and enabled = "+trueVal, group, modelName).
		Order("priority desc, channel_id asc").
		First(&ability).Error
	if err != nil {
		return nil, err
	}
	channel := model.Channel{Id: ability.ChannelId}
	err = model.DB.Omit("key").First(&channel, "id = ?", ability.ChannelId).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func GetGroupModels(ctx context.Context, group string) ([]string, error) {
	groupCol := `"group"`
	trueVal := "true"
	var models []string
	err := model.DB.Model(&model.Ability{}).Distinct("model").Where(groupCol+" = ? and enabled = "+trueVal, group).Pluck("model", &models).Error
	if err != nil {
		return nil, err
	}
	sort.Strings(models)
	return models, nil
}
