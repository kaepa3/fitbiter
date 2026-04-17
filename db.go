package main

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// トークン情報を保存する構造体
type FitbitAuth struct {
	ID           uint      `gorm:"primaryKey"`
	AccessToken  string    `gorm:"not null"`
	RefreshToken string    `gorm:"not null"`
	Expiry       time.Time `gorm:"not null"`
	UpdatedAt    time.Time
}

// 日ごとの活動量を保存するテーブル
type DailyActivity struct {
	ID            uint    `gorm:"primaryKey"`
	Date          string  `gorm:"uniqueIndex"`
	Steps         int     // 歩数
	Calories      int     // 消費カロリー
	Distance      float64 // 距離(km)
	HeartRateRest int     // 安静時心拍数
	SleepMinutes  int     // 睡眠時間(分)
	UpdatedAt     time.Time
}

// DBの初期化
func initDB() *gorm.DB {
	dsn := "host=localhost user=user password=password dbname=fitbit_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB接続失敗:", err)
	}
	db.AutoMigrate(&FitbitAuth{}, &DailyActivity{})
	return db
}
