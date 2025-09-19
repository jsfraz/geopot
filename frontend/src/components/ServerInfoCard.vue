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
    <div class="relative w-full h-full overflow-hidden">
        <Transition name="fade">
            <div v-if="!isLoaded"
                class="absolute top-0 left-0 w-full h-full flex justify-center items-center z-10 bg-black/33">
                <VueSpinnerPulse size="10" :color="spinnerColor" />
            </div>
        </Transition>
        <div>
            <div
                class="border-1 border-hacker bg-hackerbg rounded-lg p-3 flex flex-col items-start justify-center h-full">
                <p class="text-2xl font-bold text-white">{{ value ? value.ipAddress : '' }}</p>
                <p class="text-xl font-bold text-white">{{ value ? value.cityName + ' - ' + value.countryName : '' }}</p>
                <p class="text-xl font-bold text-white">{{ value ? value.regionName : '' }}</p>
            </div>
        </div>
    </div>
</template>

<style scoped></style>