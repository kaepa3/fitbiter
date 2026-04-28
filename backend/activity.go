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

// 履歴前取得
func (a *App) syncAllHistoryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ts, err := a.getAuthenticatedSource(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	totalSynced := 0
	// 今日から開始
	currentEnd := time.Now()

	for {
		// 30日前を計算
		currentStart := currentEnd.AddDate(0, 0, -30)

		startStr := currentStart.Format("2006-01-02")
		endStr := currentEnd.Format("2006-01-02")

		log.Printf("[BULK] Syncing: %s to %s", startStr, endStr)

		// 指定期間のデータを取得
		count, err := a.fetchRangeData(ctx, ts, startStr, endStr)
		if err != nil {
			log.Printf("[ERROR] Bulk sync failed at %s: %v", startStr, err)
			break
		}

		// カロリーはBMRで下限（1777等）があるため、
		// 保存されたデータの「歩数の合計」が0なら、そこが利用開始前と判断する
		var totalSteps int
		a.DB.Model(&DailyActivity{}).
			Where("date BETWEEN ? AND ?", startStr, endStr).
			Select("SUM(steps)").
			Row().Scan(&totalSteps)

		if totalSteps == 0 {
			log.Printf("[FINISH] 指定期間 %s 〜 %s の歩数が0のため、同期を終了します", startStr, endStr)
			break
		}

		totalSynced += count

		// 次のループのために、終了日をさらに30日前にずらす
		currentEnd = currentStart.AddDate(0, 0, -1)

		// 🚨 安全策: あまりに古いデータ（例: 10年以上前）やループ回数に上限を設けるのもあり
		if currentEnd.Year() < 2010 {
			break
		}

		// Fitbit APIのレートリミットを考慮し、少し待機を入れるとより安全
		time.Sleep(500 * time.Millisecond)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"success", "total_synced": %d}`, totalSynced)
}

func (a *App) syncTodayHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. ソースの取得
	ts, err := a.getAuthenticatedSource(ctx)
	if err != nil {
		log.Printf("[ERROR] Auth failed: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized) // 401を返す
		return
	}

	// 2. 日付を日本時間に固定 (JST)
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	today := time.Now().In(jst).Format("2006-01-02")

	// 3. fetchOneDayData の戻り値を確認するように変更（前述の error を返す修正とセット）
	err = a.fetchOneDayData(ctx, ts, today)
	if err != nil {
		log.Printf("[ERROR] Fetch failed for %s: %v", today, err)
		// トークンエラー（invalid_grant等）なら401、それ以外は500
		http.Error(w, "Sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success"}`))
}

// ある日のデータを取得する
func (app *App) fetchOneDayData(ctx context.Context, ts oauth2.TokenSource, date string) error {
	client := oauth2.NewClient(ctx, ts)

	activity := app.getDayDataIfExist(date)

	// 1. アクティビティ (歩数・カロリー・距離) 取得
	err, steps, calories, distance := getDayStepCalorie(client, date)
	if err == nil {
		activity.Steps = steps
		activity.Calories = calories
		activity.Distance = distance
	} else {
		return fmt.Errorf("[%s] 歩数/カロリー取得失敗: %v\n", date, err)
	}

	// 2. 安静時心拍数 取得
	err, heartRateRest := getDayHeartRateRest(client, date)
	if err == nil {
		activity.HeartRateRest = heartRateRest
	} else {
		return fmt.Errorf("[%s] 心拍数取得失敗: %v\n", date, err)
	}

	// 3. 睡眠データ 取得
	err, sleep := getDaySleep(client, date)
	if err == nil {
		activity.SleepMinutes = sleep
	} else {
		return fmt.Errorf("[%s] 睡眠取得失敗: %v\n", date, err)
	}

	// 4. 【追加】体重データ 取得
	err, weight := getDayWeight(client, date)
	if err == nil && weight > 0 {
		activity.Weight = weight
	} else if err != nil {
		// 体重は取れない日もあるため、エラーで止めずにログだけ出す
		log.Printf("[%s] 体重取得失敗または記録なし: %v\n", date, err)
	}

	// 3. DBへ保存（Saveを使用）
	// GORMの Save は、IDが存在すればUpdate、なければInsertを自動で判別してくれます
	if err := app.DB.Save(&activity).Error; err != nil {
		log.Printf("[%s] DB保存失敗: %v\n", date, err)
	}
	log.Printf("【取得・更新完了】 %s: %d歩, %d分睡眠, 安静時心拍%d\n",
		date, activity.Steps, activity.SleepMinutes, activity.HeartRateRest)
	return nil
}

