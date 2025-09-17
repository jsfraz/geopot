<script setup lang="ts">
import { ref, onMounted } from 'vue';
import Globe from 'globe.gl';

const globeContainer = ref<HTMLElement | null>(null);
let globeInstance: any = null;

onMounted(() => {
  // https://github.com/vasturiano/globe.gl/blob/master/example/hexed-polygons/index.html
  fetch('assets/ne_110m_admin_0_countries.geojson').then(res => res.json()).then(countries => {
    if (globeContainer.value) {
      // Globe instance
      globeInstance = new Globe(globeContainer.value)
        // .globeImageUrl('//cdn.jsdelivr.net/npm/three-globe/example/img/earth-dark.jpg')
        .globeImageUrl('assets/earth-dark.jpg')
        .backgroundColor('rgba(0,0,0,0)')
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
        /*
        .width(globeContainer.value.offsetWidth)
        .height(globeContainer.value.offsetHeight);
        */
    }
  });
});
</script>

<template>
  <header>
  </header>

  <main>
    <div ref="globeContainer" class="globe-container"></div>
  </main>
</template>

<style scoped>
.globe-container {
  width: 100%;
  height: 100%;
}
</style>
