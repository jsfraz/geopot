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
        animations: { enabled: true, speed: 600 },
    },
    plotOptions: {
        bar: {
            horizontal: true,
            borderRadius: 2,
            barHeight: '60%',
            distributed: false,
        },
    },
    dataLabels: {
        enabled: true,
        formatter: (val: number) => {
            return val ? `${val.toFixed(1)}%` : '';
        },
        style: { fontFamily: 'FiraCodeNerdFontMono-Regular, monospace', fontSize: '10px', fontWeight: 600, colors: ['#39ff14'] },
        offsetX: 5,
    },
    colors: ['#39ff14'],
    fill: {
        type: 'gradient' as const,
        gradient: {
            shade: 'dark',
            type: "horizontal",
            shadeIntensity: 0.5,
            gradientToColors: ['#00ff00'],
            inverseColors: true,
            opacityFrom: 0.85,
            opacityTo: 0.3,
            stops: [0, 100]
        }
    },
    grid: {
        show: false,
    },
    xaxis: {
        categories: [] as string[],
        labels: {
            style: { colors: '#5a8a5a', fontFamily: 'FiraCodeNerdFontMono-Regular, monospace', fontSize: '11px', fontWeight: 600 },
            formatter: (val: string) => {
                const num = Number(val);
                return isNaN(num) ? val : `${num.toFixed(0)}%`;
            }
        },
        axisBorder: { show: false },
        axisTicks: { show: false },
    },
    yaxis: {
        labels: {
            style: { colors: '#39ff14', fontFamily: 'FiraCodeNerdFontMono-Regular, monospace', fontSize: '12px', fontWeight: 700 },
            maxWidth: 140,
        },
    },
    tooltip: {
        theme: 'dark',
        style: { fontSize: '10px', fontFamily: 'FiraCodeNerdFontMono-Regular, monospace' },
        y: { 
            title: { formatter: () => '' },
            formatter: (val: number, opts: any) => {
                const originalIndex = entries.value.length - 1 - opts.dataPointIndex;
                const count = entries.value[originalIndex]?.count;
                return count ? `${count.toLocaleString()} connections` : `${val}%`;
            } 
        },
    },
});

onMounted(() => {
    new StatsApi(props.apiConfiguration).getTopCountries({ limit: props.limit }).subscribe({
        next: (data: ModelsTopEntry[]) => {
            entries.value = data;
            // Reverse so top entry is at top of horizontal bar chart
            const reversed = [...data].reverse();
            chartOptions.value = {
                ...chartOptions.value,
                xaxis: {
                    ...chartOptions.value.xaxis,
                    categories: reversed.map(e => e.label ?? ''),
                },
            };
            series.value = [{ data: reversed.map(e => e.percentage ?? 0) }];
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
    <div class="hud-card relative w-full h-full flex flex-col rounded-sm overflow-hidden p-2">
        <div class="px-2 pt-2.5 pb-1.5 shrink-0 border-b border-hacker/20 mb-2 bg-hacker/5">
            <p class="text-[11px] font-bold uppercase text-hacker tracking-[0.2em] font-mono italic">
                <span class="mr-1 text-hacker">▶</span> Top Countries
            </p>
        </div>
        <Transition name="fade">
            <div v-if="!isLoaded" class="absolute inset-0 flex justify-center items-center z-10 bg-black/50">
                <VueSpinnerPulse size="10" :color="spinnerColor" />
            </div>
        </Transition>
        <div class="flex-1 min-h-0">
            <VueApexCharts type="bar" height="100%" :options="chartOptions" :series="series" />
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
