package main

import (
	"context"
	"encoding/json"
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
	http.Redirect(w, r, "http://localhost:5173", http.StatusTemporaryRedirect)
}

func (a *App) getAuthStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var auth FitbitAuth
	// DBに1件でもレコードがあるか確認
	result := a.DB.First(&auth)

	status := map[string]interface{}{
		"is_authenticated": result.Error == nil,
	}

	if result.Error == nil {
		status["updated_at"] = auth.UpdatedAt
	}

	json.NewEncoder(w).Encode(status)
}

func (a *App) getActivities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// クエリパラメータの取得
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	// バリデーション（簡易版）
	if from == "" || to == "" {
		http.Error(w, "from and to parameters are required", http.StatusBadRequest)
		return
	}

	var activities []DailyActivity
	// PostgreSQLに対して期間指定でクエリを実行
	// "date" カラムが from 以上 to 以下 のデータを取得
	result := a.DB.Where("date BETWEEN ? AND ?", from, to).Order("date asc").Find(&activities)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(activities)
}
