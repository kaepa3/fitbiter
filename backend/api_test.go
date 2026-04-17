// api_test.go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchFitbitAPI_Success(t *testing.T) {
	// 1. 偽のFitbitサーバーを立ち上げる
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 期待されるURLパスかチェック
		assert.Equal(t, "/1/user/-/activities/date/2026-04-18.json", r.URL.Path)

		// 偽のJSONレスポンスを返す
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"summary": {"steps": 8500, "caloriesOut": 2100}}`)
	}))
	defer mockServer.Close()

	// 2. 実際にテスト対象の関数を呼ぶ（URLはモックサーバーに向ける）
	var res FitbitActivityResponse
	client := mockServer.Client() // モックに通信できる特殊なクライアント

	err := fetchFitbitAPI(client, mockServer.URL+"/1/user/-/activities/date/2026-04-18.json", &res)

	// 3. 検証
	assert.NoError(t, err)
	assert.Equal(t, 8500, res.Summary.Steps)
	assert.Equal(t, 2100, res.Summary.Calories)
}
