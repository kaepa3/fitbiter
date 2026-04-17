package main

import (
	"log"
	"os"
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
	ID            uint      `gorm:"primaryKey" json:"id"`
	Date          string    `gorm:"uniqueIndex" json:"date"`
	Steps         int       `json:"steps"`
	Calories      int       `json:"calories"`
	Distance      float64   `json:"distance"`
	HeartRateRest int       `json:"heart_rate_rest"`
	SleepMinutes  int       `json:"sleep_minutes"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// DBの初期化
func initDB() *gorm.DB {
	// 環境変数 DB_DSN から接続情報を取得する
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// 環境変数がない場合はローカル（localhost）をデフォルトにする
		dsn = "host=localhost user=user password=password dbname=fitbit_db port=5432 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("DB接続失敗:", err)
	}
	db.AutoMigrate(&FitbitAuth{}, &DailyActivity{})
	return db
}
