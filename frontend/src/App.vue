<!-- TODO Make standalone globe component -->

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import Globe, { type GlobeInstance } from 'globe.gl';
import { useWebSocket } from '@vueuse/core';
import { VueSpinnerPacman } from 'vue3-spinners';
import type { WSMessage } from './types/ws_message';

// Globe setup
const websocketUrl = import.meta.env.DEV ? 'ws://localhost:8080/ws' : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;
const spinnerColor = ref('#20c20e');
const isGlobeLoaded = ref(false);
const globeContainer = ref<HTMLElement | null>(null);
let globe: GlobeInstance | null = null;
// Marker
// https://www.svgrepo.com/svg/484996/server-network-part-2?edit=true
const markerSvg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" xml:space="preserve" fill="#979696" stroke="#979696" stroke-width="2.56"><g stroke-width="0"/><g stroke-linecap="round" stroke-linejoin="round"/><path style="fill:#979696" d="M232.254 69.157h42.982v377.465h-42.982zM56.146 446.588l20.715 42.976h155.373v-42.976z"/><path style="fill:#979696" d="M275.21 446.588v42.976h159.901l20.715-42.976z"/><path style="fill:#979696" d="M232.234 446.588h42.977v42.977h-42.977zM511.972 7.837v105.05a7.79 7.79 0 0 1-7.8 7.8H7.8a7.79 7.79 0 0 1-7.8-7.8V7.837A7.79 7.79 0 0 1 7.8.038h496.372a7.79 7.79 0 0 1 7.8 7.799zm0 140.481v105.05c0 4.315-3.485 7.883-7.8 7.883H7.8c-4.315 0-7.8-3.568-7.8-7.883v-105.05a7.79 7.79 0 0 1 7.8-7.8h496.372a7.79 7.79 0 0 1 7.8 7.8zm0 140.564v105.05a7.79 7.79 0 0 1-7.8 7.799H7.8a7.79 7.79 0 0 1-7.8-7.799v-105.05a7.79 7.79 0 0 1 7.8-7.799h496.372a7.79 7.79 0 0 1 7.8 7.799z"/><path style="fill:#fff" d="M492.427 6.264H19.545c-7.351 0-13.31 5.959-13.31 13.31v81.539c0 7.351 5.959 13.309 13.31 13.309h472.882c7.351 0 13.31-5.959 13.31-13.309v-81.54c0-7.351-5.959-13.309-13.31-13.309zm0 140.526H19.545c-7.351 0-13.31 5.959-13.31 13.31v81.539c0 7.351 5.959 13.31 13.31 13.31h472.882c7.351 0 13.31-5.959 13.31-13.31V160.1c0-7.351-5.959-13.31-13.31-13.31zm0 140.528H19.545c-7.351 0-13.31 5.959-13.31 13.31v81.539c0 7.351 5.959 13.31 13.31 13.31h472.882c7.351 0 13.31-5.959 13.31-13.31v-81.539c0-7.352-5.959-13.31-13.31-13.31z"/><path style="fill:#979696" d="M57.355 26.558H43.829a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V28.674a2.116 2.116 0 0 0-2.116-2.116zm0 25.75H43.829a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V54.424a2.116 2.116 0 0 0-2.116-2.116zm0 25.751H43.829a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V80.175a2.116 2.116 0 0 0-2.116-2.116zm31.666-51.501H75.495a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V28.674a2.116 2.116 0 0 0-2.116-2.116zm0 25.75H75.495a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V54.424a2.116 2.116 0 0 0-2.116-2.116zm0 25.751H75.495a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V80.175a2.116 2.116 0 0 0-2.116-2.116zm31.666-51.501h-13.525a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.525a2.116 2.116 0 0 0 2.116-2.116V28.674a2.115 2.115 0 0 0-2.116-2.116zm0 25.75h-13.525a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.525a2.116 2.116 0 0 0 2.116-2.116V54.424a2.116 2.116 0 0 0-2.116-2.116zm0 25.751h-13.525a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.525a2.116 2.116 0 0 0 2.116-2.116V80.175a2.115 2.115 0 0 0-2.116-2.116zm31.667-51.501h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V28.674a2.116 2.116 0 0 0-2.116-2.116zm0 25.75h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V54.424a2.117 2.117 0 0 0-2.116-2.116zm0 25.751h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V80.175a2.115 2.115 0 0 0-2.116-2.116zm31.666-51.501h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V28.674a2.116 2.116 0 0 0-2.116-2.116zm0 25.75h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V54.424a2.117 2.117 0 0 0-2.116-2.116zm0 25.751h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V80.175a2.115 2.115 0 0 0-2.116-2.116zm41.084-28.316h100.213v21.202H225.104z"/><circle style="fill:#43b471" cx="369.338" cy="61.198" r="19.487"/><circle style="fill:#d3d340" cx="416.663" cy="61.198" r="19.487"/><circle style="fill:#d15075" cx="463.989" cy="61.198" r="19.487"/><path style="fill:#979696" d="M57.355 167.084H43.829a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V169.2a2.116 2.116 0 0 0-2.116-2.116zm0 25.751H43.829a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm0 25.75H43.829a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm31.666-51.501H75.495a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V169.2a2.116 2.116 0 0 0-2.116-2.116zm0 25.751H75.495a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm0 25.75H75.495a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm31.666-51.501h-13.525a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.525a2.116 2.116 0 0 0 2.116-2.116V169.2a2.116 2.116 0 0 0-2.116-2.116zm0 25.751h-13.525a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.525a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm0 25.75h-13.525a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.525a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm31.667-51.501h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V169.2a2.117 2.117 0 0 0-2.116-2.116zm0 25.751h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.117 2.117 0 0 0-2.116-2.116zm0 25.75h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.117 2.117 0 0 0-2.116-2.116zm31.666-51.501h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V169.2a2.117 2.117 0 0 0-2.116-2.116zm0 25.751h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.117 2.117 0 0 0-2.116-2.116zm0 25.75h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.117 2.117 0 0 0-2.116-2.116zm41.084-28.316h100.213v21.202H225.104z"/><circle style="fill:#43b471" cx="369.338" cy="201.725" r="19.487"/><circle style="fill:#d3d340" cx="416.663" cy="201.725" r="19.487"/><circle style="fill:#d15075" cx="463.989" cy="201.725" r="19.487"/><path style="fill:#979696" d="M57.355 307.611H43.829a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm0 25.751H43.829a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm0 25.75H43.829a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm31.666-51.501H75.495a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V169.2a2.116 2.116 0 0 0-2.116-2.116zm0 25.751H75.495a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm0 25.75H75.495a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm31.666-51.501h-13.525a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.525a2.116 2.116 0 0 0 2.116-2.116V169.2a2.116 2.116 0 0 0-2.116-2.116zm0 25.751h-13.525a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.525a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm0 25.75h-13.525a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.948 2.116 2.116 2.116h13.525a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.116 2.116 0 0 0-2.116-2.116zm31.667-51.501h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V169.2a2.117 2.117 0 0 0-2.116-2.116zm0 25.751h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.117 2.117 0 0 0-2.116-2.116zm0 25.75h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.117 2.117 0 0 0-2.116-2.116zm31.666-51.501h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116V169.2a2.117 2.117 0 0 0-2.116-2.116zm0 25.751h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.117 2.117 0 0 0-2.116-2.116zm0 25.75h-13.526a2.116 2.116 0 0 0-2.116 2.116v13.548c0 1.169.947 2.116 2.116 2.116h13.526a2.116 2.116 0 0 0 2.116-2.116v-13.548a2.117 2.117 0 0 0-2.116-2.116zm41.084-28.316h100.213v21.202H225.104z"/><circle style="fill:#43b471" cx="369.338" cy="342.252" r="19.487"/><circle style="fill:#d3d340" cx="416.663" cy="342.252" r="19.487"/><circle style="fill:#d15075" cx="463.989" cy="342.252" r="19.487"/><g style="opacity:.5"><path style="opacity:.07;fill:#040000" d="M275.236 261.251v19.832h228.935c4.315 0 7.8 3.486 7.8 7.799v105.05a7.79 7.79 0 0 1-7.8 7.799H275.236v44.891h180.559l-20.661 42.983H76.837L56.425 447.12l-.249-.497h176.078v-44.891H98.992l55.512-55.512 20.91-20.827 44.31-44.31h12.53v-12.53l57.089-57.089 21.077-21.16 43.48-43.48 4.647-4.647 1.66-1.659h143.966a7.79 7.79 0 0 1 7.8 7.8v105.05c0 4.315-3.485 7.883-7.8 7.883zM504.171 0h-3.403L380.083 120.685h124.088a7.83 7.83 0 0 0 7.829-7.828V7.829A7.83 7.83 0 0 0 504.171 0z"/></g></svg>`
let serverMarkerData: { lat: number; lng: number; size: number; color: string; } | null = null;
// Arcs and rings
const ARC_REL_LEN = 0.5;
const FLIGHT_TIME = 1250;
const RINGS_MAX_R = 7;
const RING_PROPAGATION_SPEED = 4;
const NUM_RINGS = 3;
const countryCoordinatesToColor = new Map<string, string>();
const HIGH_CONTRAST_COLORS = [
  '#FF3D00',
  '#00E676',
  '#2979FF',
  '#FFFF00',
  '#AA00FF',
  '#FF9100',
  '#00BCD4',
  '#EEFF41'
];
// Rotation
let autoRotate = true;
const rotationSpeed = 5;
let lastTime: number | null = null;

