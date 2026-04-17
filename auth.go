package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/fitbit"
)

var conf = &oauth2.Config{
	Scopes:      []string{"activity", "heartrate", "sleep", "profile"},
	RedirectURL: "http://localhost:8080/callback",
	Endpoint:    fitbit.Endpoint,
}

// .envを呼び出します。
func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("読み込み出来ませんでした: %v", err)
	}

	// .envの SAMPLE_MESSAGEを取得して、messageに代入します。
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	conf.ClientID = clientID
	conf.ClientSecret = clientSecret
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	url := conf.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "トークン交換失敗", http.StatusInternalServerError)
		return
	}

	auth := FitbitAuth{
		ID:           1,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	if err := db.Save(&auth).Error; err != nil {
		log.Println("DB保存失敗:", err)
		return
	}

	fmt.Fprintf(w, "保存完了！これでもうブラウザ認証は不要です。")
}
