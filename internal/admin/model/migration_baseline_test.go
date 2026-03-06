package model

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRunMainBaselineMigrationWithDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:main_baseline_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := runMainBaselineMigrationWithDB(db); err != nil {
		t.Fatalf("first baseline run failed: %v", err)
	}
	if err := runMainBaselineMigrationWithDB(db); err != nil {
		t.Fatalf("second baseline run failed: %v", err)
	}

	if !db.Migrator().HasTable(&Channel{}) {
		t.Fatalf("expected channels table to exist")
	}
	if !db.Migrator().HasTable(&ModelProvider{}) {
		t.Fatalf("expected model_providers table to exist")
	}
	if !db.Migrator().HasTable(&ModelProviderModel{}) {
		t.Fatalf("expected provider_models table to exist")
	}
	if !db.Migrator().HasTable(&ChannelModel{}) {
		t.Fatalf("expected channel_models table to exist")
	}
	if !db.Migrator().HasTable(&ChannelProtocolCatalog{}) {
		t.Fatalf("expected channel_protocol table to exist")
	}

	var providerCount int64
	if err := db.Model(&ModelProvider{}).Count(&providerCount).Error; err != nil {
		t.Fatalf("count model providers failed: %v", err)
	}
	if providerCount == 0 {
		t.Fatalf("expected seeded model providers, got 0")
	}

	var protocolCount int64
	if err := db.Model(&ChannelProtocolCatalog{}).Count(&protocolCount).Error; err != nil {
		t.Fatalf("count channel protocols failed: %v", err)
	}
	if protocolCount == 0 {
		t.Fatalf("expected seeded channel protocols, got 0")
	}
}

func TestRunLogBaselineMigrationWithDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:log_baseline_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := runLogBaselineMigrationWithDB(db); err != nil {
		t.Fatalf("run log baseline failed: %v", err)
	}
	if !db.Migrator().HasTable(&Log{}) {
		t.Fatalf("expected logs table to exist")
	}
}