// WebSocket connection
useWebSocket(websocketUrl,
  {
    autoReconnect: true,
    onConnected(_) {
      console.log('WebSocket connected');
    },
    onDisconnected(_, event: CloseEvent) {
      console.log('WebSocket disconnected:', event.reason);
    },
    onError(_, event: Event) {
      console.error('WebSocket error:', event);
    },
    onMessage: (_, event: MessageEvent) => {
      // Parse message
      const msg: WSMessage = JSON.parse(event.data);
      console.log('WebSocket message received:', msg.type, msg.data);

      // Message type handling
      switch (msg.type) {
        case 'attacker':
          if (serverMarkerData && globe && msg.data.ipVersion != 0) {
            // Get country color based on coordinates
            const key = `${msg.data.latitude.toFixed(1)},${msg.data.longitude.toFixed(1)}`;
            if (!countryCoordinatesToColor.has(key) && msg.data.countryName) {
              // Use country name to generate color - same method as for polygons
              getCountryColorFromCoordinates(msg.data.latitude, msg.data.longitude, msg.data.countryName);
            }
            // Show arc from attacker to server
            emitArc(
              { lat: msg.data.latitude, lng: msg.data.longitude },
              { lat: serverMarkerData.lat, lng: serverMarkerData.lng }
            );
          }
          break;
        case 'server_info':
          // Set marker data
          serverMarkerData = {
            lat: msg.data.latitude,
            lng: msg.data.longitude,
            size: 1,
            color: '',
          };
          // Create marker on globe
          setServerLocationMarker();
          break;
        default:
          console.warn('Unknown message type:', msg.type);
      }
    },
  }
);

