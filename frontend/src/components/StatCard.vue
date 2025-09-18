<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { VueSpinnerPulse } from 'vue3-spinners';

// Define props with default values
const props = defineProps({
    spinnerColor: { type: String, default: '#20c20e' },
    title: { type: String, default: 'Text' },
    value: { type: [Number, String], default: '0' },
});

const isLoaded = ref(false);

// Simulace načítání dat
onMounted(() => {
    const randomDelay = Math.random() * 5000;
    setTimeout(() => {
        isLoaded.value = true;
    }, randomDelay);
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