<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { apiGet } from '../api.js'

const allConnections = ref([])
const total = ref(0)
const loading = ref(true)
const searchQuery = ref('')
const stateFilter = ref('all')
const sourceFilter = ref('all')
const sortKey = ref('bytes_down')
const sortDir = ref(-1)
let timer = null
let searchDebounce = null

// Virtual scroll
const scrollContainer = ref(null)
const scrollTop = ref(0)
const containerHeight = ref(600)
const itemHeight = 120 // approximate card height
const bufferCount = 6

onMounted(() => {
  loadData()
  timer = setInterval(loadData, 3000)
})
onUnmounted(() => { if (timer) clearInterval(timer); if (searchDebounce) clearTimeout(searchDebounce) })

async function loadData() {
  try {
    const json = await apiGet('/api/connections')
    allConnections.value = json.connections || []
    total.value = json.total || 0
  } catch (e) { /* ignore */ }
  loading.value = false
}

// Full client-side filter + sort
const filteredConnections = computed(() => {
  let list = allConnections.value
  if (stateFilter.value !== 'all') {
    list = list.filter(c => c.state === stateFilter.value)
  }
  if (sourceFilter.value !== 'all') {
    list = list.filter(c => c.source === sourceFilter.value)
  }
  const q = searchQuery.value.toLowerCase().trim()
  if (q) {
    list = list.filter(c =>
      (c.source_ip || '').includes(q) ||
      (c.dest_ip || '').includes(q) ||
      String(c.dest_port).includes(q) ||
      (c.domain || '').includes(q) ||
      (c.process || '').includes(q) ||
      (c.outbound || '').includes(q) ||
      (c.dialer || '').includes(q)
    )
  }
  return list.slice().sort((a, b) => {
    let va, vb
    switch (sortKey.value) {
      case 'bytes_down': va = a.download_rate || 0; vb = b.download_rate || 0; break
      case 'bytes_up':   va = a.upload_rate || 0;   vb = b.upload_rate || 0;   break
      case 'duration':   va = a.duration_seconds || 0; vb = b.duration_seconds || 0; break
      case 'domain':     va = (a.domain || '').toLowerCase(); vb = (b.domain || '').toLowerCase(); return va.localeCompare(vb) * sortDir.value
      default:           return 0
    }
    return (va - vb) * sortDir.value
  })
})

// Virtual scroll computed
const totalVirtualHeight = computed(() => filteredConnections.value.length * itemHeight)

const visibleRange = computed(() => {
  const start = Math.max(0, Math.floor(scrollTop.value / itemHeight) - bufferCount)
  const visibleCount = Math.ceil(containerHeight.value / itemHeight)
  const end = Math.min(filteredConnections.value.length, start + visibleCount + bufferCount * 2)
  return { start, end }
})

const visibleItems = computed(() => {
  const { start, end } = visibleRange.value
  return filteredConnections.value.slice(start, end).map((item, i) => ({
    item,
    index: start + i,
    style: { position: 'absolute', top: (start + i) * itemHeight + 'px', left: 0, right: 0 }
  }))
})

function onScroll(e) {
  scrollTop.value = e.target.scrollTop
}

function onSearchInput() {
  if (searchDebounce) clearTimeout(searchDebounce)
  searchDebounce = setTimeout(() => {
    scrollTop.value = 0
  }, 300)
}

function toggleSort(key) {
  if (sortKey.value === key) {
    sortDir.value *= -1
  } else {
    sortKey.value = key
    sortDir.value = -1
  }
}

async function closeOne(id) {
  const t = localStorage.getItem('dae_token') || ''
  try {
    await fetch('/api/connections/' + encodeURIComponent(id), { method: 'DELETE', headers: t ? { 'X-Token': t } : {} })
    await loadData()
  } catch (e) { console.error(e) }
}

async function closeAll() {
  const t = localStorage.getItem('dae_token') || ''
  try {
    await fetch('/api/connections/', { method: 'DELETE', headers: t ? { 'X-Token': t } : {} })
    await loadData()
  } catch (e) { console.error(e) }
}

function formatDuration(s) {
  if (!s || s < 1) return ''
  if (s < 60) return Math.round(s) + 's'
  if (s < 3600) return Math.round(s / 60) + 'm'
  return Math.round(s / 3600) + 'h'
}

function formatBytes(b) {
  if (!b || b === 0) return '0 B'
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(1024))
  return parseFloat((b / Math.pow(1024, i)).toFixed(2)) + ' ' + u[i]
}

function sortLabel(key) {
  if (sortKey.value !== key) return ''
  return sortDir.value > 0 ? ' ▲' : ' ▼'
}
</script>

