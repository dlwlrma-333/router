package model

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRunDropChannelGroupColumnMigrationWithDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:drop_channel_group_column_test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.Exec(`CREATE TABLE channels (
		id TEXT PRIMARY KEY,
		"group" TEXT DEFAULT '',
		models TEXT DEFAULT '',
		status INTEGER DEFAULT 1,
		priority BIGINT DEFAULT 0
	)`).Error; err != nil {
		t.Fatalf("create channels table failed: %v", err)
	}
	if err := db.Exec(`CREATE TABLE group_model_channels (
		"group" TEXT NOT NULL,
		model TEXT NOT NULL,
		channel_id TEXT NOT NULL,
		enabled BOOLEAN DEFAULT TRUE,
		priority BIGINT DEFAULT 0,
		PRIMARY KEY ("group", model, channel_id)
	)`).Error; err != nil {
		t.Fatalf("create group_model_channels table failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO channels (id, "group", models, status, priority) VALUES
		('c1', 'alpha,beta', 'gpt-4,gpt-3.5-turbo', 1, 10),
		('c2', 'beta', 'claude-3-5-sonnet', 2, 20)
	`).Error; err != nil {
		t.Fatalf("insert channels failed: %v", err)
	}

	if err := runDropChannelGroupColumnMigrationWithDB(db); err != nil {
		t.Fatalf("run migration failed: %v", err)
	}

	if db.Migrator().HasColumn("channels", "group") {
		t.Fatalf("expected channels.group to be dropped")
	}

	var count int64
	if err := db.Model(&Ability{}).Count(&count).Error; err != nil {
		t.Fatalf("count abilities failed: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected 5 abilities, got %d", count)
	}

	var disabledCount int64
	if err := db.Model(&Ability{}).Where("enabled = ?", false).Count(&disabledCount).Error; err != nil {
		t.Fatalf("count disabled abilities failed: %v", err)
	}
	if disabledCount != 1 {
		t.Fatalf("expected 1 disabled ability, got %d", disabledCount)
	}
}
