// src/api/client.ts
const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

export async function apiFetch<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${BASE_URL}${endpoint}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!response.ok) {
    // ここで共通のエラーハンドリング（ログ出力や認証切れチェック）を行う
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || 'API通信に失敗しました');
  }

  return response.json();
}
