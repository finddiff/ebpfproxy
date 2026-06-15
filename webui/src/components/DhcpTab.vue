<script setup>
import { apiGet } from '../api.js'
import { ref, onMounted } from 'vue'

const leases = ref([])
const error = ref('')
const loading = ref(true)

onMounted(async () => {
  try {
    const json = await apiGet('/api/dhcp')
    leases.value = json.leases || []
    error.value = json.error || ''
  } catch (e) {
    error.value = 'Failed to load DHCP data'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="dhcp">
    <h2>DHCP Leases</h2>
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="error && leases.length === 0" class="empty">{{ error }}</div>
    <div v-else-if="leases.length === 0" class="empty">No DHCP leases found</div>
    <table v-else class="table">
      <thead>
        <tr>
          <th>IP Address</th>
          <th>MAC Address</th>
          <th>Hostname</th>
          <th>Expires</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(lease, idx) in leases" :key="idx">
          <td><code>{{ lease.ip }}</code></td>
          <td><code>{{ lease.mac }}</code></td>
          <td>{{ lease.hostname }}</td>
          <td>{{ lease.expires_at }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.dhcp h2 { margin-bottom: 16px; color: #c9d1d9; }
.loading, .empty { color: #8b949e; padding: 20px; text-align: center; }
.table { width: 100%; border-collapse: collapse; }
.table th, .table td { padding: 10px 14px; text-align: left; border-bottom: 1px solid #21262d; }
.table th { font-size: 11px; color: #8b949e; text-transform: uppercase; }
.table td { font-size: 14px; }
.table code { font-family: 'SF Mono', monospace; color: #79c0ff; font-size: 13px; }
</style>
