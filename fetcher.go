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

// 歩数（期間用）
type FitbitStepsResponse struct {
	ActivitiesSteps []struct {
		DateTime string `json:"dateTime"`
		Value    string `json:"value"` // なぜか文字列で返ってくるので注意
	} `json:"activities-steps"`
}

// 自動でフェッチする
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
			today := time.Now().Format("2006-01-02")
			fetchOneDayData(ctx, ts, today)
		case <-ctx.Done():
			return
		}
	}
}

// カロリーとステップを取得
func getDayStepCalorie(client *http.Client, date string) (err error, steps int, calories int, distance float64) {
	reqURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/date/%s.json", date)
	if resp, err := client.Get(reqURL); err == nil {
		var res FitbitActivityResponse
		json.NewDecoder(resp.Body).Decode(&res)
		steps = res.Summary.Steps
		calories = res.Summary.Calories
		for _, d := range res.Summary.Distances {
			if d.Activity == "total" {
				distance = d.Distance
			}
		}
		resp.Body.Close()
	}
	return
}

// 睡眠の情報を取得
func getDaySleep(client *http.Client, date string) (err error, sleep int) {
	sleepURL := fmt.Sprintf("https://api.fitbit.com/1.2/user/-/sleep/date/%s.json", date)
	if resp, err := client.Get(sleepURL); err == nil {
		var res FitbitSleepResponse
		json.NewDecoder(resp.Body).Decode(&res)
		sleep = res.Summary.TotalMinutesAsleep
		resp.Body.Close()
	}
	return err, sleep
}

func getDayHeartRateRest(client *http.Client, date string) (err error, rest int) {
	// 安静時心拍数の取得
	hrURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/heart/date/%s/1d.json", date)
	hrResp, err := client.Get(hrURL)
	if err == nil {
		defer hrResp.Body.Close()
		var hrResult FitbitHeartResponse
		if err := json.NewDecoder(hrResp.Body).Decode(&hrResult); err == nil {
			if len(hrResult.ActivitiesHeart) > 0 {
				// 安静時心拍数をセット
				rest = hrResult.ActivitiesHeart[0].Value.RestingHeartRate
			}
		}
	}
	return
}

// ある日のデータを取得する
func fetchOneDayData(ctx context.Context, ts oauth2.TokenSource, date string) {
	client := oauth2.NewClient(ctx, ts)

	activity := DailyActivity{Date: date}

	// 1. アクティビティ (歩数・カロリー・距離) 取得
	err, steps, calories, distance := getDayStepCalorie(client, date)
	if err == nil {
		activity.Steps = steps
		activity.Calories = calories
		activity.Distance = distance
	}

	// 2. 安静時心拍数 取得
	err, heartRateRest := getDayHeartRateRest(client, date)
	if err == nil {
		activity.HeartRateRest = heartRateRest
	}

	// 3. 睡眠データ 取得
	err, sleep := getDaySleep(client, date)
	if err == nil {
		activity.SleepMinutes = sleep
	}

	// 4. DBにまとめてUpsert
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{"steps", "calories", "distance", "sleep_minutes", "updated_at"}),
	}).Create(&activity)

	fmt.Printf("【全部取得完了】 %s: %d歩, %d分睡眠\n", date, activity.Steps, activity.SleepMinutes)
}

// 期間指定
func fetchRangeData(ctx context.Context, ts oauth2.TokenSource, start string, end string) {
	client := oauth2.NewClient(ctx, ts)

	// 1. 歩数データの期間取得
	stepsURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/steps/date/%s/%s.json", start, end)
	resp, err := client.Get(stepsURL)
	if err != nil {
		log.Println("過去データ取得失敗:", err)
		return
	}
	defer resp.Body.Close()

	var result FitbitStepsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}

	// 2. 取得した全データをDBに保存
	for _, data := range result.ActivitiesSteps {
		steps, _ := strconv.Atoi(data.Value)
		activity := DailyActivity{
			Date:  data.DateTime,
			Steps: steps,
		}

		// 既存データがある場合は上書きしない（または更新する）設定
		db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{"steps", "updated_at"}),
		}).Create(&activity)
	}

	fmt.Printf("✅ 過去30日分（%s 〜 %s）のデータを同期しました\n", start, end)
}
