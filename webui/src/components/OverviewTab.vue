<script setup>
import { ref, watch, onMounted, inject } from 'vue'
import { apiGet } from '../api.js'

const overviewData = inject('overviewData', ref({}))

const data = ref({
  cpu_percent: 0,
  mem_used: 0,
  mem_total: 0,
  mem_percent: 0,
  connections: 0,
  udp_sessions: 0,
  upload_rate: 0,
  download_rate: 0,
  upload_total: 0,
  download_total: 0,
  net_sent: 0,
  net_recv: 0,
  uptime: 0,
  load_1: 0,
  load_5: 0,
  load_15: 0,
  traffic_samples: [],
})

watch(() => overviewData.value, (val) => {
  if (val) data.value = val
}, { deep: true, immediate: true })

onMounted(async () => {
  try {
    const json = await apiGet('/api/overview')
    data.value = json
  } catch (e) { /* ignore */ }
})

function formatBytes(b) {
  if (!b || b === 0) return '0 B'
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(1024))
  return parseFloat((b / Math.pow(1024, i)).toFixed(2)) + ' ' + u[i]
}

function formatDuration(s) {
  if (!s) return '0m'
  const d = Math.floor(s / 86400)
  const h = Math.floor((s % 86400) / 3600)
  const m = Math.floor((s % 3600) / 60)
  return d > 0 ? `${d}d ${h}h ${m}m` : h > 0 ? `${h}h ${m}m` : `${m}m`
}
</script>

<template>
  <div class="overview">
    <h2>System Overview</h2>
    <div class="stats-grid">
      <div class="stat highlight">
        <div class="stat-label">CPU</div>
        <div class="stat-value">{{ data.cpu_percent?.toFixed(1) ?? 0 }}%</div>
      </div>
      <div class="stat highlight">
        <div class="stat-label">Memory</div>
        <div class="stat-value">{{ formatBytes(data.mem_used) }} / {{ formatBytes(data.mem_total) }}</div>
        <div class="stat-sub">{{ data.mem_percent?.toFixed(1) ?? 0 }}%</div>
      </div>
      <div class="stat highlight">
        <div class="stat-label">Uptime</div>
        <div class="stat-value">{{ formatDuration(data.uptime) }}</div>
      </div>
      <div class="stat highlight">
        <div class="stat-label">Load Average</div>
        <div class="stat-value" style="font-size:18px">
          {{ data.load_1?.toFixed(1) ?? 0 }} / {{ data.load_5?.toFixed(1) ?? 0 }} / {{ data.load_15?.toFixed(1) ?? 0 }}
        </div>
      </div>
    </div>

    <h3 style="margin-top:20px">Network Traffic</h3>
    <div class="traffic-row">
      <div class="traffic-card upload">
        <div class="traffic-label">Upload Rate</div>
        <div class="traffic-rate">▲ {{ formatBytes(data.upload_rate) }}/s</div>
      </div>
      <div class="traffic-card download">
        <div class="traffic-label">Download Rate</div>
        <div class="traffic-rate">▼ {{ formatBytes(data.download_rate) }}/s</div>
      </div>
      <div class="traffic-card total-up">
        <div class="traffic-label">Upload Total</div>
        <div class="traffic-rate">{{ formatBytes(data.upload_total) }}</div>
      </div>
      <div class="traffic-card total-down">
        <div class="traffic-label">Download Total</div>
        <div class="traffic-rate">{{ formatBytes(data.download_total) }}</div>
      </div>
    </div>

    <h3 style="margin-top:20px">Connections</h3>
    <div class="stats-grid">
      <div class="stat">
        <div class="stat-label">TCP Connections</div>
        <div class="stat-value">{{ data.connections ?? 0 }}</div>
      </div>
      <div class="stat">
        <div class="stat-label">UDP Sessions</div>
        <div class="stat-value">{{ data.udp_sessions ?? 0 }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.overview h2 {
  margin-bottom: 16px;
  color: #c9d1d9;
}
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 14px;
}
.stat {
  background: #0d1117;
  border: 1px solid #30363d;
  border-radius: 8px;
  padding: 16px;
}
.stat-label {
  font-size: 11px;
  color: #8b949e;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 6px;
}
.stat-value {
  font-size: 22px;
  font-weight: 600;
  color: #58a6ff;
}
.stat-sub {
  font-size: 13px;
  color: #8b949e;
  margin-top: 4px;
}
.stat.highlight {
  border-color: #1f6feb40;
}
h3 {
  color: #8b949e;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.traffic-row {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 12px;
}
.traffic-card {
  background: #0d1117;
  border: 1px solid #30363d;
  border-radius: 8px;
  padding: 16px;
  text-align: center;
}
.traffic-label {
  font-size: 11px;
  color: #8b949e;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 6px;
}
.traffic-rate {
  font-size: 18px;
  font-weight: 600;
}
.upload .traffic-rate { color: #3fb950; }
.download .traffic-rate { color: #58a6ff; }
.total-up .traffic-rate { color: #3fb950; }
.total-down .traffic-rate { color: #58a6ff; }
</style>