<template>
  <div class="connections">
    <h2>Active Connections</h2>

    <div class="toolbar">
      <input v-model="searchQuery" class="search-box" placeholder="Filter by IP, domain, process..." spellcheck="false" @input="onSearchInput" />

      <div class="filter-group">
        <button v-for="s in ['all','syn_sent','established','closing','closed']" :key="s"
          :class="['filter-btn', { active: stateFilter === s }]"
          @click="stateFilter = s">{{ s === 'all' ? 'All' : s }}</button>
      </div>

      <div class="filter-group">
        <button v-for="s in ['all','userspace','kernel']" :key="s"
          :class="['filter-btn', { active: sourceFilter === s }]"
          @click="sourceFilter = s">{{ s === 'all' ? 'All' : s.charAt(0).toUpperCase() + s.slice(1) }}</button>
      </div>

      <div class="sort-group">
        <span class="sort-label">Sort:</span>
        <button :class="['sort-btn', { active: sortKey === 'bytes_down' }]" @click="toggleSort('bytes_down')">↓ Rate{{ sortLabel('bytes_down') }}</button>
        <button :class="['sort-btn', { active: sortKey === 'bytes_up' }]" @click="toggleSort('bytes_up')">↑ Rate{{ sortLabel('bytes_up') }}</button>
        <button :class="['sort-btn', { active: sortKey === 'duration' }]" @click="toggleSort('duration')">Dur{{ sortLabel('duration') }}</button>
        <button :class="['sort-btn', { active: sortKey === 'domain' }]" @click="toggleSort('domain')">Domain{{ sortLabel('domain') }}</button>
      </div>

      <div class="header-actions">
        <span class="count">{{ filteredConnections.length }} / {{ total }} conns</span>
        <button class="refresh-btn" @click="loadData">Refresh</button>
        <button class="close-all-btn" @click="closeAll" :disabled="!allConnections.length">Close All</button>
      </div>
    </div>

    <div v-if="loading" class="loading">Loading...</div>
    <div v-else-if="filteredConnections.length === 0" class="empty">No connections match</div>
    <div v-else ref="scrollContainer" class="conn-scroll" @scroll="onScroll">
      <div class="conn-virtual" :style="{ height: totalVirtualHeight + 'px' }">
        <div v-for="{ item: conn, style } in visibleItems" :key="conn.id || style.top" :style="style" class="conn-card-wrapper">
          <div :class="['conn-card', { closed: conn.state === 'closed' }]">
            <div class="conn-top">
              <span :class="['proto-badge', conn.protocol]">{{ conn.protocol }}</span>
              <span v-if="conn.domain" class="domain">{{ conn.domain }}</span>
              <span v-if="conn.state !== 'established' && conn.state !== 'active'" class="state-badge" :class="conn.state">{{ conn.state }}</span>
              <span class="dur">{{ formatDuration(conn.duration_seconds) }}</span>
              <button v-if="conn.state !== 'closed'" class="close-btn" @click="closeOne(conn.id)" title="Close connection">×</button>
            </div>
            <div class="conn-path">
              <code>{{ conn.source_ip }}:{{ conn.source_port }}</code>
              <span class="arrow">→</span>
              <code>{{ conn.dest_ip }}:{{ conn.dest_port }}</code>
            </div>
            <div v-if="conn.outbound || conn.dialer || conn.policy || conn.process" class="conn-meta">
              <span v-if="conn.rule_index >= 0" class="meta-tag rule">#{{ conn.rule_index }}</span>
              <span v-if="conn.outbound" class="meta-tag outbound">out: {{ conn.outbound }}</span>
              <span v-if="conn.dialer" class="meta-tag dialer">d: {{ conn.dialer }}</span>
              <span v-if="conn.policy" class="meta-tag policy">{{ conn.policy }}</span>
              <span v-if="conn.network" class="meta-tag network">{{ conn.network }}</span>
              <span v-if="conn.process" class="meta-tag process">{{ conn.process }}</span>
              <span v-if="conn.mac" class="meta-tag mac">{{ conn.mac }}</span>
              <span v-if="conn.dscp" class="meta-tag dscp">dscp:{{ conn.dscp }}</span>
            </div>
            <div class="conn-stats">
              <span :class="['up', { dim: !conn.upload_rate }]">↑ {{ formatBytes(conn.upload_rate || 0) }}/s</span>
              <span :class="['down', { dim: !conn.download_rate }]">↓ {{ formatBytes(conn.download_rate || 0) }}/s</span>
              <span class="total-up">↑ {{ formatBytes(conn.upload_bytes || 0) }} total</span>
              <span class="total-down">↓ {{ formatBytes(conn.download_bytes || 0) }} total</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.connections { display: flex; flex-direction: column; height: 100%; }
