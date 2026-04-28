<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import TheHeader from './components/TheHeader.vue'
import ActivityChart from './components/ActivityChart.vue'
import ActivitySearchForm from './components/ActivitySearchForm.vue'
import SimpleBarChart from './components/SimpleBarChart.vue'
import { activityService, type DailyActivity } from './api/activitySercive.ts'


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
const activityData = ref<DailyActivity[]>([]) // 取得したデータを格納
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
//sub
const currentView = ref<'dashboard' | 'maintenance'>('dashboard')
const isBulkSyncing = ref(false)

const execBulkSync = async () => {
    if (!confirm('過去の全データを同期します。よろしいですか？')) return

    isBulkSyncing.value = true
    try {
        // バックエンド側で「0件になるまでループ」するAPIを叩く
        await activityService.syncAllHistory()
        alert('全ての過去データの同期が完了しました。')
    } catch (error) {
        console.error('Bulk sync error:', error)
        alert('同期中にエラーが発生しました。詳細はログを確認してください。')
    } finally {
        isBulkSyncing.value = false
    }
}
</script>

<template>
    <div class="min-h-screen bg-slate-950 text-slate-200 antialiased selection:bg-blue-500/30">
        <TheHeader :is-authenticated="isAuthenticated" :last-updated="lastUpdated" v-model:currentView="currentView" />

        <template v-if="currentView === 'dashboard'">
            <main class="p-8 max-w-6xl mx-auto">
                <template v-if="isAuthenticated">
                    <ActivitySearchForm :is-loading="isLoading" v-model="range" @fetch-today="handleSyncRequest" />

                    <div
                        class="relative mt-8 p-6 bg-slate-900 border border-slate-700 rounded-xl shadow-2xl min-h-[400px]">
                        <h2 class="text-xl font-semibold mb-6 text-green-400">Activity Trend</h2>

                        <div v-if="isLoading"
                            class="absolute inset-0 bg-slate-900/50 backdrop-blur-sm flex items-center justify-center z-10 rounded-xl">
                            <div
                                class="animate-spin h-10 w-10 border-4 border-blue-500 border-t-transparent rounded-full">
                            </div>
                        </div>

                        <ActivityChart v-if="activityData.length > 0" :data="activityData" />
                        <div class="grid grid-cols-1 md:grid-cols-2 gap-8 mt-8">

                            <div class="p-6 bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl">
                                <h3 class="text-amber-400 text-sm font-black uppercase tracking-tighter mb-4">Calories
                                </h3>
                                <SimpleBarChart :data="activityData" data-key="calories" label="kcal" color="#fbbf24" />
                            </div>

                            <div class="p-6 bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl">
                                <h3 class="text-indigo-400 text-sm font-black uppercase tracking-tighter mb-4">Sleep
                                    (Hours)
                                </h3>
                                <SimpleBarChart :data="activityData" data-key="sleep_minutes" label="hours"
                                    color="#818cf8" :is-sleep="true" />
                            </div>

                            <div class="p-6 bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl">
                                <h3 class="text-rose-400 text-sm font-black uppercase tracking-tighter mb-4">
                                    Weight (kg)
                                </h3>
                                <SimpleBarChart :data="activityData" data-key="weight" label="kg" color="#fb7185" />
                            </div>

                        </div>
                    </div>
                </template>
            </main>
        </template>
        <template v-else-if="currentView === 'maintenance'">
            <div class="max-w-xl mx-auto space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
                <div class="text-center space-y-2">
                    <h2 class="text-4xl font-black italic tracking-tighter text-amber-500">
                        SYSTEM MAINTENANCE
                    </h2>
                    <p class="text-slate-500 text-sm font-bold uppercase tracking-[0.2em]">
                        Deep History Sync
                    </p>
                </div>

                <div class="p-1 bg-gradient-to-br from-amber-500/20 to-transparent rounded-3xl">
                    <div
                        class="p-8 bg-slate-900 border border-slate-800 rounded-[calc(1.5rem-1px)] shadow-2xl space-y-8">

                        <div class="flex items-center gap-4 border-b border-slate-800 pb-6">
                            <div class="h-12 w-12 bg-amber-500/10 rounded-xl flex items-center justify-center">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-amber-500" fill="none"
                                    viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                        d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                                </svg>
                            </div>
                            <div>
                                <h3 class="text-xl font-bold text-white">Full History Retrieval</h3>
                                <p class="text-slate-400 text-xs mt-1">過去に遡って全てのデータを同期します</p>
                            </div>
                        </div>

                        <div
                            class="text-slate-400 text-sm leading-relaxed bg-slate-950/50 p-6 rounded-xl border border-slate-800 space-y-3">
                            <p>
                                この操作は、本日から過去に向かって順番にデータを取得し、<span
                                    class="text-amber-500 font-bold">取得件数が0件になるまで</span>自動で継続します。
                            </p>
                            <p class="text-xs text-slate-500">
                                ※Fitbit APIの制限により、数年分のデータがある場合は完了まで数分かかることがあります。
                            </p>
                        </div>

                        <button @click="execBulkSync" :disabled="isBulkSyncing"
                            class="group relative w-full overflow-hidden py-6 bg-amber-600 hover:bg-amber-500 disabled:bg-slate-800 disabled:opacity-50 text-white font-black rounded-2xl transition-all shadow-[0_0_20px_rgba(217,119,6,0.2)] hover:shadow-[0_0_30px_rgba(217,119,6,0.4)] uppercase tracking-[0.2em]">
                            <span v-if="!isBulkSyncing" class="relative z-10 flex items-center justify-center gap-2">
                                Start Deep Sync
                                <svg xmlns="http://www.w3.org/2000/svg"
                                    class="h-5 w-5 group-hover:rotate-180 transition-transform duration-500" fill="none"
                                    viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                                </svg>
                            </span>
                            <span v-else class="relative z-10 flex items-center justify-center gap-3">
                                <div
                                    class="animate-spin h-5 w-5 border-2 border-white border-t-transparent rounded-full">
                                </div>
                                Syncing History...
                            </span>
                        </button>
                    </div>
                </div>

                <p class="text-center text-slate-600 text-[10px] font-medium tracking-widest uppercase">
                    Data recovery mode active
                </p>
            </div>
        </template>
    </div>
</template>
