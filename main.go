package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"golang.org/x/oauth2"
)

func main() {
	loadEnv()
	initDB()
	// 起動時にDBを確認
	var auth FitbitAuth
	if err := db.First(&auth, 1).Error; err == nil {
		fmt.Println("既にDBにトークンがあります。リフレッシュして使用します。")
		// ここで自動取得ジョブなどを動かす
		go startAutoFetch(context.Background(), auth)
	} else {
		fmt.Println("トークンがありません。ブラウザで http://localhost:8080/login にアクセスしてください。")
	}

	// 1. 認可画面へのリダイレクト
	http.HandleFunc("/login", handleLogin)
	// 2. コールバックの処理
	http.HandleFunc("/callback", handleCallback)
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		ts := conf.TokenSource(ctx, &oauth2.Token{
			AccessToken:  auth.AccessToken,
			RefreshToken: auth.RefreshToken,
			Expiry:       auth.Expiry,
		})
		today := time.Now().Format("2006-01-02")
		fetchOneDayData(ctx, ts, today)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var activities []DailyActivity
		db.Order("date desc").Limit(7).Find(&activities) // 直近1週間分
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, activities)
	})
	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		var a FitbitAuth
		db.First(&a, 1)
		ts := conf.TokenSource(r.Context(), &oauth2.Token{
			AccessToken: a.AccessToken, RefreshToken: a.RefreshToken, Expiry: a.Expiry,
		})
		// 今日と1ヶ月前の日付を計算
		end := time.Now().Format("2006-01-02")
		start := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		fetchRangeData(r.Context(), ts, start, end)
		fmt.Fprint(w, "過去1ヶ月分の同期を開始しました（ログを確認してください）")
	})
	fmt.Println("Server started at http://localhost:8080/login")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
