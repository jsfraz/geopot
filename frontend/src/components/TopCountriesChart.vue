<script setup lang="ts">
import { ref, onMounted } from 'vue';
import VueApexCharts from 'vue3-apexcharts';
import type ApexCharts from 'apexcharts';
import { VueSpinnerPulse } from 'vue3-spinners';
import { Configuration } from '@/api/runtime';
import { StatsApi } from '@/api/apis';
import type { ModelsTopEntry } from '@/api/models';

const props = defineProps({
    apiConfiguration: { type: Configuration, required: true },
    spinnerColor: { type: String, default: '#20c20e' },
    limit: { type: Number, default: 12 },
});

const isLoaded = ref(false);
const entries = ref<ModelsTopEntry[]>([]);

const series = ref<{ data: number[] }[]>([{ data: [] }]);
const chartOptions = ref<ApexCharts.ApexOptions>({
    chart: {
        type: 'bar' as const,
        background: 'transparent',
        toolbar: { show: false },
        animations: { enabled: true, speed: 400 },
    },
    plotOptions: {
        bar: {
            horizontal: true,
            borderRadius: 3,
            barHeight: '65%',
            distributed: false,
        },
    },
    dataLabels: {
        enabled: true,
        formatter: (_val: number, opts: any) => {
            const pct = entries.value[opts.dataPointIndex]?.percentage;
            return pct ? `${pct.toFixed(1)}%` : '';
        },
        style: { fontFamily: 'monospace', fontSize: '11px', colors: ['#20c20e'] },
        offsetX: 5,
    },
    colors: ['#20c20e'],
    fill: { opacity: 0.85 },
    grid: {
        borderColor: '#1a3a1a',
        xaxis: { lines: { show: false } },
        yaxis: { lines: { show: false } },
    },
    xaxis: {
        categories: [] as string[],
        labels: {
            style: { colors: '#5a8a5a', fontFamily: 'monospace', fontSize: '10px' },
        },
        axisBorder: { color: '#1a3a1a' },
        axisTicks: { show: false },
    },
    tooltip: {
        theme: 'dark',
        style: { fontFamily: 'monospace' },
        y: { formatter: (val: number) => val.toLocaleString() },
    },
});

onMounted(() => {
    new StatsApi(props.apiConfiguration).getTopCountries(props.limit).subscribe({
        next: (data: ModelsTopEntry[]) => {
            entries.value = data;
            // Reverse so top entry is at top of horizontal bar chart
            const reversed = [...data].reverse();
            chartOptions.value = {
                ...chartOptions.value,
                xaxis: {
                    ...chartOptions.value.xaxis,
                    categories: reversed.map(e => e.label),
                },
            };
            series.value = [{ data: reversed.map(e => e.count) }];
        },
        error: (err) => {
            console.error('TopCountriesChart error:', err);
            isLoaded.value = true;
        },
        complete: () => { isLoaded.value = true; },
    });
});
</script>

<template>
    <div class="border border-hacker bg-hackerbg rounded-lg relative w-full h-full flex flex-col overflow-hidden">
        <div class="px-3 pt-3 pb-1 shrink-0">
            <p class="text-xs font-medium uppercase text-gray-400 tracking-widest font-mono">Top Countries</p>
        </div>
        <Transition name="fade">
            <div v-if="!isLoaded"
                class="absolute inset-0 flex justify-center items-center z-10 bg-black/30">
                <VueSpinnerPulse size="10" :color="spinnerColor" />
            </div>
        </Transition>
        <div class="flex-1 min-h-0">
            <VueApexCharts
                type="bar"
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
