package model

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type legacyChannelGroupRow struct {
	Id       string `gorm:"column:id"`
	Group    string `gorm:"column:group"`
	Models   string `gorm:"column:models"`
	Status   int    `gorm:"column:status"`
	Priority *int64 `gorm:"column:priority"`
}

func runDropChannelGroupColumnMigrationWithDB(tx *gorm.DB) error {
	if tx == nil {
		return fmt.Errorf("database handle is nil")
	}
	if !tx.Migrator().HasColumn("channels", "group") {
		return nil
	}

	legacyRows := make([]legacyChannelGroupRow, 0)
	if err := tx.Table("channels").
		Select(`id, "group", models, status, priority`).
		Find(&legacyRows).Error; err != nil {
		return err
	}

	abilities := make([]Ability, 0)
	for _, row := range legacyRows {
		channelID := strings.TrimSpace(row.Id)
		if channelID == "" {
			continue
		}
		groups := parseGroupNamesFromCSV(row.Group)
		models := normalizeModelNames(strings.Split(row.Models, ","))
		for _, group := range groups {
			for _, modelName := range models {
				abilities = append(abilities, Ability{
					Group:     group,
					Model:     modelName,
					ChannelId: channelID,
					Enabled:   row.Status == ChannelStatusEnabled,
					Priority:  row.Priority,
				})
			}
		}
	}
	if len(abilities) > 0 {
		if err := tx.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&abilities).Error; err != nil {
			return err
		}
	}

	if tx.Dialector.Name() == "sqlite" {
		return tx.Exec(`ALTER TABLE channels DROP COLUMN "group"`).Error
	}
	return tx.Migrator().DropColumn("channels", "group")
}
