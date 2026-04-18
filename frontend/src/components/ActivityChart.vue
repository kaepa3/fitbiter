<script setup lang="ts">
import {
    Chart as ChartJS,
    Title, Tooltip, Legend,
    LineElement, BarElement, PointElement,
    CategoryScale, LinearScale,
    LineController, BarController,
    type ChartData, type ChartOptions // 型をインポート
} from 'chart.js'
import { Chart } from 'vue-chartjs'
import { computed } from 'vue'

ChartJS.register(
    Title, Tooltip, Legend,
    LineElement, BarElement, PointElement,
    CategoryScale, LinearScale,
    LineController, BarController
)

const props = defineProps<{
    data: any[]
}>()

// ChartData 型を明示。 'bar' | 'line' とすることで混合を許可
const chartData = computed<ChartData<'bar' | 'line'>>(() => {
    if (!props.data || props.data.length === 0) {
        return { labels: [], datasets: [] }
    }

    return {
        labels: props.data.map(d => d.date),
        datasets: [
            {
                type: 'bar' as const,
                label: 'Steps',
                data: props.data.map(d => d.steps),
                backgroundColor: 'rgba(59, 130, 246, 0.5)',
                yAxisID: 'y',
            },
            {
                type: 'line' as const,
                label: 'Heart Rate',
                data: props.data.map(d => d.heart_rate_rest),
                borderColor: '#f43f5e',
                backgroundColor: '#f43f5e',
                yAxisID: 'y1',
                tension: 0.4,
            }
        ]
    }
})

// ChartOptions 型を明示
const chartOptions: ChartOptions<'bar' | 'line'> = {
    responsive: true,
    maintainAspectRatio: false,
    scales: {
        y: {
            type: 'linear' as const, // 文字列ではなく型を固定
            display: true,
            position: 'left',
            title: { display: true, text: 'Steps' }
        },
        y1: {
            type: 'linear' as const,
            display: true,
            position: 'right',
            grid: { drawOnChartArea: false },
            title: { display: true, text: 'BPM' },
            suggestedMin: 40,
            suggestedMax: 100
        }
    }
}
</script>

<template>
    <div class="h-[400px]">
        <Chart v-if="chartData.labels && chartData.labels.length > 0" type="line" :data="chartData"
            :options="chartOptions" />
    </div>
</template>
