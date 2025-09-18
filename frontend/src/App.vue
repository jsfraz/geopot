<!-- TODO Make standalone globe component -->

<script setup lang="ts">
import { useWebSocket } from '@vueuse/core';
import type { WSMessage } from './types/ws_message';
import Globe from './components/Globe.vue';
import { ref } from 'vue';

// WebSocket URL
const websocketUrl = import.meta.env.DEV ? 'ws://localhost:8080/ws' : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;
// Globe component
const globe = ref<InstanceType<typeof Globe> | null>(null);

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

      // Globe WebsocketMessage handler
      if (globe.value) {
        globe.value.wsMessage(msg);
      }
    },
  }
);
</script>

<template>
  <Globe ref="globe" :autoRotateDefault="false"/>
</template>
