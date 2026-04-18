<script setup lang="ts">
import {
    Chart as ChartJS,
    Title, Tooltip, Legend,
    BarElement,
    CategoryScale,
    LinearScale,
    type ChartData,
    type ChartOptions
} from 'chart.js'
import { Bar } from 'vue-chartjs'
import { computed } from 'vue'

ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale)

interface Props {
    data: any[]
    dataKey: string
    color: string
    label: string
    isSleep?: boolean // 睡眠時間（分→時間）の変換用
}

const props = defineProps<Props>()

const chartData = computed<ChartData<'bar'>>(() => {
    if (!props.data || props.data.length === 0) {
        return { labels: [], datasets: [] }
    }

    return {
        labels: props.data.map(d => d.date),
        datasets: [
            {
                label: props.label,
                // isSleep が true なら 60 で割って時間に変換、それ以外はそのまま
                data: props.data.map(d => {
                    const val = d[props.dataKey] || 0
                    return props.isSleep ? Number((val / 60).toFixed(1)) : val
                }),
                backgroundColor: props.color,
                borderRadius: 6,
            }
        ]
    }
})

const chartOptions: ChartOptions<'bar'> = {
    responsive: true,
    maintainAspectRatio: false,
    scales: {
        y: {
            beginAtZero: true,
            grid: { color: 'rgba(71, 85, 105, 0.2)' },
            ticks: { color: '#94a3b8' }
        },
        x: {
            ticks: { color: '#94a3b8' }
        }
    },
    plugins: {
        legend: {
            display: false // サブグラフなので凡例は消してスッキリさせる
        }
    }
}
</script>

<template>
    <div class="h-[250px]">
        <Bar v-if="data.length > 0" :data="chartData" :options="chartOptions" />
    </div>
</template>
