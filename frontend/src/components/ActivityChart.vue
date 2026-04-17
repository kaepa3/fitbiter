<script setup lang="ts">
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
} from 'chart.js'
import { Line } from 'vue-chartjs'
import { computed } from 'vue'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend
)

const props = defineProps<{
  data: any[]
}>()

// 届いたデータを Chart.js 用の形式に変換
const chartData = computed(() => ({
  labels: props.data.map(d => d.date),
  datasets: [
    {
      label: 'Steps',
      backgroundColor: '#3b82f6', // blue-500
      borderColor: '#3b82f6',
      data: props.data.map(d => d.steps),
      tension: 0.3 // グラフを滑らかに
    }
  ]
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  scales: {
    y: {
      beginAtZero: true,
      grid: { color: '#334155' } // slate-700
    },
    x: {
      grid: { display: false }
    }
  },
  plugins: {
    legend: { labels: { color: '#f8fafc' } } // slate-50
  }
}
</script>

<template>
  <div class="h-64 md:h-96">
    <Line :data="chartData" :options="chartOptions" />
  </div>
</template>
