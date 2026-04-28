<!-- components/BaseChart.vue -->
<template>
    <div class="relative h-full w-full">
        <!-- vue-chartjs を使っている場合の例 -->
        <Chart :type="type" :data="chartData" :options="chartOptions" />
    </div>
</template>

<script setup>
import { Chart } from 'vue-chartjs'
// Chart.js の必要なモジュールの登録などはここで1回だけやる

defineProps({
    type: { type: String, default: 'bar' }, // 'bar', 'line' などを切り替え
    chartData: { type: Object, required: true },
    chartOptions: { type: Object, required: true }
})
</script>
