<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { map, latLng, tileLayer, type HeatLayer } from "leaflet";
import { type Map } from "leaflet";
import { heatLayer } from "leaflet";
import "leaflet.heat";
import { VueSpinnerPacman } from 'vue3-spinners';
import { Configuration } from '@/api/runtime';
import { StatsApi } from '@/api/apis';
import { type ModelsLatLng } from '@/api';

// Define props with default values
const props = defineProps({
    apiConfiguration: { type: Configuration, required: true },
    spinnerColor: { type: String, default: '#20c20e' },
});

// Heatmap setup
const isHeatmapLoaded = ref(false);
const heatmapContainer = ref<HTMLElement | null>(null);
let heatmap: Map | null = null;
let heat: HeatLayer = heatLayer([], {
    radius: 25,
    blur: 15,
    maxZoom: 15,
    minOpacity: 0.4
});

onMounted(() => {
    if (heatmapContainer.value) {
        heatmap = map(heatmapContainer.value, {
            zoom: 1,
            minZoom: 1,
            maxZoom: 15,
            center: latLng(52.6182256, 1.3723268),     // Norwich UK
            layers: [
                tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'),
                heat
            ]
        }).whenReady(() => {
            // Load data
            if (props.apiConfiguration) {
                new StatsApi(props.apiConfiguration).getAllLatLng().subscribe({
                    next: (data: ModelsLatLng[]) => {
                        if (data.length > 0) {
                            heat.setLatLngs(data.map(item => [item.latitude, item.longitude, 1]));
                        }
                    },
                    error: (error) => {
                        console.error('Error fetching data for heatmap:', error);
                        isHeatmapLoaded.value = true;
                    }, 
                    complete: () => {
                        isHeatmapLoaded.value = true;
                    }
                });
            } else {
                isHeatmapLoaded.value = true;
            }
        });
    }
});
</script>

<template>
    <div class="border-1 border-hacker bg-hackerbg rounded-lg relative w-full h-full overflow-hidden">
        <Transition name="fade">
            <div v-if="!isHeatmapLoaded"
                class="absolute top-0 left-0 w-full h-full flex justify-center items-center z-10 bg-black/33">
                <VueSpinnerPacman size="20" :color="spinnerColor" />
            </div>
        </Transition>
        <div ref="heatmapContainer" class="w-full h-full"></div>
    </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.5s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}

:deep(canvas) {
    width: 100% !important;
    height: 100% !important;
}
</style>