<script setup lang="ts">
import { VueDatePicker } from '@vuepic/vue-datepicker'
import '@vuepic/vue-datepicker/dist/main.css'

// 親コンポーネントへ「検索実行」を通知するための定義
const range = defineModel<[Date, Date]>()
defineProps<{
  modelValue: [Date, Date]
  isLoading: boolean
}>()

const emit = defineEmits<{
  (e: 'search', range: { from: string; to: string }): void
}>()

const handleDateChange = (val: [Date, Date]) => {
  emit('update:modelValue', val)
}

</script>

<template>
  <div class="mt-8 p-6 bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl max-w-sm mx-auto">
    <label class="block text-[10px] font-black uppercase tracking-widest text-slate-500 mb-3 text-center">
      Select Period
    </label>
    <div class="relative">
      <VueDatePicker v-model="range" range dark :enable-time-picker="false" :auto-apply="true"
        @update:model-value="handleDateChange" placeholder="Select Range" />
      <div v-if="isLoading" class="absolute -right-8 top-2">
        <div class="animate-spin h-4 w-4 border-2 border-blue-500 border-t-transparent rounded-full"></div>
      </div>
    </div>
  </div>
</template>
