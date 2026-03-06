package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yeying-community/router/common/helper"
	"gorm.io/gorm"
)

const (
	ChannelCapabilityProfilesTableName = "channel_capability_profiles"
	ChannelCapabilityResponses         = "responses"
)

type ChannelCapabilityProfile struct {
	ChannelId     string `json:"channel_id" gorm:"primaryKey;type:char(36);index"`
	Capability    string `json:"capability" gorm:"primaryKey;type:varchar(32)"`
	ClientProfile string `json:"client_profile" gorm:"primaryKey;type:varchar(64)"`
	Enabled       bool   `json:"enabled" gorm:"not null;default:true"`
	UpdatedAt     int64  `json:"updated_at" gorm:"bigint"`
}

func (ChannelCapabilityProfile) TableName() string {
	return ChannelCapabilityProfilesTableName
}

type ChannelCapabilityProfileRule struct {
	Capability    string `json:"capability"`
	ClientProfile string `json:"client_profile"`
	Enabled       bool   `json:"enabled"`
}

func NormalizeChannelCapabilityName(capability string) string {
	return strings.TrimSpace(strings.ToLower(capability))
}

func NormalizeChannelCapabilityProfileRules(rules []ChannelCapabilityProfileRule) []ChannelCapabilityProfileRule {
	if len(rules) == 0 {
		return []ChannelCapabilityProfileRule{}
	}
	normalized := make([]ChannelCapabilityProfileRule, 0, len(rules))
	seen := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		capability := NormalizeChannelCapabilityName(rule.Capability)
		clientProfile := NormalizeClientProfileName(rule.ClientProfile)
		if capability == "" || clientProfile == "" {
			continue
		}
		if !rule.Enabled {
			continue
		}
		key := capability + "::" + clientProfile
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, ChannelCapabilityProfileRule{
			Capability:    capability,
			ClientProfile: clientProfile,
			Enabled:       true,
		})
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].Capability != normalized[j].Capability {
			return normalized[i].Capability < normalized[j].Capability
		}
		return normalized[i].ClientProfile < normalized[j].ClientProfile
	})
	return normalized
}

func HydrateChannelWithCapabilityProfiles(db *gorm.DB, channel *Channel) error {
	if channel == nil {
		return nil
	}
	return HydrateChannelsWithCapabilityProfiles(db, []*Channel{channel})
}

func HydrateChannelsWithCapabilityProfiles(db *gorm.DB, channels []*Channel) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	channelIDs := make([]string, 0, len(channels))
	normalizedChannels := make([]*Channel, 0, len(channels))
	for _, channel := range channels {
		if channel == nil {
			continue
		}
		channel.Id = strings.TrimSpace(channel.Id)
		if channel.Id == "" {
			channel.SetCapabilityProfiles(nil)
			continue
		}
		channelIDs = append(channelIDs, channel.Id)
		normalizedChannels = append(normalizedChannels, channel)
	}
	if len(normalizedChannels) == 0 {
		return nil
	}
	rowsByChannelID, err := loadChannelCapabilityProfileRowsByChannelIDs(db, channelIDs)
	if err != nil {
		return err
	}
	for _, channel := range normalizedChannels {
		channel.SetCapabilityProfiles(rowsByChannelID[channel.Id])
	}
	return nil
}

func ReplaceChannelCapabilityProfilesWithDB(db *gorm.DB, channelID string, rules []ChannelCapabilityProfileRule) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}
	normalizedChannelID := strings.TrimSpace(channelID)
	if normalizedChannelID == "" {
		return nil
	}
	normalizedRules := NormalizeChannelCapabilityProfileRules(rules)
	now := helper.GetTimestamp()
	rows := make([]ChannelCapabilityProfile, 0, len(normalizedRules))
	for _, rule := range normalizedRules {
		enabled := rule.Enabled
		if !enabled {
			enabled = true
		}
		rows = append(rows, ChannelCapabilityProfile{
			ChannelId:     normalizedChannelID,
			Capability:    rule.Capability,
			ClientProfile: rule.ClientProfile,
			Enabled:       enabled,
			UpdatedAt:     now,
		})
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("channel_id = ?", normalizedChannelID).Delete(&ChannelCapabilityProfile{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

func loadChannelCapabilityProfileRowsByChannelIDs(db *gorm.DB, channelIDs []string) (map[string][]ChannelCapabilityProfileRule, error) {
	rowsByChannelID := make(map[string][]ChannelCapabilityProfileRule)
	normalizedIDs := normalizeTrimmedValuesPreserveOrder(channelIDs)
	if len(normalizedIDs) == 0 {
		return rowsByChannelID, nil
	}
	rows := make([]ChannelCapabilityProfile, 0)
	if err := db.Where("channel_id IN ?", normalizedIDs).
		Order("channel_id asc, capability asc, client_profile asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		channelID := strings.TrimSpace(row.ChannelId)
		capability := NormalizeChannelCapabilityName(row.Capability)
		clientProfile := NormalizeClientProfileName(row.ClientProfile)
		if channelID == "" || capability == "" || clientProfile == "" || !row.Enabled {
			continue
		}
		rowsByChannelID[channelID] = append(rowsByChannelID[channelID], ChannelCapabilityProfileRule{
			Capability:    capability,
			ClientProfile: clientProfile,
			Enabled:       true,
		})
	}
	for channelID, rules := range rowsByChannelID {
		rowsByChannelID[channelID] = NormalizeChannelCapabilityProfileRules(rules)
	}
	return rowsByChannelID, nil
}
