<script setup lang="ts">
import { useWebSocket } from '@vueuse/core';
import type { WSMessage } from './types/ws_message';
import Globe from './components/Globe.vue';
import { onMounted, ref } from 'vue';
import StatCard from './components/StatCard.vue';

// WebSocket URL
const websocketUrl = import.meta.env.DEV ? 'ws://localhost:8080/ws' : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;
// Globe component
const globe = ref<InstanceType<typeof Globe> | null>(null);

onMounted(() => {
  /*
  new UserApi().createUser({}).subscribe({
    next: (response) => {
      console.log('User created:', response);
    },
    error: (error) => {
      console.error('Error creating user:', error);
    },
  });
  */
});

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
  <!-- Stat cards -->
  <div class="w-full h-1/5 flex">
    <div class="flex-1 py-4 px-2">
      <StatCard :title="'Total connections'" />
    </div>
    <div class="flex-1 py-4 px-2">
      <StatCard />
    </div>
    <div class="flex-1 py-4 px-2">
      <StatCard />
    </div>
    <div class="flex-1 py-4 px-2">
      <StatCard />
    </div>
    <div class="flex-1 py-4 px-2">
      <StatCard />
    </div>
  </div>
</template>
