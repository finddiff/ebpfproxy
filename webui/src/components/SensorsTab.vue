<script setup>
import { apiGet } from '../api.js'
import { ref, onMounted } from 'vue'

const sensors = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    const json = await apiGet('/api/sensors')
    sensors.value = json.sensors || []
  } catch (e) { /* ignore */ }
  loading.value = false
})

function sensorColor(type) {
  switch (type) {
    case 'temperature': return '#f0883e'
    case 'fan': return '#3fb950'
    case 'voltage': return '#a371f7'
    default: return '#8b949e'
  }
}
</script>

<template>
  <div class="sensors">
    <h2>System Sensors</h2>
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="sensors.length === 0" class="empty">No sensor data available</div>
    <div v-else class="sensors-grid">
      <div v-for="(s, idx) in sensors" :key="idx" class="sensor-card">
        <div class="sensor-name">{{ s.name }}</div>
        <div v-if="s.type === 'info'" class="sensor-info">{{ s.unit }}</div>
        <div v-else class="sensor-value" :style="{ color: sensorColor(s.type) }">
          {{ s.value.toFixed(1) }}<span class="unit"> {{ s.unit }}</span>
        </div>
        <div class="sensor-type">{{ s.type }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sensors h2 { margin-bottom: 16px; color: #c9d1d9; }
.loading, .empty { color: #8b949e; padding: 20px; text-align: center; }
.sensors-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 12px; }
.sensor-card { background: #0d1117; border: 1px solid #30363d; border-radius: 8px; padding: 14px; }
.sensor-name { font-size: 13px; color: #c9d1d9; margin-bottom: 6px; }
.sensor-value { font-size: 28px; font-weight: 600; }
.sensor-info { font-size: 13px; color: #8b949e; word-break: break-all; }
.unit { font-size: 14px; color: #8b949e; }
.sensor-type { font-size: 10px; color: #484f58; text-transform: uppercase; margin-top: 4px; }
</style>
