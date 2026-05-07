package model

import "gorm.io/gorm"

func migrateGroupModelRoutesTableWithDB(tx *gorm.DB) error {
	if tx == nil {
		return nil
	}
	if tx.Migrator().HasTable("group_model_channels") && !tx.Migrator().HasTable(GroupModelRoutesTableName) {
		if err := tx.Migrator().RenameTable("group_model_channels", GroupModelRoutesTableName); err != nil {
			return err
		}
	}
	return tx.AutoMigrate(&GroupModelRoute{})
}
