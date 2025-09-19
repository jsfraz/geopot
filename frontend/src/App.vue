<script setup lang="ts">
import { useWebSocket } from '@vueuse/core';
import type { WSMessage } from './types/ws_message';
import Globe from './components/Globe.vue';
import StatCard from './components/StatCard.vue';
import { Configuration, StatsApi } from './api';
import { ref } from 'vue';

// WebSocket URL
const websocketUrl = import.meta.env.DEV ? 'ws://localhost:8080/ws' : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;
// Globe component
const globe = ref<InstanceType<typeof Globe> | null>(null);
// API client config and instances
const apiConfig = new Configuration({
  basePath: import.meta.env.DEV ? 'http://localhost:8080' : window.location.origin,
});
const statsApi = new StatsApi(apiConfig);
// StatCards
const totalConnectionsCard = ref<InstanceType<typeof StatCard> | null>(null);

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

      // TODO move message type logic here

      switch (msg.type) {
        case 'attacker':
          // TODO Globe arc
          // Increase total connections stat card
          if (totalConnectionsCard.value) {
            totalConnectionsCard.value.increaseValue();
          }
          break;

        default:
          console.warn('Unknown WebSocket message type:', msg.type);
          break;
      }

      // Globe WebsocketMessage handler
      if (globe.value) {
        globe.value.wsMessage(msg);
      }
    },
  }
);
</script>

<template>
  <div class="flex flex-col h-screen">
    <!-- Stat cards -->
    <div class="w-full h-1/6 flex p-4 gap-4">
      <!-- Total connections -->
      <div class="flex-1">
        <StatCard ref="totalConnectionsCard" :title="'Total connections'" :observable="statsApi.getTotalConnectionCount()" />
      </div>
      <div class="flex-1">
        <StatCard />
      </div>
      <div class="flex-1">
        <StatCard />
      </div>
      <div class="flex-1">
        <StatCard />
      </div>
      <div class="flex-1">
        <StatCard />
      </div>
      <div class="flex-1">
        <StatCard />
      </div>
    </div>
    <!-- Something and globe -->
    <div class="w-full h-2/5 flex px-4 gap-4">
      <!-- Something - 4/6 width -->
      <div class="w-4/6">
      </div>
      <!-- Globe - 2/6 width -->
      <div class="w-2/6">
        <Globe ref="globe" />
      </div>
    </div>
    <!-- Something and heatmap -->
    <div class="w-full h-2/5 flex px-4 p-4 gap-4">
      <!-- Something - 4/6 width -->
      <div class="w-4/6">
      </div>
      <!-- Globe - 2/6 width -->
      <div class="border-1 border-hacker bg-hackerbg rounded-lg w-2/6 flex items-center justify-center">
        <p class="text-white">Heatmap will be implemented here</p>
      </div>
    </div>
  </div>
</template>
