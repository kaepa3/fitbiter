package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"gorm.io/gorm/clause"
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

// ある日のデータを取得する
func (app *App) fetchOneDayData(ctx context.Context, ts oauth2.TokenSource, date string) {
	client := oauth2.NewClient(ctx, ts)

	activity := app.getDayDataIfExist(date)

	// 1. アクティビティ (歩数・カロリー・距離) 取得
	err, steps, calories, distance := getDayStepCalorie(client, date)
	if err == nil {
		activity.Steps = steps
		activity.Calories = calories
		activity.Distance = distance
	} else {
		log.Printf("[%s] 歩数/カロリー取得失敗: %v\n", date, err)
	}

	// 2. 安静時心拍数 取得
	err, heartRateRest := getDayHeartRateRest(client, date)
	if err == nil {
		activity.HeartRateRest = heartRateRest
	} else {
		log.Printf("[%s] 心拍数取得失敗: %v\n", date, err)
	}

	// 3. 睡眠データ 取得
	err, sleep := getDaySleep(client, date)
	if err == nil {
		activity.SleepMinutes = sleep
	} else {
		log.Printf("[%s] 睡眠取得失敗: %v\n", date, err)
	}

	// 3. DBへ保存（Saveを使用）
	// GORMの Save は、IDが存在すればUpdate、なければInsertを自動で判別してくれます
	if err := app.DB.Save(&activity).Error; err != nil {
		log.Printf("[%s] DB保存失敗: %v\n", date, err)
	} else {
		fmt.Printf("【取得・更新完了】 %s: %d歩, %d分睡眠, 安静時心拍%d\n",
			date, activity.Steps, activity.SleepMinutes, activity.HeartRateRest)
	}
}

// 過去の特定期間のデータを一気に取得・保存する（Appのメソッドとして定義）
func (app *App) fetchRangeData(ctx context.Context, ts oauth2.TokenSource, start string, end string) {
	client := oauth2.NewClient(ctx, ts)

	// 1. 歩数（Steps）の取得
	stepsURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/steps/date/%s/%s.json", start, end)
	var stepsRes FitbitStepsRangeResponse
	if err := fetchFitbitAPI(client, stepsURL, &stepsRes); err != nil {
		log.Printf("歩数期間データの取得失敗: %v", err)
	}

	// 2. カロリー（Calories）の取得
	calURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/calories/date/%s/%s.json", start, end)
	var calRes FitbitCaloriesRangeResponse
	if err := fetchFitbitAPI(client, calURL, &calRes); err != nil {
		log.Printf("カロリー期間データの取得失敗: %v", err)
	}
	// 3. 安静時心拍数の取得
	hrURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/heart/date/%s/%s.json", start, end)
	var hrRes FitbitHeartRangeResponse
	if err := fetchFitbitAPI(client, hrURL, &hrRes); err != nil {
		log.Printf("心拍数期間データの取得失敗: %v", err)
	}

	// 4. 睡眠の取得 (バージョン1.2)
	sleepURL := fmt.Sprintf("https://api.fitbit.com/1.2/user/-/sleep/date/%s/%s.json", start, end)
	var sleepRes FitbitSleepRangeResponse
	if err := fetchFitbitAPI(client, sleepURL, &sleepRes); err != nil {
		log.Printf("睡眠期間データの取得失敗: %v", err)
	}

	// 3. データを日付ごとに整理するためのマップを作成
	dailyMap := make(map[string]*DailyActivity)

	// 既存データをDBから一括取得してマップに詰める（データ消失対策）
	var existingActivities []DailyActivity
	app.DB.Where("date >= ? AND date <= ?", start, end).Find(&existingActivities)
	for i := range existingActivities {
		act := existingActivities[i]
		dailyMap[act.Date] = &act
	}

	// 4. APIで取れた歩数をマップにマージ
	for _, data := range stepsRes.ActivitiesSteps {
		if _, exists := dailyMap[data.DateTime]; !exists {
			dailyMap[data.DateTime] = &DailyActivity{Date: data.DateTime}
		}
		val, _ := strconv.Atoi(data.Value)
		dailyMap[data.DateTime].Steps = val
	}

	// 5. APIで取れたカロリーをマップにマージ
	for _, data := range calRes.ActivitiesCalories {
		if _, exists := dailyMap[data.DateTime]; !exists {
			dailyMap[data.DateTime] = &DailyActivity{Date: data.DateTime}
		}
		val, _ := strconv.Atoi(data.Value)
		dailyMap[data.DateTime].Calories = val
	}

	// APIで取れた心拍数をマップにマージ
	for _, data := range hrRes.ActivitiesHeart {
		if _, exists := dailyMap[data.DateTime]; !exists {
			dailyMap[data.DateTime] = &DailyActivity{Date: data.DateTime}
		}
		if data.Value.RestingHeartRate > 0 {
			dailyMap[data.DateTime].HeartRateRest = data.Value.RestingHeartRate
		}
	}

	// APIで取れた睡眠をマップにマージ（昼寝などで複数回ある場合は加算）
	for _, data := range sleepRes.Sleep {
		if _, exists := dailyMap[data.DateOfSleep]; !exists {
			dailyMap[data.DateOfSleep] = &DailyActivity{Date: data.DateOfSleep}
		}
		dailyMap[data.DateOfSleep].SleepMinutes += data.MinutesAsleep
	}
	// 6. マップのデータをスライス（配列）に変換
	var updateTargets []DailyActivity
	for _, act := range dailyMap {
		updateTargets = append(updateTargets, *act)
	}

	// 7. 一括保存（バルクアップサート）
	if len(updateTargets) > 0 {
		err := app.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "date"}},
			// 更新対象に heart_rate_rest と sleep_minutes を追加
			DoUpdates: clause.AssignmentColumns([]string{"steps", "calories", "heart_rate_rest", "sleep_minutes", "updated_at"}),
		}).Create(&updateTargets).Error

		if err != nil {
			log.Println("期間データの一括保存失敗:", err)
		} else {
			fmt.Printf("✅ 過去データ（%s 〜 %s）の同期完了: %d件処理しました\n", start, end, len(updateTargets))
		}
	}
}
