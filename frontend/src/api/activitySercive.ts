// src/api/activityService.ts
import { apiFetch } from './client';

// 型定義（バックエンドの構造体に合わせる）
export interface DailyActivity {
  date: string;
  steps: number;
  calories: number;
}

// APIからのレスポンス型を定義（LSPの補完が効くようになります）
interface AuthStatus {
  is_authenticated: boolean
  updated_at?: string
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
  }
};
