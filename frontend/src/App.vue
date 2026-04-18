<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import TheHeader from './components/TheHeader.vue'
import ActivityChart from './components/ActivityChart.vue'
import ActivitySearchForm from './components/ActivitySearchForm.vue'
import SimpleBarChart from './components/SimpleBarChart.vue'
import { activityService } from './api/activitySercive.ts'


// 状態管理用のリアクティブ変数
const isAuthenticated = ref(false)
const lastUpdated = ref('')
const isLoading = ref(true)

// デフォの日付
const today = new Date()
const lastWeek = new Date()
lastWeek.setDate(today.getDate() - 7)

// range の初期値としてセット
const range = ref<[Date, Date]>([lastWeek, today])

// range の変更を監視して、自動でフェッチ
watch(range, (newRange) => {
  if (newRange && newRange[0] && newRange[1]) {
    fetchActivityData({
      from: newRange[0], to: newRange[1]
    })
  }
}, { deep: true }) // 配列の中身の変化を検知するために deep 指定


// 起動時の認証
onMounted(async () => {
  try {
    const data = await activityService.fetchAuthStatus();
    fetchActivityData({ from: lastWeek, to: today })

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
const fetchActivityData = async (dataRange: { from: Date; to: Date }) => {
  isLoading.value = true
  try {
    const data = await activityService.fetchRange(dataRange.from, dataRange.to)
    activityData.value = data || []
  } catch (error) {
    console.error('Fetch error:', error)
  } finally {
    isLoading.value = false
  }
}

// 今日のデータを取得する
const handleSyncRequest = async () => {

  isLoading.value = true;
  try {
    const data = await activityService.syncToday()
    if (data.status != "success") {
      console.log(data)
    }
    await fetchActivityData({ from: lastWeek, to: today })
  } catch (err) {
    console.error(err);
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div class="min-h-screen bg-slate-950 text-slate-200 antialiased selection:bg-blue-500/30">
    <TheHeader :is-authenticated="isAuthenticated" :last-updated="lastUpdated" />

    <main class="p-8 max-w-6xl mx-auto">
      <template v-if="isAuthenticated">
        <ActivitySearchForm :is-loading="isLoading" v-model="range" @fetch-today="handleSyncRequest" />

        <div class="relative mt-8 p-6 bg-slate-900 border border-slate-700 rounded-xl shadow-2xl min-h-[400px]">
          <h2 class="text-xl font-semibold mb-6 text-green-400">Activity Trend</h2>

          <div v-if="isLoading"
            class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm flex items-center justify-center z-10 rounded-xl">
            <div class="animate-spin h-10 w-10 border-4 border-blue-500 border-t-transparent rounded-full">
            </div>
          </div>

          <ActivityChart v-if="activityData.length > 0" :data="activityData" />
          <div class="grid grid-cols-1 md:grid-cols-2 gap-8 mt-8">

            <div class="p-6 bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl">
              <h3 class="text-amber-400 text-sm font-black uppercase tracking-tighter mb-4">Calories</h3>
              <SimpleBarChart :data="activityData" data-key="calories" label="kcal" color="#fbbf24" />
            </div>

            <div class="p-6 bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl">
              <h3 class="text-indigo-400 text-sm font-black uppercase tracking-tighter mb-4">Sleep (Hours)
              </h3>
              <SimpleBarChart :data="activityData" data-key="sleep_minutes" label="hours" color="#818cf8"
                :is-sleep="true" />
            </div>

          </div>
        </div>
      </template>
    </main>
  </div>
</template>
