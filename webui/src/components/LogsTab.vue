<script setup>
import { apiGet } from '../api.js'
import { ref, watch, onMounted, onUnmounted, nextTick, inject } from 'vue'

const logEntry = inject('logEntry', null)

const filter = ref('')
const entries = ref([])
const autoScroll = ref(true)
const logContainer = ref(null)

onMounted(async () => {
  await loadLogs()
})

watch(() => logEntry, (entry) => {
  if (!entry) return
  const filterLower = filter.value.toLowerCase()
  if (filterLower && !entry.text.toLowerCase().includes(filterLower)) {
    return
  }
  entries.value.push(entry)
  if (entries.value.length > 1000) {
    entries.value = entries.value.slice(-500)
  }
  if (autoScroll.value) {
    nextTick(() => {
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    })
  }
})

async function loadLogs() {
  try {
    const params = filter.value ? '?filter=' + encodeURIComponent(filter.value) : ''
    const json = await apiGet('/api/logs' + params)
    entries.value = json.entries || []
    nextTick(() => {
      if (logContainer.value && autoScroll.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    })
  } catch (e) { /* ignore */ }
}

watch(filter, () => {
  loadLogs()
})

function clearLogs() {
  entries.value = []
}
</script>

<template>
  <div class="logs">
    <div class="logs-header">
      <h2>Logs</h2>
      <div class="logs-controls">
        <button class="control-btn" @click="autoScroll = !autoScroll" :class="{ active: autoScroll }">
          Auto-scroll
        </button>
        <button class="control-btn" @click="clearLogs">Clear</button>
      </div>
    </div>
    <input
      v-model="filter"
      class="filter-input"
      placeholder="Filter by keyword..."
      spellcheck="false"
    />
    <div ref="logContainer" class="log-container">
      <div v-if="entries.length === 0" class="empty">No log entries</div>
      <div v-for="(entry, idx) in entries" :key="idx" class="log-entry">
        <span class="log-time">{{ entry.time }}</span>
        <span class="log-text">{{ entry.text }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.logs h2 { margin: 0; color: #c9d1d9; }
.logs-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.logs-controls { display: flex; gap: 8px; }
.control-btn {
  padding: 5px 12px;
  background: #21262d;
  color: #8b949e;
  border: 1px solid #30363d;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.15s;
}
.control-btn:hover { background: #30363d; color: #c9d1d9; }
.control-btn.active { background: #1f6feb20; color: #58a6ff; border-color: #1f6feb40; }
.filter-input {
  width: 100%;
  padding: 8px 12px;
  background: #0d1117;
  color: #c9d1d9;
  border: 1px solid #30363d;
  border-radius: 6px;
  font-size: 13px;
  margin-bottom: 12px;
  outline: none;
  box-sizing: border-box;
}
.filter-input:focus { border-color: #1f6feb; }
.log-container {
  background: #0d1117;
  border: 1px solid #30363d;
  border-radius: 8px;
  height: 500px;
  overflow-y: auto;
  padding: 8px 0;
}
.empty { color: #8b949e; padding: 20px; text-align: center; }
.log-entry {
  display: flex;
  gap: 12px;
  padding: 3px 14px;
  font-size: 12px;
  font-family: 'SF Mono', 'Fira Code', monospace;
  border-bottom: 1px solid #161b22;
}
.log-entry:hover { background: #161b22; }
.log-time { color: #484f58; white-space: nowrap; min-width: 60px; }
.log-text { color: #c9d1d9; word-break: break-all; }
</style>
