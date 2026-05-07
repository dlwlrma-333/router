package model

import "gorm.io/gorm"

func dropProviderModelCapabilitiesWithDB(tx *gorm.DB) error {
	if tx == nil {
		return nil
	}
	if tx.Migrator().HasTable(ProviderModelsTableName) && tx.Migrator().HasColumn(ProviderModelsTableName, "capabilities") {
		if err := tx.Migrator().DropColumn(ProviderModelsTableName, "capabilities"); err != nil {
			return err
		}
	}
	return nil
}