onMounted(() => {
  // https://github.com/vasturiano/globe.gl/blob/master/example/hexed-polygons/index.html
  fetch('assets/ne_110m_admin_0_countries.geojson').then(res => res.json()).then(countries => {
    if (globeContainer.value) {
      // Globe instance
      globe = new Globe(globeContainer.value)
        // .globeImageUrl('//cdn.jsdelivr.net/npm/three-globe/example/img/earth-dark.jpg')
        .globeImageUrl('assets/earth-dark.jpg')
        .backgroundColor('rgba(0,0,0,0)')
        // Polygons
        .hexPolygonsData(countries.features)
        .hexPolygonResolution(4)
        .hexPolygonMargin(0.2)
        .hexPolygonUseDots(false)
        // .hexPolygonColor(() => `#${Math.round(Math.random() * Math.pow(2, 24)).toString(16).padStart(6, '0')}`)
        .hexPolygonColor((feature: any) => {
          // Name of the country
          const name = feature.properties.NAME || feature.properties.name || 'unknown';

          // Convert name to number (using hashing)
          let hash = 0;
          for (let i = 0; i < name.length; i++) {
            hash = name.charCodeAt(i) + ((hash << 5) - hash);
          }

          // Convert number to color
          let color = '#';
          for (let i = 0; i < 3; i++) {
            const value = (hash >> (i * 8)) & 0xFF;
            color += value.toString(16).padStart(2, '0');
          }

          return color;
        })
        // Arcs and rings
        .arcColor((arc: { color?: string }) => arc.color || 'darkOrange')
        .arcDashLength(ARC_REL_LEN)
        .arcDashGap(2)
        .arcDashInitialGap(1)
        .arcDashAnimateTime(FLIGHT_TIME)
        .arcsTransitionDuration(0)
        .ringColor((ring: { color?: string }) => {
          // If the ring has a custom color, use it
          if (ring.color) {
            // Extract RGB values from the color (assuming rgb or hex format)
            let r = 255, g = 100, b = 50; // default values

            // For rgb format
            if (typeof ring.color === 'string' && ring.color.startsWith('rgb')) {
              const match = ring.color.match(/rgb\((\d+),\s*(\d+),\s*(\d+)\)/);
              if (match) {
                r = parseInt(match[1]);
                g = parseInt(match[2]);
                b = parseInt(match[3]);
              }
            }
            // For hex format
            else if (typeof ring.color === 'string' && ring.color.startsWith('#')) {
              const hex = ring.color.substring(1);
              r = parseInt(hex.substring(0, 2), 16);
              g = parseInt(hex.substring(2, 4), 16);
              b = parseInt(hex.substring(4, 6), 16);
            }

            // Return RGBA color with fading alpha
            return (t: number) => `rgba(${r},${g},${b},${1 - t})`;
          }

          // Or use default color
          return (t: number) => `rgba(255,100,50,${1 - t})`;
        })
        .ringMaxRadius(RINGS_MAX_R)
        .ringPropagationSpeed(RING_PROPAGATION_SPEED)
        .ringRepeatPeriod(FLIGHT_TIME * ARC_REL_LEN / NUM_RINGS)
        .onGlobeReady(() => {
          // Create marker on globe
          setServerLocationMarker();
          // Rotate globe
          startRotation();
          isGlobeLoaded.value = true;
        })
        .onGlobeClick(() => {
          startStopRotation();
        })
        .onHexPolygonClick((_) => {
          startStopRotation();
        });
    }
  });

  // Listen for window resize events
  window.addEventListener('resize', handleResize);
});

