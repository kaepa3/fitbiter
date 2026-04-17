// fetcher_test.go
package main

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// テスト用のAppを生成するヘルパー関数
func setupTestApp() *App {
	// メモリ上で動作する一時的なSQLiteを使用
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&DailyActivity{})

	return &App{
		DB: db,
		// Confは今回使わないのでnilでもOK
	}
}

func TestDB_SaveActivity(t *testing.T) {
	app := setupTestApp()

	// テストデータ
	act := DailyActivity{Date: "2026-04-18", Steps: 10000}
	app.DB.Save(&act)

	// 検証（Assert）
	var saved DailyActivity
	app.DB.First(&saved, "date = ?", "2026-04-18")

	assert.Equal(t, 10000, saved.Steps, "歩数が正しく保存されていること")
}
