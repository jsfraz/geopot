<script setup lang="ts">
import type { ModelsNumberValue } from '@/api/models';
import { Observable } from 'rxjs';
import { ref, onMounted, onUnmounted } from 'vue';
import { VueSpinnerPulse } from 'vue3-spinners';

// Define props with default values
const props = defineProps({
    spinnerColor: { type: String, default: '#20c20e' },
    title: { type: String, default: 'Stat Card' },
    observableNumber: { type: Observable<ModelsNumberValue>, default: null },
    refreshInterval: { type: Number, default: null },
});

// Expose public methods
defineExpose({
    increaseNumberValue,
    increaseStringsValue
});

// Loading state and values
const isLoaded = ref(false);
const value = ref(0);
// Session tracking for unique strings (used by increaseStringsValue)
const sessionSeenStrings = new Set<string>();
let intervalId: number | null = null;

onMounted(() => {
    if (!props.observableNumber) {
        console.error('StatCard "' + props.title + '": observableNumber prop is required.');
        return;
    }
    // Set up refresh interval if provided
    if (props.refreshInterval) {
        if (props.refreshInterval <= 0) {
            console.error('StatCard "' + props.title + '": refreshInterval must be greater than 0.');
            return;
        }
        // Initial load
        loadObservables();
        intervalId = window.setInterval(loadObservables, props.refreshInterval);
    } else {
        // Initial load
        loadObservables();
    }
});

// Clear interval on unmount
onUnmounted(() => {
    if (intervalId !== null) {
        clearInterval(intervalId);
        intervalId = null;
    }
});

// Load data from observables
function loadObservables() {
    // Load data from observable - result is number
    if (props.observableNumber) {
        props.observableNumber.subscribe({
            next: (data) => {
                console.log('StatCard "' + props.title + '" data received:', data);
                // Value from backend.
                // We do NOT clear sessionSeenStrings here because it tracks what we have seen
                // live from the WebSocket. In a more complex scenario, we might want to reconcile.
                value.value = data.value;
            },
            error: (error) => {
                console.error('Error loading StatCard "' + props.title + '" data:', error);
                isLoaded.value = true;
            },
            complete: () => {
                isLoaded.value = true;
            }
        });
    }
}

// Increases value by 1
function increaseNumberValue() {
    if (!isLoaded.value) return;

    value.value += 1;
}

// Increases strings value by adding new string if not exists (in session) and increases number value
function increaseStringsValue(newString: string) {
    if (!isLoaded.value) return;

    if (!sessionSeenStrings.has(newString)) {
        sessionSeenStrings.add(newString);
        increaseNumberValue();
    }
}
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
            <div class="hud-card p-5 flex flex-col items-start justify-center h-full w-full rounded-sm overflow-hidden">
                <p class="text-[11px] font-bold mb-1.5 uppercase text-hacker/70 tracking-[0.2em] font-mono">{{ title }}
                </p>
                <div class="flex items-baseline gap-2">
                    <p class="text-4xl font-bold text-white glow-text-green tabular-nums tracking-tighter">{{
                        value.toLocaleString() }}</p>
                </div>
                <!-- Subtle decorative line -->
                <div class="w-16 h-0.5 bg-hacker/30 mt-2"></div>
            </div>
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