// Function to handle window resize
function handleResize() {
  if (globe) {
    // Update globe size
    globe.width(window.innerWidth);
    globe.height(window.innerHeight);
  }
}

// Remove event listener on component unmount
onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
  stopRotation();
});

// Function to set server location marker
function setServerLocationMarker() {
  if (globe && globe.htmlElementsData().length === 0) {
    // https://github.com/vasturiano/globe.gl/blob/master/example/html-markers/index.html
    globe
      .htmlElementsData(serverMarkerData ? [serverMarkerData] : [])
      .htmlElement(() => {
        const el = document.createElement('div');
        el.innerHTML = markerSvg;
        el.style.width = '30px';
        el.style.height = '30px';
        el.style.pointerEvents = 'auto';
        return el;
      });

    // Focus the camera on this position
    if (serverMarkerData) {
      globe.pointOfView({
        lat: serverMarkerData.lat,
        lng: serverMarkerData.lng,
        altitude: 2.5
      }, 2500);
    }
  }
}

function emitArc(start: { lat: number, lng: number }, end: { lat: number, lng: number }) {
  // Choose arc color based on start coordinates for consistency
  const hash = (start.lat.toFixed(1) + start.lng.toFixed(1)).split('').reduce(
    (acc, char) => acc + char.charCodeAt(0), 0
  );
  const arcColor = HIGH_CONTRAST_COLORS[hash % HIGH_CONTRAST_COLORS.length];
  
  // Add and remove arc after 1 cycle
  const arc = {
    startLat: start.lat,
    startLng: start.lng,
    endLat: end.lat,
    endLng: end.lng,
    color: arcColor
  };

  globe!.arcsData([...globe!.arcsData(), arc]);
  setTimeout(() => globe!.arcsData(globe!.arcsData().filter(d => d !== arc)), FLIGHT_TIME * 2);

  // Add and remove start rings
  const srcRing = { lat: start.lat, lng: start.lng, color: arcColor };
  globe!.ringsData([...globe!.ringsData(), srcRing]);
  setTimeout(() => globe!.ringsData(globe!.ringsData().filter(r => r !== srcRing)), FLIGHT_TIME * ARC_REL_LEN);

  // Add and remove target rings
  setTimeout(() => {
    const targetRing = { lat: end.lat, lng: end.lng, color: arcColor };
    globe!.ringsData([...globe!.ringsData(), targetRing]);
    setTimeout(() => globe!.ringsData(globe!.ringsData().filter(r => r !== targetRing)), FLIGHT_TIME * ARC_REL_LEN);
  }, FLIGHT_TIME);
}

