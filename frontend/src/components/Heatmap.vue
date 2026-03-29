<script setup lang="ts">
import { ref, onMounted } from 'vue';
import * as L from "leaflet";
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

// Expose public methods
defineExpose({
    addPoint
});

// Heatmap setup
const isHeatmapLoaded = ref(false);
const heatmapContainer = ref<HTMLElement | null>(null);
let heatmap: L.Map | null = null;
let heat: any = null;

onMounted(() => {
    if (heatmapContainer.value) {
        // Map initialization
        heat = (L as any).heatLayer([], {
            radius: 35,
            blur: 20,
            maxZoom: 15,
            minOpacity: 0.1,
            max: 1.0 // This will be adjusted based on data
        });

        // Prevents the leaflet-heat component from crashing if it attempts to redraw the heatmap when the window is minimized to the point where the height or width of the div is 0
        const originalRedraw = heat._redraw;
        heat._redraw = function () {
            if (this._map && (this._map.getSize().x <= 0 || this._map.getSize().y <= 0)) {
                return;
            }
            if (originalRedraw) {
                originalRedraw.call(this);
            }
        };

        heatmap = L.map(heatmapContainer.value, {
            zoom: 1,
            minZoom: 1,
            maxZoom: 15,
            center: L.latLng(25, 10.5),
            layers: [
                L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'),
                heat
            ]
        }).whenReady(() => {
            // Load data
            if (props.apiConfiguration) {
                new StatsApi(props.apiConfiguration).getAllLatLng().subscribe({
                    next: (data: ModelsLatLng[]) => {
                        if (data.length > 0) {
                            // Find max intensity for normalization
                            const maxIntensity = Math.max(...data.map(d => d.intensity || 1));
                            // Set max much lower than peak to make "hot" zones redder
                            heat.setOptions({
                                max: maxIntensity * 0.5,
                                radius: 25,
                                blur: 10,
                                minOpacity: 0.3
                            });

                            const points = data.map(item => [item.latitude, item.longitude, item.intensity || 1] as [number, number, number]);
                            heat.setLatLngs(points);
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

// Adds a new point to the heatmap
function addPoint(lat: number, lng: number) {
    if (!heatmap || !heat) return;
    heat.addLatLng([lat, lng, 1]);
}
</script>

<template>
    <div class="border border-hacker bg-hackerbg rounded-lg relative w-full h-full overflow-hidden min-h-2.5 min-w-2.5">
        <Transition name="fade">
            <div v-if="!isHeatmapLoaded"
                class="absolute top-0 left-0 w-full h-full flex justify-center items-center z-10 bg-black/33">
                <VueSpinnerPacman size="20" :color="spinnerColor" />
            </div>
        </Transition>
        <div ref="heatmapContainer" class="w-full h-full min-h-2.5 min-w-2.5"></div>
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
</style>