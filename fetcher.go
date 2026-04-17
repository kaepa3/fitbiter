package main

import (
	"context"
	"encoding/json"
	"fmt"
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

func startAutoFetch(ctx context.Context, auth FitbitAuth) {
	ts := conf.TokenSource(ctx, &oauth2.Token{
		AccessToken:  auth.AccessToken,
		RefreshToken: auth.RefreshToken,
		Expiry:       auth.Expiry,
	})

	ticker := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-ticker.C:
			fetchAllData(ctx, ts)
		case <-ctx.Done():
			return
		}
	}
}

func fetchAllData(ctx context.Context, ts oauth2.TokenSource) {
	client := oauth2.NewClient(ctx, ts)
	today := time.Now().Format("2006-01-02")

	activity := DailyActivity{Date: today}

	// 1. アクティビティ (歩数・カロリー・距離) 取得
	reqURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/date/%s.json", today)
	if resp, err := client.Get(reqURL); err == nil {
		var res FitbitActivityResponse
		json.NewDecoder(resp.Body).Decode(&res)
		activity.Steps = res.Summary.Steps
		activity.Calories = res.Summary.Calories
		for _, d := range res.Summary.Distances {
			if d.Activity == "total" {
				activity.Distance = d.Distance
			}
		}
		resp.Body.Close()
	}

	// 2. 安静時心拍数 取得
	hrURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/heart/date/%s/1d.json", today)
	if resp, err := client.Get(hrURL); err == nil {
		// ※パース処理は省略 (構造体に合わせて実装)
		resp.Body.Close()
	}

	// 3. 睡眠データ 取得
	sleepURL := fmt.Sprintf("https://api.fitbit.com/1.2/user/-/sleep/date/%s.json", today)
	if resp, err := client.Get(sleepURL); err == nil {
		var res FitbitSleepResponse
		json.NewDecoder(resp.Body).Decode(&res)
		activity.SleepMinutes = res.Summary.TotalMinutesAsleep
		resp.Body.Close()
	}

	// 4. DBにまとめてUpsert
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"steps", "calories", "distance", "sleep_minutes", "updated_at"}),
	}).Create(&activity)

	fmt.Printf("【全部取得完了】 %s: %d歩, %d分睡眠\n", today, activity.Steps, activity.SleepMinutes)
}
