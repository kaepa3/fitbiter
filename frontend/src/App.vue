<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import TheHeader from './components/TheHeader.vue'
import ActivityChart from './components/ActivityChart.vue'
import ActivitySearchForm from './components/ActivitySearchForm.vue'

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

// デフォの日付
const today = new Date()
const lastWeek = new Date()
lastWeek.setDate(today.getDate() - 7)
// range の初期値としてセット
const range = ref<[Date, Date]>([lastWeek, today])
// range の変更を監視して、自動でフェッチ


watch(range, (newRange) => {
  console.log('Range changed:', newRange) // これがコンソールに出るか確認
  if (newRange && newRange[0] && newRange[1]) {
    fetchActivityData({
      from: newRange[0].toISOString().split('T')[0],
      to: newRange[1].toISOString().split('T')[0]
    })
  }
}, { deep: true }) // 配列の中身の変化を検知するために deep 指定


onMounted(async () => {
  try {
    const response = await fetch(`${API_BASE_URL}/api/auth/status`)
    if (!response.ok) throw new Error('Network response was not ok')

    const data: AuthStatus = await response.json()
    fetchActivityData({
      from: lastWeek.toISOString().split('T')[0],
      to: today.toISOString().split('T')[0]
    })
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

// 引数を受け取る形に変更
const activityData = ref<any[]>([]) // 取得したデータを格納
const fetchActivityData = async (dataRange: { from: string; to: string }) => {
  isLoading.value = true
  try {
    const params = new URLSearchParams(dataRange)
    const response = await fetch(`${API_BASE_URL}/api/activities?${params}`)
    if (!response.ok) throw new Error('Failed to fetch activities')

    const data = await response.json()
    activityData.value = data || []
  } catch (error) {
    console.error('Fetch error:', error)
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-200 antialiased selection:bg-blue-500/30">
    <TheHeader :is-authenticated="isAuthenticated" :last-updated="lastUpdated" />

    <main class="p-8 max-w-6xl mx-auto">
      <template v-if="isAuthenticated">
        <ActivitySearchForm :is-loading="isLoading" v-model="range" />

        <div class="relative mt-8 p-6 bg-slate-900 border border-slate-700 rounded-xl shadow-2xl min-h-[400px]">
          <h2 class="text-xl font-semibold mb-6 text-green-400">Activity Trend</h2>

          <div v-if="isLoading"
            class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm flex items-center justify-center z-10 rounded-xl">
            <div class="animate-spin h-10 w-10 border-4 border-blue-500 border-t-transparent rounded-full"></div>
          </div>

          <ActivityChart v-if="activityData.length > 0" :data="activityData" />
          <div v-else-if="!isLoading" class="text-slate-500 text-center py-20">
            データを選択して Fetch を押してください
          </div>
        </div>
      </template>
    </main>
  </div>
</template>
