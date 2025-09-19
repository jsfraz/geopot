<script setup lang="ts">
import type { ModelsNumberValue, ModelsStringsValue } from '@/api/models';
import { Observable } from 'rxjs';
import { ref, onMounted } from 'vue';
import { VueSpinnerPulse } from 'vue3-spinners';

// Define props with default values
const props = defineProps({
    spinnerColor: { type: String, default: '#20c20e' },
    title: { type: String, default: 'Stat Card' },
    observableNumber: { type: Observable<ModelsNumberValue>, default: null },
    observableStrings: { type: Observable<ModelsStringsValue>, default: null }
});

// Expose public methods
defineExpose({
    increaseNumberValue,
    increaseStringsValue
});

// Loading state and values
const isLoaded = ref(false);
const value = ref(0);
const stringsValue = ref([] as string[]);

onMounted(() => {
    // Observable validation
    if (props.observableNumber && props.observableStrings) {
        console.error('StatCard "' + props.title + '": Both observableNumber and observableStrings props are set. Please provide only one.');
        isLoaded.value = true;
        return;
    }
    if (!props.observableNumber && !props.observableStrings) {
        console.error('StatCard "' + props.title + '": Neither observableNumber nor observableStrings prop is set. Please provide one.');
        isLoaded.value = true;
        return;
    }

    // Load data from observable - result is number
    if (props.observableNumber) {
        props.observableNumber.subscribe({
            next: (data) => {
                console.log('StatCard "' + props.title + '" data received:', data);
                value.value = data.value;
            },
            error: (error) => {
                // TODO show error state
                console.error('Error loading StatCard "' + props.title + '" data:', error);
                isLoaded.value = true;
            },
            complete: () => {
                isLoaded.value = true;
            }
        });
    }

    // Load data from observable - result is array of strings
    if (props.observableStrings) {
        props.observableStrings.subscribe({
            next: (data) => {
                console.log('StatCard "' + props.title + '" data received:', data);
                stringsValue.value = data.value;
                value.value = data.value.length;
            },
            error: (error) => {
                // TODO show error state
                console.error('Error loading StatCard "' + props.title + '" data:', error);
                isLoaded.value = true;
            },
            complete: () => {
                isLoaded.value = true;
            }
        });
    }
});

// Increases value by 1
function increaseNumberValue() {
    value.value += 1;
}

// Increases strings value by adding new string if not exists and increases number value
function increaseStringsValue(newString: string) {
   if (!stringsValue.value.includes(newString)) {
       stringsValue.value.push(newString);
       increaseNumberValue();
   }
}
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
                class="border-1 border-hacker bg-hackerbg rounded-lg p-4 flex flex-col items-start justify-center h-full">
                <p class="text-lg font-medium mb-2 uppercase text-gray-300">{{ title }}</p>
                <p class="text-5xl font-bold text-white">{{ value }}</p>
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