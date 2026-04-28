<!-- components/BaseChart.vue -->
<script setup>
import { Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
  PointElement, // ← これを追加！
  LineElement   // ← これを追加！
} from 'chart.js'

// ここにも追加して登録します
ChartJS.register(
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
  PointElement, // ← これを追加！
  LineElement   // ← これを追加！
)

defineProps({
  type: { type: String, default: 'bar' },
  chartData: { type: Object, required: true },
  chartOptions: { type: Object, required: true }
})
</script>

<template>
  <div class="relative h-full w-full min-h-[300px]">
    <!-- 複合グラフでも大元は Bar コンポーネントを使えばOKです -->
    <Bar :data="chartData" :options="chartOptions" />
  </div>
</template>
