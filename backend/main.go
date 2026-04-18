package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// アプリケーション全体で使い回す「依存関係」をまとめた構造体
type App struct {
	DB     *gorm.DB
	Conf   *oauth2.Config
	AppCfg *Config
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
	//------------------------------------------------------------------------
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		ts := conf.TokenSource(ctx, &oauth2.Token{
			AccessToken:  auth.AccessToken,
			RefreshToken: auth.RefreshToken,
			Expiry:       auth.Expiry,
		})
		today := time.Now().Format("2006-01-02")
		app.fetchOneDayData(ctx, ts, today)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var activities []DailyActivity
		db.Order("date desc").Limit(30).Find(&activities) // 直近1週間分
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, activities)
	})
	mux.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		var a FitbitAuth
		db.First(&a, 1)
		ts := conf.TokenSource(r.Context(), &oauth2.Token{
			AccessToken: a.AccessToken, RefreshToken: a.RefreshToken, Expiry: a.Expiry,
		})
		// 今日と1ヶ月前の日付を計算
		end := time.Now().Format("2006-01-02")
		start := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		app.fetchRangeData(r.Context(), ts, start, end)
		fmt.Fprint(w, "過去1ヶ月分の同期を開始しました（ログを確認してください）")
	})
	//------------------------------------------------------------------------

	fmt.Println("Server started at http://localhost:8080/login")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}
