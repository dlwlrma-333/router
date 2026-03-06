package model

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBuildDefaultChannelProtocolCatalog(t *testing.T) {
	items := buildDefaultChannelProtocolCatalog(1700000000)
	if len(items) == 0 {
		t.Fatalf("expected default channel protocols, got empty")
	}
	foundOpenAI := false
	for _, item := range items {
		if item.ProtocolID == 1 && item.Name == "openai" && item.Label == "OpenAI" {
			foundOpenAI = true
		}
	}
	if !foundOpenAI {
		t.Fatalf("expected openai channel protocol in defaults")
	}
}

func TestSyncChannelProtocolCatalogWithDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := syncChannelProtocolCatalogWithDB(db); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if err := syncChannelProtocolCatalogWithDB(db); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	var count int64
	if err := db.Model(&ChannelProtocolCatalog{}).Count(&count).Error; err != nil {
		t.Fatalf("count channel protocols failed: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected seeded channel protocols, got 0")
	}
}
