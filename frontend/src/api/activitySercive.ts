// src/api/activityService.ts
import { apiFetch } from './client';

// バックエンドの DailyActivity 構造体に対応
export interface DailyActivity {
  id?: number;
  date: string; // YYYY-MM-DD
  steps: number;
  calories: number;
  distance: number;
  heart_rate_rest: number;
  sleep_minutes: number;
  updated_at?: string;
  weight: number;
}

// APIの認証ステータス用
export interface AuthStatus {
  is_authenticated: boolean;
  updated_at?: string;
}

// 同期レスポンス用
export interface SyncResponse {
  status: string;
}
// 全取得のレスポンス
export interface SyncAllHistoryResponse {
  status: string;
  total_synced: string;
}

export const activityService = {
  // AuthStatusの取得
  async fetchAuthStatus(): Promise<AuthStatus> {
    return apiFetch<AuthStatus>('/api/auth/status');
  },
  // 期間指定でデータを取得
  async fetchRange(from: Date, to: Date): Promise<DailyActivity[]> {
    const fromStr = from.toISOString().split('T')[0]
    const toStr = to.toISOString().split('T')[0]
    return apiFetch<DailyActivity[]>(`/api/activities?from=${fromStr}&to=${toStr}`);
  },

  // 1日分のデータを更新・取得
  async syncToday(): Promise<{ status: string }> {
    // 戻り値が必要なくても apiFetch を通すことで、エラーチェックが自動で行われる
    return apiFetch<{ status: string }>(`/api/activities/today/sync`);
  },
  // 全データの取得
  async syncAllHistory(): Promise<SyncAllHistoryResponse> {
    // 戻り値が必要なくても apiFetch を通すことで、エラーチェックが自動で行われる
    return apiFetch<SyncAllHistoryResponse>(`/api/activities/all/sync`);
  }
};
