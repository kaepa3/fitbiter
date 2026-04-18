package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/fitbit"
)

// .envを呼び出します。
func initOAuth() *oauth2.Config {
	// リダイレクトURLを環境変数から取得
	redirectURL := os.Getenv("OAUTH_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/callback" // デフォルト
	}
	return &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"activity", "heartrate", "sleep", "profile"},
		RedirectURL:  redirectURL,
		Endpoint:     fitbit.Endpoint,
	}
}

func (app *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("login")
	url := app.Conf.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *App) handleCallback(w http.ResponseWriter, r *http.Request) {
	log.Printf("callback")
	code := r.URL.Query().Get("code")
	// 1. 認可コードをトークンに交換
	token, err := app.Conf.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		http.Error(w, "トークン交換失敗", http.StatusInternalServerError)
		return
	}

	// 2. ID: 1 のレコードを Upsert (Save)
	// 既存の ID:1 があれば更新、なければ作成
	auth := FitbitAuth{
		ID:           1,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	if err := app.DB.Save(&auth).Error; err != nil {
		log.Printf("DB保存失敗: %v", err)
		http.Error(w, "DB保存失敗", http.StatusInternalServerError)
		return
	}

	log.Println("【成功】新しいトークンを取得し、DB(ID:1)を更新しました。")

	// 3. フロントエンド（Vite）へ戻す
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

func (app *App) getAuthenticatedSource(ctx context.Context) (oauth2.TokenSource, error) {
	var auth FitbitAuth
	if err := app.DB.First(&auth, 1).Error; err != nil {
		return nil, err
	}

	// DBから読み出した時点のトークン
	initialToken := &oauth2.Token{
		AccessToken:  auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		Expiry:       auth.Expiry,
	}

	// 🔴 腐りを防ぐポイント:
	// oauth2.ReuseTokenSource を「一番外側」に置くことで、
	// 1回のリフレッシュ成功後、その新しいトークンをメモリに保持し、
	// 同一インスタンス内での 2 回目以降の呼び出しでは通信させないようにします。

	baseSrc := app.Conf.TokenSource(ctx, initialToken)
	updateSrc := &tokenUpdateSource{
		src: baseSrc,
		app: app,
	}

	// initialToken を渡すことで、最初からその値を使い、切れていれば updateSrc を叩く
	return oauth2.ReuseTokenSource(initialToken, updateSrc), nil
}

// DB更新用のラッパー構造体
type tokenUpdateSource struct {
	src oauth2.TokenSource
	app *App
}

func (s *tokenUpdateSource) Token() (*oauth2.Token, error) {
	// 1. リフレッシュ前のトークン状態を確認（セキュリティのため末尾のみ）
	var current FitbitAuth
	s.app.DB.First(&current, 1)

	// 2. 実際に Token() を呼び出す（ここでリフレッシュ通信が発生）
	token, err := s.src.Token()
	if err != nil {
		// ここでエラーが出た場合、Fitbitに送ったトークンが「腐っていた」ことが確定する
		log.Printf("[ERROR-SYNC] Refresh failed. Used token tail: ...%s, Error: %v",
			current.RefreshToken[len(current.RefreshToken)-8:], err)
		return nil, err
	}

	// 4. DB更新
	err = s.app.DB.Model(&FitbitAuth{}).Where("id = ?", 1).Updates(map[string]interface{}{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry,
	}).Error
	if err != nil {
		log.Printf("[ERROR-SYNC] DB Update Failed: %v", err)
	}

	return token, nil
}
