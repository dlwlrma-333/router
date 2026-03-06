package model

import (
	"sort"
	"strings"

	"github.com/yeying-community/router/common/helper"
)

const (
	ClientProfilesTableName = "client_profiles"

	ClientProfileCodexCLI   = "codex_cli"
	ClientProfileClaudeCode = "claude_code"
	ClientProfileGeminiCLI  = "gemini_cli"
	ClientProfileGenericAPI = "generic_api"
	ClientProfileAny        = "any"

	ClientProfileMatchModeContains = "contains"
	ClientProfileMatchModePrefix   = "prefix"
	ClientProfileMatchModeExact    = "exact"
	ClientProfileMatchModeFallback = "fallback"
	ClientProfileMatchModeWildcard = "wildcard"
)

type ClientProfile struct {
	Name        string `json:"name" gorm:"primaryKey;type:varchar(64)"`
	DisplayName string `json:"display_name" gorm:"type:varchar(128);default:''"`
	MatchMode   string `json:"match_mode" gorm:"type:varchar(32);default:'contains'"`
	MatchValue  string `json:"match_value" gorm:"type:text"`
	SortOrder   int    `json:"sort_order" gorm:"type:int;not null;default:1000"`
	Enabled     bool   `json:"enabled" gorm:"not null;default:true"`
	UpdatedAt   int64  `json:"updated_at" gorm:"bigint"`
}

func (ClientProfile) TableName() string {
	return ClientProfilesTableName
}

func NormalizeClientProfileName(name string) string {
	normalized := strings.TrimSpace(strings.ToLower(name))
	switch normalized {
	case "codex", "codex-cli":
		return ClientProfileCodexCLI
	case "claude", "claude-code":
		return ClientProfileClaudeCode
	case "gemini", "gemini-cli":
		return ClientProfileGeminiCLI
	case "generic", "generic-api", "api":
		return ClientProfileGenericAPI
	case "all", "*":
		return ClientProfileAny
	default:
		return normalized
	}
}

func BuildDefaultClientProfiles(now int64) []ClientProfile {
	if now <= 0 {
		now = helper.GetTimestamp()
	}
	return []ClientProfile{
		{
			Name:        ClientProfileCodexCLI,
			DisplayName: "Codex CLI",
			MatchMode:   ClientProfileMatchModeContains,
			MatchValue:  "codex-cli",
			SortOrder:   10,
			Enabled:     true,
			UpdatedAt:   now,
		},
		{
			Name:        ClientProfileClaudeCode,
			DisplayName: "Claude Code",
			MatchMode:   ClientProfileMatchModeContains,
			MatchValue:  "claude-code",
			SortOrder:   20,
			Enabled:     true,
			UpdatedAt:   now,
		},
		{
			Name:        ClientProfileGeminiCLI,
			DisplayName: "Gemini CLI",
			MatchMode:   ClientProfileMatchModeContains,
			MatchValue:  "gemini-cli",
			SortOrder:   30,
			Enabled:     true,
			UpdatedAt:   now,
		},
		{
			Name:        ClientProfileGenericAPI,
			DisplayName: "Generic API",
			MatchMode:   ClientProfileMatchModeFallback,
			MatchValue:  "",
			SortOrder:   900,
			Enabled:     true,
			UpdatedAt:   now,
		},
		{
			Name:        ClientProfileAny,
			DisplayName: "Any Client",
			MatchMode:   ClientProfileMatchModeWildcard,
			MatchValue:  "",
			SortOrder:   1000,
			Enabled:     true,
			UpdatedAt:   now,
		},
	}
}

func NormalizeClientProfiles(profiles []ClientProfile) []ClientProfile {
	if len(profiles) == 0 {
		return []ClientProfile{}
	}
	normalized := make([]ClientProfile, 0, len(profiles))
	seen := make(map[string]struct{}, len(profiles))
	for _, profile := range profiles {
		name := NormalizeClientProfileName(profile.Name)
		if name == "" {
			name = NormalizeClientProfileName(profile.DisplayName)
		}
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		profile.Name = name
		profile.DisplayName = strings.TrimSpace(profile.DisplayName)
		if profile.DisplayName == "" {
			profile.DisplayName = name
		}
		profile.MatchMode = strings.TrimSpace(strings.ToLower(profile.MatchMode))
		profile.MatchValue = strings.TrimSpace(strings.ToLower(profile.MatchValue))
		if profile.MatchMode == "" {
			profile.MatchMode = ClientProfileMatchModeContains
		}
		normalized = append(normalized, profile)
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].SortOrder != normalized[j].SortOrder {
			return normalized[i].SortOrder < normalized[j].SortOrder
		}
		return normalized[i].Name < normalized[j].Name
	})
	return normalized
}

func ResolveClientProfileByUserAgent(userAgent string, profiles []ClientProfile) string {
	normalizedUA := strings.TrimSpace(strings.ToLower(userAgent))
	fallback := ""
	for _, profile := range NormalizeClientProfiles(profiles) {
		if !profile.Enabled {
			continue
		}
		matchValue := strings.TrimSpace(strings.ToLower(profile.MatchValue))
		switch profile.MatchMode {
		case ClientProfileMatchModeExact:
			if matchValue != "" && normalizedUA == matchValue {
				return profile.Name
			}
		case ClientProfileMatchModePrefix:
			if matchValue != "" && strings.HasPrefix(normalizedUA, matchValue) {
				return profile.Name
			}
		case ClientProfileMatchModeContains:
			if matchValue != "" && strings.Contains(normalizedUA, matchValue) {
				return profile.Name
			}
		case ClientProfileMatchModeFallback:
			if fallback == "" {
				fallback = profile.Name
			}
		}
	}
	return fallback
}
