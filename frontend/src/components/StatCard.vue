<script setup lang="ts">
import type { ModelsValue } from '@/api';
import { Observable } from 'rxjs';
import { ref, onMounted } from 'vue';
import { VueSpinnerPulse } from 'vue3-spinners';

// Define props with default values
const props = defineProps({
    spinnerColor: { type: String, default: '#20c20e' },
    title: { type: String, required: true, default: 'Stat Card' },
    observable: { type: Observable<ModelsValue>, required: true, default: null },
});

const isLoaded = ref(false);
const value = ref(0);

onMounted(() => {
    // Load data from observable
    if (props.observable) {
        props.observable.subscribe({
            next: (data) => {
                console.log('StatCard data received:', data);
                value.value = data.value;
                isLoaded.value = true;
            },
            error: (error) => {
                // TODO show error state
                console.error('Error loading StatCard data:', error);
                isLoaded.value = true;
            },
        });
    }
});
</script>

<template>
    <div class="relative w-full h-full overflow-hidden">
        <Transition name="fade">
            <div v-if="!isLoaded"
                class="absolute top-0 left-0 w-full h-full flex justify-center items-center z-10 bg-black/50">
                <VueSpinnerPulse size="10" :color="spinnerColor" />
            </div>
        </Transition>
        <div>
            <div
                class="border-1 border-hacker rounded-lg p-4 flex flex-col items-start justify-center h-full bg-hackerbg">
                <p class="text-lg font-medium mb-2 uppercase text-gray-300">{{ title }}</p>
                <p class="text-5xl font-bold text-white">{{ value }}</p>
            </div>
        </div>
    </div>
</template>

<style scoped>
/* Ponecháváme pouze transition efekt, zbytek řeší Tailwind */
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>