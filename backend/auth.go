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

// .envを呼び出します。
func initOAuth() *oauth2.Config {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("読み込み出来ませんでした: %v", err)
	}

	return &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"activity", "heartrate", "sleep", "profile"},
		RedirectURL:  "http://localhost:8080/callback",
		Endpoint:     fitbit.Endpoint,
	}
}

func (app *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	url := app.Conf.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *App) handleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := app.Conf.Exchange(context.Background(), code)
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

	if err := app.DB.Save(&auth).Error; err != nil {
		log.Println("DB保存失敗:", err)
		return
	}

	fmt.Fprintf(w, "保存完了！これでもうブラウザ認証は不要です。")
}
