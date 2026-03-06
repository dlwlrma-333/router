package model

import (
	"strings"

	"gorm.io/gorm"
)

func syncClientProfilesWithDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	defaults := NormalizeClientProfiles(BuildDefaultClientProfiles(0))
	if len(defaults) == 0 {
		return nil
	}
	existing := make([]ClientProfile, 0)
	if err := db.Find(&existing).Error; err != nil {
		return err
	}
	existingByName := make(map[string]ClientProfile, len(existing))
	for _, profile := range existing {
		existingByName[NormalizeClientProfileName(profile.Name)] = profile
	}
	for _, profile := range defaults {
		name := NormalizeClientProfileName(profile.Name)
		if name == "" {
			continue
		}
		current, ok := existingByName[name]
		if ok {
			if strings.TrimSpace(current.DisplayName) == "" {
				current.DisplayName = profile.DisplayName
			}
			if strings.TrimSpace(current.MatchMode) == "" {
				current.MatchMode = profile.MatchMode
			}
			if strings.TrimSpace(current.MatchValue) == "" {
				current.MatchValue = profile.MatchValue
			}
			if current.SortOrder == 0 {
				current.SortOrder = profile.SortOrder
			}
			if err := db.Save(&current).Error; err != nil {
				return err
			}
			continue
		}
		if err := db.Create(&profile).Error; err != nil {
			return err
		}
	}
	return nil
}

func ListEnabledClientProfilesWithDB(db *gorm.DB) ([]ClientProfile, error) {
	if db == nil {
		return nil, nil
	}
	rows := make([]ClientProfile, 0)
	if err := db.Where("enabled = ?", true).
		Order("sort_order asc, name asc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return NormalizeClientProfiles(rows), nil
}
