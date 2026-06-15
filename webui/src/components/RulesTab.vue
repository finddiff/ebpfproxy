<script setup>
import { apiGet } from '../api.js'
import { ref, onMounted } from 'vue'

const rules = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    const json = await apiGet('/api/rules')
    rules.value = json.rules || []
  } catch (e) { /* ignore */ }
  loading.value = false
})
</script>

<template>
  <div class="rules">
    <h2>Routing Rules</h2>
    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="rules.length === 0" class="empty">No routing rules defined</div>
    <div v-else class="rule-list">
      <div v-for="rule in rules" :key="rule.index" class="rule-card">
        <div class="rule-header">
          <span class="rule-idx">#{{ rule.index }}</span>
          <span class="rule-detail">{{ rule.rule }}</span>
          <span class="arrow">→</span>
          <span :class="['outbound-tag', getOutboundClass(rule.outbound)]">{{ rule.outbound }}</span>
          <span v-if="rule.not" class="flag not">NOT</span>
          <span v-if="rule.must" class="flag must">MUST</span>
          <span v-if="rule.mark" class="flag mark">mark:0x{{ rule.mark.toString(16) }}</span>
        </div>
        <div class="rule-meta">
          <span class="meta-label">{{ rule.match_type }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
function getOutboundClass(name) {
  if (name === 'direct') return 'direct'
  if (name === 'block') return 'block'
  return 'group'
}
</script>

<style scoped>
.rules h2 { margin-bottom: 16px; color: #c9d1d9; }
.loading, .empty { color: #8b949e; padding: 20px; text-align: center; }
.rule-list { display: flex; flex-direction: column; gap: 6px; }
.rule-card { background: #0d1117; border: 1px solid #30363d; border-radius: 6px; padding: 10px 14px; }
.rule-header { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.rule-idx { font-size: 11px; color: #484f58; min-width: 24px; font-family: monospace; }
.rule-detail { font-family: monospace; font-size: 13px; color: #d2a8ff; }
.arrow { color: #484f58; font-size: 14px; }
.outbound-tag { font-size: 12px; font-weight: 600; padding: 2px 10px; border-radius: 4px; }
.outbound-tag.direct { background: #3fb95020; color: #3fb950; }
.outbound-tag.block { background: #f8514920; color: #f85149; }
.outbound-tag.group { background: #1f6feb20; color: #58a6ff; }
.flag { font-size: 10px; padding: 1px 5px; border-radius: 3px; font-weight: 600; }
.flag.not { background: #f8514920; color: #f85149; }
.flag.must { background: #f0883e20; color: #f0883e; }
.flag.mark { background: #21262d; color: #8b949e; font-family: monospace; }
.rule-meta { margin-top: 2px; }
.meta-label { font-size: 10px; color: #484f58; text-transform: uppercase; }
</style>
