package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
		fetchAllData(ctx, ts)
	})

	fmt.Println("Server started at http://localhost:8080/login")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
