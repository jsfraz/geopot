<script setup lang="ts">
import { useWebSocket } from '@vueuse/core';
import Globe from './components/Globe.vue';
import Heatmap from './components/Heatmap.vue';
import StatCard from './components/StatCard.vue';
import ConnectionsChart from './components/ConnectionsChart.vue';
import TopCountriesChart from './components/TopCountriesChart.vue';
import CredentialsTable from './components/CredentialsTable.vue';
import LiveFeed from './components/LiveFeed.vue';
import { Configuration, StatsApi } from './api';
import { ref } from 'vue';
import type { ModelsConnection } from './api/models/ModelsConnection';
import ServerInfoCard from './components/ServerInfoCard.vue';

// WebSocket URL
const websocketUrl = import.meta.env.DEV ? 'ws://localhost:8080/ws' : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`;
// Globe component
const globe = ref<InstanceType<typeof Globe> | null>(null);
// Heatmap component
const heatmap = ref<InstanceType<typeof Heatmap> | null>(null);
// LiveFeed component
const liveFeed = ref<InstanceType<typeof LiveFeed> | null>(null);
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
      if (globe.value && msg.latitude != 0 && msg.longitude != 0) {
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
      // Add point to heatmap
      if (heatmap.value && msg.latitude != 0 && msg.longitude != 0) {
        heatmap.value.addPoint(msg.latitude, msg.longitude);
      }
      // Add entry to live feed
      if (liveFeed.value) {
        liveFeed.value.addEntry(msg);
      }
    },
  }
);
</script>

<template>
  <div
    class="flex flex-col min-h-screen xl:h-screen xl:overflow-hidden bg-hackerbg font-mono selection:bg-hacker/30 selection:text-white">
    <!-- Header/Stat cards row -->
    <div class="w-full xl:h-[18%] flex flex-col xl:flex-row p-4 gap-4 shrink-0">
      <div class="grid grid-cols-1 sm:grid-cols-2 xl:flex xl:flex-1 gap-4">
        <!-- Total connections -->
        <div class="flex-1">
          <StatCard ref="totalConnectionsCard" title="Total Connections"
            :observableNumber="statsApi.getTotalConnectionCount()" />
        </div>
        <div class="flex-1">
          <StatCard ref="totalUniqueIpsCard" title="Unique IPs"
            :observableNumber="statsApi.getUniqueIPCount()" />
        </div>
        <div class="flex-1">
          <StatCard ref="totalUniqueCountriesCard" title="Unique Countries"
            :observableNumber="statsApi.getUniqueCountryCount()" />
        </div>
        <div class="flex-1">
          <StatCard ref="last24HourConnectionsCard" title="Last 24H"
            :observableNumber="statsApi.getLast24HourConnections()" :refreshInterval="60000" />
        </div>
      </div>
      <div class="w-full xl:w-2/7">
        <ServerInfoCard :apiConfiguration="apiConfig" />
      </div>
    </div>

    <!-- Row 2: Connections chart + Globe -->
    <div class="w-full xl:h-[41%] flex flex-col xl:flex-row px-4 gap-4 mt-0 shrink-0">
      <!-- Connections over time -->
      <div class="w-full xl:flex-1 h-64 xl:h-auto">
        <ConnectionsChart :apiConfiguration="apiConfig" />
      </div>
      <!-- Globe -->
      <div class="w-full xl:w-2/7 h-72 xl:h-auto">
        <Globe :apiConfiguration="apiConfig" ref="globe" />
      </div>
    </div>

    <!-- Row 3: Bottom panels + Heatmap -->
    <div class="w-full xl:h-[41%] flex flex-col xl:flex-row px-4 pt-0 pb-4 gap-4 mt-4">
      <!-- Left: Countries + LiveFeed + Credentials -->
      <div class="w-full xl:flex-1 flex flex-col xl:flex-row gap-4 h-full">
        <!-- Top Countries -->
        <div class="w-full xl:w-1/3 h-64 xl:h-auto">
          <TopCountriesChart :apiConfiguration="apiConfig" :limit="12" />
        </div>
        <!-- Live Feed -->
        <div class="w-full xl:w-1/3 h-64 xl:h-auto">
          <LiveFeed :apiConfiguration="apiConfig" ref="liveFeed" />
        </div>
        <!-- Credentials Table -->
        <div class="w-full xl:w-1/3 h-64 xl:h-auto">
          <CredentialsTable :apiConfiguration="apiConfig" />
        </div>
      </div>
      <!-- Heatmap -->
      <div class="w-full xl:w-2/7 h-72 xl:h-auto">
        <Heatmap :apiConfiguration="apiConfig" ref="heatmap" />
      </div>
    </div>
  </div>
</template>
