<script setup>
import { apiGet } from '../api.js'
import { ref, onMounted } from 'vue'

const upstreams = ref([])
const reqRules = ref([])
const respRules = ref([])
const rawConfig = ref('')
const loading = ref(true)

onMounted(async () => {
  try {
    const json = await apiGet('/api/dns')
    upstreams.value = json.upstreams || []
    reqRules.value = json.request_rules || []
    respRules.value = json.response_rules || []
    rawConfig.value = json.raw_config || ''
  } catch (e) { /* ignore */ }
  loading.value = false
})
</script>

<template>
  <div class="dns">
    <h2>DNS Configuration</h2>
    <div v-if="loading" class="loading">Loading...</div>
    <template v-else>
      <div class="section" v-if="upstreams.length > 0">
        <h3>Upstream Servers</h3>
        <div class="upstream-list">
          <div v-for="u in upstreams" :key="u.tag" class="upstream-card">
            <div class="up-tag">{{ u.tag }}</div>
            <code class="up-link">{{ u.link }}</code>
          </div>
        </div>
      </div>

      <div class="section" v-if="reqRules.length > 0">
        <h3>Request Routing</h3>
        <div class="rule-list">
          <div v-for="(r, i) in reqRules" :key="i" class="dnssl-rule">
            <span class="dnssl-detail">{{ r.rule }}</span>
            <span class="dnssl-arrow">→</span>
            <span class="dnssl-out">{{ r.upstream }}</span>
          </div>
        </div>
      </div>

      <div class="section" v-if="respRules.length > 0">
        <h3>Response Routing</h3>
        <div class="rule-list">
          <div v-for="(r, i) in respRules" :key="i" class="dnssl-rule">
            <span class="dnssl-detail">{{ r.rule }}</span>
            <span class="dnssl-arrow">→</span>
            <span class="dnssl-out">{{ r.upstream }}</span>
          </div>
        </div>
      </div>

      <div class="section" v-if="upstreams.length === 0 && reqRules.length === 0">
        <div class="empty">No DNS configuration found in config file</div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.dns h2 { margin-bottom: 16px; color: #c9d1d9; }
.loading, .empty { color: #8b949e; padding: 20px; text-align: center; }
.section { margin-bottom: 24px; }
.section h3 { color: #8b949e; font-size: 13px; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 10px; }
.upstream-list { display: flex; flex-direction: column; gap: 6px; }
.upstream-card { display: flex; align-items: center; gap: 12px; background: #0d1117; border: 1px solid #30363d; border-radius: 6px; padding: 10px 14px; }
.up-tag { font-size: 14px; font-weight: 600; color: #58a6ff; min-width: 100px; }
.up-link { font-family: monospace; font-size: 12px; color: #79c0ff; word-break: break-all; }
.rule-list { display: flex; flex-direction: column; gap: 4px; }
.dnssl-rule { display: flex; align-items: center; gap: 8px; font-size: 13px; padding: 6px 10px; background: #0d1117; border: 1px solid #21262d; border-radius: 4px; }
.dnssl-detail { font-family: monospace; color: #d2a8ff; }
.dnssl-arrow { color: #484f58; }
.dnssl-out { font-weight: 600; color: #58a6ff; }
</style>