.connections h2 { margin: 0 0 12px 0; color: #c9d1d9; font-size: 18px; flex-shrink: 0; }
.toolbar { display: flex; flex-wrap: wrap; gap: 10px; align-items: center; margin-bottom: 8px; flex-shrink: 0; }
.search-box {
  flex: 1; min-width: 180px; max-width: 280px; padding: 7px 12px;
  background: #0d1117; color: #c9d1d9; border: 1px solid #30363d;
  border-radius: 6px; font-size: 13px; outline: none;
}
.search-box:focus { border-color: #1f6feb; }
.filter-group { display: flex; gap: 3px; }
.filter-btn {
  padding: 5px 10px; background: #21262d; color: #8b949e;
  border: 1px solid #30363d; border-radius: 4px; cursor: pointer; font-size: 11px;
}
.filter-btn:hover { background: #30363d; color: #c9d1d9; }
.filter-btn.active { background: #1f6feb; color: #fff; border-color: #1f6feb; }
.sort-group { display: flex; gap: 3px; align-items: center; }
.sort-label { font-size: 11px; color: #484f58; }
.sort-btn {
  padding: 5px 8px; background: #21262d; color: #8b949e;
  border: 1px solid #30363d; border-radius: 4px; cursor: pointer; font-size: 11px;
}
.sort-btn:hover { background: #30363d; color: #c9d1d9; }
.sort-btn.active { background: #1f6feb20; color: #58a6ff; border-color: #1f6feb40; }
.header-actions { display: flex; gap: 8px; align-items: center; margin-left: auto; }
.count { font-size: 12px; color: #8b949e; white-space: nowrap; }
.refresh-btn, .close-all-btn { padding: 5px 12px; background: #21262d; color: #8b949e; border: 1px solid #30363d; border-radius: 6px; cursor: pointer; font-size: 12px; }
.refresh-btn:hover { background: #30363d; color: #c9d1d9; }
.close-all-btn:hover { background: #f8514920; color: #f85149; border-color: #f8514940; }
.close-all-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.loading, .empty { color: #8b949e; padding: 30px; text-align: center; flex-shrink: 0; }

.conn-scroll { flex: 1; overflow-y: auto; min-height: 0; }
.conn-virtual { position: relative; width: 100%; }
.conn-card-wrapper { padding: 0 2px 8px 2px; box-sizing: border-box; }
.conn-card { background: #0d1117; border: 1px solid #30363d; border-radius: 8px; padding: 12px 14px; }
.conn-card:hover { border-color: #1f6feb; }
.conn-card.closed { opacity: 0.45; border-color: #21262d; }
.conn-top { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.proto-badge { font-size: 10px; font-weight: 700; padding: 2px 6px; border-radius: 4px; min-width: 28px; text-align: center; }
.proto-badge.tcp { background: #1f6feb20; color: #58a6ff; }
.proto-badge.udp { background: #3fb95020; color: #3fb950; }
.domain { color: #d2a8ff; font-size: 13px; font-weight: 500; }
.state-badge { font-size: 10px; padding: 1px 6px; border-radius: 4px; font-weight: 600; }
.state-badge.syn_sent { background: #f0883e20; color: #f0883e; }
.state-badge.closing { background: #f8514920; color: #f85149; }
.dur { color: #484f58; font-size: 11px; }
.close-btn { margin-left: auto; width: 24px; height: 24px; border: none; background: transparent; color: #8b949e; font-size: 18px; cursor: pointer; border-radius: 4px; display: flex; align-items: center; justify-content: center; padding: 0; line-height: 1; }
.close-btn:hover { background: #f8514920; color: #f85149; }
.conn-path { display: flex; align-items: center; gap: 6px; margin-bottom: 8px; }
.conn-path code { font-family: monospace; color: #c9d1d9; font-size: 12px; }
.arrow { color: #484f58; font-size: 14px; }
.conn-meta { display: flex; flex-wrap: wrap; gap: 4px; }
.meta-tag { font-size: 10px; padding: 1px 6px; border-radius: 4px; font-weight: 500; }
.meta-tag.rule { background: #f0883e20; color: #f0883e; font-weight: 700; }
.meta-tag.outbound { background: #1f6feb20; color: #58a6ff; }
.meta-tag.dialer { background: #3fb95020; color: #3fb950; }
.meta-tag.policy { background: #a371f720; color: #a371f7; }
.meta-tag.network { background: #f0883e20; color: #f0883e; }
.meta-tag.process { background: #21262d; color: #8b949e; }
.meta-tag.mac { background: #21262d; color: #8b949e; font-family: monospace; }
.meta-tag.dscp { background: #21262d; color: #484f58; }
.conn-stats { display: flex; gap: 12px; margin-top: 4px; font-size: 11px; flex-wrap: wrap; }
.up { color: #3fb950; }
.down { color: #58a6ff; }
.up.dim, .down.dim { color: #30363d; }
.total-up, .total-down { font-size: 10px; color: #484f58; }
</style>
