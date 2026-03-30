<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import VueApexCharts from 'vue3-apexcharts';
import type ApexCharts from 'apexcharts';
import { VueSpinnerPulse } from 'vue3-spinners';
import { Configuration } from '@/api/runtime';
import { StatsApi } from '@/api/apis';
import type { ModelsHourlyStat } from '@/api/models';

// FIXME chart indicator

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
        animations: { enabled: true, speed: 800, animateGradually: { enabled: true } },
        zoom: { enabled: false },
        dropShadow: {
            enabled: true,
            top: 0,
            left: 0,
            blur: 4,
            color: '#39ff14',
            opacity: 0.4
        }
    },
    dataLabels: { enabled: false },
    stroke: {
        curve: 'smooth' as const,
        width: 3.5,
        colors: ['#39ff14'],
    },
    fill: {
        type: 'gradient' as const,
        gradient: {
            shadeIntensity: 1,
            opacityFrom: 0.5,
            opacityTo: 0.05,
            stops: [0, 100],
            colorStops: [
                { offset: 0, color: '#39ff14', opacity: 0.45 },
                { offset: 100, color: '#39ff14', opacity: 0 },
            ],
        },
    },
    grid: {
        show: true,
        borderColor: 'rgba(57, 255, 20, 0.08)',
        strokeDashArray: 4,
        xaxis: { lines: { show: true } },
        yaxis: { lines: { show: false } },
    },
    xaxis: {
        type: 'datetime' as const,
        labels: {
            style: { colors: '#5a8a5a', fontFamily: 'FiraCodeNerdFontMono-Regular, monospace', fontSize: '11px', fontWeight: 600 },
            datetimeUTC: false,
        },
        axisBorder: { show: false },
        axisTicks: { show: false },
    },
    yaxis: {
        labels: {
            style: { colors: '#5a8a5a', fontFamily: 'FiraCodeNerdFontMono-Regular, monospace', fontSize: '11px', fontWeight: 600 },
            formatter: (val: number) => val.toLocaleString(),
        },
    },
    tooltip: {
        theme: 'dark',
        x: { format: 'HH:mm dd.MM' },
        style: { fontSize: '10px', fontFamily: 'FiraCodeNerdFontMono-Regular, monospace' },
        marker: { show: false },
    },
    markers: {
        size: 0,
        colors: ['#20c20e'],
        strokeColors: '#020d01',
        strokeWidth: 2,
        hover: { size: 4 }
    },
    colors: ['#20c20e'],
});

function loadData(hours: number) {
    isLoaded.value = false;
    new StatsApi(props.apiConfiguration).getHourlyStats({ hours: hours }).subscribe({
        next: (data: ModelsHourlyStat[]) => {
            series.value = [{
                name: 'Connections',
                data: data.map(d => ({
                    x: new Date(d.bucket ?? '').getTime(),
                    y: d.count ?? 0,
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
    <div class="hud-card relative w-full h-full flex flex-col rounded-sm overflow-hidden p-2">
        <!-- Header -->
        <div
            class="flex items-center justify-between px-2 pt-2.5 pb-1.5 shrink-0 border-b border-hacker/20 mb-2 bg-hacker/5">
            <p class="text-[11px] font-bold uppercase text-hacker tracking-[0.2em] font-mono italic">
                <span class="mr-1 text-hacker">▶</span> Connections Timeline
            </p>
            <div class="flex gap-2">
                <button v-for="h in [24, 168, 720]" :key="h" @click="activeRange = h" :class="[
                    'text-[10px] px-2.5 py-1 rounded-sm font-mono font-bold border transition-all duration-300',
                    activeRange === h
                        ? 'border-hacker text-hacker bg-hacker/20 glow-text-green shadow-[0_0_15px_rgba(57,255,20,0.2)]'
                        : 'border-white/10 text-gray-400 hover:border-hacker/40 hover:text-hacker/60'
                ]">
                    {{ h === 24 ? '24H' : h === 168 ? '7D' : '30D' }}
                </button>
            </div>
        </div>
        <!-- Spinner -->
        <Transition name="fade">
            <div v-if="!isLoaded" class="absolute inset-0 flex justify-center items-center z-10 bg-black/50">
                <VueSpinnerPulse size="10" :color="spinnerColor" />
            </div>
        </Transition>
        <!-- Chart -->
        <div class="flex-1 min-h-0">
            <VueApexCharts type="area" height="100%" :options="chartOptions" :series="series" />
        </div>
    </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
