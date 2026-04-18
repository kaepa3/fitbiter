package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// 複数データをまとめるための構造体
type FitbitSummary struct {
	Steps     int
	Calories  int
	Distance  float64
	HeartRate int
}

// 睡眠データ用
type FitbitSleepResponse struct {
	Summary struct {
		TotalMinutesAsleep int `json:"totalMinutesAsleep"`
	} `json:"summary"`
}

// アクティビティ概要用 (歩数・カロリー・距離がいっぺんに取れる)
type FitbitActivityResponse struct {
	Summary struct {
		Steps     int `json:"steps"`
		Calories  int `json:"caloriesOut"`
		Distances []struct {
			Activity string  `json:"activity"`
			Distance float64 `json:"distance"`
		} `json:"distances"`
	} `json:"summary"`
}

// 安静時心拍数
type FitbitHeartResponse struct {
	ActivitiesHeart []struct {
		DateTime string `json:"dateTime"`
		Value    struct {
			RestingHeartRate int `json:"restingHeartRate"`
		} `json:"value"`
	} `json:"activities-heart"`
}

// --- 期間取得用の構造体 ---
type TimeSeriesData struct {
	DateTime string `json:"dateTime"`
	Value    string `json:"value"` // 期間APIは数値を文字列で返してくる
}

type FitbitStepsRangeResponse struct {
	ActivitiesSteps []TimeSeriesData `json:"activities-steps"`
}

type FitbitCaloriesRangeResponse struct {
	ActivitiesCalories []TimeSeriesData `json:"activities-calories"`
}
type FitbitHeartRangeResponse struct {
	ActivitiesHeart []struct {
		DateTime string `json:"dateTime"`
		Value    struct {
			RestingHeartRate int `json:"restingHeartRate"`
		} `json:"value"`
	} `json:"activities-heart"`
}

type FitbitSleepRangeResponse struct {
	Sleep []struct {
		DateOfSleep   string `json:"dateOfSleep"`
		MinutesAsleep int    `json:"minutesAsleep"`
	} `json:"sleep"`
}

// 自動でフェッチする
func (app *App) startAutoFetch(ctx context.Context, auth FitbitAuth) {
	ts := app.Conf.TokenSource(ctx, &oauth2.Token{
		AccessToken:  auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		Expiry:       auth.Expiry,
	})

	ticker := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-ticker.C:
			today := time.Now().Format("2006-01-02")
			app.fetchOneDayData(ctx, ts, today)
		case <-ctx.Done():
			return
		}
	}
}

// 共通のAPI呼び出し＆エラーチェック関数
func fetchFitbitAPI(client *http.Client, url string, target interface{}) error {
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("通信エラー: %w", err)
	}
	defer resp.Body.Close()

	// 🚨 必須チェック: 200 OK 以外は弾く
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests { // 429エラー
			// Fitbitはヘッダーにリセットまでの秒数を教えてくれる
			reset := resp.Header.Get("Fitbit-Rate-Limit-Reset")
			return fmt.Errorf("Rate Limit超過(429): リセットまで約 %s 秒", reset)
		}
		return fmt.Errorf("APIエラー: HTTP %d", resp.StatusCode)
	}

	// 成功時のみJSONをデコード（ポインタを渡して書き込んでもらう）
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("JSONパース失敗: %w", err)
	}

	return nil
}

// カロリーとステップを取得
func getDayStepCalorie(client *http.Client, date string) (err error, steps int, calories int, distance float64) {
	url := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/date/%s.json", date)
	var res FitbitActivityResponse

	if err := fetchFitbitAPI(client, url, &res); err != nil {
		return err, 0, 0, 0
	}

	for _, d := range res.Summary.Distances {
		if d.Activity == "total" {
			distance = d.Distance
		}
	}
	return nil, res.Summary.Steps, res.Summary.Calories, distance
}

// 睡眠の情報を取得
func getDaySleep(client *http.Client, date string) (err error, sleep int) {
	url := fmt.Sprintf("https://api.fitbit.com/1.2/user/-/sleep/date/%s.json", date)
	var res FitbitSleepResponse
	if err := fetchFitbitAPI(client, url, &res); err != nil {
		return err, 0 // 失敗した時はエラーだけを返す
	}
	return err, sleep
}

// 安静時心拍数を取得
func getDayHeartRateRest(client *http.Client, date string) (err error, rest int) {
	url := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/heart/date/%s/1d.json", date)
	var res FitbitHeartResponse

	if err := fetchFitbitAPI(client, url, &res); err != nil {
		return err, 0
	}
	if len(res.ActivitiesHeart) > 0 {
		return nil, res.ActivitiesHeart[0].Value.RestingHeartRate
	}
	return fmt.Errorf("心拍データが空です"), 0
}

// 情報取得
func (app *App) getDayDataIfExist(date string) DailyActivity {
	var activity DailyActivity
	if err := app.DB.Where("date = ?", date).First(&activity).Error; err != nil {
		// 見つからない場合は初期化
		activity = DailyActivity{Date: date}
	}
	return activity
}
