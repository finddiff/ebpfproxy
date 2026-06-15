<script setup>
import { ref, onMounted } from 'vue'
import { apiGet, apiPut } from '../api.js'

const groups = ref([])
const loading = ref(true)

onMounted(async () => {
  await loadData()
})

async function loadData() {
  try {
    const json = await apiGet('/api/proxy')
    groups.value = json.groups || []
  } catch (e) { /* ignore */ }
  loading.value = false
}

async function selectProxy(groupName, serverIndex) {
  try {
    await apiPut('/api/proxy/select', { group: groupName, server_index: serverIndex })
    // Update local state immediately
    for (const g of groups.value) {
      if (g.name === groupName) {
        for (let i = 0; i < g.servers.length; i++) {
          g.servers[i].selected = (i === serverIndex)
        }
      }
    }
  } catch (e) { console.error('Failed to select proxy:', e) }
}

function isSelectPolicy(policy) {
  return policy && policy.startsWith('select')
}
</script>

<template>
  <div class="proxy">
    <h2>Proxy Groups & Servers</h2>
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="groups.length === 0" class="empty">No proxy groups configured</div>
    <div v-else>
      <div v-for="group in groups" :key="group.name" class="group-card">
        <div class="group-header">
          <h3>{{ group.name }}</h3>
          <span class="policy-badge">{{ group.policy }}</span>
        </div>
        <table class="table">
          <thead>
            <tr>
              <th v-if="isSelectPolicy(group.policy)" style="width:40px">Sel</th>
              <th>Server</th>
              <th>Address</th>
              <th>Protocol</th>
              <th>Latency</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(srv, idx) in group.servers" :key="idx" :class="{ selected: srv.selected }">
              <td v-if="isSelectPolicy(group.policy)">
                <input
                  type="radio"
                  :name="'select-' + group.name"
                  :checked="srv.selected"
                  @change="selectProxy(group.name, idx)"
                />
              </td>
              <td>{{ srv.name }}</td>
              <td><code>{{ srv.address }}</code></td>
              <td><span class="proto-tag">{{ srv.protocol }}</span></td>
              <td>{{ srv.latency_ms ? srv.latency_ms.toFixed(0) + ' ms' : '-' }}</td>
              <td>
                <span :class="['status-badge', srv.alive ? 'alive' : 'dead']">
                  {{ srv.alive ? 'Alive' : 'Dead' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.proxy h2 { margin-bottom: 16px; color: #c9d1d9; }
.loading, .empty { color: #8b949e; padding: 20px; text-align: center; }
.group-card { background: #0d1117; border: 1px solid #30363d; border-radius: 8px; padding: 16px; margin-bottom: 16px; }
.group-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.group-header h3 { margin: 0; color: #c9d1d9; font-size: 16px; }
.policy-badge { font-size: 11px; background: #1f6feb20; color: #58a6ff; padding: 2px 8px; border-radius: 10px; }
.table { width: 100%; border-collapse: collapse; }
.table th, .table td { padding: 8px 12px; text-align: left; border-bottom: 1px solid #21262d; font-size: 13px; }
.table th { font-size: 11px; color: #8b949e; text-transform: uppercase; }
.table td { vertical-align: middle; }
.table code { font-family: 'SF Mono', monospace; color: #79c0ff; font-size: 12px; }
.proto-tag { font-size: 11px; background: #30363d; color: #8b949e; padding: 1px 6px; border-radius: 4px; text-transform: uppercase; }
.status-badge { font-size: 11px; padding: 2px 8px; border-radius: 10px; font-weight: 600; }
.status-badge.alive { background: #3fb95020; color: #3fb950; }
.status-badge.dead { background: #f8514920; color: #f85149; }
tr.selected { background: rgba(31, 111, 235, 0.08); }
input[type="radio"] { accent-color: #1f6feb; cursor: pointer; }
</style>
