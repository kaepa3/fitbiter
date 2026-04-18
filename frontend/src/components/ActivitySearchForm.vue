<script setup lang="ts">
import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'

// 親コンポーネントへ「検索実行」を通知するための定義
const range = defineModel<[Date, Date]>()

defineProps<{
    isLoading: boolean
}>()

// 同期ボタン用のイベントだけを定義
const emit = defineEmits<{
    (e: 'fetch-today'): void
}>()
</script>

<template>
    <div class="mt-8 p-6 bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl max-w-sm mx-auto">
        <label class="block text-[10px] font-black uppercase tracking-widest text-slate-500 mb-4 text-center">
            Select Period
        </label>

        <div class="flex flex-col items-center gap-4 w-full">

            <div class="w-full">
                <VueDatePicker v-model="range" range dark :auto-apply="true" :time-config="{ enableTimePicker: false }"
                    :formats="{ input: 'yyyy/MM/dd' }" placeholder="Select Range" />
            </div>

            <button @click="emit('fetch-today')" :disabled="isLoading"
                class="w-full px-4 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 disabled:opacity-50 text-white font-bold rounded-xl transition-all duration-200 shadow-lg flex items-center justify-center gap-2">
                <span v-if="!isLoading">Sync Today's Data</span>
                <span v-else>Syncing...</span>

                <div v-if="isLoading"
                    class="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full"></div>
            </button>
        </div>
    </div>
</template>
<style scoped>
/* ライブラリ内部の入力要素をターゲットにする */
:deep(.dp__input) {
    text-align: center;
    /* 左側のアイコンと文字が重ならないようにパディングを調整 */
    padding-inline-start: 35px;
}

/* プレースホルダー（Select Range）も中央寄せにする場合 */
:deep(.dp__input::placeholder) {
    text-align: center;
}
</style>
