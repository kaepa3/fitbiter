package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// アプリケーション全体で使い回す「依存関係」をまとめた構造体
type App struct {
	DB     *gorm.DB
	Conf   *oauth2.Config
	AppCfg *Config
	Mu     sync.Mutex // リフレッシュ処理排他用のMutexを追加
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("読み込み出来ませんでした: %v", err)
	}

	db := initDB()
	conf := initOAuth()
	appCfg := appCfgInit()

	// 2. Appインスタンスの生成
	app := &App{
		DB:     db,
		Conf:   conf,
		AppCfg: appCfg,
	}
	// 起動時にDBを確認
	var auth FitbitAuth
	if err := db.First(&auth, 1).Error; err == nil {
		fmt.Println("既にDBにトークンがあります。リフレッシュして使用します。")
		// ここで自動取得ジョブなどを動かす
		go app.startAutoFetch(context.Background(), auth)
	} else {
		fmt.Println("トークンがありません。ブラウザで http://localhost:8080/login にアクセスしてください。")
	}
	mux := http.NewServeMux()
	// 1. 認可画面へのリダイレクト
	mux.HandleFunc("/login", app.handleLogin)
	mux.HandleFunc("/api/auth/login", app.handleLogin)
	// 2. コールバックの処理
	mux.HandleFunc("/callback", app.handleCallback)
	mux.HandleFunc("/api/auth/status", app.getAuthStatus)
	mux.HandleFunc("/api/activities", app.getActivities)
	mux.HandleFunc("/api/activities/today/sync", app.syncTodayHandler)
	mux.HandleFunc("/api/activities/all/sync", app.syncAllHistoryHandler)

	fmt.Println("Server started at http://localhost:8080/login")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}