// Get country color from coordinates (with caching)
function getCountryColorFromCoordinates(lat: number, lng: number, countryName?: string): string {
  // Create a simple key for mapping - rounded coordinates
  const key = `${lat.toFixed(1)},${lng.toFixed(1)}`;

  // If we already know the color for these coordinates, return it
  if (countryCoordinatesToColor.has(key)) {
    return countryCoordinatesToColor.get(key)!;
  }

  // Use the same algorithm as in hexPolygonColor
  // If we have a country name, use it; otherwise, use the coordinates
  const name = countryName || `${lat},${lng}`;
  // Convert name to number (using hashing)
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }

  // Convert number to color - use the same algorithm as in hexPolygonColor
  let color = '#';
  for (let i = 0; i < 3; i++) {
    const value = (hash >> (i * 8)) & 0xFF;
    color += value.toString(16).padStart(2, '0');
  }
  // Store in map
  countryCoordinatesToColor.set(key, color);

  return color;
}

// Function to update rotation
function updateRotation(time: number) {
  if (autoRotate && globe) {
    if (lastTime !== null) {
      const deltaT = time - lastTime;
      const rotation = globe.pointOfView();
      globe.pointOfView({
        ...rotation,
        lng: rotation.lng + (rotationSpeed * deltaT / 1000) % 360
      });
    }
    lastTime = time;
    requestAnimationFrame(updateRotation);
  }
}

// Start rotation
function startRotation() {
  autoRotate = true;
  lastTime = null;
  requestAnimationFrame(updateRotation);
}

// Stop rotation
function stopRotation() {
  autoRotate = false;
  lastTime = null;
}

// Start or stop rotation
function startStopRotation() {
  autoRotate = !autoRotate;
  if (!autoRotate) {
    lastTime = null;
  } else {
    requestAnimationFrame(updateRotation);
  }
}
</script>

<template>
  <div class="fixed top-0 left-0 w-screen h-screen overflow-hidden">
    <Transition name="fade">
      <div v-if="!isGlobeLoaded"
        class="absolute top-0 left-0 w-full h-full flex justify-center items-center z-10 bg-black/50">
        <VueSpinnerPacman size="20" :color="spinnerColor" />
      </div>
    </Transition>
    <div ref="globeContainer" class="w-full h-full"></div>
  </div>
</template>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.5s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
