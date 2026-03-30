<script setup lang="ts">
import type { ModelsConnection } from '@/api';
import { StatsApi } from '@/api/apis';
import { Configuration } from '@/api/runtime';
import { ref, onMounted } from 'vue';
import { VueSpinnerPulse } from 'vue3-spinners';

// Define props with default values
const props = defineProps({
    apiConfiguration: { type: Configuration, required: true },
    spinnerColor: { type: String, default: '#20c20e' },
});

// Loading state and values
const isLoaded = ref(false);
let value = ref<ModelsConnection | null>(null);

onMounted(() => {
    if (props.apiConfiguration) {
        new StatsApi(props.apiConfiguration).getServerInfo().subscribe({
            next: (data) => {
                value.value = data;
            },
            error: (error) => {
                // TODO show error state
                console.error('Error fetching server info:', error);
                isLoaded.value = true;
            },
            complete: () => {
                isLoaded.value = true;
            }
        });
    }
});

</script>

<template>
    <div class="relative w-full h-full">
        <Transition name="fade">
            <div v-if="!isLoaded"
                class="absolute top-0 left-0 w-full h-full flex justify-center items-center z-10 bg-black/50">
                <VueSpinnerPulse size="10" :color="spinnerColor" />
            </div>
        </Transition>
        <div class="h-full">
            <div
                class="hud-card p-5 flex flex-col items-start justify-center h-full w-full rounded-sm overflow-hidden">
                <p class="text-[11px] font-bold mb-2.5 uppercase text-hacker/70 tracking-[0.2em] font-mono italic">
                    <span class="mr-1 text-hacker">▶</span> System Identity
                </p>
                <div class="space-y-1.5">
                    <p class="text-3xl font-bold text-white glow-text-green tabular-nums">{{ value ? value.ipAddress : '---.---.---.---' }}</p>
                    <div class="flex items-center gap-2 text-sm font-mono font-bold">
                        <span class="text-hacker">{{ value ? value.countryName : 'Searching...' }}</span>
                        <span class="text-gray-600">//</span>
                        <span class="text-gray-400 text-xs">{{ value ? value.cityName : 'N/A' }}</span>
                    </div>
                    <p class="text-[11px] text-hacker font-mono font-bold tracking-wider uppercase bg-hacker/10 px-2 py-0.5 inline-block border border-hacker/20">{{ value ? value.regionName : 'Initializing...' }}</p>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped></style>