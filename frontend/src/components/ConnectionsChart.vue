<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import VueApexCharts from 'vue3-apexcharts';
import type ApexCharts from 'apexcharts';
import { VueSpinnerPulse } from 'vue3-spinners';
import { Configuration } from '@/api/runtime';
import { StatsApi } from '@/api/apis';
import type { ModelsHourlyStat } from '@/api/models';

const props = defineProps({
    apiConfiguration: { type: Configuration, required: true },
    spinnerColor: { type: String, default: '#20c20e' },
});

const isLoaded = ref(false);
const activeRange = ref(24);
const series = ref<{ name: string; data: { x: number; y: number }[] }[]>([{ name: 'Connections', data: [] }]);

const chartOptions = ref<ApexCharts.ApexOptions>({
    chart: {
        type: 'area' as const,
        background: 'transparent',
        toolbar: { show: false },
        animations: { enabled: true, speed: 400 },
        zoom: { enabled: false },
    },
    dataLabels: { enabled: false },
    stroke: {
        curve: 'smooth',
        width: 2,
        colors: ['#20c20e'],
    },
    fill: {
        type: 'gradient',
        gradient: {
            shadeIntensity: 1,
            opacityFrom: 0.45,
            opacityTo: 0.02,
            stops: [0, 100],
            colorStops: [
                { offset: 0, color: '#20c20e', opacity: 0.45 },
                { offset: 100, color: '#20c20e', opacity: 0.02 },
            ],
        },
    },
    grid: {
        borderColor: '#1a3a1a',
        strokeDashArray: 3,
        xaxis: { lines: { show: false } },
    },
    xaxis: {
        type: 'datetime',
        labels: {
            style: { colors: '#5a8a5a', fontFamily: 'monospace', fontSize: '11px' },
            datetimeUTC: false,
        },
        axisBorder: { color: '#1a3a1a' },
        axisTicks: { color: '#1a3a1a' },
    },
    yaxis: {
        labels: {
            style: { colors: '#5a8a5a', fontFamily: 'monospace', fontSize: '11px' },
            formatter: (val: number) => val.toLocaleString(),
        },
    },
    tooltip: {
        theme: 'dark',
        x: { format: 'HH:mm dd.MM' },
        style: { fontFamily: 'monospace' },
    },
    markers: { size: 0 },
    colors: ['#20c20e'],
});

function loadData(hours: number) {
    isLoaded.value = false;
    new StatsApi(props.apiConfiguration).getHourlyStats(hours).subscribe({
        next: (data: ModelsHourlyStat[]) => {
            series.value = [{
                name: 'Connections',
                data: data.map(d => ({
                    x: new Date(d.bucket).getTime(),
                    y: d.count,
                })),
            }];
        },
        error: (err) => {
            console.error('ConnectionsChart error:', err);
            isLoaded.value = true;
        },
        complete: () => {
            isLoaded.value = true;
        },
    });
}

onMounted(() => loadData(activeRange.value));
watch(activeRange, (v) => loadData(v));
</script>

<template>
    <div class="border border-hacker bg-hackerbg rounded-lg relative w-full h-full flex flex-col overflow-hidden">
        <!-- Header -->
        <div class="flex items-center justify-between px-3 pt-3 pb-1 shrink-0">
            <p class="text-xs font-medium uppercase text-gray-400 tracking-widest font-mono">Connections over time</p>
            <div class="flex gap-1">
                <button v-for="h in [24, 168, 720]" :key="h"
                    @click="activeRange = h"
                    :class="[
                        'text-xs px-2 py-0.5 rounded font-mono border transition-colors',
                        activeRange === h
                            ? 'border-hacker text-hacker bg-hacker/10'
                            : 'border-gray-700 text-gray-500 hover:border-gray-500 hover:text-gray-300'
                    ]">
                    {{ h === 24 ? '24H' : h === 168 ? '7D' : '30D' }}
                </button>
            </div>
        </div>
        <!-- Spinner -->
        <Transition name="fade">
            <div v-if="!isLoaded"
                class="absolute inset-0 flex justify-center items-center z-10 bg-black/30">
                <VueSpinnerPulse size="10" :color="spinnerColor" />
            </div>
        </Transition>
        <!-- Chart -->
        <div class="flex-1 min-h-0">
            <VueApexCharts
                type="area"
                height="100%"
                :options="chartOptions"
                :series="series"
            />
        </div>
    </div>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
