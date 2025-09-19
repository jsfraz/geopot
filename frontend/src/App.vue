<script setup lang="ts">
import { useWebSocket } from '@vueuse/core';
import Globe from './components/Globe.vue';
import StatCard from './components/StatCard.vue';
import { Configuration, StatsApi } from './api';
import { ref } from 'vue';
import type { ModelsConnection } from './api/models/ModelsConnection';

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
const totalUniqueIpsCard = ref<InstanceType<typeof StatCard> | null>(null);
const totalUniqueCountriesCard = ref<InstanceType<typeof StatCard> | null>(null);
const last24HourConnectionsCard = ref<InstanceType<typeof StatCard> | null>(null);

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
      const msg: ModelsConnection = JSON.parse(event.data);
      console.log('WebSocket message received:', msg);

      // Globe WebsocketMessage handler
      if (globe.value && msg.latitude && msg.longitude) {
        globe.value.emitArc(msg);
      }
      // Increase total connections stat card
      if (totalConnectionsCard.value) {
        totalConnectionsCard.value.increaseNumberValue();
      }
      // Add unique IP to total unique IPs stat card
      if (totalUniqueIpsCard.value) {
        totalUniqueIpsCard.value.increaseStringsValue(msg.ipAddress);
      }
      // Add unique country to total unique countries stat card
      if (totalUniqueCountriesCard.value && msg.countryName != "") {
        totalUniqueCountriesCard.value.increaseStringsValue(msg.countryName);
      }
      // Increase last 24 hours stat card
      if (last24HourConnectionsCard.value) {
        last24HourConnectionsCard.value.increaseNumberValue();
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
        <StatCard ref="totalConnectionsCard" :title="'Total connections'"
          :observableNumber="statsApi.getTotalConnectionCount()" />
      </div>
      <div class="flex-1">
        <StatCard ref="totalUniqueIpsCard" :title="'Unique IPs'"
          :observableStrings="statsApi.getAllUniqueIPAddresses()" />
      </div>
      <div class="flex-1">
        <StatCard ref="totalUniqueCountriesCard" :title="'Unique countries'"
          :observableStrings="statsApi.getAllUniqueCountries()" />
      </div>
      <div class="flex-1">
        <StatCard ref="last24HourConnectionsCard" :title="'Last 24 hours'"
          :observableNumber="statsApi.getLast24HourConnections()" :refreshInterval="60000" />
      </div>
      <div class="w-2/7">
        <StatCard />
      </div>
    </div>
    <!-- Something and globe -->
    <div class="w-full h-2/5 flex px-4 gap-4">
      <!-- Something -->
      <div class="w-5/7">
      </div>
      <!-- Globe -->
      <div class="w-2/7">
        <Globe :apiConfiguration="apiConfig" ref="globe" />
      </div>
    </div>
    <!-- Something and heatmap -->
    <div class="w-full h-2/5 flex px-4 p-4 gap-4">
      <!-- Something -->
      <div class="w-5/7">
      </div>
      <!-- Globe -->
      <div class="border-1 border-hacker bg-hackerbg rounded-lg w-2/7 flex items-center justify-center">
        <p class="text-white">Heatmap will be implemented here</p>
      </div>
    </div>
  </div>
</template>
