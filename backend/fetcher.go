package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
	Sleep []struct {
		MinutesAsleep int `json:"minutesAsleep"` // 分
	} `json:"sleep"`
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

// 体重とBMI
type FitbitWeightRangeResponse struct {
	Weight []struct {
		BMI    float64 `json:"bmi"`
		Date   string  `json:"date"`
		Weight float64 `json:"weight"`
	} `json:"weight"`
}

// 自動でフェッチする
func (app *App) startAutoFetch(ctx context.Context, auth FitbitAuth) {
	ticker := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-ticker.C:
			ts, err := app.getAuthenticatedSource(ctx)
			if err == nil {
				start := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
				end := time.Now().Format("2006-01-02")
				app.fetchRangeData(ctx, ts, start, end)
			} else {
				fmt.Println(err)
			}
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
	// 安定している v1.2 を使用し、Summary の合計値を採用
	url := fmt.Sprintf("https://api.fitbit.com/1.2/user/-/sleep/date/%s.json", date)
	var res FitbitSleepResponse
	if err := fetchFitbitAPI(client, url, &res); err != nil {
		return err, 0
	}
	// Summaryから合計分を返す
	return nil, res.Summary.TotalMinutesAsleep
}

// 1日の体重データを取得する関数
func getDayWeight(client *http.Client, date string) (error, float64) {
	// 期間ではなく 1日指定 のURL
	weightURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/body/log/weight/date/%s.json", date)

	var weightRes struct {
		Weight []struct {
			Weight float64 `json:"weight"`
		} `json:"weight"`
	}

	if err := fetchFitbitAPI(client, weightURL, &weightRes); err != nil {
		return err, 0
	}

	// データが存在すれば返す（1日に複数回測った場合は最新のものを採用）
	if len(weightRes.Weight) > 0 {
		latestWeight := weightRes.Weight[len(weightRes.Weight)-1].Weight
		return nil, latestWeight
	}

	// 記録がない日は 0 を返す
	return nil, 0
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