// 過去の特定期間のデータを一気に取得・保存する（Appのメソッドとして定義）
func (app *App) fetchRangeData(ctx context.Context, ts oauth2.TokenSource, start string, end string) (int, error) {
	client := oauth2.NewClient(ctx, ts)

	// 取得チェック
	var existingDates []string
	app.DB.Model(&DailyActivity{}).
		Where("date BETWEEN ? AND ?", start, end).
		Pluck("date", &existingDates)

	// 全ての日付がすでに存在していれば、APIを叩かずに終了
	// 30日分すべて揃っているかチェック（簡略化のため件数で比較）
	if len(existingDates) >= 31 { // 30日+α
		return 0, nil
	}

	// 1. 歩数（Steps）の取得
	stepsURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/steps/date/%s/%s.json", start, end)
	var stepsRes FitbitStepsRangeResponse
	if err := fetchFitbitAPI(client, stepsURL, &stepsRes); err != nil {
		return 0, fmt.Errorf("歩数期間データの取得失敗: %v", err)
	}

	// 2. カロリー（Calories）の取得
	calURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/calories/date/%s/%s.json", start, end)
	var calRes FitbitCaloriesRangeResponse
	if err := fetchFitbitAPI(client, calURL, &calRes); err != nil {
		return 0, fmt.Errorf("カロリー期間データの取得失敗: %v", err)
	}
	// 3. 安静時心拍数の取得
	hrURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/activities/heart/date/%s/%s.json", start, end)
	var hrRes FitbitHeartRangeResponse
	if err := fetchFitbitAPI(client, hrURL, &hrRes); err != nil {
		return 0, fmt.Errorf("心拍数期間データの取得失敗: %v", err)
	}

	// 4. 睡眠の取得 (バージョン1)
	sleepURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/sleep/date/%s/%s.json", start, end)
	var sleepRes FitbitSleepRangeResponse
	if err := fetchFitbitAPI(client, sleepURL, &sleepRes); err != nil {
		return 0, fmt.Errorf("睡眠期間データの取得失敗: %v", err)
	}
	weightURL := fmt.Sprintf("https://api.fitbit.com/1/user/-/body/log/weight/date/%s/%s.json", start, end)
	var weightRes FitbitWeightRangeResponse
	if err := fetchFitbitAPI(client, weightURL, &weightRes); err != nil {
		return 0, fmt.Errorf("体重データの取得失敗: %v", err)
	}

	// 3. データを日付ごとに整理するためのマップを作成
	dailyMap := make(map[string]*DailyActivity)

	// 既存データをDBから一括取得してマップに詰める（データ消失対策）
	var existingActivities []DailyActivity
	app.DB.Where("date >= ? AND date <= ?", start, end).Find(&existingActivities)
	for i := range existingActivities {
		act := existingActivities[i]
		act.SleepMinutes = 0
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
	for _, s := range sleepRes.Sleep {
		if act, ok := dailyMap[s.DateOfSleep]; ok {
			act.SleepMinutes = s.MinutesAsleep
		}
	}
	// 取得した体重データを dailyMap に流し込む
	for _, w := range weightRes.Weight {
		if act, ok := dailyMap[w.Date]; ok {
			act.Weight = w.Weight
		}
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
			DoUpdates: clause.AssignmentColumns([]string{"steps", "calories", "distance", "heart_rate_rest", "sleep_minutes", "weight", "updated_at"}),
		}).Create(&updateTargets).Error
		if err != nil {
			return 0, fmt.Errorf("期間データの一括保存失敗:%v", err)
		}
	}
	// 実際に保存・更新した日数を返す
	return len(updateTargets), nil
}
