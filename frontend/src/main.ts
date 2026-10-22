import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'

// VueApexCharts is imported by the chart components themselves, so registering
// it globally would pull ApexCharts into the initial bundle for nothing.
createApp(App).mount('#app')
