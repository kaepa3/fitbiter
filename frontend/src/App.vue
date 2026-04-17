<script setup lang="ts">
import { ref, onMounted } from 'vue'
import ActivityChart from './components/ActivityChart.vue'

// APIからのレスポンス型を定義（LSPの補完が効くようになります）
interface AuthStatus {
  is_authenticated: boolean
  updated_at?: string
}

// 状態管理用のリアクティブ変数
const isAuthenticated = ref(false)
const lastUpdated = ref('')
const isLoading = ref(true)

// バックエンドのURL（Docker Compose環境に合わせて調整）
const API_BASE_URL = 'http://localhost:8080'

onMounted(async () => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/auth/status`)
    if (!response.ok) throw new Error('Network response was not ok')

    const data: AuthStatus = await response.json()

    // 取得したデータを反映
    isAuthenticated.value = data.is_authenticated
    if (data.updated_at) {
      lastUpdated.value = new Date(data.updated_at).toLocaleString()
    }
  } catch (error) {
    console.error('Failed to fetch auth status:', error)
  } finally {
    isLoading.value = false
  }
})

// Fitbit連携を開始する関数
const startAuthorize = () => {
  // バックエンドのログインエンドポイントへリダイレクト
  window.location.href = `${API_BASE_URL}/login`
}
const dateRange = ref({
  from: new Date().toISOString().split('T')[0], // 今日をデフォルトに
  to: new Date().toISOString().split('T')[0]
})

const activityData = ref<any[]>([]) // 取得したデータを格納

const fetchActivityData = async () => {
  isLoading.value = true
  try {
    // クエリパラメータ付きでリクエスト (例: /api/activities?from=2026-04-01&to=2026-04-18)
    const params = new URLSearchParams({
      from: dateRange.value.from,
      to: dateRange.value.to
    })

    const response = await fetch(`${API_BASE_URL}/api/activities?${params}`)
    if (!response.ok) throw new Error('Failed to fetch activities')

    const data = await response.json()
    activityData.value = data // 取得したデータを保持
    console.log('Fetched data:', data)
  } catch (error) {
    console.error('Fetch error:', error)
  } finally {
    isLoading.value = false
  } // ここにバックエンドのAPI（/api/activity 等）を叩く処理を追加していく
}

</script>

<template>
  <div class="p-8 font-sans">
    <h1 class="text-2xl font-bold mb-4">Fitbiter Dashboard</h1>

    <div v-if="isLoading" class="text-gray-500">
      Loading status...
    </div>

    <div v-else>
      <div class="flex items-center gap-2 mb-6">
        <span :class="isAuthenticated ? 'bg-green-500' : 'bg-red-500'"
          class="w-3 h-3 rounded-full animate-pulse"></span>
        <span class="font-mono">
          Status: {{ isAuthenticated ? 'CONNECTED' : 'DISCONNECTED' }}
        </span>
      </div>

      <div v-if="!isAuthenticated">
        <button @click="startAuthorize"
          class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded shadow-lg transition">
          Authorize with Fitbit
        </button>
      </div>

      <div v-else class="text-sm text-gray-400">
        Last synced: {{ lastUpdated }}
      </div>
    </div>
    <div v-if="isAuthenticated" class="mt-8 p-6 bg-slate-900 border border-slate-700 rounded-xl shadow-2xl">
      <h2 class="text-xl font-semibold mb-6 text-blue-400">Activity Data Lookup</h2>

      <div class="flex flex-col md:flex-row gap-6 items-end">
        <div class="flex flex-col gap-2 w-full md:w-auto">
          <label class="text-xs font-bold uppercase tracking-wider text-slate-400">From</label>
          <input type="date" v-model="dateRange.from"
            class="bg-slate-800 border border-slate-600 text-white rounded-lg p-3 focus:ring-2 focus:ring-blue-500 outline-none transition">
        </div>

        <div class="flex flex-col gap-2 w-full md:w-auto">
          <label class="text-xs font-bold uppercase tracking-wider text-slate-400">To</label>
          <input type="date" v-model="dateRange.to"
            class="bg-slate-800 border border-slate-600 text-white rounded-lg p-3 focus:ring-2 focus:ring-blue-500 outline-none transition">
        </div>

        <button @click="fetchActivityData"
          class="w-full md:w-auto bg-blue-600 hover:bg-blue-500 text-white font-bold py-3 px-8 rounded-lg shadow-lg transform active:scale-95 transition">
          Fetch Data
        </button>
      </div>
    </div>
    <div v-if="activityData.length > 0" class="mt-8 p-6 bg-slate-900 border border-slate-700 rounded-xl shadow-2xl">
      <h2 class="text-xl font-semibold mb-6 text-green-400">Activity Trend</h2>
      <ActivityChart :data="activityData" />
    </div>
  </div>
</template>
